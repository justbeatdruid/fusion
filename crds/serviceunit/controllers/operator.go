package controllers

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"strings"
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
	FissionHost string
	FissionPort int
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


type EnvReqInfo struct {
	Metadata struct {
		Name string `json:"name"`
		Namespace string `json:"namespace"`
	} `json:"metadata"`
	Spec struct {
		Version int `json:"version"`
		Runtime struct {
			Image string `json:"image"`
		} `json:"runtime"`
		Builder struct {
			Image string `json:"image"`
			Command string `json:"command"`
		} `json:"builder"`
		Poolsize int `json:"poolsize"`   //3
		Keeparchive bool `json:"keeparchive"` //false
	} `json:"spec"`
}
type PkgRefInfoReq struct {
	Metadata struct {
		Name string `json:"name"`
		Namespace string `json:"namespace"`
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
	Spec struct {
		Environment struct {
			Namespace string `json:"namespace"`
			Name string `json:"name"`
		} `json:"environment"`
		Source struct {
			Type string `json:"type"`
			Literal []byte  `json:"literal"`
			Checksum struct {
			} `json:"checksum"`
		} `json:"source"`
		Deployment struct {
			Type string `json:"type"`
			Literal []byte  `json:"literal"`
			Checksum struct {
			} `json:"checksum"`
		} `json:"deployment"`
		BuildCommand string `json:"buildcmd,omitempty"`
	} `json:"spec"`
	Status struct {
		BuildStatus string `json:"buildstatus"`
		BuildLogs string `json:"buildLogs"`
	} `json:"status"`
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

const (
	FissionRouterPort       = 80
	FissionRouter           = "router.fission"
	EnvUrl                  = "/v2/environments"         //成功
	PkgUrl                  = "/v2/packages" //读取消息体失败
	FunctionUrl             = "/v2/functions" //创建Topic失败
	NodeJsImage             ="fission/node-env"
	NodeJsBuild             ="fission/node-builder"
	PythonImage             ="fission/python-env"
	PythonBuild             ="fission/python-builder"
	GoImage                 ="fission/go-env"
	GoBuild                 ="fission/go-builder"
	Command                 = "build"
	NodeJs                  = "nodejs"
	Python                  = "python"
	Go                      = "go"
	Zip                     = ".zip"
)

type FunctionReqInfo struct {
	Metadata struct {
		Name string `json:"name"`
		Namespace string `json:"namespace"`
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
	Spec struct {
		Environment struct {
			Namespace string `json:"namespace"`
			Name string `json:"name"`
		} `json:"environment"`
		Package struct {
			Packageref struct {
				Namespace string `json:"namespace"`
				Name string `json:"name"`
				Resourceversion string `json:"resourceversion"`
			} `json:"packageref"`
			FunctionName string `json:"functionName"`
		} `json:"package"`
		Secrets interface{} `json:"secrets"`
		Configmaps interface{} `json:"configmaps"`
		Resources interface{} `json:"resources"`
		InvokeStrategy struct {
			ExecutionStrategy struct {
				ExecutorType string `json:"ExecutorType"`
				MinScale int `json:"MinScale"`
				MaxScale int `json:"MaxScale"`
				TargetCPUPercent int `json:"TargetCPUPercent"`
				SpecializationTimeout int `json:"SpecializationTimeout"`
			} `json:"ExecutionStrategy"`
			StrategyType string `json:"StrategyType"`
		} `json:"InvokeStrategy"`
		FunctionTimeout int `json:"functionTimeout"`
	} `json:"spec"`
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

func NewOperator(host string, port int, fissionHost string, fissionPort int, cafile string) (*Operator, error) {
	klog.Infof("NewOperator  event:%s %d %s", host, port, cafile)
	return &Operator{
		Host:   host,
		Port:   port,
		FissionHost:   fissionHost,
		FissionPort:   fissionPort,
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
	switch db.Spec.Type {
	case nlptv1.WebService:
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
	case nlptv1.FunctionService :
		requestBody = &RequestBody{
			Name:     db.ObjectMeta.Name,
			Protocol: "http",
			Host:     FissionRouter,
			Port:     FissionRouterPort,
			TimeOut:  60000,
			WirteOut: 60000,
			ReadOut:  60000,
		}
	}
	if db.Spec.Type == nlptv1.FunctionService {
		fn, err := r.CreateFunction(db)
		if err != nil{
			return fmt.Errorf("create function error: %+v", err)
		}
		klog.V(5).Infof("create function result fn: %+v", fn)
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
	if db.Spec.Type == nlptv1.DataService || db.Spec.Type == nlptv1.FunctionService {
		(*db).Spec.KongService.Host = responseBody.Host
		(*db).Spec.KongService.Protocol = responseBody.Protocol
		(*db).Spec.KongService.Port = responseBody.Port
		(*db).Spec.KongService.ReadOut = responseBody.ReadTimeout
		(*db).Spec.KongService.WirteOut = responseBody.WriteTimeout
		(*db).Spec.KongService.TimeOut = responseBody.ConnectTimeout
	}
	(*db).Spec.KongService.ID = responseBody.ID
	return nil
}

func (r *Operator) DeleteServiceByKong(db *nlptv1.Serviceunit) (err error) {
	klog.Infof("delete service %s %s", db.ObjectMeta.Name, db.Spec.KongService.ID)
	//删除函数api
	err = r.DeleteFunction(db)
	if err != nil {
		return fmt.Errorf("delete function error: %+v",err)
	}
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	for k, v := range headers {
		request = request.Set(k, v)
	}
	id := db.Spec.KongService.ID
	klog.Infof("delete service id %s %s", id, fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, id))
	response, body, errs := request.Delete(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, id)).End()
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
	klog.Infof("update service id %s %s", id, fmt.Sprintf("%s://%s:%d%s/%s", schema, r.Host, r.Port, path, id))
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
	if db.Spec.Type == nlptv1.FunctionService {
		requestBody = &RequestBody{
			Name:     db.ObjectMeta.Name,
			Protocol: db.Spec.KongService.Protocol,
			Host:     db.Spec.KongService.Host,
			Port:     db.Spec.KongService.Port,
			TimeOut:  db.Spec.KongService.TimeOut,
			WirteOut: db.Spec.KongService.WirteOut,
			ReadOut:  db.Spec.KongService.ReadOut,
		}
		fn, err := r.UpdateFunction(db)
		if err != nil{
			return fmt.Errorf("update function error: %+v ",err)
		}
		klog.V(5).Infof("update function result fn: %+v", fn)
	}
	responseBody := &ResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for update service error: %+v", errs)
	}
	klog.V(5).Infof("update service response body: %s", string(body))
	//patch接口返回200
	if response.StatusCode != 200 {
		return fmt.Errorf("request for update service error: receive wrong status code: %s", string(body))
	}
	if db.Spec.Type == nlptv1.DataService || db.Spec.Type == nlptv1.FunctionService {
		(*db).Spec.KongService.Host = responseBody.Host
		(*db).Spec.KongService.Protocol = responseBody.Protocol
		(*db).Spec.KongService.Port = responseBody.Port
	}
	return nil
}

func GetContentsPkg(filePath string) ([]byte, error) {
	code, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading %v", filePath)
	}
	return code, nil
}
func GetLanguage(lan string) string {
	switch  lan {
	case NodeJs:
		return NodeJs
	case Python:
		return Python
	case Go:
		return Go
	}
	return NodeJs
}

func (r *Operator) CreateEnv(db *nlptv1.Serviceunit) (*FissionResInfoRsp, error) {
	klog.Infof("Enter CreateEnv :%s, Host:%s, Port:%d", db.ObjectMeta.Name, r.Host, r.Port)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, r.FissionHost, r.FissionPort, EnvUrl))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &EnvReqInfo{}
	languageInfo := db.Spec.FissionRefInfo.Language
	name := fmt.Sprintf("%v-%v", languageInfo, db.ObjectMeta.Name)
	requestBody.Metadata.Name = name
	requestBody.Metadata.Namespace =  db.ObjectMeta.Namespace
	switch languageInfo  {
	case NodeJs:
		requestBody.Spec.Runtime.Image = NodeJsImage
		requestBody.Spec.Builder.Image = NodeJsBuild
	case Python:
		requestBody.Spec.Runtime.Image = PythonImage
		requestBody.Spec.Builder.Image = PythonBuild
	case Go:
		requestBody.Spec.Runtime.Image = GoImage
		requestBody.Spec.Builder.Image = GoBuild
	}
	requestBody.Spec.Builder.Command = Command
	requestBody.Spec.Version = 2
	requestBody.Spec.Poolsize = 3
	requestBody.Spec.Keeparchive = false

	responseBody := &FissionResInfoRsp{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Errorf("request for create env error %+v", errs)
		return nil, fmt.Errorf("request for create env error: %+v", errs)
	}
	klog.V(5).Infof("create env code and body: %d %s\n", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.Errorf("create failed msg: %s\n", responseBody)
		return nil, fmt.Errorf("request for create rate error: receive wrong status code: %s", string(body))
	}
	klog.V(5).Infof("create env name: %s\n", responseBody.Name)
	return responseBody, nil
}

