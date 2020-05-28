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
	Columns []Column        `json:"columns"`
	Data    [][]interface{} `json:"data"`
	Error   Error           `json:"error"`
}

type Response struct {
	ID      string `json:"id"`
	InfoUri string `json:"infoUri"`
	NextUrl string `json:"nextUri"`
	Stats   Stats  `json:"stats"`
	Data    []Data `json:"data"`
	Error   Error  `json:"error"`
}

type Data map[string]interface{}

type Column struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type Stats struct {
	State  string `json:"state"` //FINISHED/PLANNING/QUEUED/FAILED
	Queued bool   `json:"queued"`
}
type Error struct {
	Message       string        `json:"message"`
	ErrorCode     int           `json:"errorCode"`
	ErrorName     string        `json:"errorName"`
	ErrorType     string        `json:"errorType"`
	ErrorLocation ErrorLocation `json:"errorLocation"`
	FailureInfo   FailureInfo   `json:"failureInfo"`
}

type ErrorLocation struct {
	LineNumber   int `json:"lineNumber"`
	ColumnNumber int `json:"columnNumber"`
}
type FailureInfo struct {
	Type          string        `json:"type"`
	Message       string        `json:"message"`
	Cause         Cause         `json:"cause"`
	Suppressed    interface{}   `json:"suppressed"`
	Stack         []string      `json:"stack"`
	ErrorLocation ErrorLocation `json:"errorLocation"`
}

type Cause struct {
	Type       string      `json:"type"`
	Suppressed interface{} `json:"suppressed"`
	Stack      []string    `json:"stack"`
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
	headerKey    = "X-Presto-User"
	Finished     = "FINISHED"
	Planning     = "PLANNING"
	Queued       = "QUEUED"
	Failed       = "FAILED"
)

//提交Pulsar SQL查询请求
func (c *Connector) CreateQueryRequest(sql string) (*Response, error) {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)
	request.Header.Set(headerKey, c.PrestoUser)
	url := fmt.Sprintf("%s://%s:%d%s", protocol, c.Host, c.Port, statementUrl)

	responseBody := &ResponseBody{}
	request.Header.Add("Content-Type","application/json")
	response, _, errs := request.Post(url).Type(gorequest.TypeText).Send(sql).EndStruct(responseBody)
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
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)
	//request.Header.Set(headerKey, c.PrestoUser)
	request.Header.Add("Content-Type","application/json")

	responseBody := &ResponseBody{}
	response, _, errs := request.Get(url).Retry(10, 1000*1000, http.StatusNotFound).EndStruct(responseBody)
	if errs != nil {
		return nil, fmt.Errorf("query pulsar sql query result error: %+v", errs)
	}
	if response.StatusCode == http.StatusOK {

		return c.ToResponse(responseBody), nil
	}

	return nil, fmt.Errorf("query pulsar sql query result error, status code is: %+v", response.StatusCode)
}

func (c *Connector) ToResponse(resBody *ResponseBody) *Response {
	var datas = make([]Data, 0)
	if resBody.Data != nil {
		for _, d := range resBody.Data {
			var data = make(map[string]interface{})
			for index, column := range resBody.Columns {
				key := column.Name
				value := d[index]
				data[key] = value
			}
			datas = append(datas, data)
		}
	}
	return &Response{
		ID:      resBody.ID,
		InfoUri: resBody.InfoUri,
		NextUrl: resBody.NextUrl,
		Stats:   resBody.Stats,
		Data:    datas,
		Error:   resBody.Error,
	}

}
