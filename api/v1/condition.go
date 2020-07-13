package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/cloudspannerecosystem/dynamodb-adapter/models"
	"github.com/cloudspannerecosystem/dynamodb-adapter/pkg/logger"
	"github.com/cloudspannerecosystem/dynamodb-adapter/service/services"
)

var operations = []string{" SET ", " DELETE ", " ADD ", " REMOVE "}

func between(value string, a string, b string) string {
	// Get substring between two strings.
	posFirst := strings.Index(value, a)
	if posFirst == -1 {
		return ""
	}
	posLast := strings.Index(value, b)
	if posLast == -1 {
		return ""
	}
	posFirstAdjusted := posFirst + len(a)
	if posFirstAdjusted >= posLast {
		return ""
	}
	return value[posFirstAdjusted:posLast]
}

func before(value string, a string) string {
	// Get substring before a string.
	pos := strings.Index(value, a)
	if pos == -1 {
		return ""
	}
	return value[0:pos]
}

func after(value string, a string) string {
	// Get substring after a string.
	pos := strings.LastIndex(value, a)
	if pos == -1 {
		return ""
	}
	adjustedPos := pos + len(a)
	if adjustedPos >= len(value) {
		return ""
	}
	return value[adjustedPos:len(value)]
}

func deleteEmpty(s []string) []string {
	var r []string
	for _, str := range s {
		if str != "" {
			r = append(r, str)
		}
	}
	return r
}

func parseActionValue(actionValue string, updateAtrr models.UpdateAttr, assignment bool) (map[string]interface{}, *models.UpdateExpressionCondition) {
	expr := parseUpdateExpresstion(actionValue)
	if expr != nil {
		actionValue = expr.ActionVal
		expr.AddValues = make(map[string]float64)
	}
	resp := make(map[string]interface{})
	pairs := strings.Split(actionValue, ",")
	var v []string
	for _, p := range pairs {
		var addValue float64
		status := false
		if strings.Contains(p, "+") {
			tokens := strings.Split(p, "+")
			tokens[1] = strings.TrimSpace(tokens[1])
			p = tokens[0]
			v1, ok := updateAtrr.ExpressionAttributeValues[(tokens[1])]
			if ok {
				v2, ok := v1.(float64)
				if ok {
					addValue = v2
					status = true
				} else {
					v2, ok := v1.(int64)
					if ok {
						addValue = float64(v2)
						status = true
					}
				}
			}
		}
		if strings.Contains(p, "-") {
			tokens := strings.Split(p, "-")
			tokens[1] = strings.TrimSpace(tokens[1])
			v1, ok := updateAtrr.ExpressionAttributeValues[(tokens[1])]
			if ok {
				v2, ok := v1.(float64)
				if ok {
					addValue = -v2
					status = true

				} else {
					v2, ok := v1.(int64)
					if ok {
						addValue = float64(-v2)
						status = true
					}
				}
			}
		}
		if assignment {
			v = strings.Split(p, " ")
			v = deleteEmpty(v)
		} else {
			v = strings.Split(p, "=")
		}
		v[0] = strings.Replace(v[0], " ", "", -1)
		v[1] = strings.Replace(v[1], " ", "", -1)

		if status {
			expr.AddValues[v[0]] = addValue
		}
		if updateAtrr.ExpressionAttributeNames[v[0]] != "" {
			tmp, ok := updateAtrr.ExpressionAttributeValues[v[1]]
			if ok {
				resp[updateAtrr.ExpressionAttributeNames[v[0]]] = tmp
			}
		} else {
			if strings.Contains(v[1], "%") {
				for j := 0; j < len(expr.Field); j++ {
					if strings.Contains(v[1], "%"+expr.Value[j]+"%") {
						tmp, ok := updateAtrr.ExpressionAttributeValues[expr.Value[j]]
						if ok {
							resp[v[0]] = tmp
						}
					}
				}
			} else {
				tmp, ok := updateAtrr.ExpressionAttributeValues[v[1]]
				if ok {
					resp[v[0]] = tmp
				}
			}
		}
	}
	// Merge primaryKeyMap and updateAttributes
	for k, v := range updateAtrr.PrimaryKeyMap {
		resp[k] = v
	}
	return resp, expr
}

