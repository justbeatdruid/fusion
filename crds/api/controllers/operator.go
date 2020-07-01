package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/parnurzeal/gorequest"

	nlptv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"k8s.io/klog"
)

const path string = "/routes"
const HttpTriggerUrl string =  "/v2/triggers/http"

var headers = map[string]string{
	"Content-Type": "application/json",
}
var retryStatus = []int{http.StatusBadRequest, http.StatusInternalServerError}

type Operator struct {
	Host           string
	Port           int
	KongPortalPort int
	CAFile         string
	PrometheusHost string
	PrometheusPort int
	FissionAddress FissionAddress
}

type RequestBody struct {
	Service   ServiceID `json:"service"`
	Protocols []string  `json:"protocols"`
	Paths     []string  `json:"paths"`
	StripPath bool      `json:"strip_path"`
	Hosts     []string  `json:"hosts"`
	Methods   []string  `json:"methods"`
	Name      string    `json:"name"`
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
	Methods                 []string    `json:"methods"`
	Message                 string      `json:"message"`
	Fields                  interface{} `json:"fields"`
	Code                    int         `json:"code"`
}

type ServiceID struct {
	ID string `json:"id"`
}

type JwtRequestBody struct {
	Name   string    `json:"name"`
	Config ConfigJwt `json:"config"`
}

type Config struct {
	SecretIsBase64    bool          `json:"secret_is_base64"`
	RunOnPreflight    bool          `json:"run_on_preflight"`
	URIParamNames     []string      `json:"uri_param_names"`
	KeyClaimName      string        `json:"key_claim_name"`
	HeaderNames       []string      `json:"header_names"`
	MaximumExpiration int           `json:"maximum_expiration"`
	Anonymous         interface{}   `json:"anonymous"`
	ClaimsToVerify    []string      `json:"claims_to_verify"`
	CookieNames       []interface{} `json:"cookie_names"`
}
type ConfigJwt struct {
	ClaimsToVerify []string `json:"claims_to_verify"`
}

type JwtResponseBody struct {
	CreatedAt int `json:"created_at"`
	Config    struct {
		SecretIsBase64    bool          `json:"secret_is_base64"`
		RunOnPreflight    bool          `json:"run_on_preflight"`
		URIParamNames     []string      `json:"uri_param_names"`
		KeyClaimName      string        `json:"key_claim_name"`
		HeaderNames       []string      `json:"header_names"`
		MaximumExpiration int           `json:"maximum_expiration"`
		Anonymous         interface{}   `json:"anonymous"`
		ClaimsToVerify    []string      `json:"claims_to_verify"`
		CookieNames       []interface{} `json:"cookie_names"`
	} `json:"config"`
	ID        string      `json:"id"`
	Service   interface{} `json:"service"`
	Name      string      `json:"name"`
	Protocols []string    `json:"protocols"`
	Enabled   bool        `json:"enabled"`
	RunOn     string      `json:"run_on"`
	Consumer  interface{} `json:"consumer"`
	Route     struct {
		ID string `json:"id"`
	} `json:"route"`
	Tags    interface{} `json:"tags"`
	Message string      `json:"message"`
	Fields  interface{} `json:"fields"`
	Code    int         `json:"code"`
}
type AclRequestBody struct {
	Name   string       `json:"name"`
	Config AclReqConfig `json:"config"`
}

type AclRspConfig struct {
	HideGroupsHeader bool        `json:"hide_groups_header"`
	Blacklist        interface{} `json:"blacklist"`
	Whitelist        []string    `json:"whitelist"`
}

type AclReqConfig struct {
	HideGroupsHeader bool     `json:"hide_groups_header"`
	Whitelist        []string `json:"whitelist"`
}

type AclResponseBody struct {
	CreatedAt int          `json:"created_at"`
	Config    AclRspConfig `json:"config"`
	ID        string       `json:"id"`
	Service   interface{}  `json:"service"`
	Name      string       `json:"name"`
	Protocols []string     `json:"protocols"`
	Enabled   bool         `json:"enabled"`
	RunOn     string       `json:"run_on"`
	Consumer  interface{}  `json:"consumer"`
	Route     struct {
		ID string `json:"id"`
	} `json:"route"`
	Tags    interface{} `json:"tags"`
	Message string      `json:"message"`
	Fields  interface{} `json:"fields"`
	Code    int         `json:"code"`
}

