curl localhost:8001/api/v1/serviceunits -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
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

     },
    "group": "182a3f334f835da7"
  }
}'