func (r *Operator) CreatePkgByFile(db *nlptv1.Serviceunit, env *FissionResInfoRsp) (*FissionResInfoRsp, error){
	klog.Infof("Enter CreatePkgByFile :%s, Host:%s, Port:%d", db.ObjectMeta.Name, r.Host, r.Port)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, r.FissionHost, r.FissionPort, PkgUrl))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &PkgRefInfoReq{}
	name := fmt.Sprintf("%v-%v-pkg", db.Spec.FissionRefInfo.FnName, db.ObjectMeta.Name)
	requestBody.Metadata.Name = name
	requestBody.Metadata.Namespace = db.ObjectMeta.Namespace
	requestBody.Spec.Environment.Name = env.Name
	requestBody.Spec.Environment.Namespace = db.ObjectMeta.Namespace
    //判断是否是文件还是在线编辑代码
	if len(db.Spec.FissionRefInfo.FnFile)>0 {
		if strings.Contains(db.Spec.FissionRefInfo.FnFile, Zip){
			requestBody.Spec.Source.Type = "literal"
			requestBody.Spec.Source.Literal, _ = GetContentsPkg(db.Spec.FissionRefInfo.FnFile)
			requestBody.Spec.BuildCommand = db.Spec.FissionRefInfo.BuildCmd
		}else {
			requestBody.Spec.Deployment.Type = "literal"
			requestBody.Spec.Deployment.Literal, _ = GetContentsPkg(db.Spec.FissionRefInfo.FnFile)
		}
	}else {
		requestBody.Spec.Deployment.Type = "literal"
		requestBody.Spec.Deployment.Literal = []byte(db.Spec.FissionRefInfo.FnCode)
	}
	responseBody := &FissionResInfoRsp{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Errorf("request for create pkg error %+v", errs)
		return nil, fmt.Errorf("request for create pkg error: %+v", errs)
	}
	klog.V(5).Infof("create pkg code and body: %d %s\n", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.Errorf("create pkg failed msg: %s\n", responseBody)
		return nil, fmt.Errorf("request for create rate error: receive wrong status code: %s", string(body))
	}
	klog.V(5).Infof("create pkg name: %s\n", responseBody.Name)
	return  responseBody, nil
}

