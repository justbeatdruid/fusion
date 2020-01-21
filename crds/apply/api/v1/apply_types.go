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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ApplySpec defines the desired state of Apply
type ApplySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Apply. Edit Apply_types.go to remove/update
	Name       string     `json:"name,omitempty"`
	TargetType TargetType `json:"targetType"`
	TargetID   string     `json:"targetID"`
	TargetName string     `json:"targetName"`
	AppID      string     `json:"appID"`
	AppName    string     `json:"appName"`
	ExpireAt   time.Time  `json:"expireAt"`
}

type TargetType string

const (
	Serviceunit TargetType = "serviceunit"
	Api         TargetType = "api"
)

// ApplyStatus defines the observed state of Apply
type ApplyStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status Status `json:"status"`
	Reason string `json:"reason"`
}

type Status string

const (
	Waiting  Status = "waiting"
	Admited  Status = "admitted"
	Finished Status = "finished"
	Denied   Status = "denied"
	Error    Status = "error"
)

// +kubebuilder:object:root=true

// Apply is the Schema for the applies API
type Apply struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplySpec   `json:"spec,omitempty"`
	Status ApplyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplyList contains a list of Apply
type ApplyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Apply `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Apply{}, &ApplyList{})
}
