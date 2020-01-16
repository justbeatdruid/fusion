curl localhost:8001/api/v1/serviceunits -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "newtest",
    "description": "a simple serviceunit",
    "datasources": [
      {
	"id": "57c9172db382fb70e3c8a3db5f6b3c41d9434877a4c3f55547c247960d28d289"
      }
    ]
  }
}'
