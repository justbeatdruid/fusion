curl localhost:8001/api/v1/serviceunit/create -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "xuxutest",
    "description": "a simple serviceunit",
    "type": "multi",
    "multiDatasource": [
      {
	"id": "57c9172db382fb70e3c8a3db5f6b3c41d9434877a4c3f55547c247960d28d289",
        "fields": [
          {
            "originType": "int",
            "originField": "field1",
            "serviceType": "int",
            "serviceField": "field1"
          } 
	]
      }
    ]
  }
}'
