curl localhost:8001/api/v1/serviceunits -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -H 'userId:0' -H'tenantId:74a2512097335196b3040bed704c65c1'  -d'
{
  "data": {
    "name": "func7",
    "type":"function",
    "description": "a simple serviceunit",
    "fissionRefInfo": {
        "language": "nodejs",
        "fnName": "func7",
        "fnFile" : "/data/upload/serviceunit/rsp.js",
        "entryPoint": "module.exports"
     }
  }
}'