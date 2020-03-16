curl localhost:8001/api/v1/apis -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
	-H'userId:6' -H'tenantId:3000' \
  -d'
{
  "data": {
    "name": "dwtest3",
    "apitype": "public",
    "authtype": "APPAUTH",
    "serviceunit": {
      "id": "8b49f77c55c75832"
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
      "primaryTableId": "d495cf88298e4467b74c3feb01371c03",
      "properties": [
      {
        "tableId": "9cc854934fc44e1e856690b0dd889fea",
	"tableName": "tb_aqi_grade",
        "propertyId": "1164",
	"propertyName": "grddesc",
	"physicalType": "varchar",
	"operator": "",
	"withGroupby": true
      },
      {
        "tableId": "da57f73a052545efae92253032451765",
	"tableName": "tb_monitor_place",
        "propertyId": "1093",
	"propertyName": "monpl_distdiv_nm",
	"physicalType": "varchar",
	"operator": "",
	"withGroupby": true
      },
      {
        "tableId": "d495cf88298e4467b74c3feb01371c03",
	"tableName": "tb_air_quality",
        "propertyId": "1091",
	"propertyName": "aqi_grd_id",
	"physicalType": "int",
	"operator": "sum",
	"withGroupby": false
      },
      {
        "tableId": "d495cf88298e4467b74c3feb01371c03",
	"tableName": "tb_air_quality",
        "propertyId": "1056",
	"propertyName": "id",
	"physicalType": "int",
	"operator": "sum",
	"withGroupby": false
      }
      ]
    },
    "apiQueryInfo": {
      "webParams": [
      {
        "name": "tb_aqi_grade.grddesc",
	"type": "varchar",
	"location": "query",
	"required": true,
	"valueDefault": "",
	"example": "",
	"description": ""
      }
      ]
    }
  }
}'
