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

// NamespaceSpec defines the desired state of Namespace
type NamespaceSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ID        string   `json:"id"`
	Name      string   `json:"name"` //namespace名称
	Namespace string   `json:"namespace"`
	Tenant    string   `json:"tenant"` //namespace的所属租户名称
	Url       string   `json:"url"`
	Policies  Policies `json:"policies"`
}

type Policies struct {
	AuthPolicies                          AuthPolicies             `json:"auth_policies"`
	ReplicationClusters                   []string                 `json:"replication_clusters"`
	Bundles                               BundlesData              `json:"bundles"`
	BacklogQuotaMap                       map[string]BacklogQuota  `json:"backlog_quota_map"`
	TopicDispatchRate                     map[string]DispatchRate  `json:"topicDispatchRate"`
	SubscriptionDispatchRate              map[string]DispatchRate  `json:"subscriptionDispatchRate"`
	ReplicatorDispatchRate                map[string]DispatchRate  `json:"replicatorDispatchRate"`
	ClusterSubscribeRate                  map[string]SubscribeRate `json:"clusterSubscribeRate"`
	Persistence                           PersistencePolicies      `json:"persistence"`
	DeduplicationEnabled                  bool                     `json:"deduplicationEnabled"`
	LatencyStatsSampleRate                map[string]int           `json:"latency_stats_sample_rate"`
	MessageTtlInSeconds                   int                      `json:"message_ttl_in_seconds"`
	RetentionPolicies                     RetentionPolicies        `json:"retention_policies"`
	Deleted                               bool                     `json:"deleted"`
	AntiAffinityGroup                     string                   `json:"antiAffinityGroup"`
	EncryptionRequired                    bool                     `json:"encryption_required"`
	SubscriptionAuthMode                  string                   `json:"subscription_auth_mode"` //None, Prefix
	MaxProducersPerTopic                  int                      `json:"max_producers_per_topic"`
	MaxConsumersPerTopic                  int                      `json:"max_consumers_per_topic"`
	MaxConsumersPerSubscription           int                      `json:"max_consumers_per_subscription"`
	CompactionThreshold                   int64                    `json:"compaction_threshold"`
	OffloadThreshold                      int64                    `json:"offload_threshold"`
	OffloadDeletionLagMS                  int64                    `json:"offload_deletion_lag_ms"`
	SchemaAutoUpdateCompatibilityStrategy string                   `json:"schema_auto_update_compatibility_strategy"`
	SchemaValidationEnforced              bool                     `json:"schema_validation_enforced"`
}

type AuthPolicies struct {
	NamespaceAuth         map[string][]string            `json:"namespace_auth"`
	DestinationAuth       map[string]map[string][]string `json:"destination_auth"`
	SubscriptionAuthRoles map[string][]string            `json:"subscription_auth_roles"`
}

type BundlesData struct {
	Boundaries []string `json:"boundaries"`
	NumBundles int      `json:"numBundles"`
}

type BacklogQuota struct {
	Limit  int64  `json:"limit"`
	Policy string `json:"policy"` //producer_request_hold,producer_exception,consumer_backlog_eviction
}

type DispatchRate struct {
	DispatchThrottlingRateInMsg  int   `json:"dispatchThrottlingRateInMsg"`
	DispatchThrottlingRateInByte int64 `json:"dispatchThrottlingRateInByte"`
	RatePeriodInSecond           int   `json:"ratePeriodInSecond"`
}

type SubscribeRate struct {
	SubscribeThrottlingRatePerConsumer int `json:"subscribeThrottlingRatePerConsumer"`
	RatePeriodInSecond                 int `json:"ratePeriodInSecond"`
}
type PersistencePolicies struct {
	BookkeeperEnsemble             int     `json:"bookkeeperEnsemble"`
	BookkeeperWriteQuorum          int     `json:"bookkeeperWriteQuorum"`
	BookkeeperAckQuorum            int     `json:"bookkeeperAckQuorum"`
	ManagedLedgerMaxMarkDeleteRate float64 `json:"managedLedgerMaxMarkDeleteRate"`
}

type RetentionPolicies struct {
	RetentionTimeInMinutes int `json:"retentionTimeInMinutes"`
	RetentionSizeInMB      int `json:"retentionSizeInMB"`
}
type Status string

const (
	Init     Status = "init"
	Creating Status = "creating"
	Created  Status = "created"
	Delete   Status = "delete"
	//Deleting Status = "deleting"
	Error Status = "error"
)

// NamespaceStatus defines the observed state of Namespace
type NamespaceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	Status  Status `json:"status"`
	Message string `json:"message"`
}

// +kubebuilder:object:root=true

// Namespace is the Schema for the namespaces API
type Namespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NamespaceSpec   `json:"spec,omitempty"`
	Status NamespaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NamespaceList contains a list of Namespace
type NamespaceList struct {
	metav1.TypeMeta
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Namespace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Namespace{}, &NamespaceList{})
}
