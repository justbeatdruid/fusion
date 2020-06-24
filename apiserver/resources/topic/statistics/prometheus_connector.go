package statistics

import (
	"fmt"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
	"net/http"
)

type Connector struct {
	Host string `json:"host"`
	Port int `json:"port"`
}
type Request struct {
	Query string `json:"query"`
	Start int64 `json:"start"`
	End   int64  `json:"end"`
	Step  int   `json:"step"`
}
type Response struct{
	Status string `json:"status"` //success or error
	Data Data `json:"data"`
	ErrorType  string `json:"errorType"`
	Error      string `json:"error"`
	Warnings   []string `json:"warnings"`
}

type Data struct {
	ResultType string `json:"resultType"`
	Result    []Result `json:"result"`
}

type Result struct {
	Metric map[string]string `json:"metric"`
	Value []interface{} `json:"value"`
}

const (
	QueryRangeUrl = "http://%s:%d/api/v1/query_range?query=%s&start=%d&end=%d&step=%d"
	QuerySumOverRange = "sum_over_time(pulsar_storage_size[1d])"
)

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

func (c *Connector) QueryRange(r *Request) (*Response, error) {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)

	url := fmt.Sprintf(QueryRangeUrl, c.Host, c.Port, r.Query, r.Start, r.End, r.Step)

	responseBody := &Response{}
	response, body, errs := request.Get(url).EndStruct(responseBody)
	if errs != nil {
		return nil, fmt.Errorf("query from prometheus error, response: %+v, body: %+v, error: %+v", response, body, errs)
	}

	if response.StatusCode ==  http.StatusOK {
		return responseBody, nil
	}

	return nil, fmt.Errorf("query from ")

}