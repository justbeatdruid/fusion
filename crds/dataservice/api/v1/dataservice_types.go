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

// DataserviceSpec defines the desired state of Dataservice
type DataserviceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// read data from
	Source string `json:"source"`
	// write data to
	Target string `json:"target"`
	Task   Task   `json:"task"`
}

type Task struct {
	// type of task, realtime or periodic
	Type         TaskType      `json:"type"`
	RealtimeTask *RealtimeTask `json:"realtimeTask,omitempty"`
	PeriodicTask *PeriodicTask `json:"periodicTask,omitempty"`
}

type TaskType string

type RealtimeTask struct {
	Incremental bool `json:"incremental"`
}

type PeriodicTask struct {
	Minute TimeConfig `json:"minute"`
	Hour   TimeConfig `json:"hour"`
	Date   TimeConfig `json:"date"`
	Month  TimeConfig `json:"mounth"`
	Day    TimeConfig `json:"day"`
}

// in crontab format
// 3: Value
// 2,3,4: Values
// 3-12: StartAt 3, StopAt 12
// /3: Every 3
// *: Always
type TimeConfig struct {
	Value   int   `json:"value"`
	Values  []int `json:"values"`
	StartAt int   `json:"startAt"`
	StopAt  int   `json:"stopAt"`
	Every   int   `json:"every"`
	Always  bool  `json:"always"`
}

// DataserviceStatus defines the observed state of Dataservice
type DataserviceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status    Status      `json:"status"`
	StartedAt metav1.Time `json:"startedAt"`
	StoppedAt metav1.Time `json:"stoppedAt"`
}

type Status string

// +kubebuilder:object:root=true

// Dataservice is the Schema for the dataservices API
type Dataservice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DataserviceSpec   `json:"spec,omitempty"`
	Status DataserviceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// DataserviceList contains a list of Dataservice
type DataserviceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Dataservice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Dataservice{}, &DataserviceList{})
}
