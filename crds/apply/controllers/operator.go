package controllers

import (
	"fmt"
	"net/http"
	"time"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	nlptv1 "github.com/chinamobile/nlpt/crds/apply/api/v1"
	"github.com/parnurzeal/gorequest"

	"k8s.io/klog"
)

const path string = "/consumers"

var headers = map[string]string{
	"Content-Type": "application/json",
}
var retryStatus = []int{http.StatusBadRequest, http.StatusInternalServerError}

type Operator struct {
	Host   string
	Port   int
	CAFile string
}

//
type AddWhiteRequestBody struct {
	Group string `json:"group"`
}

type AddWhiteResponseBody struct {
	CreatedAt int `json:"created_at"`
	Consumer  struct {
		ID string `json:"id"`
	} `json:"consumer"`
	ID      string      `json:"id"`
	Group   string      `json:"group"`
	Tags    interface{} `json:"tags"`
	Message string      `json:"message"`
	Fields  interface{} `json:"fields"`
	Code    int         `json:"code"`
}

/*
{"message":"UNIQUE violation detected on '{name=\"app-manager\"}'","name":"unique constraint violation","fields":{"name":"app-manager"},"code":5}
*/
type FailMsg struct {
	Message string      `json:"message"`
	Name    string      `json:"name"`
	Fields  interface{} `json:"fields"`
	Code    int         `json:"code"`
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

func NewOperator(host string, port int, cafile string) (*Operator, error) {
	klog.Infof("NewOperator  event:%s %d %s", host, port, cafile)
	return &Operator{
		Host:   host,
		Port:   port,
		CAFile: cafile,
	}, nil
}

//acl里面白名单的name为api的id
func (r *Operator) AddConsumerToAcl(db *nlptv1.Apply, api *apiv1.Api) (aclId string, err error) {
	id := api.ObjectMeta.Name
	klog.Infof("begin add consumer to acl %s", api.ObjectMeta.Name)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s%s%s", schema, r.Host, r.Port, "/consumers/", db.Spec.SourceID, "/acls"))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &AddWhiteRequestBody{
		Group: id, //whilte list group name
	}
	responseBody := &AddWhiteResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return "0", fmt.Errorf("request for add consumer to acl error: %+v", errs)
	}
	klog.V(5).Infof("add consumer to acl whitelist code: %d, body: %s ", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.V(5).Infof("add consumer to acl failed msg: %s\n", responseBody.Message)
		return "0", fmt.Errorf("request for add consumer to acl error: receive wrong status code: %s", string(body))
	}
	klog.V(5).Infof("acl consumer id: %s\n", responseBody.ID)

	if err != nil {
		return "0", fmt.Errorf("create acl error %s", responseBody.Message)
	}
	return responseBody.ID, nil
}
