curl -XPATCH localhost:8001/api/v1/apis/8e46df20ec899432 -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
{
  "data": {
    "name": "xuxutest",
    "apiType":"public",
    "authType":"APPAUTH",
    "tag":"apitest",
    "serviceunit": {
      "id": "44f747227d7e85a5",
      "name": "test",
      "group": "testgroup"
    },
    "frequency": 10,
    "KongApi":{
      "paths":[
      "/api/v1/webapi"
      ]
    },
    "apiDefineInfo":{
    "path":"/api/v1/webapi/test",
     "method":"POST",
     "cors":"false",
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
  }]
    },
  "apiReturnInfo" :{
  "normalExample":"{code:0}"
  }
  }
}'

