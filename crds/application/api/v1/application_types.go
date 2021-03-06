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
const (
	GroupLabel = "nlpt.cmcc.com/group"
)

// ApplicationSpec defines the desired state of Application
type ApplicationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Application. Edit Application_types.go to remove/update
	Name            string       `json:"name"`
	Description     string       `json:"description"`
	Group           Group        `json:"group"`
	AccessKey       string       `json:"accessKey"`
	AccessSecretKey string       `json:"accessSecretKey"`
	APIs            []Api        `json:"apis"`
	ConsumerInfo    ConsumerInfo `json:"comsumer"`
	TopicAuth       TopicAuth    `json:"topicAuth"`
	Result          Result       `json:"result"`
	DisplayStatus   DisStatus    `json:"disStatus"`
}

type Api struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ConsumerInfo struct {
	ConsumerID string `json:"id"`
	Key        string `json:"key"`
	Secret     string `json:"secret"`
	Token      string `json:"jwt"`
}

type TopicAuth struct {
	Token string `json:"jwt"`
}

// ApplicationStatus defines the observed state of Application
type ApplicationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status  Status `json:"status"`
	Message string `json:"msg"`
}

type Status string

const (
	Init    Status = "init"
	Created Status = "created"
	Delete  Status = "delete"
	Error   Status = "error"
)

type Result string

const (
	CREATING      Result = "creating"
	CREATESUCCESS Result = "createSuccess"
	CREATEFAILED  Result = "createFailed"
	UPDATESUCCESS Result = "updateSuccess"
	UPDATEFAILED  Result = "updateFailed"
	DELETING      Result = "deleting"
	DELETEFAILED  Result = "deleteFailed"
)

type DisStatus string

const (
	SuCreating    DisStatus = "创建中"
	CreateSuccess DisStatus = "创建成功"
	CreateFailed  DisStatus = "创建失败"
	UpdateSuccess DisStatus = "更新成功"
	UpdateFailed  DisStatus = "更新失败"
	DeleteFailed  DisStatus = "删除失败"
)

// +kubebuilder:object:root=true

// Application is the Schema for the applications API
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec,omitempty"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Application{}, &ApplicationList{})
}
