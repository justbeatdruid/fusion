curl localhost:8001/api/v1/apis -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
{
  "data": {
    "name": "xuxutest",
    "apiType":"public",
    "authType":"APPAUTH",
    "serviceunit": {
      "id": "16b77388f08a46b4",
      "name": "test",
      "group": "testgroup"
    },
    "frequency": 10,
    "method": "GET",
    "protocol": "HTTP",
    "returnType": "json",
    "KongApi":{
      "paths":[
      "/api/v1/webapi",
      "/nlstore/v1/"
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
    ],
  "apiAttribute" :{
  "normalExample":"{code:0}" 
  }
  }
}'