func (r *Operator) CreateFnByEnvAndPkg(db *nlptv1.Serviceunit, env *FissionResInfoRsp, pkg *FissionResInfoRsp) (*FissionResInfoRsp, error){
	klog.Infof("Enter CreateFnByEnvAndPkg :%s, Host:%s, Port:%d", db.ObjectMeta.Name, r.Host, r.Port)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s", schema, r.FissionHost, r.FissionPort, FunctionUrl))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &FunctionReqInfo{}
	requestBody.Metadata.Name = db.Spec.FissionRefInfo.FnName
	requestBody.Metadata.Namespace = db.ObjectMeta.Namespace
	requestBody.Spec.Environment.Name = env.Name
	requestBody.Spec.Environment.Namespace = env.Namespace
	requestBody.Spec.Package.Packageref.Namespace = pkg.Namespace
	requestBody.Spec.Package.Packageref.Name = pkg.Name
	requestBody.Spec.Package.Packageref.Resourceversion = pkg.ResourceVersion
	//函数入口
	requestBody.Spec.Package.FunctionName = db.Spec.FissionRefInfo.Entrypoint
	requestBody.Spec.InvokeStrategy.ExecutionStrategy.ExecutorType = "poolmgr"
	requestBody.Spec.InvokeStrategy.StrategyType = "execution"
	requestBody.Spec.FunctionTimeout = 120
	
	responseBody := &FissionResInfoRsp{}
	//判断package的状态是否完成
	for i:=1;i<15;i++{
		time.Sleep(time.Duration(i)*time.Second)
		Pkg,err := r.getPkgVersion(db)
		if err!=nil {
			return nil,fmt.Errorf("get pkgResourceVersion error: %v",err )
		}
		//单文件成功状态是none,zip包成功状态是succeeded
		if Pkg.Status.BuildStatus=="none"||Pkg.Status.BuildStatus=="succeeded"{
			break
		}
	}
	Pkg,err := r.getPkgVersion(db)
	if err!=nil {
		return nil,fmt.Errorf("get pkgResourceVersion error: %v",err )
	}
	if Pkg.Status.BuildStatus=="failed"{
		return nil, fmt.Errorf("create function error: +%v",Pkg.Status.BuildLogs)
	}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Errorf("request for create function error %+v", errs)
		return nil, fmt.Errorf("request for create function error: %+v", errs)
	}
	klog.V(5).Infof("create function code and body: %d %s\n", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.Errorf("create function failed msg: %s\n", responseBody)
		return nil, fmt.Errorf("request for create rate error: receive wrong status code: %s", string(body))
	}
	klog.V(5).Infof("create function name: %s\n", responseBody.Name)
	return  responseBody, nil
}

