curl localhost:8001/api/v1/datasources -H 'content-type:application/json' \
  -d'
{
  "data": {
    "name": "testdatasource",
    "type": "rdb",
    "rdb": {
      "type": "mysql",
      "database": "mysql",
      "connect": {
        "host": "119.3.248.187",
	"port": 3306,
	"username": "root",
	"password": "my-secret-pw"
      }
    }
  }
}'