func parseUpdateExpresstion(actionValue string) *models.UpdateExpressionCondition {
	if actionValue == "" {
		return nil
	}
	expr := new(models.UpdateExpressionCondition)
	expr.ActionVal = actionValue
	for {
		index := strings.Index(expr.ActionVal, "if_not_exists")
		if index == -1 {
			index = strings.Index(expr.ActionVal, "if_exists")
			if index == -1 {
				break
			}
			expr.Condition = append(expr.Condition, "if_exists")
		} else {
			expr.Condition = append(expr.Condition, "if_not_exists")
		}
		if len(expr.Condition) == 0 {
			break
		}
		start := -1
		end := -1
		for i := index; i < len(expr.ActionVal); i++ {
			if expr.ActionVal[i] == '(' && start == -1 {
				start = i
			}
			if expr.ActionVal[i] == ')' && end == -1 {
				end = i
				break
			}
		}
		bracketValue := expr.ActionVal[start+1 : end]
		tokens := strings.Split(bracketValue, ",")
		expr.Field = append(expr.Field, strings.TrimSpace(tokens[0]))
		v := strings.TrimSpace(tokens[1])
		expr.Value = append(expr.Value, v)
		expr.ActionVal = strings.Replace(expr.ActionVal, expr.ActionVal[index:end+1], "%"+v+"%", 1)
	}
	return expr
}

func performOperation(ctx context.Context, action string, actionValue string, updateAtrr models.UpdateAttr, oldRes map[string]interface{}) (map[string]interface{}, map[string]interface{}, error) {
	switch {
	case action == "DELETE":
		// perform delete
		m, expr := parseActionValue(actionValue, updateAtrr, true)
		res, err := services.Del(ctx, updateAtrr.TableName, updateAtrr.PrimaryKeyMap, updateAtrr.ConditionalExpression, m, expr)
		return res, m, err
	case action == "SET":
		// Update data in table
		m, expr := parseActionValue(actionValue, updateAtrr, false)
		res, err := services.Put(ctx, updateAtrr.TableName, m, expr, updateAtrr.ConditionalExpression, updateAtrr.ExpressionAttributeValues, oldRes)
		return res, m, err
	case action == "ADD":
		// Add data in table
		m, expr := parseActionValue(actionValue, updateAtrr, true)
		res, err := services.Add(ctx, updateAtrr.TableName, updateAtrr.PrimaryKeyMap, updateAtrr.ConditionalExpression, m, updateAtrr.ExpressionAttributeValues, expr, oldRes)
		return res, m, err

	case action == "REMOVE":
		res, err := services.Remove(ctx, updateAtrr.TableName, updateAtrr, actionValue, nil, oldRes)
		return res, updateAtrr.PrimaryKeyMap, err
	default:
	}
	return nil, nil, nil
}

// UpdateExpression performs an expression
func UpdateExpression(ctx context.Context, updateAtrr models.UpdateAttr) (interface{}, error) {
	updateAtrr.ExpressionAttributeNames = ChangeColumnToSpannerExpressionName(updateAtrr.TableName, updateAtrr.ExpressionAttributeNames)
	var oldRes map[string]interface{}
	if updateAtrr.ReturnValues != "NONE" {
		oldRes, _ = services.GetWithProjection(ctx, updateAtrr.TableName, updateAtrr.PrimaryKeyMap, "", nil)
	}
	var resp map[string]interface{}
	var actVal = make(map[string]interface{})
	var er error
	for k, v := range updateAtrr.ExpressionAttributeNames {
		updateAtrr.UpdateExpression = strings.ReplaceAll(updateAtrr.UpdateExpression, k, v)
		updateAtrr.ConditionalExpression = strings.ReplaceAll(updateAtrr.ConditionalExpression, k, v)
	}
	m := extractOperations(updateAtrr.UpdateExpression)
	for k, v := range m {
		res, acVal, err := performOperation(ctx, k, v, updateAtrr, oldRes)
		resp = res
		er = err
		for k, v := range acVal {
			actVal[k] = v
		}
	}
	if er == nil {
		go services.StreamDataToThirdParty(oldRes, resp, updateAtrr.TableName)
	}
	logger.LogDebug(updateAtrr.ReturnValues, resp, oldRes)
	switch updateAtrr.ReturnValues {
	case "NONE":
		return nil, er
	case "ALL_NEW":
		return ChangeResponseToOriginalColumns(updateAtrr.TableName, resp), er
	case "ALL_OLD":
		if oldRes == nil || len(oldRes) == 0 {
			return nil, er
		}
		return ChangeResponseToOriginalColumns(updateAtrr.TableName, oldRes), er
	case "UPDATED_NEW":
		var resVal = make(map[string]interface{})
		for k := range actVal {
			resVal[k] = resp[k]
		}
		return ChangeResponseToOriginalColumns(updateAtrr.TableName, resVal), er
	case "UPDATED_OLD":
		if oldRes == nil || len(oldRes) == 0 {
			return nil, er
		}
		var resVal = make(map[string]interface{})
		for k := range actVal {
			resVal[k] = oldRes[k]
		}
		return ChangeResponseToOriginalColumns(updateAtrr.TableName, resVal), er
	default:
		return ChangeResponseToOriginalColumns(updateAtrr.TableName, resp), er
	}
}

