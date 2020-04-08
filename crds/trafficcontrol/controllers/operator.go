package controllers

import (
	"fmt"
	nlptv1 "github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"
	"github.com/parnurzeal/gorequest"
	"net/http"
	"time"

	"k8s.io/klog"
)

const path string = "/plugins"

var headers = map[string]string{
	"Content-Type": "application/json",
}
var retryStatus = []int{http.StatusBadRequest, http.StatusInternalServerError}

//??
type Operator struct {
	Host           string
	Port           int
	KongPortalPort int
	CAFile         string
}

type Config struct {
	Second  int    `json:"second"`
	Minute  int    `json:"minute"`
	Hour    int    `json:"hour"`
	Day     int    `json:"day"`
	Month   int    `json:"month"`
	Year    int    `json:"year"`
	LimitBy string `json:"limit_by"`
}

type RouteID struct {
	ID string `json:"id"`
}

type RateLimitingRequestBody struct {
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

/*
{"created_at":1581781480,
"config":{"minute":null,"policy":"local","month":null,"redis_timeout":2000,"limit_by":"consumer","hide_client_headers":false,"second":5,"day":null,"redis_password":null,"year":null,"redis_database":0,"hour":10000,"redis_port":6379,"redis_host":null,"fault_tolerant":true},
"id":"78090843-a4a7-4cb3-8f64-56bf88781c90",
"service":null,
"name":"rate-limiting",
"protocols":["grpc","grpcs","http","https"],
"enabled":true,"run_on":"first",
"consumer":null,
"route":{"id":"9caa66ef-f71c-4588-b463-1efbc52ef2cd"},
"tags":null}
*/

type RateRequestBody struct {
	Name     string `json:"name"`
	Consumer struct {
		ID string `json:"id"`
	} `json:"consumer"`
	Config struct {
		Minute int `json:"minute"`
	} `json:"config"`
}

type RateResponseBody struct {
	CreatedAt int `json:"created_at"`
	Config    struct {
		Minute            int         `json:"minute"`
		Policy            string      `json:"policy"`
		Month             int         `json:"month"`
		RedisTimeout      int         `json:"redis_timeout"`
		LimitBy           string      `json:"limit_by"`
		HideClientHeaders bool        `json:"hide_client_headers"`
		Second            int         `json:"second"`
		Day               int         `json:"day"`
		RedisPassword     interface{} `json:"redis_password"`
		Year              int         `json:"year"`
		RedisDatabase     int         `json:"redis_database"`
		Hour              int         `json:"hour"`
		RedisPort         int         `json:"redis_port"`
		RedisHost         interface{} `json:"redis_host"`
		FaultTolerant     bool        `json:"fault_tolerant"`
	} `json:"config"`
	ID        string      `json:"id"`
	Service   interface{} `json:"service"`
	Name      string      `json:"name"`
	Protocols []string    `json:"protocols"`
	Enabled   bool        `json:"enabled"`
	RunOn     string      `json:"run_on"`
	Consumer  struct {
		ID string `json:"id"`
	} `json:"consumer"`
	Route struct {
		ID string `json:"id"`
	} `json:"route"`
	Tags    interface{} `json:"tags"`
	Message string      `json:"message"`
	Fields  interface{} `json:"fields"`
	Code    int         `json:"code"`
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

func (r *Operator) getRequestBody(spec nlptv1.TrafficcontrolSpec) *RateLimitingRequestBody {
	requestBody := &RateLimitingRequestBody{}
	requestBody.Name = "rate-limiting"
	requestBody.Config = make(map[string]interface{})
	switch spec.Type {
	case nlptv1.APIC:
		requestBody.Config["limit_by"] = "service"
	case nlptv1.IPC:
		requestBody.Config["limit_by"] = "ip"
	case nlptv1.APPC:
		requestBody.Config["limit_by"] = "consumer"
	default:
		requestBody.Config["limit_by"] = "consumer"
	}

	if spec.Config.Year != 0 {
		requestBody.Config["year"] = spec.Config.Year
	}
	if spec.Config.Month != 0 {
		requestBody.Config["month"] = spec.Config.Month
	}
	if spec.Config.Day != 0 {
		requestBody.Config["day"] = spec.Config.Day
	}
	if spec.Config.Hour != 0 {
		requestBody.Config["hour"] = spec.Config.Hour
	}
	if spec.Config.Minute != 0 {
		requestBody.Config["minute"] = spec.Config.Minute
	}
	if spec.Config.Second != 0 {
		requestBody.Config["second"] = spec.Config.Second
	}
	return requestBody
}

func (r *Operator) AddRouteRatelimitByKong(db *nlptv1.Trafficcontrol) (err error) {
	for index := 0; index < len(db.Spec.Apis); {
		api := db.Spec.Apis[index]
		if db.ObjectMeta.Labels[api.ID] == "true" && len(api.TrafficID) == 0 {
			id := api.KongID
			klog.Infof("begin create rate-limiting , the KongID of api is %s", id)
			request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
			schema := "http"
			request = request.Post(fmt.Sprintf("%s://%s:%d%s%s%s", schema, r.Host, r.Port, "/routes/", id, path))
			for k, v := range headers {
				request = request.Set(k, v)
			}
			klog.Infof(fmt.Sprintf("====%s://%s:%d%s%s%s", schema, r.Host, r.Port, "/routes/", id, path))
			request = request.Retry(3, 5*time.Second, retryStatus...)
			requestBody := r.getRequestBody(db.Spec)
			responseBody := &RateResponseBody{}
			response, body, errs := request.Send(requestBody).EndStruct(responseBody)
			if len(errs) > 0 {
				db.Spec.Apis[index].Result = nlptv1.BINDFAILED
				return fmt.Errorf("request for add rate-limiting error: %+v", errs)
			}
			klog.Infof("creation rate-limiting code: %d, body: %s ", response.StatusCode, string(body))
			if response.StatusCode != 201 {
				db.Spec.Apis[index].Result = nlptv1.BINDFAILED
				klog.Infof("create rate-limiting failed msg: %s\n", responseBody.Message)
				return fmt.Errorf("request for create rate-limiting error: receive wrong status code: %s", string(body))
			}
			(*db).Spec.Apis[index].Result = nlptv1.SUCCESS
			(*db).Spec.Apis[index].TrafficID = responseBody.ID

		}
		index = index + 1
	}
	return nil
}

func (r *Operator) AddSpecialAppRateLimit(db *nlptv1.Trafficcontrol, consumer []string) (err error) {
	for index := 0; index < len(db.Spec.Apis); index = index + 1 {
		api := db.Spec.Apis[index]
		if db.ObjectMeta.Labels[api.ID] == "true" && len(api.SpecialID) == 0 {
			id := api.KongID
			//(*db).Spec.Apis[index].Result = nlptv1.SUCCESS
			klog.Infof("begin create rate-limiting by conusmer %s %s", id, consumer)
			for i := 0; i < len(consumer); i++ {
				var rateId string
				rateId, err := r.AddSpecialAppRateLimitByKong(db, id, consumer[i], i)
				if err != nil {
					(*db).Spec.Apis[index].Result = nlptv1.BINDFAILED
					klog.Infof("create app %s rate-limiting failed", db.Spec.Config.Special[i].ID)
					return fmt.Errorf("create rate-limiting error %+v", err)
				}
				db.Spec.Apis[index].SpecialID = append(db.Spec.Apis[index].SpecialID, rateId)
			}
			klog.Infof("create rate-limiting result is %s", db.Spec.Apis[index].Result)
			(*db).Spec.Apis[index].Result = nlptv1.SUCCESS
		}
	}
	return nil
}
func (r *Operator) DeleteSpecialAppRateLimit(db *nlptv1.Trafficcontrol) (err error) {
	for index := 0; index < len(db.Spec.Apis); {
		api := db.Spec.Apis[index]
		if db.ObjectMeta.Labels[api.ID] == "false" && len(api.SpecialID) != 0 {
			id := api.KongID
			(*db).Spec.Apis[index].Result = nlptv1.SUCCESS
			klog.Infof("begin delete rate-limiting by conusmer %s", id)
			for i := 0; i < len(api.SpecialID); {
				err := r.DeleteRateLimitByKong(api.SpecialID[i])
				if err != nil {
					(*db).Spec.Apis[index].Result = nlptv1.UNBINDFAILED
					klog.Infof("delete %s rate-limiting failed", db.Spec.Config.Special[i].ID)
					return fmt.Errorf("delete rate-limiting error %+v", err)
				}
				api.SpecialID = append(api.SpecialID[:i], api.SpecialID[i+1:]...)
			}
			klog.Infof("delete rate-limiting result is ok")
			(*db).Spec.Apis = append((*db).Spec.Apis[:index], (*db).Spec.Apis[index+1:]...)
			klog.Infof("delete rate-limiting result is ok %v", (*db).Spec.Apis)
		} else {
			index = index + 1
		}
	}
	return nil
}

func (r *Operator) AddSpecialAppRateLimitByKong(db *nlptv1.Trafficcontrol, routeId string, consumerId string, index int) (id string, err error) {
	klog.Infof("begin create rate with consumer %s:%s", routeId, consumerId)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s%s%s", schema, r.Host, r.Port, "/routes/", routeId, "/plugins"))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &RateRequestBody{}
	requestBody.Name = "rate-limiting"
	requestBody.Config.Minute = db.Spec.Config.Special[index].Minute
	requestBody.Consumer.ID = consumerId
	responseBody := &RateResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Infof("create rate by consumer error: %+v", errs)
		return "", fmt.Errorf("request for create rate by conusmer error: %+v", errs)
	}
	klog.Infof("create rate by conusmer response body: %s", string(body))
	if response.StatusCode != 201 {
		klog.Infof("create failed msg: %s\n", responseBody.Message)
		return "", fmt.Errorf("request for create rate error: receive wrong status code: %s", string(body))
	}
	klog.Infof("app rate limite ID==: %s\n", responseBody.ID)
	return responseBody.ID, nil
}

func (r *Operator) DeleteRateLimitByKong(pluginId string) (err error) {
	klog.Infof("begin delete rate limit %s", pluginId)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	for k, v := range headers {
		request = request.Set(k, v)
	}
	response, body, errs := request.Delete(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, "/plugins", pluginId)).End()
	request = request.Retry(3, 5*time.Second, retryStatus...)

	if len(errs) > 0 {
		return fmt.Errorf("request for delete plugin error: %+v", errs)
	}

	klog.V(5).Infof("delete plugin response code: %d%s", response.StatusCode, string(body))
	if response.StatusCode != 204 {
		return fmt.Errorf("request for delete api error: receive wrong status code: %d", response.StatusCode)
	}
	return nil
}
func (r *Operator) DeleteRouteLimitByKong(db *nlptv1.Trafficcontrol) (err error) {
	for index := 0; index < len(db.Spec.Apis); {
		api := db.Spec.Apis[index]
		if db.ObjectMeta.Labels[api.ID] == "false" && len(api.TrafficID) != 0 {
			trafficID := api.TrafficID //route_id
			klog.Infof("begin delete rate-limiting , the TrafficID of api is %s", trafficID)
			request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
			schema := "http"
			for k, v := range headers {
				request = request.Set(k, v)
			}
			response, body, errs := request.Delete(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, trafficID)).End()
			request = request.Retry(3, 5*time.Second, retryStatus...)

			if len(errs) > 0 {
				db.Spec.Apis[index].Result = nlptv1.UNBINDFAILED
				return fmt.Errorf("request for delete rate-limiting error: %+v", errs)
			}
			klog.V(5).Infof("delete rate-limiting response code: %d%s", response.StatusCode, string(body))
			if response.StatusCode != 204 {
				db.Spec.Apis[index].Result = nlptv1.UNBINDFAILED
				return fmt.Errorf("request for delete rate-limiting error: receive wrong status code: %d", response.StatusCode)
			}
			db.Spec.Apis[index].Result = nlptv1.SUCCESS
			db.Spec.Apis[index].TrafficID = ""
			db.Spec.Apis = append(db.Spec.Apis[:index], db.Spec.Apis[index+1:]...)
		} else {
			index = index + 1
		}
	}
	return nil
}

