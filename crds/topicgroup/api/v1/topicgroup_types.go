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

// TopicgroupSpec defines the desired state of Topicgroup
type TopicgroupSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Name      string    `json:"name"` //namespace名称
	Policies  *Policies `json:"policies,omitempty"`
	Available bool      `json:"available"` //资源是否可用
}
type Policies struct {
	RetentionPolicies           *RetentionPolicies        `json:"retention_policies,omitempty"` //消息保留策略
	MessageTtlInSeconds         *int                      `json:"message_ttl_in_seconds"`       //未确认消息的最长保留时长
	BacklogQuota                *map[string]BacklogQuota  `json:"backlog_quota_map"`
	Bundles                     *Bundles                  `json:"bundles"` //key:destination_storage
	TopicDispatchRate           *map[string]DispatchRate  `json:"topicDispatchRate"`
	SubscriptionDispatchRate    *map[string]DispatchRate  `json:"subscriptionDispatchRate"`
	ClusterSubscribeRate        *map[string]SubscribeRate `json:"clusterSubscribeRate"`
	Persistence                 *PersistencePolicies      `json:"persistence"` //Configuration of bookkeeper persistence policies.
	DeduplicationEnabled        *bool                     `json:"deduplicationEnabled"`
	EncryptionRequired          *bool                     `json:"encryption_required"`
	SubscriptionAuthMode        *string                   `json:"subscription_auth_mode"` //None/Prefix
	MaxProducersPerTopic        *int                      `json:"max_producers_per_topic"`
	MaxConsumersPerTopic        *int                      `json:"max_consumers_per_topic"`
	MaxConsumersPerSubscription *int                      `json:"max_consumers_per_subscription"`
	CompactionThreshold         *int64                    `json:"compaction_threshold"`
	OffloadThreshold            *int64                    `json:"offload_threshold"`
	OffloadDeletionLagMs        *int64                    `json:"offload_deletion_lag_ms"`
	IsAllowAutoUpdateSchema     *bool                     `json:"is_allow_auto_update_schema"`
	SchemaValidationEnforced    *bool                     `json:"schema_validation_enforced"`
	SchemaCompatibilityStrategy *string                   `json:"schema_compatibility_strategy"`
}
type Bundles struct {
	Boundaries []string `json:"boundaries"`
	NumBundles int      `json:"numBundles"`
}
type SubscribeRate struct {
	SubscribeThrottlingRatePerConsumer int `json:"subscribeThrottlingRatePerConsumer"` //默认-1
	RatePeriodInSecond                 int `json:"ratePeriodInSecond"`                 //默认30
}
type PersistencePolicies struct {
	BookkeeperEnsemble             int     `json:"bookkeeperEnsemble"`
	BookkeeperWriteQuorum          int     `json:"bookkeeperWriteQuorum"`
	BookkeeperAckQuorum            int     `json:"bookkeeperAckQuorum"`
	ManagedLedgerMaxMarkDeleteRate float64 `json:"managedLedgerMaxMarkDeleteRate"`
}
type DispatchRate struct {
	DispatchThrottlingRateInMsg  int   `json:"dispatchThrottlingRateInMsg"`  //默认：-1
	DispatchThrottlingRateInByte int64 `json:"dispatchThrottlingRateInByte"` //默认：-1
	RelativeToPublishRate        bool  `json:"relativeToPublishRate"`        /* throttles dispatch relatively publish-rate */
	RatePeriodInSecond           int   `json:"ratePeriodInSecond"`           /* by default dispatch-rate will be calculate per 1 second */

}
type RetentionPolicies struct {
	RetentionTimeInMinutes int   `json:"retentionTimeInMinutes"`
	RetentionSizeInMB      int64 `json:"retentionSizeInMB"`
}

type BacklogQuota struct {
	Limit  int64  `json:"limit"`  //未确认消息的积压大小
	Policy string `json:"policy"` //producer_request_hold,producer_exception,consumer_backlog_eviction

}
type Status string

const (
	Init     Status = "init"
	Creating Status = "creating"
	Created  Status = "created"
	Delete   Status = "delete"
	Deleting Status = "deleting"
	Error    Status = "error"
	Update   Status = "update"
	Updating Status = "updating"
	Updated  Status = "updated"
)

// TopicgroupStatus defines the observed state of Topicgroup
type TopicgroupStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status  Status `json:"status"`
	Message string `json:"message"`
}

// +kubebuilder:object:root=true

// Topicgroup is the Schema for the topicgroups API
type Topicgroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TopicgroupSpec   `json:"spec,omitempty"`
	Status TopicgroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TopicgroupList contains a list of Topicgroup
type TopicgroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Topicgroup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Topicgroup{}, &TopicgroupList{})
}
