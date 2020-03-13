package datawarehouse

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/pkg/datawarehouse/api/v1"
	"github.com/chinamobile/nlpt/pkg/logs"

	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
)

type Connector interface {
	GetExampleDatawarehouse() (*v1.Datawarehouse, error)
	QueryData(v1.Query) (v1.Result, error)
}

type httpConnector struct {
	MetadataHost string
	MetadataPort int
	DataHost     string
	DataPort     int
}

var metadataRequestBody = struct {
	User string `json:"userId"`
}{
	User: "admin",
}

var headers = map[string]string{
	"Content-Type": "application/json",
}

const metadataRequestPath = "/cmcc/data/service/dataService/metadata/getMetadataInfo"

func (c *httpConnector) GetExampleDatawarehouse() (*v1.Datawarehouse, error) {
	request := gorequest.New().SetLogger(logs.GetGoRequestLogger(6)).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, c.MetadataHost, c.MetadataPort, metadataRequestPath))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second)

	responseBody := &v1.Datawarehouse{}
	response, body, errs := request.Send(metadataRequestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return nil, fmt.Errorf("request for getting metadata error: %+v", errs)
	}
	klog.V(6).Infof("creation response body: %s", string(body))
	if response.StatusCode/100 != 2 {
		klog.V(5).Infof("create operation failed: %d %s", response.StatusCode, string(body))
		return nil, fmt.Errorf("request for getting metadata error: receive wrong status code: %s", string(body))
	}
	if responseBody == nil {
		return nil, fmt.Errorf("get null response")
	}
	if len(responseBody.Databases) == 0 {
		return nil, fmt.Errorf("cannot get one database from response")
	}
	return responseBody, nil
}

const dataRequestPath = "/cmcc/data/service/SqlAssemble/getQueryResult"

//const dataRequestPath = "/cmcc/data/service/dataService/query/getQueryResult"

type WappedResult struct {
	Success bool      `json:"success"`
	Code    int       `json:"code"`
	Message string    `json:"message"`
	Data    v1.Result `json:"data"`
}

func (c *httpConnector) QueryData(q v1.Query) (v1.Result, error) {
	q.UserID = "admin"
	if len(q.DatabaseName) == 0 {
		return v1.Result{}, fmt.Errorf("database not set")
	}
	request := gorequest.New().SetLogger(logs.GetGoRequestLogger(6)).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, c.DataHost, c.DataPort, dataRequestPath))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second)

	responseBody := &WappedResult{}
	response, body, errs := request.Send(&q).EndStruct(responseBody)
	if len(errs) > 0 {
		return v1.Result{}, fmt.Errorf("request for quering data error: %+v", errs)
	}
	klog.V(6).Infof("creation response body: %s", string(body))
	if response.StatusCode/100 != 2 {
		klog.V(5).Infof("create operation failed: %d %s", response.StatusCode, string(body))
		return v1.Result{}, fmt.Errorf("request for quering data error: receive wrong status code: %s", string(body))
	}

	return responseBody.Data, nil
}

func NewConnector(metadatahost string, metadataport int, datahost string, dataport int) Connector {
	return &httpConnector{
		MetadataHost: metadatahost,
		MetadataPort: metadataport,
		DataHost:     datahost,
		DataPort:     dataport,
	}
}

type fakeConnector struct{}

func (fakeConnector) GetExampleDatawarehouse() (*v1.Datawarehouse, error) {
	b := []byte(example)
	result := &v1.Datawarehouse{}
	if err := json.Unmarshal(b, result); err != nil {
		return nil, fmt.Errorf("unmarshal datawarehouse struct error: %+v", err)
	}
	return result, nil
}

func (fakeConnector) QueryData(v1.Query) (v1.Result, error) {
	return v1.Result{}, nil
}

func FakeConnector() Connector {
	return fakeConnector{}
}