//UpdateRouteLimitByKong...
func (r *Operator) UpdateRouteLimitByKong(db *nlptv1.Trafficcontrol) (err error) {
	for index, value := range db.Spec.Apis {
		if len(value.TrafficID) != 0 {
			request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
			for k, v := range headers {
				request = request.Set(k, v)
			}
			request = request.Patch(fmt.Sprintf("http://%s:%d%s/%s", r.Host, r.Port, path, value.TrafficID))
			request = request.Retry(3, 5*time.Second, retryStatus...)
			requestBody := r.getRequestBody(db.Spec)
			response, body, errs := request.Send(requestBody).EndStruct(&RateResponseBody{})
			if len(errs) > 0 || response.StatusCode != 200 {
				klog.Infof("request for update route error: %+v, response:%v,body:%v", errs, response, string(body))
				db.Spec.Apis[index].Result = nlptv1.UPDATEFAILED
				return fmt.Errorf("request for update route error: %+v, code:%v", errs, response.StatusCode)
			} else {
				db.Spec.Apis[index].Result = nlptv1.SUCCESS
			}
		}
	}
	return nil
}

//UpdateSpecialAppRateLimitByKong...
func (r *Operator) UpdateSpecialAppRateLimitConfig(db *nlptv1.Trafficcontrol, consumer []string) (err error) {
	for index, value := range db.Spec.Apis {
		for j := range consumer {
			if j < len(value.SpecialID) {
				//patch
				klog.Infof("begin update rate-limiting by consumer %s %s", value.KongID, consumer)
				request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
				request = request.Patch(fmt.Sprintf("http://%s:%d/routes/%s/plugins/%s", r.Host, r.Port, value.KongID, value.SpecialID[j]))
				for k, v := range headers {
					request = request.Set(k, v)
				}
				request = request.Retry(3, 5*time.Second, retryStatus...)
				requestBody := &RateRequestBody{}
				requestBody.Name = "rate-limiting"
				requestBody.Config.Minute = db.Spec.Config.Special[j].Minute
				requestBody.Consumer.ID = consumer[j]
				response, body, errs := request.Send(requestBody).EndStruct(&RateResponseBody{})
				if len(errs) > 0 || response.StatusCode != 200 {
					klog.Infof("update rate by consumer error: %+v", errs)
					db.Spec.Apis[index].Result = nlptv1.UPDATEFAILED
					return fmt.Errorf("request for update rate by conusmer error: %+v, code:%v, body:%v", errs, response.StatusCode, string(body))
				}
			}
			if j >= len(db.Spec.Apis[index].SpecialID) {
				//添加多余的app
				klog.Infof("update need create rate-limiting %s", consumer[j])
				var rateId string
				rateId, err := r.AddSpecialAppRateLimitByKong(db, value.KongID, consumer[j], j)
				if err != nil {
					(*db).Spec.Apis[index].Result = nlptv1.UPDATEFAILED
					klog.Infof("create by update app %s rate-limiting failed", db.Spec.Config.Special[j].ID)
					return fmt.Errorf("create by update rate-limiting error %+v", err)
				}
				db.Spec.Apis[index].SpecialID = append(db.Spec.Apis[index].SpecialID, rateId)
			}
		}
		if len(consumer) < len(db.Spec.Apis[index].SpecialID) {
			//delete 删除多余的app
			klog.Infof("update need delete rate-limiting %d and %d", len(consumer), len(value.SpecialID))
			for i := len(consumer); i < len(db.Spec.Apis[index].SpecialID); {
				err := r.DeleteRateLimitByKong(db.Spec.Apis[index].SpecialID[i])
				if err != nil {
					(*db).Spec.Apis[index].Result = nlptv1.UPDATEFAILED
					klog.Infof("delete by update %s rate-limiting failed", db.Spec.Config.Special[i].ID)
					return fmt.Errorf("delete by update rate-limiting error %+v", err)
				}
				db.Spec.Apis[index].SpecialID = append(db.Spec.Apis[index].SpecialID[:i], db.Spec.Apis[index].SpecialID[i+1:]...)
			}
			klog.Infof("delete by update rate-limiting result is ok")
		}
		db.Spec.Apis[index].Result = nlptv1.SUCCESS
	}
	return nil
}

