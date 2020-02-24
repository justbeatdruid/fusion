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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
type Api struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	//app api ip user 类型记录的kong插件id
	TrafficID string `json:"trafficID"`
	//特殊应用 记录的kong插件id列表
	SpecialID []string `json:"specialID"`
	KongID    string   `json:"kongID"`
	Result    Result   `json:"result"`
	Detail    string   `json:"detail"`
}
type Result string

const (
	INIT    Result = "init"
	FAILED  Result = "failed"
	SUCCESS Result = "success"
)

type ConfigInfo struct {
	Year    int       `json:"year"`
	Month   int       `json:"month"`
	Day     int       `json:"day"`
	Hour    int       `json:"hour"`
	Minute  int       `json:"minute"`
	Second  int       `json:"second"`
	Special []Special `json:"special"`
}

const MAXNUM int = 5

type Special struct {
	ID     string `json:"id"`
	Minute int    `json:"minute"`
}

// TrafficcontrolSpec defines the desired state of Trafficcontrol
type TrafficcontrolSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ID          string     `json:"ID"`
	Name        string     `json:"name"`
	Type        LimitType  `json:"type"`
	Config      ConfigInfo `json:"config"`
	User        string     `json:"user"`
	Apis        []Api      `json:"apis"`
	Description string     `json:"description"`
	//KongInfo
	//KongService KongServiceInfo `json:"kongServiceInfo"`
}

type LimitType string

const (
	APIC     LimitType = "api"
	IPC      LimitType = "ip"
	APPC     LimitType = "app"
	USERC    LimitType = "user"
	SPECAPPC LimitType = "specialapp"
)

type Status string

const (
	Init      Status = "init"
	Bind      Status = "bind"      //绑定api
	Binding   Status = "binding"   //启动绑定api
	Binded    Status = "binded"    //绑定结束
	UnBind    Status = "unbind"    //绑定api
	UnBinding Status = "unbinding" //启动绑定api
	UnBinded  Status = "unbinded"  //绑定结束
	Delete    Status = "delete"
	Deleting  Status = "deleting"
	Error     Status = "error"
	Update    Status = "update" //更新流程策略时更新插件
	Updating  Status = "updating"
	Updated   Status = "updated"
)

// TrafficcontrolStatus defines the observed state of Trafficcontrol
type TrafficcontrolStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status    Status      `json:"status"`
	UpdatedAt metav1.Time `json:"time.Time"`
	APICount  int         `json:"apiCount"`
	Published bool        `json:"published"`
	Message   string      `json:"message"`
}

// +kubebuilder:object:root=true

// Trafficcontrol is the Schema for the trafficcontrols API
type Trafficcontrol struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TrafficcontrolSpec   `json:"spec,omitempty"`
	Status TrafficcontrolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TrafficcontrolList contains a list of Trafficcontrol
type TrafficcontrolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Trafficcontrol `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Trafficcontrol{}, &TrafficcontrolList{})
}
