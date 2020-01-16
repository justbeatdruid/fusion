curl localhost:8001/api/v1/apis -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "xuxutest",
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
