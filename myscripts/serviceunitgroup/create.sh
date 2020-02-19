curl localhost:8001/api/v1/serviceunitgroups -H 'content-type:application/json' -v \
  -d'
{
  "data": {
    "name": "testgroup",
    "description": "this is a test serviceunit group"
  }
}'
