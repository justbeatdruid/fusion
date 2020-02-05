package controllers

import (
	"fmt"
	"net/http"
	"time"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
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
func (r *Operator) AddConsumerToAcl(db *nlptv1.Apply, api *apiv1.Api) (err error) {
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
		return fmt.Errorf("request for add consumer to acl error: %+v", errs)
	}
	klog.V(5).Infof("add consumer to acl whitelist code: %d, body: %s ", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.V(5).Infof("add consumer to acl failed msg: %s\n", responseBody.Message)
		return fmt.Errorf("request for add consumer to acl error: receive wrong status code: %s", string(body))
	}
	(*db).Spec.AclID = responseBody.ID
	klog.V(5).Infof("acl consumer id: %s\n", responseBody.ID)

	if err != nil {
		return fmt.Errorf("create acl error %s", responseBody.Message)
	}
	return nil
}

//解绑API
func (r *Operator) DeleteConsumerFromAcl(db *nlptv1.Apply, app *appv1.Application) (err error) {
	klog.Infof("delete consumer from acl %s.", db.Spec.AclID)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	for k, v := range headers {
		request = request.Set(k, v)
	}
	klog.Infof("delete consumer is %s", fmt.Sprintf("%s://%s:%d%s%s%s%s", schema, r.Host, r.Port,
		"/consumers/", app.Spec.ConsumerInfo.ConsumerID, "/acls/", db.Spec.AclID))
	response, body, errs := request.Delete(fmt.Sprintf("%s://%s:%d%s%s%s%s", schema, r.Host, r.Port,
		"/consumers/", app.Spec.ConsumerInfo.ConsumerID, "/acls/", db.Spec.AclID)).End()
	klog.Infof("delete consumer from acl response code: %d %s", response.StatusCode, string(body))
	request = request.Retry(3, 5*time.Second, retryStatus...)

	if len(errs) > 0 {
		return fmt.Errorf("request for delete consumer from acl error: %+v", errs)
	}

	klog.V(5).Infof("delete consumer response code: %d%s", response.StatusCode, string(body))
	if response.StatusCode != 204 {
		return fmt.Errorf("request for delete consumer error: receive wrong status code: %d", response.StatusCode)
	}
	return nil
}
