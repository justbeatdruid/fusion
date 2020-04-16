/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	dwv1 "github.com/chinamobile/nlpt/crds/api/datawarehouse/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const ServiceunitLabel = "nlpt.cmcc.com/serviceunit"
const applicationLabel = "nlpt.cmcc.com/application"

func ApplicationLabel(id string) string {
	return strings.Join([]string{applicationLabel, id}, ".")
}

func IsApplicationLabel(l string) bool {
	match, _ := regexp.MatchString(fmt.Sprintf("%s.([0-9a-f]{16})", applicationLabel), l)
	return match
}

func GetIDFromApplicationLabel(l string) string {
	if !IsApplicationLabel(l) {
		return "wronglabel"
	}
	return l[len(applicationLabel)+2:]
}

// ApiSpec defines the desired state of Api
type ApiSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name           string        `json:"name"`
	Description    string        `json:"description"`
	Serviceunit    Serviceunit   `json:"serviceunit"`
	Applications   []Application `json:"applications"`
	Users          []User        `json:"users"`
	ApiType        ApiType       `json:"apiType"` //API类型
	AuthType       AuthType      `json:"authType"`
	Tags           string        `json:"tags"`
	ApiBackendType string        `json:"apiBackendType"`
	//data api
	Frequency  int        `json:"frequency"`
	Method     Method     `json:"method"`
	Protocol   Protocol   `json:"protocol"`
	ReturnType ReturnType `json:"returnType"`
	// Data API related attributes
	// Simple RDB API
	RDBQuery *RDBQuery `json:"rdbQuery,omitempty"`
	// Datawarehouse API, define a query
	DataWarehouseQuery *dwv1.DataWarehouseQuery `json:"datawarehouseQuery,omitempty"`
	//web api
	ApiDefineInfo ApiDefineInfo `json:"apiDefineInfo"`
	KongApi       KongApiInfo   `json:"kongApi"`
	ApiReturnInfo ApiReturnInfo `json:"apiReturnInfo"`
	ApiQueryInfo  ApiQueryInfo  `json:"apiQueryInfo"`
	//api publishInfo
	PublishInfo PublishInfo `json:"publishInfo"`

	Traffic     Traffic     `json:"traffic"`
	Restriction Restriction `json:"restriction"`
}

type Serviceunit struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Group  string `json:"group"`
	KongID string `json:"kongID"`
	Type   string `json:"Type"`
	Host   string `json:"Host"`
	Port   int    `json:"Port"`
	//API的协议从服务单元获取
	Protocol string `json:"protocol"`
}

type Traffic struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	SpecialID []string `json:"specialID"`
}

type Restriction struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PublishInfo struct {
	Version      string `json:"version"`
	Host         string `json:"host"`
	PublishCount int    `json:"publishCount"`
}

type Application struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Group string `json:"group"`
	AclID string `json:"aclId"`
}

type User struct {
	Name     string `json:"name"`
	InCharge bool   `json:"inCharge"`
}

type Method string

const (
	POST   Method = "POST"
	GET    Method = "GET"
	DELETE Method = "DELETE"
	PUT    Method = "PUT"
	OPTION Method = "OPTION"
	LIST   Method = "LIST"
	PATCH  Method = "PATCH"
)

type ApiType string

const (
	PUBLIC  ApiType = "public"
	PRIVATE ApiType = "private"
)

type AuthType string

const (
	NOAUTH  AuthType = "NOAUTH"
	APPAUTH AuthType = "APPAUTH"
)

type Protocol string

const (
	HTTP  Protocol = "HTTP"
	HTTPS Protocol = "HTTPS"
)

type ReturnType string

const (
	Json ReturnType = "json"
)

type RDBQuery struct {
	Table       string       `json:"table"`
	QueryFields []QueryField `json:"queryFields"`
}

type QueryField struct {
	Field       string `json:"field"`
	Type        string `json:"type"`
	Operator    string `json:"operator"`
	Description string `json:"description"`
}

func RDBParameterFromQuery(q QueryField) ApiParameter {
	ap := ApiParameter{
		Name:        q.Field,
		Type:        ParameterType(q.Type),
		Description: q.Description,
	}
	return ap
}

type ApiQueryInfo struct {
	WebParams []WebParams `json:"webParams"`
}

//define api
type ApiDefineInfo struct {
	Path      string   `json:"path"`
	MatchMode string   `json:"matchMode"`
	Method    Method   `json:"method"`
	Protocol  Protocol `json:"protocol"` //直接从服务单元里面获取不需要前台传入
	Cors      string   `json:"cors"`
}

//define api return
type ApiReturnInfo struct {
	NormalExample  string `json:"normalExample"`
	FailureExample string `json:"failureExample"`
}

