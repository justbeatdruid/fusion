
curl -XPUT localhost:8001/api/v1/apis -H 'content-type:application/json'  -v -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
{
  "data": {
    "operation": "delete",
    "apis": [
    { "id": "eb3988260f3c3ab2"
   },
   {"id":"910fc00e905b9bdf" }

    ]
  }
}'