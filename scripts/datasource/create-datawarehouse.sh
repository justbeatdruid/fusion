curl localhost:8001/api/v1/datasources -H'content-type:application/json' \
  -H 'userId:0' -H'tenantId:default' \
  -d'
{
  "data":{
    "name": "test123",
    "type": "datawarehouse",
    "datawarehouse": {
      "databaseName": "aaa",
      "subjectName": "bbb"
    }
  }
}'
