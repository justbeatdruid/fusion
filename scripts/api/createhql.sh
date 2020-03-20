curl localhost:8001/api/v1/apis -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
{
  "data": {
    "name": "dwtest3",
    "apitype": "public",
    "authtype": "APPAUTH",
    "serviceunit": {
      "id": "d3dd3dfd4dd6e19c"
    },
    "apiDefineInfo": {
      "path": "",
      "matchMode": "",
      "method": "GET",
      "protocol": "HTTP",
      "cors": ""
    },
    "frequency": 10,
    "method": "GET",
    "protocol": "HTTP",
    "returnType": "json",
    "datawarehouseQuery": {
      "type": "hql",
      "database": "ecosystem",
      "hql": "select * from tb_aqi_grade limit 10"
    },
    "apiQueryInfo": {
      "webParams": [
      ]
    }
  }
}'
