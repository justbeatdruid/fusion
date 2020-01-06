curl localhost:8001/api/v1/api/create -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "xuxutest",
    "serviceunit": {
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
    "method": "POST",
    "protocol": "HTTP",
    "returnType": "json",
    "parameters": [
      {
        "name": "para1",
	"type": "int",
	"description": "this is first parameter",
	"example": "0",
	"required": true
      },
      {
        "name": "para2",
	"type": "string",
	"description": "this is second parameter",
	"example": "hello",
	"required": false
      }
    ]
  }
}'
