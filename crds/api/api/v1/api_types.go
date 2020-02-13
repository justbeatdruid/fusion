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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const ServiceunitLabel = "serviceunit"
const applicationLabel = "application"

func ApplicationLabel(id string) string {
	return strings.Join([]string{applicationLabel, id}, "/")
}

func IsApplicationLabel(l string) bool {
	match, _ := regexp.MatchString(fmt.Sprintf("%s/([0-9a-f]{16})", applicationLabel), l)
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

	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Serviceunit  Serviceunit   `json:"serviceunit"`
	Applications []Application `json:"applications"`
	Users        []User        `json:"users"`
	Frequency    int           `json:"frequency"`
	Method       Method        `json:"method"`
	Protocol     Protocol      `json:"protocol"`
	ReturnType   ReturnType    `json:"returnType"`
	ApiFields    []Field       `json:"apiFields"`
	WebParams    []WebParams   `json:"webParams"`
	KongApi      KongApiInfo   `json:"kongApi"`
	PublishInfo  PublishInfo   `json:"publishInfo"`
	ApiType      ApiType       `json:"apiType"` //API类型
	AuthType     AuthType      `json:"authType"`
	Traffic      Traffic       `json:"traffic"`
}

type KongApiInfo struct {
	//Kong变量
	//A list of domain names that match this Route. With form-encoded, the notation is hosts[]=example.com&hosts[]=foo.test. With JSON, use an Array.
	Hosts         []string `json:"hosts"`
	Paths         []string `json:"paths"`
	Headers       []string `json:"Headers"`
	HttpsCode     int      `json:"https_redirect_status_code"`
	RegexPriority int      `json:"regex_priority"`
	StripPath     bool     `json:"strip_path"`
	PreserveHost  bool     `json:"preserve_host"`
	Snis          []string `json:"snis"`
	Protocols     []string `json:"protocols"`
	KongID        string   `json:"kong_id"`
	JwtID         string   `json:"jwt_id"`
	AclID         string   `json:"acl_id"`
}

type Serviceunit struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Group  string `json:"group"`
	KongID string `json:"kongID"`
	Type   string `json:"Type"`
	Host   string `json:"Host"`
	Port   int    `json:"Port"`
}

type Traffic struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PublishInfo struct {
	Version string `json:"version"`
	Host    string `json:"host"`
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
	NOAUTH  AuthType = "NOAUTR"
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

type Field struct {
	TableName     string        `json:"table"`
	OriginType    ParameterType `json:"originType"`
	OriginField   string        `json:"originField"`
	ServiceType   ParameterType `json:"serviceType"`
	ServiceField  string        `json:"serviceField"`
	ParameterInfo *ApiParameter `json:"parameterInfo,omitempty"`
}

type ApiParameter struct {
	Name        string        `json:"name"`
	Type        ParameterType `json:"type"`
	Operator    Operator      `json:"operator"`
	Example     string        `json:"example"`
	Description string        `json:"description"`
	Required    bool          `json:"required"`
}

type ParameterType string

const (
	Int    ParameterType = "int"
	Bool   ParameterType = "bool"
	Float  ParameterType = "float"
	String ParameterType = "string"
)

func (f Field) Validate() error {
	for k, v := range map[string]string{
		"origin field":  f.OriginField,
		"service field": f.ServiceField,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	for k, v := range map[string]ParameterType{
		"origin type":  f.OriginType,
		"service type": f.ServiceType,
	} {
		switch v {
		case Int, Bool, Float, String:
		default:
			return fmt.Errorf("%s type is unknown: %s", k, v)
		}
	}
	return nil
}

type WebParams struct {
	Name        string        `json:"name"`     //必须
	Type        ParameterType `json:"type"`     //必须
	Location    LocationType  `json:"location"` //必须
	Required    bool          `json:"required"` //必须
	DefValue    interface{}   `json:"valueDefault"`
	Example     interface{}   `json:"example"`
	Description string        `json:"description"`
	ValidEnable int           `json:"alidEnable"`
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
	Status           Status      `json:"status"`
	AccessLink       AccessLink  `json:"access"`
	UpdatedAt        metav1.Time `json:"updatedAt"`
	ReleasedAt       metav1.Time `json:"releasedAt"`
	ApplicationCount int         `json:"applicationCount"`
	CalledCount      int         `json:"calledCount"`
	Message          string      `json:"msg"`

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
	Init     Status = "init"
	Creating Status = "creating"
	Created  Status = "created"
	Offing   Status = "offing"
	Offline  Status = "offline"
	Delete   Status = "delete"
	Deleting Status = "deleting"
	Error    Status = "error"
)

type Publish string

const (
	UnPublished Publish = "unPublished"
	Published   Publish = "published"
	//Offline       Publish = "offline"
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
