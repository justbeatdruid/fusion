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
	"time"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RestrictionSpec defines the desired state of Restriction
type RestrictionSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ID          string     `json:"ID"`
	Name        string     `json:"name"`
	Type        LimitType  `json:"type"`
	Action      Action     `json:"action"`
	Config      ConfigInfo `json:"config"`
	User        string     `json:"user"`
	Apis        []Api      `json:"apis"`
	Description string     `json:"description"`
	//KongInfo
	//KongService KongServiceInfo `json:"kongServiceInfo"`
}

type ConfigInfo struct {
	Ip   string `json:"ip"`
	User string `json:"user"`
}

type Action string

const (
	WHITE Action = "white"
	BLACK Action = "black"
)

type LimitType string

const (
	IP   LimitType = "ip"
	USER LimitType = "user"
)

type Api struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	KongID string `json:"kongID"`
	//ip 限制 kong插件id
	PluginID string `json:"pluginID"`
	//特殊应用 记录的kong插件id列表
	Result Result `json:"result"`
	Detail string `json:"detail"`
}
type Result string

const (
	INIT    Result = "init"
	FAILED  Result = "failed"
	SUCCESS Result = "success"
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
type RestrictionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status    Status    `json:"status"`
	UpdatedAt time.Time `json:"time.Time"`
	APICount  int       `json:"apiCount"`
	Published bool      `json:"published"`
	Message   string    `json:"message"`
}

// +kubebuilder:object:root=true

// Restriction is the Schema for the restrictions API
type Restriction struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RestrictionSpec   `json:"spec,omitempty"`
	Status RestrictionStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RestrictionList contains a list of Restriction
type RestrictionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Restriction `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Restriction{}, &RestrictionList{})
}
