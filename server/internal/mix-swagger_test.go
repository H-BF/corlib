package internal

import (
	"encoding/json"
	"testing"

	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

var swagger1 = `
{
  "swagger": "2.0",
  "info": {
    "title": "api/api.proto",
    "version": "version not set"
  },
  "tags": [
    {
      "name": "HealthcheckWorker"
    }
  ],
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "paths": {
    "/v1/IsHttpAdvancedCheckOk": {
      "post": {
        "operationId": "HealthcheckWorker_IsHttpAdvancedCheckOk",
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/healthcheckIsOk"
            }
          },
          "default": {
            "description": "An unexpected error response.",
            "schema": {
              "$ref": "#/definitions/rpcStatus"
            }
          }
        },
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/healthcheckHttpAdvancedData"
            }
          }
        ],
        "tags": [
          "HealthcheckWorker"
        ]
      }
    },
    "/v1/IsHttpCheckOk": {
      "post": {
        "operationId": "HealthcheckWorker_IsHttpCheckOk",
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/healthcheckIsOk"
            }
          },
          "default": {
            "description": "An unexpected error response.",
            "schema": {
              "$ref": "#/definitions/rpcStatus"
            }
          }
        },
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/healthcheckHttpData"
            }
          }
        ],
        "tags": [
          "HealthcheckWorker"
        ]
      }
    },
    "/v1/IsHttpsCheckOk": {
      "post": {
        "operationId": "HealthcheckWorker_IsHttpsCheckOk",
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/healthcheckIsOk"
            }
          },
          "default": {
            "description": "An unexpected error response.",
            "schema": {
              "$ref": "#/definitions/rpcStatus"
            }
          }
        },
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/healthcheckHttpsData"
            }
          }
        ],
        "tags": [
          "HealthcheckWorker"
        ]
      }
    },
    "/v1/IsIcmpCheckOk": {
      "post": {
        "operationId": "HealthcheckWorker_IsIcmpCheckOk",
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/healthcheckIsOk"
            }
          },
          "default": {
            "description": "An unexpected error response.",
            "schema": {
              "$ref": "#/definitions/rpcStatus"
            }
          }
        },
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/healthcheckIcmpData"
            }
          }
        ],
        "tags": [
          "HealthcheckWorker"
        ]
      }
    },
    "/v1/IsTcpCheckOk": {
      "post": {
        "operationId": "HealthcheckWorker_IsTcpCheckOk",
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/healthcheckIsOk"
            }
          },
          "default": {
            "description": "An unexpected error response.",
            "schema": {
              "$ref": "#/definitions/rpcStatus"
            }
          }
        },
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/healthcheckTcpData"
            }
          }
        ],
        "tags": [
          "HealthcheckWorker"
        ]
      }
    }
  },
  "definitions": {
    "healthcheckHttpAdvancedData": {
      "type": "object",
      "properties": {
        "healthcheckType": {
          "type": "string"
        },
        "healthcheckAddress": {
          "type": "string"
        },
        "nearFieldsMode": {
          "type": "boolean"
        },
        "userDefinedData": {
          "type": "object",
          "additionalProperties": {
            "type": "string"
          }
        },
        "timeout": {
          "type": "string"
        },
        "fwmark": {
          "type": "string",
          "format": "int64"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "healthcheckHttpData": {
      "type": "object",
      "properties": {
        "healthcheckAddress": {
          "type": "string"
        },
        "timeout": {
          "type": "string"
        },
        "fwmark": {
          "type": "string",
          "format": "int64"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "healthcheckHttpsData": {
      "type": "object",
      "properties": {
        "healthcheckAddress": {
          "type": "string"
        },
        "timeout": {
          "type": "string"
        },
        "fwmark": {
          "type": "string",
          "format": "int64"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "healthcheckIcmpData": {
      "type": "object",
      "properties": {
        "ipS": {
          "type": "string"
        },
        "timeout": {
          "type": "string"
        },
        "fwmark": {
          "type": "string",
          "format": "int64"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "healthcheckIsOk": {
      "type": "object",
      "properties": {
        "isOk": {
          "type": "boolean"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "healthcheckTcpData": {
      "type": "object",
      "properties": {
        "healthcheckAddress": {
          "type": "string"
        },
        "timeout": {
          "type": "string"
        },
        "fwmark": {
          "type": "string",
          "format": "int64"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "protobufAny": {
      "type": "object",
      "properties": {
        "typeUrl": {
          "type": "string"
        },
        "value": {
          "type": "string",
          "format": "byte"
        }
      }
    },
    "rpcStatus": {
      "type": "object",
      "properties": {
        "code": {
          "type": "integer",
          "format": "int32"
        },
        "message": {
          "type": "string"
        },
        "details": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/protobufAny"
          }
        }
      }
    }
  }
}
`
var swagger2 = `
{
  "swagger": "2.0",
  "info": {
    "title": "api/api.proto",
    "version": "version not set"
  },
  "tags": [
    {
      "name": "HealthcheckWorker2"
    }
  ],
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "paths": {
    "/v2/IsHttpAdvancedCheckOk": {
      "post": {
        "operationId": "HealthcheckWorker2_IsHttpAdvancedCheckOk",
        "responses": {
          "200": {
            "description": "A successful response.",
            "schema": {
              "$ref": "#/definitions/healthcheckIsOk"
            }
          },
          "default": {
            "description": "An unexpected error response.",
            "schema": {
              "$ref": "#/definitions/rpcStatus"
            }
          }
        },
        "parameters": [
          {
            "name": "body",
            "in": "body",
            "required": true,
            "schema": {
              "$ref": "#/definitions/healthcheckHttpAdvancedData"
            }
          }
        ],
        "tags": [
          "HealthcheckWorker2"
        ]
      }
    }
  },
  "definitions": {
    "healthcheckHttpAdvancedData": {
      "type": "object",
      "properties": {
        "healthcheckType": {
          "type": "string"
        },
        "healthcheckAddress": {
          "type": "string"
        },
        "nearFieldsMode": {
          "type": "boolean"
        },
        "userDefinedData": {
          "type": "object",
          "additionalProperties": {
            "type": "string"
          }
        },
        "timeout": {
          "type": "string"
        },
        "fwmark": {
          "type": "string",
          "format": "int64"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "healthcheckHttpData": {
      "type": "object",
      "properties": {
        "healthcheckAddress": {
          "type": "string"
        },
        "timeout": {
          "type": "string"
        },
        "fwmark": {
          "type": "string",
          "format": "int64"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "healthcheckHttpsData": {
      "type": "object",
      "properties": {
        "healthcheckAddress": {
          "type": "string"
        },
        "timeout": {
          "type": "string"
        },
        "fwmark": {
          "type": "string",
          "format": "int64"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "healthcheckIcmpData": {
      "type": "object",
      "properties": {
        "ipS": {
          "type": "string"
        },
        "timeout": {
          "type": "string"
        },
        "fwmark": {
          "type": "string",
          "format": "int64"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "healthcheckIsOk": {
      "type": "object",
      "properties": {
        "isOk": {
          "type": "boolean"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "healthcheckTcpData": {
      "type": "object",
      "properties": {
        "healthcheckAddress": {
          "type": "string"
        },
        "timeout": {
          "type": "string"
        },
        "fwmark": {
          "type": "string",
          "format": "int64"
        },
        "id": {
          "type": "string"
        }
      }
    },
    "protobufAny": {
      "type": "object",
      "properties": {
        "typeUrl": {
          "type": "string"
        },
        "value": {
          "type": "string",
          "format": "byte"
        }
      }
    },
    "rpcStatus": {
      "type": "object",
      "properties": {
        "code": {
          "type": "integer",
          "format": "int32"
        },
        "message": {
          "type": "string"
        },
        "details": {
          "type": "array",
          "items": {
            "$ref": "#/definitions/protobufAny"
          }
        }
      }
    }
  }
}
`

func Test_SwaggerCompose(t *testing.T) {
	var sw1 *spec.Swagger
	var sw2 *spec.Swagger
	err := json.Unmarshal([]byte(swagger1), &sw1)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	err = json.Unmarshal([]byte(swagger2), &sw2)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	err = SwaggerComposer{}.Compose(sw1, sw2)
	assert.NoError(t, err)
	if err != nil {
		return
	}
	assert.Equal(t, 2, len(sw1.Tags))
	_, found := sw1.Paths.Paths["/v2/IsHttpAdvancedCheckOk"]
	assert.Equal(t, true, found)
}