type CorsRequestBody struct {
	Name string `json:"name"`
}

type CorsResponseBody struct {
	CreatedAt int `json:"created_at"`
	Config    struct {
		Methods           []interface{} `json:"methods"`
		ExposedHeaders    interface{}   `json:"exposed_headers"`
		MaxAge            interface{}   `json:"max_age"`
		Headers           interface{}   `json:"headers"`
		Origins           interface{}   `json:"origins"`
		Credentials       bool          `json:"credentials"`
		PreflightContinue bool          `json:"preflight_continue"`
	} `json:"config"`
	ID        string      `json:"id"`
	Service   interface{} `json:"service"`
	Enabled   bool        `json:"enabled"`
	Protocols []string    `json:"protocols"`
	Name      string      `json:"name"`
	Consumer  interface{} `json:"consumer"`
	Route     struct {
		ID string `json:"id"`
	} `json:"route"`
	Tags    interface{} `json:"tags"`
	Message string      `json:"message"`
	Fields  interface{} `json:"fields"`
	Code    int         `json:"code"`
}

type PrometheusRequestBody struct {
	Name string `json:"name"`
}

type PrometheusResponseBody struct {
	CreatedAt int `json:"created_at"`
	Config    struct {
	} `json:"config"`
	ID        string      `json:"id"`
	Service   interface{} `json:"service"`
	Name      string      `json:"name"`
	Protocols []string    `json:"protocols"`
	Enabled   bool        `json:"enabled"`
	RunOn     string      `json:"run_on"`
	Consumer  interface{} `json:"consumer"`
	Route     struct {
		ID string `json:"id"`
	} `json:"route"`
	Tags    interface{} `json:"tags"`
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Fields  interface{} `json:"fields"`
}

type Metric struct {
	Name            string `json:"__name__"`
	Endpoint        string `json:"endpoint"`
	ExportedService string `json:"exported_service"`
	Instance        string `json:"instance"`
	Job             string `json:"job"`
	Namespace       string `json:"namespace"`
	Pod             string `json:"pod"`
	Route           string `json:"route"`
	Service         string `json:"service"`
	Type            string `json:"type"`
}

type PrometheusResult struct {
	Metric Metric        `json:"metric"`
	Value  []interface{} `json:"value"`
}

type PrometheusData struct {
	ResultType string             `json:"resultType"`
	Result     []PrometheusResult `json:"result"`
}

type PrometheusResponse struct {
	Status string         `json:"status"`
	Data   PrometheusData `json:"data"`
}

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

type FissionAddress struct {
	ControllerHost string    `json:"controllerHost"`
	ControllerPort int      `json:"controllerPort"`
}

type RouteReqInfo struct {
	Metadata struct {
		Name string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Host string `json:"host"`
		Relativeurl string `json:"relativeurl"`
		Method string `json:"method"`
		Functionref struct {
			Type string `json:"type"`
			Name string `json:"name"`
			Functionweights interface{} `json:"functionweights"`
		} `json:"functionref"`
		Createingress bool `json:"createingress"`
		Ingressconfig struct {
			Annotations interface{} `json:"annotations"`
			Path string `json:"path"`
			Host string `json:"host"`
			TLS string `json:"tls"`
		} `json:"ingressconfig"`
	} `json:"spec"`
}

