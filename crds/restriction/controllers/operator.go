package controllers

import (
	"fmt"
	nlptv1 "github.com/chinamobile/nlpt/crds/restriction/api/v1"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
	"net/http"
	"time"
)

const path string = "/plugins"

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

type Config struct {
	WhiteList []string `json:"whitelist"`
	BlackList []string `json:"blacklist"`
}

type RouteID struct {
	ID string `json:"id"`
}

type RestrictionRequestBody struct {
	Name   string `json:"name"`
	Config Config `json:"config"`
}

/*
{
"created_at":1582774255,
"config":{
  "whitelist":["54.13.21.1","143.1.0.0\/24"],
  "blacklist":null
  },
"id":"3bdc6d03-681e-4b2b-930b-df13a5ba4628",
"service":null,
"name":"ip-restriction",
"protocols":["grpc","grpcs","http","https"],
"enabled":true,
"run_on":"first",
"consumer":null,
"route":{
  "id":"c3efc4a8-e7c0-4ecc-9f54-d40aa33291bb"
  },
"tags":null
}
*/
type RestrictionResponseBody struct {
	CreatedAt int         `json:"created_at"`
	Config    Config      `json:"config"`
	ID        string      `json:"id"`
	Service   interface{} `json:"service"`
	Name      string      `json:"name"`
	Protocols []string    `json:"protocols"`
	Enabled   bool        `json:"enabled"`
	RunOn     string      `json:"run_on"`
	Consumer  interface{} `json:"consumer"`
	RouteId   RouteID     `json:"routeid"`
	Tags      interface{} `json:"tags"`
	Message   string      `json:"message"`
	Fields    interface{} `json:"fields"`
	Code      int         `json:"code"`
}

/*
{"message":"UNIQUE violation detected on '{service=null,name=\"rate-limiting\",route={id=\"9caa66ef-f71c-4588-b463-1efbc52ef2cd\"},consumer=null}'",
"name":"unique constraint violation",
"fields":{"service":null,
"name":"rate-limiting","route":{"id":"9caa66ef-f71c-4588-b463-1efbc52ef2cd"},"consumer":null},
"code":5}
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

func (r *Operator) AddRestrictionByKong(db *nlptv1.Restriction) (err error) {
	for index := 0; index < len(db.Spec.Apis); {
		apiSource := db.Spec.Apis[index]
		if len(apiSource.PluginID) == 0 {
			routeId := apiSource.KongID
			klog.Infof("begin create restriction , the routeID is %s", routeId)
			request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
			schema := "http"
			request = request.Post(fmt.Sprintf("%s://%s:%d%s%s/%s", schema, "192.168.1.207", 30081, "/routes/", routeId, path))
			for k, v := range headers {
				request = request.Set(k, v)
			}
			request = request.Retry(3, 5*time.Second, retryStatus...)
			requestBody := &RestrictionRequestBody{}
			requestBody.Name = "ip-restriction"
			requestBody.Config.WhiteList = append(requestBody.Config.WhiteList, "54.13.21.1")
			requestBody.Config.WhiteList = append(requestBody.Config.WhiteList, "143.1.0.0/24")
			requestBody.Config.WhiteList = append(requestBody.Config.WhiteList, "54.13.21.2")
			requestBody.Config.WhiteList = append(requestBody.Config.WhiteList, "144.1.0.0/24")

			responseBody := &RestrictionResponseBody{}
			response, body, errs := request.Send(requestBody).EndStruct(responseBody)
			if len(errs) > 0 {
				db.Spec.Apis[index].Result = nlptv1.FAILED
				return fmt.Errorf("request for add restriction error: %+v", errs)
			}
			klog.V(5).Infof("creation restriction code: %d, body: %s ", response.StatusCode, string(body))
			if response.StatusCode != 201 {
				db.Spec.Apis[index].Result = nlptv1.FAILED
				klog.V(5).Infof("create restriction failed msg: %s\n", responseBody.Message)
				return fmt.Errorf("request for create restriction error: receive wrong status code: %s", string(body))
			}
			db.Spec.Apis[index].Result = nlptv1.SUCCESS
			(*db).Spec.Apis[index].PluginID = responseBody.ID
		}
		index = index + 1
	}
	return nil
}

//DeleteRestrictionByKong
func (r *Operator) DeleteRestrictionByKong(db *nlptv1.Restriction) (err error) {
	for index := 0; index < len(db.Spec.Apis); {
		apiSource := db.Spec.Apis[index]
		if len(apiSource.PluginID) != 0 {
			restrictionID := apiSource.PluginID //_id
			klog.Infof("begin delete rate-limiting , the RestrictionID is %s", restrictionID)
			request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
			schema := "http"
			for k, v := range headers {
				request = request.Set(k, v)
			}
			response, body, errs := request.Delete(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, restrictionID)).End()
			request = request.Retry(3, 5*time.Second, retryStatus...)

			if len(errs) > 0 {
				db.Spec.Apis[index].Result = nlptv1.FAILED
				return fmt.Errorf("request for delete ip-restriction error: %+v", errs)
			}
			klog.V(5).Infof("delete ip-restriction response code: %d%s", response.StatusCode, string(body))
			if response.StatusCode != 204 {
				db.Spec.Apis[index].Result = nlptv1.FAILED
				return fmt.Errorf("request for delete ip-restriction error: receive wrong status code: %d", response.StatusCode)
			}
			db.Spec.Apis[index].Result = nlptv1.SUCCESS
			//  apis[index]
			db.Spec.Apis = append(db.Spec.Apis[:index], db.Spec.Apis[index+1:]...)
		} else {
			db.Spec.Apis = append(db.Spec.Apis[:index], db.Spec.Apis[index+1:]...)
		}
	}
	return nil
}
