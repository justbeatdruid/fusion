curl localhost:8001/api/v1/apis -H 'content-type:application/json' -H 'X-auth-Token:8c06d2a4d87a8e4df280589889d85f67e6d755b2fb33538fd724ba6f700361a0' \
  -d'
{
  "data": {
    "name": "dwtest3",
    "apitype": "public",
    "authtype": "APPAUTH",
    "serviceunit": {
      "id": "c716e0f728e9358b"
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
      "primaryTableId": "54d4a267dae14cf1a560e497986177e7",
      "properties": [
      {
        "tableId": "54d4a267dae14cf1a560e497986177e7",
	"tableName": "fat_se_dwd_outpatient_diagnosis_di",
        "propertyId": "2933",
	"propertyName": "diagnosis_lx_code",
	"physicalType": "varchar"
      },
      {
        "tableId": "54d4a267dae14cf1a560e497986177e7",
	"tableName": "fat_se_dwd_outpatient_diagnosis_di",
        "propertyId": "2934",
	"propertyName": "diagnosis_lx",
	"physicalType": "varchar"
      }
      ]
    },
    "apiQueryInfo": {
      "webParams": [
      {
        "name": "id",
	"type": "int",
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
