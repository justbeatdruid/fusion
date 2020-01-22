package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/parnurzeal/gorequest"

	nlptv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"

	"k8s.io/klog"
)

const path string = "/services"

var headers = map[string]string{
	"Content-Type": "application/json",
}
var retryStatus = []int{http.StatusBadRequest, http.StatusInternalServerError}

type Operator struct {
	Host   string
	Port   int
	CAFile string
}

type RequestBody struct {
	Name     string `json:"name"`
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	TimeOut  int    `json:"connect_timeout"`
	WirteOut int    `json:"write_timeout"`
	ReadOut  int    `json:"read_timeout"`
}

// {"success":true,"code":200,"message":"请求成功","data":{"targetDbType":"MySQL","targetDbIp":"192.168.100.103","targetDbPort":"3306","targetDbUser":"root","targetDbPass":"Pass1234","targetDbName":["POSTGRESQL_public","POSTGRESQL_testschema"]},"pageInfo":null,"ext":null}
type ResponseBody struct {
	Host              string      `json:"host"`
	CreatedAt         int         `json:"created_at"`
	ConnectTimeout    int         `json:"connect_timeout"`
	ID                string      `json:"id"`
	Protocol          string      `json:"protocol"`
	Name              string      `json:"name"`
	ReadTimeout       int         `json:"read_timeout"`
	Port              int         `json:"port"`
	Path              string      `json:"path"`
	UpdatedAt         int         `json:"updated_at"`
	Retries           int         `json:"retries"`
	WriteTimeout      int         `json:"write_timeout"`
	Tags              []string    `json:"tags"`
	ClientCertificate interface{} `json:"client_certificate"`
	Message           string      `json:"message"`
	Fields            interface{} `json:"fields"`
	Code              int         `json:"code"`
}

/*
{"host":"apps",
"created_at":1578378841,
"connect_timeout":60000,
"id":"f15d6a13-a65c-44b0-a8b9-274896562654",
"protocol":"http","name":"apps","read_timeout":60000,"port":80,"path":null,"updated_at":1578378841,"retries":5,"write_timeout":60000,"tags":null,"client_certificate":null}r
*/
type ServiceData struct {
	Host              string      `json:"host"`
	CreatedAt         int         `json:"created_at"`
	ConnectTimeout    int         `json:"connect_timeout"`
	ID                string      `json:"id"`
	Protocol          string      `json:"protocol"`
	Name              string      `json:"name"`
	ReadTimeout       int         `json:"read_timeout"`
	Port              int         `json:"port"`
	Path              string      `json:"path"`
	UpdatedAt         int         `json:"updated_at"`
	Retries           int         `json:"retries"`
	WriteTimeout      int         `json:"write_timeout"`
	Tags              []string    `json:"tags"`
	ClientCertificate interface{} `json:"client_certificate"`
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

type DataOrFailMsg struct {
	Message FailMsg
	Data    ServiceData
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

func (r *Operator) CreateServiceByKong(db *nlptv1.Serviceunit) (err error) {
	klog.Infof("Enter CreateServiceByKong name:%s, Host:%s, Port:%d", db.ObjectMeta.Name, r.Host, r.Port)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, r.Host, r.Port, path))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	//TODO 服务的地址信息 : 数据服务使用fusion-apiserver web后端使用传入的值
	requestBody := &RequestBody{
		Name:     db.ObjectMeta.Name,
		Protocol: "http",
		Host:     "fusion-apiserver",
		Port:     8001,
		TimeOut:  60000,
		WirteOut: 60000,
		ReadOut:  60000,
	}
	if db.Spec.Type == nlptv1.WebService {

		requestBody = &RequestBody{
			Name:     db.ObjectMeta.Name,
			Protocol: db.Spec.KongService.Protocol,
			Host:     db.Spec.KongService.Host,
			Port:     db.Spec.KongService.Port,
			TimeOut:  db.Spec.KongService.TimeOut,
			WirteOut: db.Spec.KongService.WirteOut,
			ReadOut:  db.Spec.KongService.ReadOut,
		}
		if requestBody.Port == 0 {
			requestBody.Port = 80
		}
		if requestBody.TimeOut == 0 {
			requestBody.TimeOut = 60000
		}
		if requestBody.WirteOut == 0 {
			requestBody.WirteOut = 60000
		}
		if requestBody.ReadOut == 0 {
			requestBody.ReadOut = 60000
		}
	}
	responseBody := &ResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for create service error: %+v", errs)
	}
	klog.V(5).Infof("create service response body: %s", string(body))
	if response.StatusCode != 201 {
		return fmt.Errorf("request for create service error: receive wrong status code: %s", string(body))
	}
	if db.Spec.Type == nlptv1.DataService {
		(*db).Spec.KongService.Host = responseBody.Host
		(*db).Spec.KongService.Protocol = responseBody.Protocol
		(*db).Spec.KongService.Port = responseBody.Port
	}
	(*db).Spec.KongService.ID = responseBody.ID
	return nil
}

