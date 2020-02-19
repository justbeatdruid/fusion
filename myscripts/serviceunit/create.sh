curl localhost:8001/api/v1/serviceunits -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "testserviceunit",
    "description": "this is a test serviceunit",
    "type": "data",
    "datasources": {
      "id": "36ed8f501e0118a8"
    },
    "group": "182a3f334f835da7"
  }
}'
