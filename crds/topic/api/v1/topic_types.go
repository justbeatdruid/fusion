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

// TopicSpec defines the desired state of Topic
type TopicSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Topic. Edit Topic_types.go to remove/update
	Name      string `json:"name"`
	Tenant string `json:"tenant"`
	TopicNamespace string `json:"topicNamespace"`
	Namespace string `json:"namespace"`
	Partition int `json:"partition"`    //topic的分区数量，不指定时默认为1，指定partition大于1，则该topic的消息会被多个broker处理
	IsNonPersistent bool `json:"isNonPersistent""` //topic是否不持久化
}

// TopicStatus defines the observed state of Topic
type TopicStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status Status `json:"status"`
	Message string `json:"message"`
}

type Status string
const (
	Init     Status = "init"
	Creating Status = "creating"
	Created  Status = "created"
	//Delete   Status = "delete"
	//Deleting Status = "deleting"
	Error    Status = "error"
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
