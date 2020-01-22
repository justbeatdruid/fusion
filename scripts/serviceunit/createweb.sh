curl localhost:8001/api/v1/serviceunits -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "sunyu1",
    "type": "web",
    "description": "a simple serviceunit",
    "kongService": {

        "host": "192.168.7.200",
        "port": 8080,
        "protocol" : "http"

     }
  }
}'