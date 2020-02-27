curl -XPATCH localhost:8001/api/v1/apis/174ee0374aeea338 -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "xuxutest",
    "apiType":"private",
    "authType":"APPAUTH",
    "serviceunit": {
      "id": "16b77388f08a46b4",
      "name": "test",
      "group": "testgroup"
    },
    "frequency": 10,
    "method": "POST",
    "protocol": "HTTP",
    "returnType": "json",
    "KongApi":{
      "paths":[
      "/api/v1/test",
      "/nlstore/v1/test"
      ]
    },
    "webParams": [
      {
        "name": "para1",
	"type": "int",
	"location": "path",
	"description": "this is first parameter",
	"valueDefault":"0",
	"example": "0",
	"required": true
      },
      {
        "name": "para2",
	"type": "string",
	"location": "query",
	"description": "this is second parameter",
	"example": "hello",
	"valueDefault":"hello",
	"required": false
  }
    ]
  }
}'
