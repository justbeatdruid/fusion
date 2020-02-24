curl localhost:8001/api/v1/serviceunits -H 'content-type:application/json' \
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