//define webbackend info
type KongApiInfo struct {
	//Kong变量
	//A list of domain names that match this Route. With form-encoded, the notation is hosts[]=example.com&hosts[]=foo.test. With JSON, use an Array.
	Hosts         []string `json:"hosts"`
	Paths         []string `json:"paths"` //kong 是数组 界面是字符串
	Headers       []string `json:"headers"`
	Methods       []string `json:"methods"`
	HttpsCode     int      `json:"https_redirect_status_code"`
	RegexPriority int      `json:"regex_priority"`
	StripPath     bool     `json:"strip_path"`
	PreserveHost  bool     `json:"preserve_host"`
	Snis          []string `json:"snis"`
	Protocols     []string `json:"protocols"`
	KongID        string   `json:"kong_id"`
	JwtID         string   `json:"jwt_id"`
	AclID         string   `json:"acl_id"`
	CorsID        string   `json:"cors_id"`
	PrometheusID  string   `json:"prometheus_id"`
}

type ApiBind struct {
	ID string `json:"id"`
}

func ParameterFromDataWarehouseQuery(q dwv1.QueryProperty) ApiParameter {
	ap := ApiParameter{
		Name: fmt.Sprintf("%s.%s", q.TableName, q.PropertyName),
		Type: ParameterType(q.PhysicalType),
	}
	return ap
}

type ParameterType string

const (
	Int    ParameterType = "number"
	Bool   ParameterType = "bool"
	String ParameterType = "string"
)

func (f QueryField) Validate() error {
	for k, v := range map[string]string{
		"field": f.Field,
		"type":  f.Type,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	return nil
}

type ApiParameter = WebParams

type WebParams struct {
	Name        string        `json:"name"`     //必须
	Type        ParameterType `json:"type"`     //必须
	Location    LocationType  `json:"location"` //必须
	Required    bool          `json:"required"` //必须
	Operator    string        `json:"operator"`
	DefValue    string        `json:"valueDefault"`
	Example     string        `json:"example"`
	Description string        `json:"description"`
	ValidEnable int           `json:"validEnable"`
	MinNum      int           `json:"minNum"`
	MaxNum      int           `json:"maxNum"`
	MinSize     int           `json:"minSize"`
	MaxSize     int           `json:"maxSize"`
}

type LocationType string

const (
	Path   LocationType = "path"
	Query  LocationType = "query"
	Header LocationType = "header"
	Body   LocationType = "body"
)

type Operator string

const (
	Equal Operator = "="
	More  Operator = ">"
	Less  Operator = "<"
	In    Operator = "in"
	Like  Operator = "like"
)

// ApiStatus defines the observed state of Api
type ApiStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status           Status        `json:"status"`
	Action           Action        `json:"action"`
	PublishStatus    PublishStatus `json:"publishStatus"`
	AccessLink       AccessLink    `json:"access"`
	UpdatedAt        metav1.Time   `json:"updatedAt"`
	ReleasedAt       metav1.Time   `json:"releasedAt"`
	ApplicationCount int           `json:"applicationCount"`
	CalledCount      int           `json:"calledCount"`
	Message          string        `json:"msg"`

	Applications map[string]ApiApplicationStatus `json:"applications"`
}

type AccessLink string

func (a AccessLink) Parse() error {
	if !(strings.HasPrefix(string(a), "http://") || strings.HasPrefix(string(a), "https://")) {
		return fmt.Errorf("url should start with a scheme")
	}
	if _, err := url.Parse(string(a)); err != nil {
		return err
	}
	return nil
}

type Status string

const (
	Init    Status = "init"
	Running Status = "running"
	Success Status = "success"
	Error   Status = "error"
)

type Action string

const (
	Create  Action = "create"  //create api
	Update  Action = "update"  //update api
	Delete  Action = "delete"  //delete api
	Publish Action = "publish" //publish api
	Offline Action = "offline" //offline api
	UnBind  Action = "unbind"  //unbind api
	Bind    Action = "bind"    //unbind api
)

type PublishStatus string

// only update when exec publish or exec offline
const (
	UnRelease PublishStatus = "unRelease" //未发布
	Released  PublishStatus = "released"  //已发布
	Offlined  PublishStatus = "offlined"  //已下线
)

type ApiApplicationStatus struct {
	AppID            string `json:"appID"`
	BindingSucceeded bool   `json:"bindingSucceeded"`
	Message          string `json:"message"`
}

// +kubebuilder:object:root=true

// Api is the Schema for the apis API
type Api struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApiSpec   `json:"spec,omitempty"`
	Status ApiStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApiList contains a list of Api
type ApiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Api `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Api{}, &ApiList{})
}