type FissionResInfoRsp struct {
	Name string `json:"name"`
	Namespace string `json:"namespace"`
	SelfLink string `json:"selfLink"`
	UID string `json:"uid"`
	ResourceVersion string `json:"resourceVersion"`
	Generation int `json:"generation"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
}

func NewOperator(host string, port int, portal int, cafile string, prometheusHost string, prometheusPort int, address FissionAddress) (*Operator, error) {
	klog.Infof("NewOperator  event:%s %d %s", host, port, cafile)
	return &Operator{
		Host:           host,
		Port:           port,
		KongPortalPort: portal,
		CAFile:         cafile,
		PrometheusHost: prometheusHost,
		PrometheusPort: prometheusPort,
		FissionAddress: address,
	}, nil
}

func (r *Operator) CreateRouteByKong(db *nlptv1.Api, su *suv1.Serviceunit) (err error) {
	klog.Infof("Enter CreateRouteByKong name:%s, Host:%s, Port:%d", db.Name, r.Host, r.Port)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, r.Host, r.Port, path))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	//API创建时路由信息 数据后端使用 /api/v1/apis/{id}/data web后端使用传入参数
	protocols := []string{}
	methods := []string{}
	paths := []string{}
	//替换ID
	id := db.Spec.Serviceunit.KongID
	requestBody := &RequestBody{}
	requestBody.Service = ServiceID{id}
	//请求协议及方法使用公共请求参数
	requestBody.Name = db.ObjectMeta.Name
	//设置为true会删除前缀 route When matching a Route via one of the paths, strip the matching prefix from the upstream request URL. Defaults to true.
	requestBody.StripPath = false

	switch db.Spec.Serviceunit.Type {
	case string(suv1.DataService):
		//data  后端服务协议使用服务单元的协议信息 method使用API请求中定义的
		methods = append(methods, strings.ToUpper(string(db.Spec.ApiDefineInfo.Method)))
		protocols = append(protocols, strings.ToLower(string(db.Spec.Serviceunit.Protocol)))
		requestBody.Protocols = protocols
		requestBody.Methods = methods
		queryPath := fmt.Sprintf("%s%s%s%s%s", "/api/v1/apis/", db.ObjectMeta.Namespace, "/", db.ObjectMeta.Name, "/data")
		paths = append(paths, queryPath)
		requestBody.Paths = paths
	case string(suv1.WebService):
		////web 类型使用 KongApi.Method KongApi.Protocol 后端服务协议使用服务单元的协议信息 method使用API请求中定义的
		requestBody.Paths = db.Spec.KongApi.Paths
		methods = append(methods, strings.ToUpper(string(db.Spec.ApiDefineInfo.Method)))
		protocols = append(protocols, strings.ToLower(string(db.Spec.Serviceunit.Protocol)))
		requestBody.Protocols = protocols
		requestBody.Methods = methods
		if len(db.Spec.KongApi.Hosts) != 0 {
			requestBody.Hosts = db.Spec.KongApi.Hosts
		}
	case string(suv1.FunctionService):
		//function 需要先再fission创建httptrigger 后端服务协议使用服务单元的协议信息 method使用API请求中定义的
		if err := r.CreateRouteByFission(db, su); err != nil {
			return fmt.Errorf("request for create route by fission error: %+v", err)
		}
		methods = append(methods, strings.ToUpper(string(db.Spec.ApiDefineInfo.Method)))
		protocols = append(protocols, strings.ToLower(string(db.Spec.Serviceunit.Protocol)))
		requestBody.Protocols = protocols
		requestBody.Methods = methods
		queryPath := fmt.Sprintf("%s%s%s%s%s", "/api/v1/apis/", db.ObjectMeta.Namespace, "/", db.ObjectMeta.Name, "/function")
		paths = append(paths, queryPath)
		requestBody.Paths = paths

	}
	responseBody := &ResponseBody{}
	klog.Infof("begin send create route requeset body: %+v", responseBody)
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for create route error: %+v", errs)
	}
	if response.StatusCode != 201 {
		return fmt.Errorf("request for create route error: receive wrong status code: %s", string(body))
	}

	(*db).Spec.KongApi.Hosts = responseBody.Hosts
	(*db).Spec.KongApi.Protocols = responseBody.Protocols
	(*db).Spec.KongApi.Paths = responseBody.Paths
	(*db).Spec.KongApi.Methods = responseBody.Methods
	(*db).Spec.KongApi.KongID = responseBody.ID

	if (*db).Spec.AuthType == nlptv1.APPAUTH {
		//在route上创建jwt及acl插件
		if err := r.AddRouteJwtByKong(db); err != nil {
			return fmt.Errorf("request for add route jwt error: %+v", err)
		}
		if err := r.AddRouteAclByKong(db); err != nil {
			return fmt.Errorf("request for add route acl error: %+v", err)
		}
	}
	if (*db).Spec.ApiDefineInfo.Cors == "true" {
		if err := r.AddRouteCorsByKong(db); err != nil {
			return fmt.Errorf("request for add route acl error: %+v", err)
		}
	}
	return nil
}

func (r *Operator) AddRouteJwtByKong(db *nlptv1.Api) (err error) {
	id := db.Spec.KongApi.KongID
	klog.Infof("begin create jwt %s", id)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s%s%s", schema, r.Host, r.Port, "/routes/", id, "/plugins"))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	//TODO 后续确定超时策略后再添加token超时机制
	//verList := []string{"exp"}
	//configBody := &ConfigJwt{
	//ClaimsToVerify: verList, //校验的参数列表 exp
	//}
	requestBody := &JwtRequestBody{
		Name: "jwt", //插件名称
		//Config: *configBody,
	}
	responseBody := &JwtResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for create consumer error: %+v", errs)
	}
	klog.V(5).Infof("creation jwt code: %d, body: %s ", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.V(5).Infof("create jwt failed msg: %s\n", responseBody.Message)
		return fmt.Errorf("request for create jwt error: receive wrong status code: %s", string(body))
	}
	(*db).Spec.KongApi.JwtID = responseBody.ID

	if err != nil {
		return fmt.Errorf("create jwt error %s", responseBody.Message)
	}
	return nil
}

func (r *Operator) AddRouteAclByKong(db *nlptv1.Api) (err error) {
	id := db.Spec.KongApi.KongID
	klog.Infof("begin create acl %s", id)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s%s%s", schema, r.Host, r.Port, "/routes/", id, "/plugins"))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	whiteList := []string{db.ObjectMeta.Name}
	configBody := &AclReqConfig{
		Whitelist:        whiteList, //
		HideGroupsHeader: true,
	}
	requestBody := &AclRequestBody{
		Name:   "acl", //插件名称
		Config: *configBody,
	}
	responseBody := &AclResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for create acl error: %+v", errs)
	}
	klog.V(5).Infof("create acl code: %d, body: %s ", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.V(5).Infof("create acl failed msg: %s\n", responseBody.Message)
		return fmt.Errorf("request for create acl error: receive wrong status code: %s", string(body))
	}
	(*db).Spec.KongApi.AclID = responseBody.ID

	if err != nil {
		return fmt.Errorf("create acl error %s", responseBody.Message)
	}
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

//绑定api
func (r *Operator) AddConsumerToAcl(appId string, api *nlptv1.Api) (aclId string, err error) {
	id := api.ObjectMeta.Name
	klog.Infof("begin add consumer to acl %s", api.ObjectMeta.Name)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s%s%s", schema, r.Host, r.Port, "/consumers/", appId, "/acls"))
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

//解绑API
func (r *Operator) DeleteConsumerFromAcl(aclId string, comId string) (err error) {
	klog.Infof("delete consumer from acl %s.", comId)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	for k, v := range headers {
		request = request.Set(k, v)
	}
	klog.Infof("delete consumer is %s", fmt.Sprintf("%s://%s:%d%s%s%s%s", schema, r.Host, r.Port,
		"/consumers/", comId, "/acls/", aclId))
	response, body, errs := request.Delete(fmt.Sprintf("%s://%s:%d%s%s%s%s", schema, r.Host, r.Port,
		"/consumers/", comId, "/acls/", aclId)).End()
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

func (r *Operator) AddRouteCorsByKong(db *nlptv1.Api) (err error) {
	id := db.Spec.KongApi.KongID
	klog.Infof("begin create cors %s", id)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s%s%s", schema, r.Host, r.Port, "/routes/", id, "/plugins"))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &CorsRequestBody{
		Name: "cors", //插件名称
	}
	responseBody := &CorsResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for create cors error: %+v", errs)
	}
	klog.V(5).Infof("create cors code: %d, body: %s ", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.V(5).Infof("create cors failed msg: %s\n", responseBody.Message)
		return fmt.Errorf("request for create cors error: receive wrong status code: %s", string(body))
	}
	(*db).Spec.KongApi.CorsID = responseBody.ID

	if err != nil {
		return fmt.Errorf("create cors error %s", responseBody.Message)
	}
	return nil
}

func (r *Operator) getRouteInfoFromKong(db *nlptv1.Api) (rsp *ResponseBody, err error) {
	klog.Infof("begin get route info from kong %s %s", db.Spec.Name, db.Spec.KongApi.KongID)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	responseBody := &ResponseBody{}
	response, body, errs := request.Get(fmt.Sprintf("%s://%s:%d%s%s", schema, r.Host, r.Port, "/routes/", db.Spec.KongApi.KongID)).EndStruct(responseBody)
	if len(errs) > 0 {
		return nil, fmt.Errorf("request for get route info error: %+v", errs)

	}
	if response.StatusCode != 200 {
		return nil, fmt.Errorf("request for get route error: receive wrong status code: %s", string(body))
	}

	klog.Infof("get route info return: code:%d ,body:%s", response.StatusCode, string(body))
	klog.Infof("get route info methods:%v, paths:%v, hosts:%v", responseBody.Methods, responseBody.Paths, responseBody.Hosts)
	return responseBody, nil
}

func SliceEq(a, b []string) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		klog.Infof("SliceEq nil false %v:%v", a, b)
		return false
	}
	if len(a) != len(b) {
		klog.Infof("SliceEq len false %v:%v", a, b)
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			klog.Infof("SliceEq ele false %v:%v", a, b)
			return false
		}
	}
	return true
}

func (r *Operator) UpdateRouteInfoFromKong(db *nlptv1.Api, isOffline bool) (err error) {
	klog.Infof("begin update route info %s:%s", db.Spec.Name, db.Spec.KongApi.KongID)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"

	request = request.Patch(fmt.Sprintf("%s://%s:%d%s%s", schema, r.Host, r.Port, "/routes/", db.Spec.KongApi.KongID))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	protocols := []string{}
	methods := []string{}
	paths := []string{}
	id := db.Spec.Serviceunit.KongID
	requestBody := &RequestBody{}
	requestBody.Service = ServiceID{id}
	requestBody.Protocols = protocols
	if len(db.Spec.KongApi.Methods) != 0 {
		requestBody.Methods = db.Spec.KongApi.Methods
	}
	requestBody.Name = db.ObjectMeta.Name
	//设置为true会删除前缀 route When matching a Route via one of the paths, strip the matching prefix from the upstream request URL. Defaults to true.
	requestBody.StripPath = isOffline
	if db.Spec.Serviceunit.Type == "data" {
		methods = append(methods, strings.ToUpper(string(db.Spec.ApiDefineInfo.Method)))
		protocols = append(protocols, strings.ToLower(string(db.Spec.Serviceunit.Protocol)))
		requestBody.Protocols = protocols
		requestBody.Methods = methods
		queryPath := fmt.Sprintf("%s%s%s%s%s", "/api/v1/apis/", db.ObjectMeta.Namespace, "/", db.ObjectMeta.Name, "/data")
		paths = append(paths, queryPath)
		requestBody.Paths = paths
	} else {
		methods = append(methods, strings.ToUpper(string(db.Spec.ApiDefineInfo.Method)))
		protocols = append(protocols, strings.ToLower(string(db.Spec.Serviceunit.Protocol)))
		requestBody.Protocols = protocols
		requestBody.Methods = methods
		requestBody.Paths = db.Spec.KongApi.Paths
		if len(db.Spec.KongApi.Hosts) != 0 {
			requestBody.Hosts = db.Spec.KongApi.Hosts
		}
	}
	responseBody := &ResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Infof("request for update route error: %+v", errs)
		return fmt.Errorf("request for update route error: %+v", errs)
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("request for update route error: receive wrong status code: %s", string(body))
	}

	klog.Infof("update route info return: code:%d ,body:%s", response.StatusCode, string(body))
	klog.Infof("update route info methods:%v, paths:%v, hosts:%v", responseBody.Methods, responseBody.Paths, responseBody.Hosts)
	return nil
}

func (r *Operator) DeletePluginByKong(pluginId string) (err error) {
	klog.Infof("begin delete plugin %s", pluginId)
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

func (r *Operator) UpdateRouteByKong(db *nlptv1.Api) (err error) {
	klog.Infof("Enter UpdateRouteByKong name:%s, Host:%s, Port:%d", db.Name, r.Host, r.Port)
	rsp, errs := r.getRouteInfoFromKong(db)
	if errs != nil {
		return fmt.Errorf("request for get route info error: %+v", errs)
	}
	if SliceEq(rsp.Hosts, db.Spec.KongApi.Hosts) && SliceEq(rsp.Paths, db.Spec.KongApi.Paths) &&
		SliceEq(rsp.Methods, db.Spec.KongApi.Methods) && SliceEq(rsp.Protocols, db.Spec.KongApi.Protocols) &&
		rsp.StripPath == false {
		klog.Infof("no need update route info")
	} else {
		if errs := r.UpdateRouteInfoFromKong(db, false); errs != nil {
			return fmt.Errorf("request for get route info error: %+v", errs)
		}
	}
	//更新鉴权方式 无鉴权到APP鉴权 添加插件
	if len(db.Spec.KongApi.JwtID) == 0 && db.Spec.AuthType == nlptv1.APPAUTH {
		if err := r.AddRouteJwtByKong(db); err != nil {
			klog.Infof("request for add route jwt error: %+v", err)
			return fmt.Errorf("request for add route jwt error: %+v", err)
		}
		if err := r.AddRouteAclByKong(db); err != nil {
			klog.Infof("request for add route acl error: %+v", err)
			return fmt.Errorf("request for add route acl error: %+v", err)
		}
	}
	//更新鉴权方式 APP鉴权到无鉴权 删除插件
	if len(db.Spec.KongApi.JwtID) != 0 && db.Spec.AuthType == nlptv1.NOAUTH {
		if err := r.DeletePluginByKong(db.Spec.KongApi.JwtID); err != nil {
			klog.Infof("request for delete route jwt error: %+v", err)
			return fmt.Errorf("request for delete route jwt error: %+v", err)
		}
		(*db).Spec.KongApi.JwtID = ""

		if err := r.DeletePluginByKong(db.Spec.KongApi.AclID); err != nil {
			klog.Infof("request for delete route acl error: %+v", err)
			return fmt.Errorf("request for delete route acl error: %+v", err)
		}
		(*db).Spec.KongApi.AclID = ""
	}
	return nil
}

func (r *Operator) SyncApiCountFromPrometheus(m map[string]int) error {
	klog.Infof("sync api count from kong.")
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	responseBody := &PrometheusResponse{}

	response, body, errs := request.Get(fmt.Sprintf("%s://%s:%d%s", schema, r.PrometheusHost, r.PrometheusPort, "/api/v1/query?query=kong_latency_count{type=\"kong\"}")).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Infof("sync api count from kong error %+v.", errs)
		return fmt.Errorf("get api count error %+v", errs)

	}

	klog.Infof("SyncApiCountFromPrometheus: %d %s\n", response.StatusCode, string(body))
	if response.StatusCode != 200 {
		return fmt.Errorf("request for get api count error: wrong status code: %d", response.StatusCode)
	}
	result := responseBody.Data.Result
	if result == nil || len(result) == 0 {
		klog.Warning("sync api count from prometheus null.")
		return nil
	}
	for index := range result {
		route := responseBody.Data.Result[index].Metric.Route
		num := responseBody.Data.Result[index].Value[1].(string)
		count, _ := strconv.Atoi(num)
		m[route] = count
	}
	route := responseBody.Data.Result[0].Metric.Route
	num := responseBody.Data.Result[0].Value[1].(string)
	klog.Infof("SyncApiCountFromPrometheus ROUTE:  %s %s\n", route, num)
	klog.Infof("SyncApiCountFromPrometheus Result:  %+v", m)
	return nil
}

func (r *Operator) syncApiFailedCountFromKong(m map[string]int) error {
	klog.Infof("sync api failed count from kong.")
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	responseBody := &PrometheusResponse{}

	response, body, errs := request.Get(fmt.Sprintf("%s://%s:%d%s", schema, r.PrometheusHost, r.PrometheusPort,
		"/api/v1/query?query=kong_http_status{code!~\"2.*\"}")).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Errorf("sync api failed count from kong error %+v.", errs)
		return fmt.Errorf("get api failed count error %+v", errs)

	}
	klog.Infof("syncApiFailedCountFromKong: %d %s\n", response.StatusCode, string(body))
	result := responseBody.Data.Result
	if result == nil || len(result) == 0 {
		klog.Warning("sync api failed count from prometheus null.")
		return nil
	}
	for index := range result {
		route := responseBody.Data.Result[index].Metric.Route
		num := responseBody.Data.Result[index].Value[1].(string)
		count, _ := strconv.Atoi(num)
		m[route] = count
	}
	route := responseBody.Data.Result[0].Metric.Route
	num := responseBody.Data.Result[0].Value[1].(string)
	klog.Infof("syncApiFailedCountFromKong ROUTE:  %s%s\n", route, num)
	klog.Infof("syncApiFailedCountFromKong Result Map:  %+v", m)
	return nil
}

func (r *Operator) syncLatencyCountFromKong(m map[string]int) error {
	klog.Infof("sync api latency count from kong.")
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	responseBody := &PrometheusResponse{}

	response, body, errs := request.Get(fmt.Sprintf("%s://%s:%d%s", schema, r.PrometheusHost, r.PrometheusPort,
		"/api/v1/query?query=kong_latency_sum/kong_latency_count{type=\"request\"}")).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Infof("sync api latency from kong error %+v.", errs)
		return fmt.Errorf("get api latency error %+v", errs)

	}
	klog.Infof("syncApiLatencyCountFromKong: %d %s\n", response.StatusCode, string(body))
	result := responseBody.Data.Result
	if result == nil || len(result) == 0 {
		klog.Warning("sync api latency from prometheus null.")
		return nil
	}
	for index := range result {
		route := responseBody.Data.Result[index].Metric.Route
		num := responseBody.Data.Result[index].Value[1].(string)
		count, _ := strconv.ParseFloat(num, 64)
		intCount := int(count)
		klog.V(5).Infof("syncApiLatencyCountFromKong ROUTE count:  %s%s%d\n", route, num, intCount)
		m[route] = intCount
	}
	route := responseBody.Data.Result[0].Metric.Route
	num := responseBody.Data.Result[0].Value[1].(string)
	klog.Infof("syncApiLatencyCountFromKong ROUTE:  %s%s\n", route, num)
	klog.Infof("syncApiLatencyCountFromKong Result:  %+v", m)
	return nil
}

func (r *Operator) syncApiCallFrequencyFromKong(m map[string]int) error {
	klog.Infof("sync api call frequency count from kong.")
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	responseBody := &PrometheusResponse{}
	response, body, errs := request.Get(fmt.Sprintf("%s://%s:%d%s", schema, r.PrometheusHost, r.PrometheusPort,
		"/api/v1/query?query=rate(kong_latency_count{type=\"kong\"}[5m])")).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Infof("sync api call frequency from kong error %+v.", errs)
		return fmt.Errorf("get api call frequency error %+v", errs)
	}
	klog.Infof("sync api call frequency from kong response: %d %s\n", response.StatusCode, string(body))
	result := responseBody.Data.Result
	if result == nil || len(result) == 0 {
		klog.Warning("sync api call frequency from prometheus null.")
		return nil
	}
	for index := range result {
		route := responseBody.Data.Result[index].Metric.Route
		num := responseBody.Data.Result[index].Value[1].(string)
		count, _ := strconv.ParseFloat(num, 64)
		count = math.Floor(count*60 + 0.5)
		intCount := int(count)
		klog.V(5).Infof("sync api call frequency from kong route count:  %f\n", count)
		m[route] = intCount
	}
	route := responseBody.Data.Result[0].Metric.Route
	num := responseBody.Data.Result[0].Value[1].(string)
	klog.Infof("sync api call frequency from kong route info:  %s%s\n", route, num)
	klog.Infof("sync api call frequency from kong result:  %+v", m)
	return nil
}

func (r *Operator) CreateRouteByFission(db *nlptv1.Api, su *suv1.Serviceunit) (err error) {
	klog.Infof("Enter CreateRouteByFission name:%s, Host:%s, Port:%d", db.Name, r.Host, r.Port)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, r.FissionAddress.ControllerHost, r.FissionAddress.ControllerPort, HttpTriggerUrl))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	methods := strings.ToUpper(string(db.Spec.ApiDefineInfo.Method))
	funcPath := fmt.Sprintf("%s%s%s%s%s", "/api/v1/apis/", db.ObjectMeta.Namespace, "/", db.ObjectMeta.Name, "/function")

	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &RouteReqInfo{}
	requestBody.Metadata.Name = db.ObjectMeta.Name
	requestBody.Metadata.Namespace = db.ObjectMeta.Namespace
	requestBody.Spec.Functionref.Type = "name"
	requestBody.Spec.Functionref.Name = su.Spec.FissionRefInfo.FnName
	requestBody.Spec.Relativeurl = funcPath
	requestBody.Spec.Method = methods
	requestBody.Spec.Createingress = false
	responseBody := &FissionResInfoRsp{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Errorf("send  create route by fission error %+v", errs)
		return fmt.Errorf("request for create route by fission error: %+v", errs)
	}

	klog.V(5).Infof("creation route by fission response code and body: %d, %s",  response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.Errorf("create route by fission failed msg: %s\n", responseBody)
		return fmt.Errorf("request for create route by fission error: receive wrong body: %s", string(body))
	}
	klog.V(5).Infof("ID==: %s\n", responseBody.Name)
	return nil
}

