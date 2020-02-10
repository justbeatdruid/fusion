curl localhost:8001/api/v1/serviceunitgroups -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "xuxutest",
    "description": "this is a test serviceunit group"
  }
}'
