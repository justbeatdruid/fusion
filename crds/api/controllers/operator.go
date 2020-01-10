package controllers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"

	nlptv1 "github.com/chinamobile/nlpt/crds/api/api/v1"

	"k8s.io/klog"
)

const path string = "/routes"

var headers = map[string]string{
	"Content-Type": "application/json",
}
var retryStatus = []int{http.StatusBadRequest, http.StatusInternalServerError}

type Operator struct {
	Host           string
	Port           int
	KongPortalPort int
	CAFile         string
}

type RequestBody struct {
	Service   ServiceID `json:"service"`
	Protocols []string  `json:"protocols"`
	Paths     []string  `json:"paths"`
}

// {"success":true,"code":200,"message":"请求成功","data":{"targetDbType":"MySQL","targetDbIp":"192.168.100.103","targetDbPort":"3306","targetDbUser":"root","targetDbPass":"Pass1234","targetDbName":["POSTGRESQL_public","POSTGRESQL_testschema"]},"pageInfo":null,"ext":null}
//{"id":"c128437c-6f9c-4978-8dc6-a1323b51a39e","tags":null,"updated_at":1578534015,"destinations":null,"headers":null,"protocols":["http","https"],"created_at":1578534015,"snis":null,"service":{"id":"8915796c-c401-4889-a2ed-8a69708f987f"},"name":null,"preserve_host":false,"regex_priority":0,"strip_path":true,"sources":null,"paths":["\/application"],"https_redirect_status_code":426,"hosts":null,"methods":null}
type ResponseBody struct {
	ID                      string      `json:"id"`
	Tags                    interface{} `json:"tags"`
	UpdatedAt               int         `json:"updated_at"`
	Destinations            interface{} `json:"destinations"`
	Headers                 []string    `json:"headers"`
	Protocols               []string    `json:"protocols"`
	CreatedAt               int         `json:"created_at"`
	Snis                    interface{} `json:"snis"`
	Service                 ServiceID   `json:"service"`
	Name                    interface{} `json:"name"`
	PreserveHost            bool        `json:"preserve_host"`
	RegexPriority           int         `json:"regex_priority"`
	StripPath               bool        `json:"strip_path"`
	Sources                 interface{} `json:"sources"`
	Paths                   []string    `json:"paths"`
	HTTPSRedirectStatusCode int         `json:"https_redirect_status_code"`
	Hosts                   []string    `json:"hosts"`
	Methods                 interface{} `json:"methods"`
	Message                 string      `json:"message"`
	Fields                  interface{} `json:"fields"`
	Code                    int         `json:"code"`
}

type ServiceID struct {
	ID string `json:"id"`
}

/*
{"host":"apps",
"created_at":1578378841,
"connect_timeout":60000,
"id":"f15d6a13-a65c-44b0-a8b9-274896562654",
"protocol":"http","name":"apps","read_timeout":60000,"port":80,"path":null,"updated_at":1578378841,"retries":5,"write_timeout":60000,"tags":null,"client_certificate":null}r
*/

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

func NewOperator(host string, port int, portal int, cafile string) (*Operator, error) {
	klog.Infof("NewOperator  event:%s %d %s", host, port, cafile)
	return &Operator{
		Host:           host,
		Port:           port,
		KongPortalPort: portal,
		CAFile:         cafile,
	}, nil
}

func (r *Operator) CreateRouteByKong(db *nlptv1.Api) (err error) {
	klog.Infof("Enter CreateRouteByKong name:%s, Host:%s, Port:%d", db.Name, r.Host, r.Port)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, r.Host, r.Port, path))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	//TODO API创建时路由信息 /apiquery
	protocols := []string{}
	protocols = append(protocols, strings.ToLower(string(db.Spec.Protocol)))
	paths := []string{}
	paths = append(paths, "/apiquery")
	//TODO 替换ID
	id := db.Spec.Serviceunit.KongID
	//id := "ce5e6f95-611b-4feb-96f1-f3b025876424"
	requestBody := &RequestBody{
		Service:   ServiceID{id},
		Protocols: protocols,
		Paths:     paths,
	}
	responseBody := &ResponseBody{}
	klog.Infof("begin send create route requeset body: %+v", responseBody)
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	klog.Infof("end send create route response body:%d %s", response.StatusCode, string(body))
	if len(errs) > 0 {
		return fmt.Errorf("request for create route error: %+v", errs)
	}
	if response.StatusCode != 201 {
		return fmt.Errorf("request for create route error: receive wrong status code: %s", string(body))
	}

	(*db).Spec.KongApi.Hosts = responseBody.Hosts
	(*db).Spec.KongApi.Protocols = responseBody.Protocols
	(*db).Spec.KongApi.Paths = responseBody.Paths
	(*db).Spec.KongApi.KongID = responseBody.ID
	return nil
}

func (r *Operator) DeleteRouteByKong(db *nlptv1.Api) (err error) {
        klog.Infof("delete api %s %s", db.ObjectMeta.Name, db.Spec.KongApi.KongID)
        request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
        schema := "http"
        for k, v := range headers {
                request = request.Set(k, v)
        }
        id := db.Spec.KongApi.KongID
        klog.Infof("delete api id %s %s", id, fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, id))
        response, body, errs := request.Delete(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, id)).End()
        klog.Infof("delete api response code: %d %s", response.StatusCode, string(body))
        request = request.Retry(3, 5*time.Second, retryStatus...)

        if len(errs) > 0 {
                return fmt.Errorf("request for delete api error: %+v", errs)
        }

        klog.V(5).Infof("delete api response code: %d%s", response.StatusCode, string(body))
        if response.StatusCode != 204 {
                return fmt.Errorf("request for delete api error: receive wrong status code: %d", response.StatusCode)
        }
        return nil
}