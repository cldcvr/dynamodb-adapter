package v1

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudspannerecosystem/dynamodb-adapter/config"
	"github.com/cloudspannerecosystem/dynamodb-adapter/models"
	"github.com/cloudspannerecosystem/dynamodb-adapter/pkg/errors"
	"github.com/cloudspannerecosystem/dynamodb-adapter/pkg/logger"
	"github.com/cloudspannerecosystem/dynamodb-adapter/service/services"
	"github.com/cloudspannerecosystem/dynamodb-adapter/storage"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
)

// InitVPC - Init VPC routes
func InitDBAPI(g *gin.RouterGroup) {

	r := g.Group("/")
	r.POST("/batchGet", BatchGetMeta)
	r.POST("/getWithProjection", GetMetaWithProjection)
	r.POST("/batchGetWithProjection", BatchGetMetaWithProjection)
	r.POST("/get", Get)

	r.POST("/Query", QueryTable)

	r.POST("/put", UpdateMeta)
	r.POST("/batchPut", BatchUpdateMeta)

	r.POST("/batchDelete", BatchDelete)
	r.POST("/deleteItem", DeleteItem) // return whole object
	r.POST("/deleteWithCondExpression", DeleteItem)

	r.POST("/scan", Scan)

	r.POST("/update", Update)
	r.POST("/updateAttribute", UpdateAttribute)

}

func enrichSpan(c *gin.Context, span opentracing.Span, query models.Query) opentracing.Span {
	span = span.SetTag("table", query.TableName)
	span = span.SetTag("index", query.IndexName)
	return span
}

func addParentSpanID(c *gin.Context, span opentracing.Span) opentracing.Span {
	parentSpanID := c.Request.Header.Get("X-B3-Spanid")
	traceID := c.Request.Header.Get("X-B3-Traceid")
	serviceName := c.Request.Header.Get("service-name")
	span = span.SetTag("parentSpanId", parentSpanID)
	span = span.SetTag("traceId", traceID)
	span = span.SetTag("service-name", serviceName)
	return span
}

// UpdateMeta Writes a record
// @Description Writes a record
// @Summary Writes record
// @ID put
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.Meta true "Please add request body of type models.Meta"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /put/ [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func UpdateMeta(c *gin.Context) {
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var meta models.Meta
	if err := c.ShouldBindJSON(&meta); err != nil {
		c.JSON(errors.New("ValidationException", err).HTTPResponse(meta))
	} else {
		if allow := services.MayIReadOrWrite(meta.TableName, true, "UpdateMeta"); !allow {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		logger.LogDebug(meta)
		meta.AttrMap, err = ConvertDynamoToMap(meta.TableName, meta.DynamoObject)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(meta))
			return
		}
		meta.ExpressionAttributeValues, err = ConvertDynamoToMap(meta.TableName, meta.DynamoObjectAttr)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(meta))
			return
		}

		res, err := put(c.Request.Context(), meta.TableName, meta.AttrMap, nil, meta.ConditionalExp, meta.ExpressionAttributeValues)
		if err != nil {
			c.JSON(errors.HTTPResponse(err, meta))
		} else {
			c.JSON(http.StatusOK, ChangeResponseToOriginalColumns(meta.TableName, res))
		}
	}
}

func put(ctx context.Context, tableName string, putObj map[string]interface{}, expr *models.UpdateExpressionCondition, conditionExp string, expressionAttr map[string]interface{}) (map[string]interface{}, error) {
	tableConf, err := config.GetTableConf(tableName)
	if err != nil {
		return nil, err
	}
	sKey := tableConf.SortKey
	pKey := tableConf.PartitionKey
	var oldResp = map[string]interface{}{}

	oldResp, err = storage.GetStorageInstance().SpannerGet(ctx, tableName, putObj[pKey], putObj[sKey], nil)
	if err != nil {
		return nil, err
	}
	res, err := services.Put(ctx, tableName, putObj, nil, conditionExp, expressionAttr, oldResp)
	if err != nil {
		return nil, err
	}
	go services.StreamDataToThirdParty(oldResp, res, tableName)
	return oldResp, nil
}

