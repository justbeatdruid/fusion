curl localhost:8001/api/v1/serviceunits -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
{
  "data": {
    "name": "newtest",
    "type": "data",
    "description": "a simple serviceunit",
    "datasources":
      {
	"id": "123d4b16dc745343"
      },
    "group": "182a3f334f835da7"
  }
}'