func (r *Operator) DeleteServiceByKong(db *nlptv1.Serviceunit) (err error) {
	klog.Infof("delete service %s %s", db.ObjectMeta.Name, db.Spec.KongService.ID)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	for k, v := range headers {
		request = request.Set(k, v)
	}
	id := db.Spec.KongService.ID
	klog.Infof("delete service id %s %s", id, fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, id))
	response, body, errs := request.Delete(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, id)).End()
	klog.Infof("delete service response code: %d %s", response.StatusCode, string(body))
	request = request.Retry(3, 5*time.Second, retryStatus...)

	if len(errs) > 0 {
		return fmt.Errorf("request for delete service error: %+v", errs)
	}

	klog.V(5).Infof("delete service response code: %d%s", response.StatusCode, string(body))
	if response.StatusCode != 204 {
		return fmt.Errorf("request for delete service error: receive wrong status code: %d", response.StatusCode)
	}
	return nil
}

// + update_sunyu
func (r *Operator) UpdateServiceByKong(db *nlptv1.Serviceunit) (err error) {
	klog.Infof("Enter UpdateServiceByKong name:%s, Host:%s, Port:%d", db.ObjectMeta.Name, r.Host, r.Port)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	id := db.Spec.KongService.ID
	klog.Infof("delete service id %s %s", id, fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, id))
	request = request.Patch(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, id))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	//TODO 服务的地址信息 : 数据服务使用fusion-apiserver web后端使用传入的值
	requestBody := &RequestBody{
		Name:     db.ObjectMeta.Name,
		Protocol: "http",
		Host:     "fusion-apiserver",
		Port:     8001,
		TimeOut:  60000,
		WirteOut: 60000,
		ReadOut:  60000,
	}
	if db.Spec.Type == nlptv1.WebService {
		requestBody = &RequestBody{
			Name:     db.ObjectMeta.Name,
			Protocol: db.Spec.KongService.Protocol,
			Host:     db.Spec.KongService.Host,
			Port:     db.Spec.KongService.Port,
			TimeOut:  db.Spec.KongService.TimeOut,
			WirteOut: db.Spec.KongService.WirteOut,
			ReadOut:  db.Spec.KongService.ReadOut,
		}
		if requestBody.TimeOut == 0 {
			requestBody.TimeOut = 60000
		}
		if requestBody.WirteOut == 0 {
			requestBody.WirteOut = 60000
		}
		if requestBody.ReadOut == 0 {
			requestBody.ReadOut = 60000
		}
		if requestBody.Port == 0 {
			requestBody.Port = 80
		}
	}
	responseBody := &ResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for update service error: %+v", errs)
	}
	klog.V(5).Infof("update service response body: %s", string(body))
	if response.StatusCode != 201 {
		return fmt.Errorf("request for update service error: receive wrong status code: %s", string(body))
	}
	if db.Spec.Type == nlptv1.DataService {
		(*db).Spec.KongService.Host = responseBody.Host
		(*db).Spec.KongService.Protocol = responseBody.Protocol
		(*db).Spec.KongService.Port = responseBody.Port
	}
	return nil
}