func extractOperations(updateExpression string) map[string]string {
	if updateExpression == "" {
		return nil
	}
	updateExpression = strings.TrimSpace(updateExpression)
	updateExpression = " " + updateExpression
	opsIndex := []int{}
	opsSeq := map[int]string{}
	for _, k := range operations {
		if index := strings.Index(updateExpression, k); index > -1 {
			opsSeq[index] = k
			updateExpression = strings.Replace(updateExpression, k, "%", 1)
			opsIndex = append(opsIndex, index)
		}
	}

	sort.Ints(opsIndex)
	tokens := strings.Split(updateExpression, "%")[1:]
	ops := map[string]string{}
	for i, opsIndex := range opsIndex {
		ops[strings.TrimSpace(opsSeq[opsIndex])] = tokens[i]
	}
	return ops
}

func ReplaceHashRangeExpr(query models.Query) models.Query {
	for k, v := range query.ExpressionAttributeNames {
		query.HashExp = strings.ReplaceAll(query.HashExp, k, v)
		query.FilterExp = strings.ReplaceAll(query.FilterExp, k, v)
		query.FilterExp = strings.ReplaceAll(query.FilterExp, k, v)
		query.RangeExp = strings.ReplaceAll(query.RangeExp, k, v)
	}
	return query
}

func ConvertDynamoToMap(tableName string, dynamoMap map[string]*dynamodb.AttributeValue) (map[string]interface{}, error) {
	if dynamoMap == nil || len(dynamoMap) == 0 {
		return nil, nil
	}
	rs := make(map[string]interface{})
	err := ConvertFromMap(dynamoMap, &rs, tableName)
	if err != nil {
		return nil, err
	}
	_, ok := models.TableColChangeMap[tableName]
	if ok {
		rs = ChangeColumnToSpanner(rs)
	}
	return rs, nil
}

func ConvertDynamoArrayToMapArray(tableName string, dynamoMap []map[string]*dynamodb.AttributeValue) ([]map[string]interface{}, error) {
	if dynamoMap == nil || len(dynamoMap) == 0 {
		return nil, nil
	}
	rs := make([]map[string]interface{}, len(dynamoMap))
	for i := 0; i < len(dynamoMap); i++ {
		err := ConvertFromMap(dynamoMap[i], &rs[i], tableName)
		if err != nil {
			return nil, err
		}
		_, ok := models.TableColChangeMap[tableName]
		if ok {
			rs[i] = ChangeColumnToSpanner(rs[i])
		}
	}
	return rs, nil
}

func ChangeColumnToSpannerExpressionName(tableName string, expressNameMap map[string]string) map[string]string {
	_, ok := models.TableColChangeMap[tableName]
	if !ok {
		return expressNameMap
	}

	rs := make(map[string]string)
	if expressNameMap != nil {
		for k, v := range expressNameMap {
			if v1, ok := models.ColumnToOriginalCol[v]; ok {
				rs[k] = v1
			} else {
				rs[k] = v
			}
		}
	}

	return rs
}

func ChangesArrayResponseToOriginalColumns(tableName string, obj []map[string]interface{}) []map[string]interface{} {
	_, ok := models.TableColChangeMap[tableName]
	if !ok {
		return obj
	}
	for i := 0; i < len(obj); i++ {
		obj[i] = ChangeResponseColumn(obj[i])
	}
	return obj
}

func ChangeResponseToOriginalColumns(tableName string, obj map[string]interface{}) map[string]interface{} {
	_, ok := models.TableColChangeMap[tableName]
	if !ok {
		return obj
	}
	rs := make(map[string]interface{})
	logger.LogInfo(models.ColumnToOriginalCol)
	if obj != nil {
		for k, v := range obj {

			if k1, ok := models.OriginalColResponse[k]; ok {
				rs[k1] = v
			} else {
				rs[k] = v
			}
		}
	}

	return rs
}

func ChangeResponseColumn(obj map[string]interface{}) map[string]interface{} {
	rs := make(map[string]interface{})

	if obj != nil {
		for k, v := range obj {

			if k1, ok := models.OriginalColResponse[k]; ok {
				rs[k1] = v
			} else {
				rs[k] = v
			}
		}
	}

	return rs
}
func ChangeColumnToSpanner(obj map[string]interface{}) map[string]interface{} {
	rs := make(map[string]interface{})

	if obj != nil {
		for k, v := range obj {

			if k1, ok := models.ColumnToOriginalCol[k]; ok {
				rs[k1] = v
			} else {
				rs[k] = v
			}
		}
	}

	return rs
}

