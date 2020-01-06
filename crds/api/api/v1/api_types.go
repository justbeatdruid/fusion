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
	"time"

	datav1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApiSpec defines the desired state of Api
type ApiSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name         string        `json:"name"`
	Serviceunit  Serviceunit   `json:"serviceunit"`
	Applications []Application `json:"applications"`
	Users        []User        `json:"users"`
	Frequency    int           `json:"frequency"`
	Method       Method        `json:"method"`
	Protocol     Protocol      `json:"protocol"`
	ReturnType   ReturnType    `json:"returnType"`
	Parameters   []Parameter   `json:"parameter"`
}

type Serviceunit struct {
	Name  string `json:"name"`
	Group string `json:"group"`
}

type Application struct {
	Name  string `json:"name"`
	Group string `json:"group"`
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

type Parameter struct {
	Name        string               `json:"name"`
	Type        datav1.ParameterType `json:"type"`
	Description string               `json:"description"`
	Example     string               `json:"example"`
	Required    bool                 `json:"required"`
}

// ApiStatus defines the observed state of Api
type ApiStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	UpdatedAt        time.Time `json:"updatedAt"`
	ReleasedAt       time.Time `json:"releasedAt"`
	ApplicationCount int       `json:"applicationCount"`
	CalledCount      int       `json:"calledCount"`
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