// UpdateMeta Writes a record
// @Description Writes a record
// @Summary Writes record
// @ID batch-put
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.BatchMetaUpdate true "Please add request body of type models.BatchMetaUpdate"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /batchPut/ [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func BatchUpdateMeta(c *gin.Context) {
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var batchMetaUpdate models.BatchMetaUpdate
	if err1 := c.ShouldBindJSON(&batchMetaUpdate); err1 != nil {
		c.JSON(errors.New("ValidationException", err1).HTTPResponse(batchMetaUpdate))
	} else {
		logger.LogDebug(batchMetaUpdate)
		if allow := services.MayIReadOrWrite(batchMetaUpdate.TableName, true, "BatchUpdateMeta"); !allow {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		batchMetaUpdate.ArrAttrMap, err1 = ConvertDynamoArrayToMapArray(batchMetaUpdate.TableName, batchMetaUpdate.DynamoObject)
		if err1 != nil {
			c.JSON(errors.New("ValidationException", err1).HTTPResponse(batchMetaUpdate))
			return
		}
		err2 := services.BatchPut(c.Request.Context(), batchMetaUpdate.TableName, batchMetaUpdate.ArrAttrMap)
		if err2 == nil {
			c.JSON(http.StatusOK, gin.H{})
		} else {
			c.JSON(errors.HTTPResponse(err2, batchMetaUpdate))
		}
		// for i := 0; i < len(batchMetaUpdate.ArrAttrMap); i++ {
		// }
	}
}

func queryResponse(query models.Query, c *gin.Context) {
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI())
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	var err1 error
	if allow := services.MayIReadOrWrite(query.TableName, false, ""); !allow {
		c.JSON(http.StatusOK, gin.H{})
		return
	}

	if query.Select == "COUNT" {
		query.OnlyCount = true
	}

	query.StartFrom, err1 = ConvertDynamoToMap(query.TableName, query.ExclusiveStartKey)
	if err1 != nil {
		c.JSON(errors.New("ValidationException", err1).HTTPResponse(query))
		return
	}
	query.RangeValMap, err1 = ConvertDynamoToMap(query.TableName, query.ExpressionAttributeValues)
	if err1 != nil {
		c.JSON(errors.New("ValidationException", err1).HTTPResponse(query))
		return
	}

	if query.Limit == 0 {
		query.Limit = 5000
	}
	query.ExpressionAttributeNames = ChangeColumnToSpannerExpressionName(query.TableName, query.ExpressionAttributeNames)
	query = ReplaceHashRangeExpr(query)
	res, hash, err := services.QueryAttributes(c.Request.Context(), query)
	if err == nil {
		changedOutput := ChangeQueryResponseColumn(query.TableName, res)
		if _, ok := changedOutput["Items"]; ok && changedOutput["Items"] != nil {
			changedOutput["Items"], err = ChangeMaptoDynamoMap(changedOutput["Items"])
			if err != nil {
				c.JSON(errors.HTTPResponse(err, "ItemsChangeError"))
			}
		}
		if _, ok := changedOutput["LastEvaluatedKey"]; ok && changedOutput["LastEvaluatedKey"] != nil {
			changedOutput["LastEvaluatedKey"], err = ChangeMaptoDynamoMap(changedOutput["LastEvaluatedKey"])
			if err != nil {
				c.JSON(errors.HTTPResponse(err, "LastEvaluatedKeyChangeError"))
			}
		}

		c.JSON(http.StatusOK, changedOutput)
	} else {
		c.JSON(errors.HTTPResponse(err, query))
	}
	if hash != "" {
		span = span.SetTag("qHash", hash)
	}
}

// QueryTable queries a table
// @Description Query a table
// @Summary Query a table
// @ID query-table
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.Query true "Please add request body of type models.Query"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /query/ [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func QueryTable(c *gin.Context) {
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var query models.Query
	if err := c.ShouldBindJSON(&query); err != nil {
		c.JSON(errors.New("ValidationException", err).HTTPResponse(query))
	} else {
		logger.LogInfo(query)
		queryResponse(query, c)
	}
}

//BatchGetMeta Get Batch meta
// @Description Request items in a batch.
// @Summary Query a table
// @ID query-table
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.BatchMeta true "Please add request body of type models.BatchMeta"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /batchGet/ [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func BatchGetMeta(c *gin.Context) {
	start := time.Now()
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var batchMeta models.BatchMeta
	if err := c.ShouldBindJSON(&batchMeta); err != nil {
		c.JSON(errors.New("ValidationException", err).HTTPResponse(batchMeta))
	} else {
		logger.LogDebug(batchMeta)
		if allow := services.MayIReadOrWrite(batchMeta.TableName, false, ""); !allow {
			c.JSON(http.StatusOK, []gin.H{})
			return
		}
		batchMeta.KeyArray, err = ConvertDynamoArrayToMapArray(batchMeta.TableName, batchMeta.DynamoObject)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(batchMeta))
			return
		}
		res, err := services.BatchGet(c.Request.Context(), batchMeta.TableName, batchMeta.KeyArray)
		span = span.SetTag("table", batchMeta.TableName)
		span = span.SetTag("batchRequestCount", len(batchMeta.DynamoObject))
		span = span.SetTag("batchResponseCount", len(res))
		if err == nil {
			c.JSON(http.StatusOK, ChangesArrayResponseToOriginalColumns(batchMeta.TableName, res))
		} else {
			c.JSON(errors.HTTPResponse(err, batchMeta))
		}
		if time.Since(start) > time.Second*1 {
			go fmt.Println("BatchGetCall", batchMeta)
		}
	}
}

// GetMetaWithProjection to get with projections
// @Description Get a record with projections
// @Summary Get a record with projections
// @ID get-with-projection
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.GetWithProjectionMeta true "Please add request body of type models.GetWithProjectionMeta"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /getWithProjection/ [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func GetMetaWithProjection(c *gin.Context) {
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var getWithProjectionMeta models.GetWithProjectionMeta
	if err := c.ShouldBindJSON(&getWithProjectionMeta); err != nil {
		c.JSON(errors.New("ValidationException", err).HTTPResponse(getWithProjectionMeta))
	} else {
		span.SetTag("table", getWithProjectionMeta.TableName)
		logger.LogDebug(getWithProjectionMeta)
		if allow := services.MayIReadOrWrite(getWithProjectionMeta.TableName, false, ""); !allow {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		getWithProjectionMeta.PrimaryKeyMap, err = ConvertDynamoToMap(getWithProjectionMeta.TableName, getWithProjectionMeta.DynamoObject)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(getWithProjectionMeta))
			return
		}
		getWithProjectionMeta.ExpressionAttributeNames = ChangeColumnToSpannerExpressionName(getWithProjectionMeta.TableName, getWithProjectionMeta.ExpressionAttributeNames)
		res, rowErr := services.GetWithProjection(c.Request.Context(), getWithProjectionMeta.TableName, getWithProjectionMeta.PrimaryKeyMap, getWithProjectionMeta.ProjectionExpression, getWithProjectionMeta.ExpressionAttributeNames)
		if rowErr == nil {
			c.JSON(http.StatusOK, ChangeResponseToOriginalColumns(getWithProjectionMeta.TableName, res))
		} else {
			c.JSON(errors.HTTPResponse(rowErr, getWithProjectionMeta))
		}
	}
}

// BatchGetMetaWithProjection to get with projections
// @Description Request items in a batch with projections.
// @Summary Request items in a batch with projections.
// @ID batch-get-with-projection
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.BatchGetWithProjectionMeta true "Please add request body of type models.BatchGetWithProjectionMeta"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /batchGetWithProjection/ [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func BatchGetMetaWithProjection(c *gin.Context) {
	start := time.Now()
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var batchGetWithProjectionMeta models.BatchGetWithProjectionMeta
	if err1 := c.ShouldBindJSON(&batchGetWithProjectionMeta); err1 != nil {
		c.JSON(errors.New("ValidationException", err1).HTTPResponse(batchGetWithProjectionMeta))
	} else {
		logger.LogDebug(batchGetWithProjectionMeta)
		if allow := services.MayIReadOrWrite(batchGetWithProjectionMeta.TableName, false, ""); !allow {
			c.JSON(http.StatusOK, []gin.H{})
			return
		}
		batchGetWithProjectionMeta.KeyArray, err1 = ConvertDynamoArrayToMapArray(batchGetWithProjectionMeta.TableName, batchGetWithProjectionMeta.DynamoObject)
		if err1 != nil {
			c.JSON(errors.New("ValidationException", err1).HTTPResponse(batchGetWithProjectionMeta))
			return
		}
		batchGetWithProjectionMeta.ExpressionAttributeNames = ChangeColumnToSpannerExpressionName(batchGetWithProjectionMeta.TableName, batchGetWithProjectionMeta.ExpressionAttributeNames)
		res, err2 := services.BatchGetWithProjection(c.Request.Context(), batchGetWithProjectionMeta.TableName, batchGetWithProjectionMeta.KeyArray, batchGetWithProjectionMeta.ProjectionExpression, batchGetWithProjectionMeta.ExpressionAttributeNames)
		span = span.SetTag("table", batchGetWithProjectionMeta.TableName)
		span = span.SetTag("batchRequestCount", len(batchGetWithProjectionMeta.DynamoObject))
		span = span.SetTag("batchResponseCount", len(res))
		if err2 == nil {
			c.JSON(http.StatusOK, ChangesArrayResponseToOriginalColumns(batchGetWithProjectionMeta.TableName, res))
		} else {
			c.JSON(errors.HTTPResponse(err2, batchGetWithProjectionMeta))
		}
		if time.Since(start) > time.Second*1 {
			go fmt.Println("BatchGetCall", batchGetWithProjectionMeta)
		}
	}
}

// DeleteItem  ...
// @Description Delete Item from table
// @Summary Delete Item from table
// @ID delete-row
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.Delete true "Please add request body of type models.Delete"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /deleteItem/ [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func DeleteItem(c *gin.Context) {
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var deleteItem models.Delete
	if err := c.ShouldBindJSON(&deleteItem); err != nil {
		c.JSON(errors.New("ValidationException", err).HTTPResponse(deleteItem))
	} else {
		logger.LogDebug(deleteItem)
		if allow := services.MayIReadOrWrite(deleteItem.TableName, true, "DeleteItem"); !allow {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		deleteItem.PrimaryKeyMap, err = ConvertDynamoToMap(deleteItem.TableName, deleteItem.DynamoObject)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(deleteItem))
			return
		}
		deleteItem.ExpressionAttributeValues, err = ConvertDynamoToMap(deleteItem.TableName, deleteItem.DynamoObjectAttrVal)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(deleteItem))
			return
		}
		oldRes, _ := services.GetWithProjection(c.Request.Context(), deleteItem.TableName, deleteItem.PrimaryKeyMap, "", nil)
		err := services.Delete(c.Request.Context(), deleteItem.TableName, deleteItem.PrimaryKeyMap, deleteItem.ConditionalExpression, deleteItem.ExpressionAttributeValues, nil)
		if err == nil {
			c.JSON(http.StatusOK, ChangeResponseToOriginalColumns(deleteItem.TableName, oldRes))
			go services.StreamDataToThirdParty(oldRes, nil, deleteItem.TableName)
		} else {
			c.JSON(errors.HTTPResponse(err, deleteItem))
		}
	}
}

