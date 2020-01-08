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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DatasourceSpec defines the desired state of Datasource
type DatasourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Datasource. Edit Datasource_types.go to remove/update
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Database string  `json:"database"`
	Schema   string  `json:"schema,omitempty"`
	Table    string  `json:"table"`
	Fields   []Field `json:"fields"`

	Connect ConnectInfo `json:"connect"`

	CreateUser CreateUser `json:"createUser"`
	UpdateUser UpdateUser `json:"updateUser"`
}

type Field struct {
	OriginType   ParameterType `json:"originType"`
	OriginField  string        `json:"originField"`
	ServiceType  ParameterType `json:"serviceType"`
	ServiceField string        `json:"serviceField"`
}
type CreateUser struct {
	UserId   string
	UserName string
}
type UpdateUser struct {
	UserId   string
	UserName string
}
type ParameterType string

const (
	Int    ParameterType = "int"
	Bool   ParameterType = "bool"
	Float  ParameterType = "float"
	String ParameterType = "string"
)

type ConnectInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

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

// DatasourceStatus defines the observed state of Datasource
type DatasourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status    Status    `json:"status"`
	UpdatedAt time.Time `json:"UpdatedAt"`
	CreatedAt time.Time `json:"CreatedAt"`
}
type Status string

const (
	Init     Status = "init"
	Creating Status = "creating"
	Created  Status = "created"
	Delete   Status = "delete"
	Deleting Status = "deleting"
	Error    Status = "error"
)

// +kubebuilder:object:root=true

// Datasource is the Schema for the datasources API
type Datasource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DatasourceSpec   `json:"spec,omitempty"`
	Status DatasourceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DatasourceList contains a list of Datasource
type DatasourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Datasource `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Datasource{}, &DatasourceList{})
}