func convertFrom(a *dynamodb.AttributeValue, tableName string) interface{} {
	if a.S != nil {
		return *a.S
	}

	if a.N != nil {
		if strings.ToLower(*a.N) == "infinity" || strings.ToLower(*a.N) == "-infinity" || strings.ToLower(*a.N) == "nan" {
			panic("N does not support " + *a.N + " type value")
		}
		// Number is tricky b/c we don't know which numeric type to use. Here we
		// simply try the different types from most to least restrictive.
		if n, err := strconv.ParseInt(*a.N, 10, 64); err == nil {
			return float64(n)
		}
		if n, err := strconv.ParseUint(*a.N, 10, 64); err == nil {
			return float64(n)
		}
		n, err := strconv.ParseFloat(*a.N, 64)
		if err != nil {
			panic(err)
		}
		return n
	}

	if a.BOOL != nil {
		return *a.BOOL
	}

	if a.NULL != nil {
		return nil
	}

	if a.M != nil {
		m := make(map[string]interface{})
		for k, v := range a.M {
			m[k] = convertFrom(v, tableName)
		}
		return m
	}

	if a.L != nil {
		l := make([]interface{}, len(a.L))
		for index, v := range a.L {
			l[index] = convertFrom(v, tableName)
		}
		return l
	}

	if a.B != nil {
		return a.B
	}
	if a.SS != nil {
		l := make([]interface{}, len(a.SS))
		for index, v := range a.SS {
			l[index] = *v
		}
		return l
	}
	if a.NS != nil {
		l := make([]interface{}, len(a.NS))
		for index, v := range a.NS {
			l[index], _ = strconv.ParseFloat(*v, 64)
		}
		return l
	}
	panic(fmt.Sprintf("%#v is not a supported dynamodb.AttributeValue", a))
}

func ConvertFromMap(item map[string]*dynamodb.AttributeValue, v interface{}, tableName string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(runtime.Error); ok {
				err = e
			} else if s, ok := r.(string); ok {
				err = fmt.Errorf(s)
			} else {
				err = r.(error)
			}
			item = nil
		}
	}()

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return awserr.New("SerializationError",
			fmt.Sprintf("v must be a non-nil pointer to a map[string]interface{} or struct, got %s",
				rv.Type()),
			nil)
	}
	if rv.Elem().Kind() != reflect.Struct && !(rv.Elem().Kind() == reflect.Map && rv.Elem().Type().Key().Kind() == reflect.String) {
		return awserr.New("SerializationError",
			fmt.Sprintf("v must be a non-nil pointer to a map[string]interface{} or struct, got %s",
				rv.Type()),
			nil)
	}

	m := make(map[string]interface{})
	for k, v := range item {
		m[k] = convertFrom(v, tableName)
	}

	if isTyped(reflect.TypeOf(v)) {
		err = convertToTyped(m, v)
	} else {
		rv.Elem().Set(reflect.ValueOf(m))
	}
	return err
}

func convertToTyped(in, out interface{}) error {
	b, err := json.Marshal(in)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(bytes.NewReader(b))
	return decoder.Decode(&out)
}

func isTyped(v reflect.Type) bool {
	switch v.Kind() {
	case reflect.Struct:
		return true
	case reflect.Array, reflect.Slice:
		if isTyped(v.Elem()) {
			return true
		}
	case reflect.Map:
		if isTyped(v.Key()) {
			return true
		}
		if isTyped(v.Elem()) {
			return true
		}
	case reflect.Ptr:
		return isTyped(v.Elem())
	}
	return false
}

func ChangeQueryResponseColumn(tableName string, obj map[string]interface{}) map[string]interface{} {
	_, ok := models.TableColChangeMap[tableName]
	if !ok {
		return obj
	}
	Items, ok := obj["Items"]
	if ok {
		m, ok := Items.([]map[string]interface{})
		if ok {
			obj["Items"] = ChangesArrayResponseToOriginalColumns(tableName, m)
		}
	}
	LastEvaluatedKey, ok := obj["LastEvaluatedKey"]
	if ok {
		m, ok := LastEvaluatedKey.(map[string]interface{})
		if ok {
			obj["LastEvaluatedKey"] = ChangeResponseToOriginalColumns(tableName, m)
		}
	}
	return obj
}
