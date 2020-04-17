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
	//datav1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const GroupLabel = "nlpt.cmcc.com/group"

// ServiceunitSpec defines the desired state of Serviceunit
type ServiceunitSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name         string      `json:"name"`
	Group        Group       `json:"group"`
	Type         ServiceType `json:"type"`
	DatasourceID *Datasource `json:"datasourcesID,omitempty"`
	APIs         []Api       `json:"apis"`
	Description  string      `json:"description"`
	//KongInfo
	KongService   KongServiceInfo `json:"kongServiceInfo"`
	Result        Result          `json:"result"`
	DisplayStatus DisStatus       `json:"disStatus"`
}

type KongServiceInfo struct {
	Host     string `json:"host"`
	ID       string `json:"id"`
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
	Path     string `json:"path"`
	TimeOut  int    `json:"connect_timeout"`
	WirteOut int    `json:"write_timeout"`
	ReadOut  int    `json:"read_timeout"`
}

type Datasource struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	GroupID *string `json:"group,omitempty"`

	// never stored in crds
	DataWarehouse *DataWarehouse `json:"dataWarehouse,omitempty"`
}

type DataWarehouse struct {
	DatabaseName        string `json:"databaseName"`
	DatabaseDisplayName string `json:"databaseDisplayName"`
	SubjectName         string `json:"subjectName"`
	SubjectDisplayName  string `json:"subjectDisplayName"`
}

type Group struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Api struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type ServiceType string

const (
	DataService ServiceType = "data"
	WebService  ServiceType = "web"
)

// ServiceunitStatus defines the observed state of Serviceunit
type ServiceunitStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status    Status      `json:"status"`
	UpdatedAt metav1.Time `json:"time"`
	APICount  int         `json:"apiCount"`
	Published bool        `json:"published"`
	Message   string      `json:"message"`
}

type Result string

const (
	CREATING     Result = "creating"
	CREATESUCCESS Result = "createSuccess"
	CREATEFAILED Result = "createFailed"
	UPDATING     Result = "updating"
	UPDATESUCCESS Result = "updateSuccess"
	UPDATEFAILED Result = "updateFailed"
	DELETING	Result = "deleting"
	DELETEFAILED Result = "deleteFailed"
)

type DisStatus string

const (
	SuCreating    DisStatus = "创建中"
	CreateSuccess DisStatus = "创建成功"
	CreateFailed  DisStatus = "创建失败"
	SuUpdating	  DisStatus = "跟新中"
	UpdateSuccess DisStatus = "跟新成功"
	UpdateFailed  DisStatus = "跟新失败"
	DeleteFailed  DisStatus = "删除失败"
)

type Status string

const (
	Init     Status = "init"
	Creating Status = "creating"
	Created  Status = "created"
	Delete   Status = "delete"
	Deleting Status = "deleting"
	Error    Status = "error"
	// + update_sunyu
	Update   Status = "update"
	Updating Status = "updating"
	Updated  Status = "updated"
)

// +kubebuilder:object:root=true

// Serviceunit is the Schema for the serviceunits API
type Serviceunit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceunitSpec   `json:"spec,omitempty"`
	Status ServiceunitStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ServiceunitList contains a list of Serviceunit
type ServiceunitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Serviceunit `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Serviceunit{}, &ServiceunitList{})
}
