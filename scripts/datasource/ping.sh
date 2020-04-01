curl localhost:8001/api/v1/datasources/connection -H 'content-type:application/json' \
  -d'
{
  "data": {
    "type": "rdb",
    "rdb": {
      "type": "mysql",
      "database": "mysql",
      "connect": {
        "host": "10.160.32.24",
	"port": 3306,
	"username": "root",
	"password": "123456"
      }
    }
  }
}'
