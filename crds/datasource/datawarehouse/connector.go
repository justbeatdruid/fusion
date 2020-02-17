package datawarehouse

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/crds/datasource/datawarehouse/api/v1"
	"github.com/chinamobile/nlpt/pkg/logs"

	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
)

type Connector interface {
	GetExampleDatawarehouse() (*v1.Datawarehouse, error)
}

type httpConnector struct {
	Host string
	Port int
}

var metadataRequestBody = struct {
	User string `json:"userId"`
}{
	User: "admin",
}

var headers = map[string]string{
	"Content-Type": "application/json",
}

var metadataRequestPath = "/cmcc/data/service/dataService/metadata/getMetadataInfo"

func (c *httpConnector) GetExampleDatawarehouse() (*v1.Datawarehouse, error) {
	request := gorequest.New().SetLogger(logs.GetGoRequestLogger(4)).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, c.Host, c.Port, metadataRequestPath))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second)

	responseBody := &v1.Datawarehouse{}
	response, body, errs := request.Send(metadataRequestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return nil, fmt.Errorf("request for getting metadata error: %+v", errs)
	}
	klog.V(5).Infof("creation response body: %s", string(body))
	if response.StatusCode/100 != 2 {
		klog.V(5).Infof("create operation failed: %d %s", response.StatusCode, string(body))
		return nil, fmt.Errorf("request for getting metadata error: receive wrong status code: %s", string(body))
	}

	return responseBody, nil
}

func NewConnector(host string, port int) Connector {
	return &httpConnector{
		Host: host,
		Port: port,
	}
}

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

type fakeConnector struct{}

func (fakeConnector) GetExampleDatawarehouse() (*v1.Datawarehouse, error) {
	b := []byte(example)
	result := &v1.Datawarehouse{}
	if err := json.Unmarshal(b, result); err != nil {
		return nil, fmt.Errorf("unmarshal datawarehouse struct error: %+v", err)
	}
	return result, nil
}
