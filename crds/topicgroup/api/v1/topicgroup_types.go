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

// TopicgroupSpec defines the desired state of Topicgroup
type TopicgroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Topicgroup. Edit Topicgroup_types.go to remove/update
	ID        string   `json:"id"`
	Name      string   `json:"name"` //namespace名称
	Namespace string   `json:"namespace"`
	Tenant    string   `json:"tenant"` //namespace的所属租户名称
	Url       string   `json:"url"`
	Policies  Policies `json:"policies,omitempty"`
	Available bool     `json:"available"` //资源是否可用
}
type Policies struct {
	RetentionPolicies   RetentionPolicies `json:"retentionPolicies,omitempty"` //消息保留策略
	MessageTtlInSeconds int               `json:"messageTtlInSeconds"`         //未确认消息的最长保留时长
	BacklogQuota        BacklogQuota      `json:"backlogQuota"`
	NumBundles          int               `json:"numBundles"`
}

type RetentionPolicies struct {
	RetentionTimeInMinutes int   `json:"retentionTimeInMinutes"`
	RetentionSizeInMB      int64 `json:"retentionSizeInMB"`
}

type BacklogQuota struct {
	Limit  int64  `json:"limit"`  //未确认消息的积压大小
	Policy string `json:"policy"` //producer_request_hold,producer_exception,consumer_backlog_eviction

}
type Status string

const (
	Init     Status = "init"
	Creating Status = "creating"
	Created  Status = "created"
	Delete   Status = "delete"
	Deleting Status = "deleting"
	Error    Status = "error"
	Update   Status = "update"
	Updating Status = "updating"
	Updated  Status = "updated"
)

// TopicgroupStatus defines the observed state of Topicgroup
type TopicgroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status  Status `json:"status"`
	Message string `json:"message"`
}

// +kubebuilder:object:root=true

// Topicgroup is the Schema for the topicgroups API
type Topicgroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicgroupSpec   `json:"spec,omitempty"`
	Status TopicgroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicgroupList contains a list of Topicgroup
type TopicgroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topicgroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Topicgroup{}, &TopicgroupList{})
}
