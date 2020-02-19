curl localhost:8001/api/v1/apis -H 'content-type:application/json' \
  -d'
{
	"data": {
		"name": "testapi",
		"apitype": "public",
		"authtype": "APPAUTH",
		"serviceunit": {
			"id": "88aca60ab57c22c3",
			"name": "testserviceunit",
			"group": "182a3f334f835da7"
		},
		"users": [{
			"name": "xuxu",
			"inCharge": true
		}],
		"frequency": 10,
		"method": "GET",
		"protocol": "HTTP",
		"returnType": "json",
		"dataserviceQuery": {
			"associationTable": [{
				"associationPropertyName": "aqi_grd_id",
				"associationTableName": "tb_air_quality",
				"propertyName": "id",
				"tableName": "tb_aqi_grade"
			}, {
				"associationPropertyName": "monpl_id",
				"associationTableName": "tb_air_quality",
				"propertyName": "id",
				"tableName": "tb_monitor_place"
			}],
			"databaseName": "ecosystem",
			"limitNum": 10,
			"primaryTableName": "tb_air_quality",
			"queryFieldList": [{
					"propertyName": "grddesc",
					"tableName": "tb_aqi_grade",
					"propertyDisplayName": "描述"
				}, {
					"propertyName": "monpl_distdiv_nm",
					"tableName": "tb_monitor_place",
					"propertyDisplayName": "省份名称"
				}, {
					"propertyName": "aqi_grd_id",
					"tableName": "tb_air_quality",
					"operator": "sum",
					"propertyDisplayName": "空气质量等级id"
				},
				{
					"propertyName": "id",
					"tableName": "tb_air_quality",
					"operator": "sum",
					"propertyDisplayName": "id"
				}
			],
			"groupByFieldInfo": [{
					"propertyName": "grddesc",
					"tableName": "tb_aqi_grade"
				},
				{
					"propertyName": "monpl_distdiv_nm",
					"tableName": "tb_monitor_place"
				}
			],

			"whereFieldInfo": [{
				"operator": "equal",
				"propertyName": "grddesc",
				"tableName": "tb_aqi_grade",
				"value": [
					"严重"
				],
				"parameterEnabled": true,
				"type": "string",
				"example": "严重",
				"description": "空气质量等级参数",
				"required": true
			}],
			"userId": "admin"
		}
	}
}'
