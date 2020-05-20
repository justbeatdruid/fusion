curl localhost:8001/api/v1/apis -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
   -H 'userId:12' -H 'tenantId:74a2512097335196b3040bed704c65c1' -d'
{
  "data": {
    "name": "createapi1",
    "apiType":"public",
    "authType":"APPAUTH",
    "tag":"apitest",
    "serviceunit": {
      "id": "817f29986deee503",
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
     "method":"GET",
     "protocol":"HTTP",
     "cors":"true"
     },
     "apiQueryInfo":{
     "webParams": [
      {
        "name": "para1",
        "type": "string",
        "location": "path",
        "description": "this is first parameter",
        "valueDefault":"0",
        "example": "0",
        "required": true,
        "backend":{
          "name": "bakpara1",
          "location": "path"
         }

      },
      {
        "name": "para2",
        "type": "string",
        "location": "query",
        "description": "this is second parameter",
        "example": "hello",
        "valueDefault":"hello",
        "required": false,
         "backend":{
          "name": "bakpara2",
          "location": "path"
         }

  }]
      },
  "apiReturnInfo" :{
  "normalExample":"{code:0}"
  }
  }
}'
