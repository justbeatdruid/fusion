curl localhost:8001/api/v1/applicationgroups -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "testgroup",
    "description": "this is a test application group"
  }
}'