// BatchDelete ...
// @Description Batch Delete Item from table
// @Summary Batch Delete Items from table
// @ID batch-delete-rows
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.BulkDelete true "Please add request body of type models.BulkDelete"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /batchDelete/ [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func BatchDelete(c *gin.Context) {
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var bulkDelete models.BulkDelete
	if err1 := c.ShouldBindJSON(&bulkDelete); err1 != nil {
		c.JSON(errors.New("ValidationException", err1).HTTPResponse(bulkDelete))
	} else {
		logger.LogDebug(bulkDelete)
		if allow := services.MayIReadOrWrite(bulkDelete.TableName, true, "BatchDelete"); !allow {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		bulkDelete.PrimaryKeyMapArray, err1 = ConvertDynamoArrayToMapArray(bulkDelete.TableName, bulkDelete.DynamoObject)
		if err1 != nil {
			c.JSON(errors.New("ValidationException", err1).HTTPResponse(bulkDelete))
			return
		}
		err2 := services.BatchDelete(c.Request.Context(), bulkDelete.TableName, bulkDelete.PrimaryKeyMapArray)
		if err2 == nil {
			c.JSON(http.StatusOK, []gin.H{})
		} else {
			c.JSON(errors.HTTPResponse(err2, bulkDelete))
		}
	}
}

// Get record from table
// @Description Get record from table
// @Summary Get record from table
// @ID get
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.GetMeta true "Please add request body of type models.GetMeta"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /get/ [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func Get(c *gin.Context) {
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var meta models.GetMeta
	if err := c.ShouldBindJSON(&meta); err != nil {
		c.JSON(errors.New("ValidationException", err).HTTPResponse(meta))
	} else {
		if allow := services.MayIReadOrWrite(meta.TableName, false, ""); !allow {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		meta.PrimaryKeyMap, err = ConvertDynamoToMap(meta.TableName, meta.DynamoObject)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(meta))
			return
		}
		logger.LogDebug(meta)
		resp, rowErr := services.GetWithProjection(c.Request.Context(), meta.TableName, meta.PrimaryKeyMap, "", nil)
		if rowErr == nil {
			c.JSON(http.StatusOK, ChangeResponseToOriginalColumns(meta.TableName, resp))
		} else {
			c.JSON(errors.HTTPResponse(rowErr, meta))
		}
	}
}

// Scan record from table
// @Description Scan records from table
// @Summary Scan records from table
// @ID scan
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.ScanMeta true "Please add request body of type models.ScanMeta"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /scan/ [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func Scan(c *gin.Context) {
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var meta models.ScanMeta
	if err := c.ShouldBindJSON(&meta); err != nil {
		c.JSON(errors.New("ValidationException", err).HTTPResponse(meta))
	} else {
		if allow := services.MayIReadOrWrite(meta.TableName, false, ""); !allow {
			c.JSON(http.StatusOK, gin.H{})
			return
		}

		meta.StartFrom, err = ConvertDynamoToMap(meta.TableName, meta.ExclusiveStartKey)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(meta))
			return
		}

		meta.ExpressionAttributeMap, err = ConvertDynamoToMap(meta.TableName, meta.ExpressionAttributeValues)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(meta))
			return
		}
		if meta.Select == "COUNT" {
			meta.OnlyCount = true
		}

		logger.LogDebug(meta)
		res, err := services.Scan(c.Request.Context(), meta)
		if err == nil {
			changedOutput := ChangeQueryResponseColumn(meta.TableName, res)
			if _, ok := changedOutput["Items"]; ok && changedOutput["Items"] != nil {
				itemsOutput, err := ChangeMaptoDynamoMap(changedOutput["Items"])
				if err != nil {
					c.JSON(errors.HTTPResponse(err, "ItemsChangeError"))
				}
				changedOutput["Items"] = itemsOutput["L"]
			}
			if _, ok := changedOutput["LastEvaluatedKey"]; ok && changedOutput["LastEvaluatedKey"] != nil {
				changedOutput["LastEvaluatedKey"], err = ChangeMaptoDynamoMap(changedOutput["LastEvaluatedKey"])
				if err != nil {
					c.JSON(errors.HTTPResponse(err, "LastEvaluatedKeyChangeError"))
				}
			}
			c.JSON(http.StatusOK, res)
		} else {
			c.JSON(errors.HTTPResponse(err, meta))
		}
	}
}

