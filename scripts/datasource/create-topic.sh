curl localhost:8001/api/v1/datasources -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "testdatasource",
    "type": "topic",
    "topic": "59289baac76a2086"
    }
  }
}'
