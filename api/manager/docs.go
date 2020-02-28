// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2020-02-28 18:10:14.0576964 +0100 CET m=+2.249568017

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
        "/manager/job/execute": {
            "post": {
                "description": "This method will execute the job sent via the post body",
                "consumes": [
                    "text/plain"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "jobs"
                ],
                "summary": "executes a job",
                "operationId": "execute-job",
                "parameters": [
                    {
                        "description": "Pipeline Script",
                        "name": "pipeline",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "type": "string"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/manager.SuccessResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/manager.FailureResponse"
                        }
                    }
                }
            }
        },
        "/manager/job/info/{job_id}": {
            "get": {
                "description": "This method will return a single job by it's id or an error.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "jobs"
                ],
                "summary": "find a single job by it's id",
                "operationId": "job-info",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Job ID",
                        "name": "job_id",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/manager.JobInfoResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/manager.FailureResponse"
                        }
                    }
                }
            }
        },
        "/manager/job/list": {
            "get": {
                "description": "This method will return a list of all jobs",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "jobs"
                ],
                "summary": "lists all jobs",
                "operationId": "job-list",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/manager.JobListResponse"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/manager.FailureResponse"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "manager.FailureResponse": {
            "type": "object",
            "properties": {
                "error": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "manager.JobInfoResponse": {
            "type": "object",
            "properties": {
                "job": {
                    "type": "object",
                    "$ref": "#/definitions/model.Job"
                }
            }
        },
        "manager.JobListResponse": {
            "type": "object",
            "properties": {
                "jobs": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.Job"
                    }
                }
            }
        },
        "manager.SuccessResponse": {
            "type": "object",
            "properties": {
                "status": {
                    "type": "string"
                }
            }
        },
        "model.Job": {
            "type": "object",
            "properties": {
                "createdAt": {
                    "type": "string"
                },
                "id": {
                    "type": "string"
                },
                "logs": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "pipeline": {
                    "type": "string"
                },
                "progress": {
                    "type": "integer"
                },
                "stages": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.JobStage"
                    }
                },
                "updatedAt": {
                    "type": "string"
                }
            }
        },
        "model.JobStage": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "progress": {
                    "type": "integer"
                },
                "tasks": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/model.JobStageTask"
                    }
                }
            }
        },
        "model.JobStageTask": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "integer"
                },
                "name": {
                    "type": "string"
                },
                "progress": {
                    "type": "integer"
                },
                "stderr": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "stdout": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
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
	Version:     "0.1",
	Host:        "",
	BasePath:    "/",
	Schemes:     []string{},
	Title:       "dips",
	Description: "dips manager api",
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
