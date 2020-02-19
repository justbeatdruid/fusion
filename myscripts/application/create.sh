curl localhost:8001/api/v1/applications -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "testapp",
    "group": "9966cdb0a1c18ca0",
    "description": "this is a test application"
  }
}'