//UpdateSpecialAppRateLimitByKong...
func (r *Operator) UpdateSpecialAppRateLimit(db *nlptv1.Trafficcontrol, consumer []string) (err error) {
	for index, value := range db.Spec.Apis {
		for index, v := range value.SpecialID {
			for j := range consumer {
				klog.Infof("begin update rate-limiting by conusmer %s %s", value.KongID, consumer)
				request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
				request = request.Patch(fmt.Sprintf("http://%s:%d/routes/%s/plugins/%s", r.Host, r.Port, value.KongID, v))
				for k, v := range headers {
					request = request.Set(k, v)
				}
				request = request.Retry(3, 5*time.Second, retryStatus...)
				requestBody := &RateRequestBody{}
				requestBody.Name = "rate-limiting"
				requestBody.Config.Minute = db.Spec.Config.Special[index].Minute
				requestBody.Consumer.ID = consumer[j]
				response, body, errs := request.Send(requestBody).EndStruct(&RateResponseBody{})
				if len(errs) > 0 || response.StatusCode != 200 {
					klog.Infof("update rate by consumer error: %+v", errs)
					db.Spec.Apis[index].Result = nlptv1.UPDATEFAILED
					return fmt.Errorf("request for update rate by conusmer error: %+v, code:%v, body:%v", errs, response.StatusCode, string(body))
				}
			}
		}
		db.Spec.Apis[index].Result = nlptv1.SUCCESS
	}
	return nil
}
