curl localhost:8001/api/v1/dataservices -H 'content-type:application/json' \
  -d'
{
  "data": {
    "type": "periodic",
    "cronConfig": "0 9-18/1 * * *"
  }
}'
