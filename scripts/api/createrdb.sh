curl localhost:8001/api/v1/apis -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
{
  "data": {
    "name": "rdbtest222",
    "apitype": "public",
    "authtype": "APPAUTH",
    "serviceunit": {
      "id": "f0306883cd3f47e9"
    },
    "frequency": 10,
    "method": "GET",
    "protocol": "HTTP",
    "returnType": "json",
    "rdbQuery": {
      "table": "table name",
      "queryFields" : [{
        "field": "field1",
        "type": "int",
        "operator": "",
        "description": ""
      }]
    },
    "apiQueryInfo": {
      "webParams": [
      {
        "name": "id",
	"type": "varchar",
	"location": "query",
	"required": true,
	"valueDefault": "",
	"example": "",
	"description": ""
      }
      ]
    }
  }
}'