func (r *Operator) CreateFunction(db *nlptv1.Serviceunit) (*FissionResInfoRsp, error){
	klog.Infof("Enter CreateFunction :%s, Host:%s, Port:%d", db.ObjectMeta.Name, r.Host, r.Port)
    env, err := r.CreateEnv(db)
	if err != nil {
		return nil, fmt.Errorf("request for create env error: %+v", err)
	}
	(*db).Spec.FissionRefInfo.EnvName = env.Name
	time.Sleep(5*time.Second)
	pkg, err := r.CreatePkgByFile(db, env)
	if err != nil {
		return nil, fmt.Errorf("request for create pkg error: %+v", err)
	}
	(*db).Spec.FissionRefInfo.PkgName = pkg.Name
	(*db).Spec.FissionRefInfo.PkgResourceVersion = pkg.ResourceVersion
	fn, err := r.CreateFnByEnvAndPkg(db, env, pkg)
	if err != nil {
		return nil, fmt.Errorf("request for create function error: %+v", err)
	}
	(*db).Spec.FissionRefInfo.FnResourceVersion = fn.ResourceVersion
	klog.V(5).Infof("create function name: %s\n", fn.Name)
	return  fn, nil
}

func (r *Operator) UpdateFunction(db *nlptv1.Serviceunit)(*FissionResInfoRsp,error){
	klog.Infof("Enter UpdateFunction :%s, Host:%s, Port:%d", db.ObjectMeta.Name, r.Host, r.Port)
	pkg, err := r.UpdatePkgByFile(db)
	if err!=nil{
		return nil, fmt.Errorf("request for update pkg error: %+v", err)
	}
	(*db).Spec.FissionRefInfo.PkgResourceVersion=pkg.ResourceVersion
	(*db).Spec.FissionRefInfo.PkgName=pkg.Name
    fn, err := r.UpdateFnByEnvAndPkg(db,pkg)
    if err != nil {
    	return nil, fmt.Errorf("request for update function error: %+v", err)
	}
	(*db).Spec.FissionRefInfo.FnResourceVersion=fn.ResourceVersion
	klog.V(5).Infof("update function name: %s\n", fn.Name)
	return  fn, nil
}

