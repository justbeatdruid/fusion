curl localhost:8001/api/v1/datasources/connection -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "testdatasource",
    "type": "pulsar",
    "mq": {
      "type": "public",
      "mqConnection": {
        "address": "10.160.32.24:8080,10.160.32.24:80"
      }    
    }
  }
}'
