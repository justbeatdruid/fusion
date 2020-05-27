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

type Status string

// ClientauthSpec defines the desired state of Clientauth
type ClientauthSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Name          string          `json:"name"`
	IssuedAt      int64           `json:"issuedAt"`
	ExipreAt      int64           `json:"expireAt"`
	Token         string          `json:"token"`
	AuthorizedMap *map[string]int `json:"authorizedMap,omitempty"` //已授权topic id列表
	Description   string          `json:"description,omitempty"`   //描述
}

// ClientauthStatus defines the observed state of Clientauth
type ClientauthStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status  Status `json:"status"`
	Message string `json:"message"`
	AuthorizationStatus Status  `json:"authorizationStatus"`
}

const (
	Init     Status = "init"
	Creating Status = "creating"
	Created  Status = "created"
	Delete   Status = "delete"
	Deleting Status = "deleting"
	Error    Status = "error"
	Updated  Status = "updated"
)

// +kubebuilder:object:root=true

// Clientauth is the Schema for the clientauths API
type Clientauth struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClientauthSpec   `json:"spec,omitempty"`
	Status ClientauthStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClientauthList contains a list of Clientauth
type ClientauthList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Clientauth `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Clientauth{}, &ClientauthList{})
}
