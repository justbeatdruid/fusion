curl localhost:8001/api/v1/datasources -H'content-type:application/json' \
  -H 'userId:0' -H'tenantId:default' \
  -d'
{
  "data":{
    "name": "test-datawarehouse-datasource",
    "type": "datawarehouse",
    "datawarehouse": {
      "databaseName": "BS_health_medical",
      "subjectName": "卫生健康服务"
    }
  }
}'