func (r *Operator) UpdatePkgByFile(db *nlptv1.Serviceunit)(*FissionResInfoRsp,error)  {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	//查询package的resourceversion
	Pkg,err:=r.getPkgVersion(db)
	if err!=nil{
		return nil,fmt.Errorf("get pkgResourceVersion error: %v",err)
	}
	//更新package的url
	request = request.Put(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.FissionHost, r.FissionPort, PkgUrl,db.Spec.FissionRefInfo.PkgName))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &PkgRefInfoReq{}
	requestBody.Metadata.Name = db.Spec.FissionRefInfo.PkgName
	requestBody.Metadata.Namespace = db.ObjectMeta.Namespace
	requestBody.Spec.Environment.Name = db.Spec.FissionRefInfo.EnvName
	requestBody.Spec.Environment.Namespace = db.ObjectMeta.Namespace
	requestBody.Metadata.ResourceVersion = Pkg.Metadata.ResourceVersion
	//判断是否是文件还是在线编辑
	if len(db.Spec.FissionRefInfo.FnFile)>0{
		if strings.Contains(db.Spec.FissionRefInfo.FnFile, Zip){
			requestBody.Spec.Source.Type = "literal"
			requestBody.Spec.Source.Literal, _ = GetContentsPkg(db.Spec.FissionRefInfo.FnFile)
			requestBody.Spec.BuildCommand = db.Spec.FissionRefInfo.BuildCmd
		}else {
			requestBody.Spec.Deployment.Type = "literal"
			requestBody.Spec.Deployment.Literal, _ = GetContentsPkg(db.Spec.FissionRefInfo.FnFile)
		}
	}else {
		requestBody.Spec.Deployment.Type = "literal"
		requestBody.Spec.Deployment.Literal = []byte(db.Spec.FissionRefInfo.FnCode)
	}
	responseBody := &FissionResInfoRsp{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Errorf("request for update pkg error %+v", errs)
		return nil, fmt.Errorf("request for update pkg error: %+v", errs)
	}
	klog.V(5).Infof("update pkg code and body: %d %s\n", response.StatusCode, string(body))
	if response.StatusCode != 200 {
		klog.Errorf("update pkg failed msg: %s\n", responseBody)
		return nil, fmt.Errorf("request for update rate error: receive wrong status code: %s", string(body))
	}
	klog.V(5).Infof("update pkg name: %s\n", responseBody.Name)
	return  responseBody, nil
}

func (r *Operator) UpdateFnByEnvAndPkg(db *nlptv1.Serviceunit,pkg *FissionResInfoRsp)(*FissionResInfoRsp, error){
	klog.Infof("Enter UpdateFnByEnvAndPkg :%s, Host:%s, Port:%d", db.ObjectMeta.Name, r.Host, r.Port)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Put(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.FissionHost, r.FissionPort, FunctionUrl,db.Spec.FissionRefInfo.FnName))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &FunctionReqInfo{}
	requestBody.Metadata.Name = db.Spec.FissionRefInfo.FnName
	requestBody.Metadata.Namespace = db.ObjectMeta.Namespace
	requestBody.Spec.Environment.Name = db.Spec.FissionRefInfo.EnvName
	requestBody.Spec.Environment.Namespace = db.ObjectMeta.Namespace
	requestBody.Spec.Package.Packageref.Namespace = db.ObjectMeta.Namespace
	requestBody.Spec.Package.Packageref.Name = db.Spec.FissionRefInfo.PkgName
	//函数入口
	requestBody.Spec.Package.FunctionName = db.Spec.FissionRefInfo.Entrypoint
	requestBody.Spec.InvokeStrategy.ExecutionStrategy.ExecutorType = "poolmgr"
	requestBody.Spec.InvokeStrategy.StrategyType = "execution"
	requestBody.Spec.FunctionTimeout = 120

	responseBody := &FissionResInfoRsp{}

	//判断package的状态是否完成
	for i:=1;i<15;i++{
		time.Sleep(time.Duration(i)*time.Second)
		Pkg,err := r.getPkgVersion(db)
		if err!=nil {
			return nil,fmt.Errorf("get pkgResourceVersion error: %v",err )
		}
		//单文件成功状态是none,zip包成功状态是succeeded
		if Pkg.Status.BuildStatus=="none"||Pkg.Status.BuildStatus=="succeeded"{
			requestBody.Spec.Package.Packageref.Resourceversion = Pkg.Metadata.ResourceVersion
			break
		}
	}
	//查询function的resourceversion
	Fn,err:=r.getFnVersion(db)
	if err!=nil{
		return nil,fmt.Errorf("get FnResourceVersion error: %v",err)
	}
	requestBody.Metadata.ResourceVersion = Fn.Metadata.ResourceVersion
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Errorf("request for update function error %+v", errs)
		return nil, fmt.Errorf("request for update function error: %+v", errs)
	}

	klog.V(5).Infof("update function code and body: %d %s\n", response.StatusCode, string(body))
	if response.StatusCode != 200 {
		klog.Errorf("update function failed msg: %s\n", responseBody)
		return nil, fmt.Errorf("request for update function error: receive wrong status code: %s", string(body))
	}
	klog.V(5).Infof("update function name: %s\n", responseBody.Name)
	return  responseBody, nil

}
//获取pkg的信息
func (r *Operator) getPkgVersion(db *nlptv1.Serviceunit)(*PkgRefInfoReq,error){
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	//查询package的resourceversion
	request = request.Get(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.FissionHost, r.FissionPort, PkgUrl,db.Spec.FissionRefInfo.PkgName)).Query("namespace="+db.Namespace)
	responseBody := &PkgRefInfoReq{}
	_, _, errs := request.Send("").EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Errorf("request for get pkg error %+v", errs)
		return nil, fmt.Errorf("request for get pkg error: %+v", errs)
	}
	return  responseBody,nil
}

