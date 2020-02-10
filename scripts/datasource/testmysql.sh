curl -XGET 127.0.0.1:8001/api/v1/datasources/ConnectMysql -v -H'content-type:application/json' -d'
{
  "userName": "root",
  "password": "password",
  "ip": "127.0.0.1",
  "port": "3306",
  "dbName": "mysql",
  "tableName": "mysql"
}
'
