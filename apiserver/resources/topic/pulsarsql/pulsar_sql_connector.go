package pulsarsql

import (
	"fmt"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
	"net/http"
)

type Connector struct {
	PrestoUser string `json:"prestoUser"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
}

type ResponseBody struct {
	ID      string          `json:"id"`
	InfoUri string          `json:"infoUri"`
	NextUrl string          `json:"nextUri"`
	Stats   Stats           `json:"stats"`
	Columns Column          `json:"columns"`
	Data    [][]interface{} `json:"data"`
}

type Response struct {
	ID      string `json:"id"`
	InfoUri string `json:"infoUri"`
	NextUrl string `json:"nextUri"`
	Stats   Stats  `json:"stats"`
	Data    []Data `json:"data"`
}

type Data map[string]interface{}

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Stats struct {
	State  string `json:"state"` //FINISHED/PLANNING/QUEUED
	Queued bool   `json:"queued"`
}

type requestLogger struct {
	prefix string
}

var logger = &requestLogger{}

func (r *requestLogger) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *requestLogger) Printf(format string, v ...interface{}) {
	klog.V(4).Infof(format, v...)
}

func (r *requestLogger) Println(v ...interface{}) {
	klog.V(4).Infof("%+v", v)
}

const (
	statementUrl = "/v1/statement"
	protocol     = "http"
)

//提交Pulsar SQL查询请求
func (c *Connector) CreateQueryRequest(sql string) (*Response, error) {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)
	url := fmt.Sprintf("%s://%s:%d%s", protocol, c.Host, c.Port, statementUrl)

	responseBody := &ResponseBody{}
	response, _, errs := request.Post(url).Send(sql).EndStruct(responseBody)
	if errs != nil {
		return nil, fmt.Errorf("create pulsar sql request error: %+v", errs)
	}
	if response.StatusCode == http.StatusOK {
		return c.ToResponse(responseBody), nil
	}
	return nil, fmt.Errorf("create pulsar sql request error, status code is: %+v", response.StatusCode)
}

//查询任务状态
func (c *Connector) QueryMessage(url string) (*Response, error) {
	return nil, nil
}

func (c *Connector) ToResponse(resBody *ResponseBody) *Response{
	response := Response{}

	return &response
}