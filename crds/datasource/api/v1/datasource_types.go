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

	dwv1 "github.com/chinamobile/nlpt/crds/datasource/datawarehouse/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DatasourceSpec defines the desired state of Datasource
type DatasourceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Datasource. Edit Datasource_types.go to remove/update
	Name string `json:"name"`
	Type Type   `json:"type"`

	RDB           *RDB           `json:"rdb,omitempty"`
	DataWarehouse *dwv1.Database `json:"datawarehouse,omitempty"`

	Location string `json:"localtion"`
	AuthType string `json:"authType"`
}

type Type string

func (t Type) String() string {
	return string(t)
}

const (
	RDBType           Type = "rdb"
	DataWarehouseType Type = "datawarehouse"
)

type RDB struct {
	Type     string `json:"type"`
	Database string `json:"database"`
	Schema   string `json:"schema,omitempty"`

	Connect ConnectInfo `json:"connect"`

	CreateUser User `json:"createUser"`
	UpdateUser User `json:"updateUser"`
}

func (r *RDB) Validate() error {
	if r == nil {
		return fmt.Errorf("rdb is null")
	}
	for k, v := range map[string]string{
		"database": r.Database,
		"host":     r.Connect.Host,
		"username": r.Connect.Username,
		"password": r.Connect.Password,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	if r.Connect.Port < 0 || r.Connect.Port > 65536 {
		return fmt.Errorf("invalid port %d", r.Connect.Port)
	}
	return nil
}

type User struct {
	UserId   string `json:"id"`
	UserName string `json:"name"`
}

type ConnectInfo struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// DatasourceStatus defines the observed state of Datasource
type DatasourceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status    Status      `json:"status"`
	UpdatedAt metav1.Time `json:"updatedAt"`
	CreatedAt metav1.Time `json:"createdAt"`
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
