curl localhost:8001/api/v1/apis -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
{
  "data": {
    "name": "xuxutest",
    "apitype":"public",
    "authtype":"APPAUTH",
    "serviceunit": {
      "id": "4dadb18b24b7a783",
      "name": "test",
      "group": "testgroup"
    },
    "users": [
      {
        "name": "xuxu",
        "inCharge": true
      }
    ],
    "frequency": 10,
    "method": "GET",
    "protocol": "HTTP",
    "returnType": "json",
    "parameters": [
      {
        "name": "para1",
	"type": "int",
	"operator": "equal",
	"description": "this is first parameter",
	"example": "0",
	"required": true
      },
      {
        "name": "para2",
	"type": "string",
	"operator": "equal",
	"description": "this is second parameter",
	"example": "hello",
	"required": false
      }
    ]
  }
}'
