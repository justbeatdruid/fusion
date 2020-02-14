package datawarehouse

import (
	"encoding/json"
	"fmt"

	"github.com/chinamobile/nlpt/crds/datasource/datawarehouse/api/v1"
)

var example = `
{
	"data": [{
		"databaseName": "traffic",
		"table_property": [{
			"tableName": "logical",
			"tableType": "事实逻辑表",
			"tags": ["交通", "事故"],
			"desc": "交通事故逻辑表信息",
			"property": [{
					"tableId": "2",
					"id": 1,
					"name": "TRAFFIC_ID",
					"displayName": "事故ID",
					"unique": "是",
					"dataType": "整型",
					"length": 15,
					"desc": "",
					"encryption": "不加密",
					"encrypAlgorithm": "",
					"primaryKey": "是"
				},
				{
					"id": 5,
					"name": "TRAFFIC_LOCATION",
					"displayName": "发生地点",
					"unique": "否",
					"dataType": "通用字符串",
					"length": 64,
					"desc": "事故发生地点信息",
					"encryption": "不加密",
					"encrypAlgorithm": ""
				}
			]
		}, {
			"tableName": "dimension",
			"tableType": "维度表",
			"tags": ["维度"],
			"desc": "维度信息表",
			"property": [{
					"tableId": "1",
					"id": 3,
					"name": "AA",
					"displayName": "字段A",
					"unique": "是",
					"dataType": "整型",
					"length": 15,
					"desc": "",
					"encryption": "不加密",
					"encrypAlgorithm": "",
					"primaryKey": "是"
				},
				{
					"id": 5,
					"name": "BB",
					"displayName": "字段B",
					"unique": "否",
					"dataType": "通用字符串",
					"length": 64,
					"desc": "字段B描述",
					"encryption": "不加密",
					"encrypAlgorithm": ""
				}
			]
		}]
	}]
}`

func GetExampleDatawarehouse() (*v1.Datawarehouse, error) {
	b := []byte(example)
	result := &v1.Datawarehouse{}
	if err := json.Unmarshal(b, result); err != nil {
		return nil, fmt.Errorf("unmarshal datawarehouse struct error: %+v", err)
	}
	return result, nil
}
