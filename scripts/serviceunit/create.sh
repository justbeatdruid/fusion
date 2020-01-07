curl localhost:8001/api/v1/serviceunit/create -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "xuxutest",
    "description": "a simple serviceunit",
    "type": "multi",
    "multiDatasource": [
      {
	"id": "123456",
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
