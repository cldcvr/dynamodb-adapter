// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/batchDelete/": {
            "post": {
                "description": "Batch Delete Item from table",
                "produces": [
                    "application/json"
                ],
                "summary": "Batch Delete Items from table",
                "operationId": "batch-delete-rows",
                "parameters": [
                    {
                        "description": "Please add request body of type models.BulkDelete",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.BulkDelete"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/batchGet/": {
            "post": {
                "description": "Request items in a batch.",
                "produces": [
                    "application/json"
                ],
                "summary": "Query a table",
                "operationId": "query-table",
                "parameters": [
                    {
                        "description": "Please add request body of type models.BatchMeta",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.BatchMeta"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/batchGetWithProjection/": {
            "post": {
                "description": "Request items in a batch with projections.",
                "produces": [
                    "application/json"
                ],
                "summary": "Request items in a batch with projections.",
                "operationId": "batch-get-with-projection",
                "parameters": [
                    {
                        "description": "Please add request body of type models.BatchGetWithProjectionMeta",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.BatchGetWithProjectionMeta"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/batchPut/": {
            "post": {
                "description": "Writes a record",
                "produces": [
                    "application/json"
                ],
                "summary": "Writes record",
                "operationId": "batch-put",
                "parameters": [
                    {
                        "description": "Please add request body of type models.BatchMetaUpdate",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.BatchMetaUpdate"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/deleteItem/": {
            "post": {
                "description": "Delete Item from table",
                "produces": [
                    "application/json"
                ],
                "summary": "Delete Item from table",
                "operationId": "delete-row",
                "parameters": [
                    {
                        "description": "Please add request body of type models.Delete",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.Delete"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/get/": {
            "post": {
                "description": "Get record from table",
                "produces": [
                    "application/json"
                ],
                "summary": "Get record from table",
                "operationId": "get",
                "parameters": [
                    {
                        "description": "Please add request body of type models.GetMeta",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.GetMeta"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/getWithProjection/": {
            "post": {
                "description": "Get a record with projections",
                "produces": [
                    "application/json"
                ],
                "summary": "Get a record with projections",
                "operationId": "get-with-projection",
                "parameters": [
                    {
                        "description": "Please add request body of type models.GetWithProjectionMeta",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.GetWithProjectionMeta"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/put/": {
            "post": {
                "description": "Writes a record",
                "produces": [
                    "application/json"
                ],
                "summary": "Writes record",
                "operationId": "put",
                "parameters": [
                    {
                        "description": "Please add request body of type models.Meta",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.Meta"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/query/": {
            "post": {
                "description": "Query a table",
                "produces": [
                    "application/json"
                ],
                "summary": "Query a table",
                "operationId": "query-table",
                "parameters": [
                    {
                        "description": "Please add request body of type models.Query",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.Query"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/scan/": {
            "post": {
                "description": "Scan records from table",
                "produces": [
                    "application/json"
                ],
                "summary": "Scan records from table",
                "operationId": "scan",
                "parameters": [
                    {
                        "description": "Please add request body of type models.ScanMeta",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.ScanMeta"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/scanTable/": {
            "post": {
                "description": "Scan Table",
                "produces": [
                    "application/json"
                ],
                "summary": "Scan Table",
                "operationId": "scan-table",
                "parameters": [
                    {
                        "description": "Please add request body of type models.ScanMeta",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.ScanMeta"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/update/": {
            "post": {
                "description": "updates a record in Spanner",
                "produces": [
                    "application/json"
                ],
                "summary": "updates a record in Spanner",
                "operationId": "update",
                "parameters": [
                    {
                        "description": "Please add request body of type models.UpdateAttr",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.UpdateAttr"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        },
        "/updateAttribute": {
            "post": {
                "description": "update attribute a record in Spanner",
                "produces": [
                    "application/json"
                ],
                "summary": "updates a record in Spanner",
                "operationId": "update",
                "parameters": [
                    {
                        "description": "Please add request body of type models.UpdateAttr",
                        "name": "requestBody",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/models.UpdateAttr"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "401": {
                        "description": "{\"errorMessage\":\"API access not allowed\",\"errorCode\": \"E0005\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    },
                    "500": {
                        "description": "{\"errorMessage\":\"We had a problem with our server. Try again later.\",\"errorCode\":\"E0001\"}",
                        "schema": {
                            "$ref": "#/definitions/gin.H"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "gin.H": {
            "type": "object",
            "additionalProperties": true
        },
        "models.BatchGetWithProjectionMeta": {
            "type": "object",
            "properties": {
                "dynamoObject": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "additionalProperties": {
                            "$ref": "#/definitions/dynamodb.AttributeValue"
                        }
                    }
                },
                "expressionAttributeNames": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "keyArray": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "additionalProperties": true
                    }
                },
                "projectionExpression": {
                    "type": "string"
                },
                "tableName": {
                    "type": "string"
                }
            }
        },
        "models.BatchMeta": {
            "type": "object",
            "properties": {
                "dynamoObject": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "additionalProperties": {
                            "$ref": "#/definitions/dynamodb.AttributeValue"
                        }
                    }
                },
                "keyArray": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "additionalProperties": true
                    }
                },
                "tableName": {
                    "type": "string"
                }
            }
        },
        "models.BatchMetaUpdate": {
            "type": "object",
            "properties": {
                "arrAttrMap": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "additionalProperties": true
                    }
                },
                "dynamoObject": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "additionalProperties": {
                            "$ref": "#/definitions/dynamodb.AttributeValue"
                        }
                    }
                },
                "tableName": {
                    "type": "string"
                }
            }
        },
        "models.BulkDelete": {
            "type": "object",
            "properties": {
                "dynamoObject": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "additionalProperties": {
                            "$ref": "#/definitions/dynamodb.AttributeValue"
                        }
                    }
                },
                "keyArray": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "additionalProperties": true
                    }
                },
                "tableName": {
                    "type": "string"
                }
            }
        },
        "models.Delete": {
            "type": "object",
            "properties": {
                "conditionalExpression": {
                    "type": "string"
                },
                "dynamoObject": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/dynamodb.AttributeValue"
                    }
                },
                "dynamoObjectAttrVal": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/dynamodb.AttributeValue"
                    }
                },
                "expressionAttributeValues": {
                    "type": "object",
                    "additionalProperties": true
                },
                "primaryKeyMap": {
                    "type": "object",
                    "additionalProperties": true
                },
                "tableName": {
                    "type": "string"
                }
            }
        },
        "models.GetMeta": {
            "type": "object",
            "properties": {
                "dynamoObject": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/dynamodb.AttributeValue"
                    }
                },
                "primaryKeyMap": {
                    "type": "object",
                    "additionalProperties": true
                },
                "tableName": {
                    "type": "string"
                }
            }
        },
        "models.GetWithProjectionMeta": {
            "type": "object",
            "properties": {
                "dynamoObject": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/dynamodb.AttributeValue"
                    }
                },
                "expressionAttributeNames": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "primaryKeyMap": {
                    "type": "object",
                    "additionalProperties": true
                },
                "projectionExpression": {
                    "type": "string"
                },
                "tableName": {
                    "type": "string"
                }
            }
        },
        "models.Meta": {
            "type": "object",
            "properties": {
                "attrMap": {
                    "type": "object",
                    "additionalProperties": true
                },
                "conditionalExp": {
                    "type": "string"
                },
                "dynamoObject": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/dynamodb.AttributeValue"
                    }
                },
                "dynamoObjectAttrVal": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/dynamodb.AttributeValue"
                    }
                },
                "expressionAttributeValues": {
                    "type": "object",
                    "additionalProperties": true
                },
                "tableName": {
                    "type": "string"
                }
            }
        },
        "models.Query": {
            "type": "object",
            "properties": {
                "dynamoObject": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/dynamodb.AttributeValue"
                    }
                },
                "expressionAttributeNames": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "filterExp": {
                    "type": "string"
                },
                "filterVal": {
                    "type": "object"
                },
                "filterValDDB": {
                    "type": "string"
                },
                "hasVal": {
                    "type": "object"
                },
                "hasValDDB": {
                    "type": "string"
                },
                "hashExp": {
                    "type": "string"
                },
                "indexName": {
                    "type": "string"
                },
                "limit": {
                    "type": "integer"
                },
                "onlyCount": {
                    "type": "boolean"
                },
                "projectionExpression": {
                    "type": "string"
                },
                "rangeExp": {
                    "type": "string"
                },
                "rangeVal": {
                    "type": "object"
                },
                "rangeValDDB": {
                    "type": "string"
                },
                "rangeValMap": {
                    "type": "object",
                    "additionalProperties": true
                },
                "rangeValMapDDB": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/dynamodb.AttributeValue"
                    }
                },
                "sortAscending": {
                    "type": "boolean"
                },
                "startFrom": {
                    "type": "object",
                    "additionalProperties": true
                },
                "tableName": {
                    "type": "string"
                }
            }
        },
        "models.ScanMeta": {
            "type": "object",
            "properties": {
                "dynamoObject": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/dynamodb.AttributeValue"
                    }
                },
                "limit": {
                    "type": "integer"
                },
                "startFrom": {
                    "type": "object",
                    "additionalProperties": true
                },
                "tableName": {
                    "type": "string"
                }
            }
        },
        "models.UpdateAttr": {
            "type": "object",
            "properties": {
                "attrNames": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "attrVals": {
                    "type": "object",
                    "additionalProperties": true
                },
                "conditionalExp": {
                    "type": "string"
                },
                "dynamoObject": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/dynamodb.AttributeValue"
                    }
                },
                "dynamoObjectAttrVal": {
                    "type": "object",
                    "additionalProperties": {
                        "$ref": "#/definitions/dynamodb.AttributeValue"
                    }
                },
                "primaryKeyMap": {
                    "type": "object",
                    "additionalProperties": true
                },
                "returnValues": {
                    "type": "string"
                },
                "tableName": {
                    "type": "string"
                },
                "updateExpression": {
                    "type": "string"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "1.0",
	Host:        "localhost:9050",
	BasePath:    "/v1",
	Schemes:     []string{},
	Title:       "dynamodb-adapter APIs",
	Description: "This is a API collection for dynamodb-adapter",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}