//获取function的信息
func (r *Operator) getFnVersion(db *nlptv1.Serviceunit)(*FunctionReqInfo,error){
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	//查询function的resourceversion
	request = request.Get(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.FissionHost, r.FissionPort, FunctionUrl,db.Spec.FissionRefInfo.FnName)).Query("namespace="+db.Namespace)
	responseBody := &FunctionReqInfo{}
	_, _, errs := request.Send("").EndStruct(responseBody)
	if len(errs) > 0 {
		klog.Errorf("request for get function error %+v", errs)
		return nil, fmt.Errorf("request for get function error: %+v", errs)
	}
	return  responseBody,nil
}
func (r *Operator) DeleteFunction(db *nlptv1.Serviceunit)error{
	klog.Infof("Enter CreateFunction :%s, Host:%s, Port:%d", db.ObjectMeta.Name, r.Host, r.Port)
	err := r.DeleteFn(db)
	if err != nil {
		return  fmt.Errorf("request for delete function error: %+v", err)
	}
	err = r.DeletePkg(db)
	if err != nil{
		return fmt.Errorf("request for delete package error: %+v", err)
	}
	err = r.DeleteEnv(db)
	if err != nil{
		return fmt.Errorf("request for delete env error: %+v", err)
	}
	return nil
}

func (r *Operator) DeleteFn(db *nlptv1.Serviceunit)error{
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Delete(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.FissionHost, r.FissionPort,FunctionUrl,db.Spec.FissionRefInfo.FnName)).Query("namespace="+db.Namespace)
	response, body, errs := request.Send("").End()
	if len(errs) > 0 {
		klog.Errorf("request for delete function error %+v", errs)
		return fmt.Errorf("request for delete function error: %+v", errs)
	}
	klog.V(5).Infof("delete function code: %d %s\n", response.StatusCode,body)
	//function不存在，返回404
	if response.StatusCode ==404{
		return nil
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("request for delete function error: receive wrong status code: %s", string(body))
	}
	return nil
}

func (r *Operator) DeletePkg(db *nlptv1.Serviceunit) error {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Delete(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.FissionHost, r.FissionPort,PkgUrl,db.Spec.FissionRefInfo.PkgName)).Query("namespace="+db.Namespace)
	response, body, errs := request.Send("").End()
	if len(errs) > 0 {
		klog.Errorf("request for delete package error %+v", errs)
		return fmt.Errorf("request for delete package error: %+v", errs)
	}
	klog.V(5).Infof("delete package code: %d %s\n", response.StatusCode,body)
	//package不存在，返回404
	if response.StatusCode ==404{
		return nil
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("request for delete package error: receive wrong status code: %s", string(body))
	}
	return nil
}

func (r *Operator) DeleteEnv(db *nlptv1.Serviceunit) error {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Delete(fmt.Sprintf("%s://%s:%d%s/%s", schema, r.FissionHost, r.FissionPort,EnvUrl,db.Spec.FissionRefInfo.EnvName)).Query("namespace="+db.Namespace)
	response, body, errs := request.Send("").End()
	if len(errs) > 0 {
		klog.Errorf("request for delete environment error %+v", errs)
		return fmt.Errorf("request for delete environment error: %+v", errs)
	}
	klog.V(5).Infof("delete environment code: %d %s\n", response.StatusCode,body)
	//env不存在，返回404
	if response.StatusCode ==404{
		return nil
	}
	if response.StatusCode != 200 {
		return fmt.Errorf("request for delete environment error: receive wrong status code: %s", string(body))
	}
	return nil
}