// Update updates a record in Spanner
// @Description updates a record in Spanner
// @Summary updates a record in Spanner
// @ID update
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.UpdateAttr true "Please add request body of type models.UpdateAttr"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /update/ [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func Update(c *gin.Context) {
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var updateAttr models.UpdateAttr
	if err := c.ShouldBindJSON(&updateAttr); err != nil {
		c.JSON(errors.New("ValidationException", err).HTTPResponse(updateAttr))
	} else {
		if allow := services.MayIReadOrWrite(updateAttr.TableName, true, "update"); !allow {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		updateAttr.PrimaryKeyMap, err = ConvertDynamoToMap(updateAttr.TableName, updateAttr.DynamoObject)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(updateAttr))
			return
		}
		updateAttr.ExpressionAttributeValues, err = ConvertDynamoToMap(updateAttr.TableName, updateAttr.DynamoObjectAttr)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(updateAttr))
			return
		}
		resp, err := UpdateExpression(c.Request.Context(), updateAttr)
		if err != nil {
			c.JSON(errors.HTTPResponse(err, updateAttr))
		} else {
			c.JSON(http.StatusOK, resp)
		}
	}
}

// UpdateAttribute update attribute a record in Spanner
// @Description update attribute a record in Spanner
// @Summary updates a record in Spanner
// @ID update
// @Produce  json
// @Success 200 {object} gin.H
// @Param requestBody body models.UpdateAttr true "Please add request body of type models.UpdateAttr"
// @Failure 500 {object} gin.H "{"errorMessage":"We had a problem with our server. Try again later.","errorCode":"E0001"}"
// @Router /updateAttribute [post]
// @Failure 401 {object} gin.H "{"errorMessage":"API access not allowed","errorCode": "E0005"}"
func UpdateAttribute(c *gin.Context) {
	defer PanicHandler(c)
	defer c.Request.Body.Close()
	carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
	spanContext, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders, carrier)
	if err != nil || spanContext == nil {
		logger.LogDebug(err)
	}
	span, ctx := opentracing.StartSpanFromContext(c.Request.Context(), c.Request.URL.RequestURI(), opentracing.ChildOf(spanContext))
	c.Request = c.Request.WithContext(ctx)
	defer span.Finish()
	span = addParentSpanID(c, span)
	var updateAtrr models.UpdateAttr
	if err := c.ShouldBindJSON(&updateAtrr); err != nil {
		c.JSON(errors.New("ValidationException", err).HTTPResponse(updateAtrr))

	} else {
		if allow := services.MayIReadOrWrite(updateAtrr.TableName, true, "UpdateAttribute"); !allow {
			c.JSON(http.StatusOK, gin.H{})
			return
		}
		updateAtrr.PrimaryKeyMap, err = ConvertDynamoToMap(updateAtrr.TableName, updateAtrr.DynamoObject)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(updateAtrr))
			return
		}
		updateAtrr.ExpressionAttributeValues, err = ConvertDynamoToMap(updateAtrr.TableName, updateAtrr.DynamoObjectAttr)
		if err != nil {
			c.JSON(errors.New("ValidationException", err).HTTPResponse(updateAtrr))
			return
		}
		if updateAtrr.ReturnValues == "" {
			updateAtrr.ReturnValues = "ALL_NEW"
		}
		logger.LogDebug(updateAtrr)
		resp, err := UpdateExpression(c.Request.Context(), updateAtrr)
		if err == nil {
			c.JSON(http.StatusOK, resp)
		} else {
			c.JSON(errors.HTTPResponse(err, updateAtrr))
		}
	}

}
