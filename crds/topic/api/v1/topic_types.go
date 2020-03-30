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

	"strings"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TopicSpec defines the desired state of Topic
type TopicSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Topic. Edit Topic_types.go to remove/update
	Name            string       `json:"name"`
	Tenant          string       `json:"tenant"`
	TopicGroup      string       `json:"topicGroup"`
	Namespace       string       `json:"namespace"`
	Partition       int          `json:"partition"`       //topic的分区数量，不指定时默认为1，指定partition大于1，则该topic的消息会被多个broker处理
	IsNonPersistent bool         `json:"isNonPersistent"` //topic是否不持久化
	Url             string       `json:"url"`             //topic url
	Permissions     []Permission `json:"permissions"`
}

type Actions []string

const (
	Consume = "consume"
	Produce = "produce"
)

type Permission struct {
	AuthUserID   string           `json:"authUserId"`   //对应clientauth的ID
	AuthUserName string           `json:"authUserName"` //对应clientauth的NAME
	Actions      Actions          `json:"actions"`      //授权的操作：发布、订阅或者发布+订阅
	Status       PermissionStatus `json:"status"`       //用户的授权状态，已授权、待删除、待授权
}

const (
	Granted = "granted" //已授权
	Grant   = "grant"
)

// TopicStatus defines the observed state of Topic
type TopicStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status  Status `json:"status"`
	Message string `json:"message"`
}

type PermissionStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status  string `json:"status"`
	Message string `json:"message"`
}

type Status string

const (
	Init              Status = "init"
	Creating          Status = "creating"
	Created           Status = "created"
	Delete            Status = "delete"
	Deleting          Status = "deleting"
	Error             Status = "error"
	Updating          Status = "updating"
	Updated           Status = "updated"
	Update            Status = "update"
	CascadingDelete   Status = "cascadingDelete"   //删除topicgroup时级联删除状态
	CascadingDeleting Status = "cascadingDeleting" //删除topicgroup时级联删除状态
	CascadingDeleted  Status = "cascadingDeleted"  //删除topicgroup时级联删除状态
)

// +kubebuilder:object:root=true

// Topic is the Schema for the topics API
type Topic struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicSpec   `json:"spec,omitempty"`
	Status TopicStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicList contains a list of Topic
type TopicList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topic `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Topic{}, &TopicList{})
}

func (in *Topic) GetUrl() (url string) {

	var build strings.Builder
	if in.Spec.IsNonPersistent {
		build.WriteString("non-persistent://")
	} else {
		build.WriteString("persistent://")
	}

	build.WriteString(in.Spec.Tenant)
	build.WriteString("/")
	build.WriteString(in.Spec.TopicGroup)
	build.WriteString("/")
	build.WriteString(in.Spec.Name)

	return build.String()
}
