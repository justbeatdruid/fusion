package service

import (
	"fmt"
	pulsar "github.com/chinamobile/nlpt/apiserver/resources/topicgroup/pulsar"
	"github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/names"
	"strconv"
)

const (
	producerRequestHold     = "producer_request_hold"
	producerException       = "producer_exception"
	consumerBacklogEviction = "consumer_backlog_eviction"
	NotSet                  = -12323344
	NotSetString            = "NotSet"
)

type Topicgroup struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"` //Topic分组名称
	Namespace string     `json:"namespace"`
	Policies  *Policies  `json:"policies,omitempty"` //Topic分组的策略
	CreatedAt int64      `json:"createdAt"`          //创建时间
	Users     user.Users `json:"users"`
	Status    v1.Status  `json:"status"`
	Message   string     `json:"message"`
	Available bool       `json:"available"` //是否可用
}

type Policies struct {
	RetentionPolicies           *RetentionPolicies        `json:"retentionPolicies,omitempty"` //消息保留策略
	MessageTtlInSeconds         *int                      `json:"messageTtlInSeconds"`         //未确认消息的最长保留时长
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

type AuthPolices struct {
	NamespaceAuth         NamespaceAuth         `json:"namespace_auth"`
	DestinationAuth       DestinationAuth       `json:"destination_auth"`
	SubscriptionAuthRoles SubscriptionAuthRoles `json:"subscription_auth_roles"`
}

type NamespaceAuth struct {
	//TODO 待补充
}

type DestinationAuth struct {
	//TODO 待补充
}

type SubscriptionAuthRoles struct {
	//TODO 待补充
}

type RetentionPolicies struct {
	RetentionTimeInMinutes int   `json:"retentionTimeInMinutes"`
	RetentionSizeInMB      int64 `json:"retentionSizeInMB"`
}

type BacklogQuota struct {
	Limit  int64  `json:"limit"`  //未确认消息的积压大小
	Policy string `json:"policy"` //producer_request_hold,producer_exception,consumer_backlog_eviction

}

func NewPolicies(fillDefaultValue bool) *Policies {
	bMap := make(map[string]BacklogQuota)
	var bQuota BacklogQuota
	bQuota.Policy = NotSetString
	bQuota.Limit = NotSet
	bMap["destination_storage"] = bQuota
	return &Policies{
		RetentionPolicies: &RetentionPolicies{
			RetentionSizeInMB:      NotSet,
			RetentionTimeInMinutes: NotSet,
		},
		//MessageTtlInSeconds: nil,

		BacklogQuota: &bMap,
		Bundles: &Bundles{
			Boundaries: nil,
			NumBundles: NotSet,
		},
	}

}

// only used in creation options
func ToAPI(app *Topicgroup) *v1.Topicgroup {
	crd := &v1.Topicgroup{}
	crd.TypeMeta.Kind = "Topicgroup"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = app.Namespace

	crd.Spec = v1.TopicgroupSpec{
		Name:     app.Name,
		Policies: ToPolicesApi(app.Policies),
	}

	status := app.Status
	if len(status) == 0 {
		status = v1.Init
	}
	crd.Status = v1.TopicgroupStatus{
		Status:  status,
		Message: app.Message,
	}

	crd.ObjectMeta.Labels = user.AddUsersLabels(app.Users, crd.ObjectMeta.Labels)
	return crd
}

func ToModel(obj *v1.Topicgroup) *Topicgroup {
	return &Topicgroup{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,
		Status:    obj.Status.Status,
		Message:   obj.Status.Message,
		CreatedAt: obj.ObjectMeta.CreationTimestamp.Unix(),
		Policies:  ToPolicesModel(obj.Spec.Policies),
		Users:     user.GetUsersFromLabels(obj.ObjectMeta.Labels),
		Available: obj.Spec.Available,
	}
}

func ToPolicesModel(obj *v1.Policies) *Policies {
	cRate := make(map[string]SubscribeRate)
	for k, v := range *obj.ClusterSubscribeRate {
		cRate[k] = SubscribeRate{
			SubscribeThrottlingRatePerConsumer: v.SubscribeThrottlingRatePerConsumer,
			RatePeriodInSecond:                 v.RatePeriodInSecond,
		}
	}

	sRate := make(map[string]DispatchRate)
	for k, v := range *obj.SubscriptionDispatchRate {
		sRate[k] = DispatchRate{
			DispatchThrottlingRateInMsg:  v.DispatchThrottlingRateInMsg,
			DispatchThrottlingRateInByte: v.DispatchThrottlingRateInByte,
			RelativeToPublishRate:        v.RelativeToPublishRate,
			RatePeriodInSecond:           v.RatePeriodInSecond,
		}
	}

	tRate := make(map[string]DispatchRate)
	for k, v := range *obj.TopicDispatchRate {
		sRate[k] = DispatchRate{
			DispatchThrottlingRateInMsg:  v.DispatchThrottlingRateInMsg,
			DispatchThrottlingRateInByte: v.DispatchThrottlingRateInByte,
			RelativeToPublishRate:        v.RelativeToPublishRate,
			RatePeriodInSecond:           v.RatePeriodInSecond,
		}
	}

	bMap := make(map[string]BacklogQuota)
	var backlogQuota BacklogQuota

	var objBacklogQuota = *obj.BacklogQuota
	backlogQuota.Limit = objBacklogQuota["destination_storage"].Limit
	backlogQuota.Policy = objBacklogQuota["destination_storage"].Policy
	bMap["destination_storage"] = backlogQuota

	//kubernetes不支持float64类型，因此存储的时候转成string，展示的时候再转成float64
	managedLedgerMaxMarkDeleteRate, _ := strconv.ParseFloat(obj.Persistence.ManagedLedgerMaxMarkDeleteRate, 64)

	return &Policies{
		RetentionPolicies: &RetentionPolicies{
			RetentionSizeInMB:      obj.RetentionPolicies.RetentionSizeInMB,
			RetentionTimeInMinutes: obj.RetentionPolicies.RetentionTimeInMinutes,
		},
		MessageTtlInSeconds: obj.MessageTtlInSeconds,
		BacklogQuota:        &bMap,
		Bundles: &Bundles{
			Boundaries: obj.Bundles.Boundaries,
			NumBundles: obj.Bundles.NumBundles,
		},
		SchemaCompatibilityStrategy: obj.SchemaCompatibilityStrategy,
		IsAllowAutoUpdateSchema:     obj.IsAllowAutoUpdateSchema,
		SchemaValidationEnforced:    obj.SchemaValidationEnforced,
		MaxConsumersPerSubscription: obj.MaxConsumersPerSubscription,
		MaxConsumersPerTopic:        obj.MaxConsumersPerTopic,
		MaxProducersPerTopic:        obj.MaxProducersPerTopic,
		CompactionThreshold:         obj.CompactionThreshold,
		OffloadDeletionLagMs:        obj.OffloadDeletionLagMs,
		OffloadThreshold:            obj.OffloadThreshold,
		SubscriptionAuthMode:        obj.SubscriptionAuthMode,
		EncryptionRequired:          obj.EncryptionRequired,
		Persistence: &PersistencePolicies{
			BookkeeperEnsemble:             obj.Persistence.BookkeeperEnsemble,
			BookkeeperWriteQuorum:          obj.Persistence.BookkeeperWriteQuorum,
			BookkeeperAckQuorum:            obj.Persistence.BookkeeperAckQuorum,
			ManagedLedgerMaxMarkDeleteRate: managedLedgerMaxMarkDeleteRate,
		},

		SubscriptionDispatchRate: &sRate,
		ClusterSubscribeRate:     &cRate,
		TopicDispatchRate:        &tRate,
		DeduplicationEnabled:     obj.DeduplicationEnabled,
	}
}
func ToPolicesApi(policies *Policies) *v1.Policies {
	if policies == nil {
		return nil
	}

	cRate := make(map[string]v1.SubscribeRate)
	for k, v := range *policies.ClusterSubscribeRate {
		cRate[k] = v1.SubscribeRate{
			SubscribeThrottlingRatePerConsumer: v.SubscribeThrottlingRatePerConsumer,
			RatePeriodInSecond:                 v.RatePeriodInSecond,
		}
	}

	sRate := make(map[string]v1.DispatchRate)
	for k, v := range *policies.SubscriptionDispatchRate {
		sRate[k] = v1.DispatchRate{
			DispatchThrottlingRateInMsg:  v.DispatchThrottlingRateInMsg,
			DispatchThrottlingRateInByte: v.DispatchThrottlingRateInByte,
			RelativeToPublishRate:        v.RelativeToPublishRate,
			RatePeriodInSecond:           v.RatePeriodInSecond,
		}
	}

	tRate := make(map[string]v1.DispatchRate)
	for k, v := range *policies.TopicDispatchRate {
		sRate[k] = v1.DispatchRate{
			DispatchThrottlingRateInMsg:  v.DispatchThrottlingRateInMsg,
			DispatchThrottlingRateInByte: v.DispatchThrottlingRateInByte,
			RelativeToPublishRate:        v.RelativeToPublishRate,
			RatePeriodInSecond:           v.RatePeriodInSecond,
		}
	}

	bMap := make(map[string]v1.BacklogQuota)

	if policies.BacklogQuota != nil {
		var backlopQuota v1.BacklogQuota
		backlopQuota.Limit = (*policies.BacklogQuota)["destination_storage"].Limit
		backlopQuota.Policy = (*policies.BacklogQuota)["destination_storage"].Policy
		bMap["destination_storage"] = backlopQuota
	}


	for k, v := range *policies.TopicDispatchRate {
		sRate[k] = v1.DispatchRate{
			DispatchThrottlingRateInMsg:  v.DispatchThrottlingRateInMsg,
			DispatchThrottlingRateInByte: v.DispatchThrottlingRateInByte,
			RelativeToPublishRate:        v.RelativeToPublishRate,
			RatePeriodInSecond:           v.RatePeriodInSecond,
		}
	}
	crd := v1.Policies{
		Bundles: &v1.Bundles{
			Boundaries: policies.Bundles.Boundaries,
			NumBundles: policies.Bundles.NumBundles,
		},
		MessageTtlInSeconds: policies.MessageTtlInSeconds,
		RetentionPolicies: &v1.RetentionPolicies{
			RetentionTimeInMinutes: policies.RetentionPolicies.RetentionTimeInMinutes,
			RetentionSizeInMB:      policies.RetentionPolicies.RetentionSizeInMB,
		},
		BacklogQuota:                &bMap,
		SchemaCompatibilityStrategy: policies.SchemaCompatibilityStrategy,
		IsAllowAutoUpdateSchema:     policies.IsAllowAutoUpdateSchema,
		SchemaValidationEnforced:    policies.SchemaValidationEnforced,
		MaxConsumersPerSubscription: policies.MaxConsumersPerSubscription,
		MaxConsumersPerTopic:        policies.MaxConsumersPerTopic,
		MaxProducersPerTopic:        policies.MaxProducersPerTopic,
		CompactionThreshold:         policies.CompactionThreshold,
		OffloadDeletionLagMs:        policies.OffloadDeletionLagMs,
		OffloadThreshold:            policies.OffloadThreshold,
		SubscriptionAuthMode:        policies.SubscriptionAuthMode,
		EncryptionRequired:          policies.EncryptionRequired,
		Persistence: &v1.PersistencePolicies{
			BookkeeperEnsemble:             policies.Persistence.BookkeeperEnsemble,
			BookkeeperWriteQuorum:          policies.Persistence.BookkeeperWriteQuorum,
			BookkeeperAckQuorum:            policies.Persistence.BookkeeperAckQuorum,
			ManagedLedgerMaxMarkDeleteRate: strconv.FormatFloat(policies.Persistence.ManagedLedgerMaxMarkDeleteRate, 'f', -1, 64)},
		SubscriptionDispatchRate: &sRate,
		ClusterSubscribeRate:     &cRate,
		TopicDispatchRate:        &tRate,
		DeduplicationEnabled:     policies.DeduplicationEnabled,
	}

	return &crd
}

func ToListModel(items *v1.TopicgroupList) []*Topicgroup {
	var app = make([]*Topicgroup, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (p *Policies) Validate() error {
	//TODO 参数校验待验证
	if p != nil {
		if *p.MessageTtlInSeconds < pulsar.DefaultMessageTTlInSeconds && *p.MessageTtlInSeconds != NotSet {
			return fmt.Errorf("messageTtlInSeconds is invalid: %d", p.MessageTtlInSeconds)
		}

		if p.RetentionPolicies.RetentionTimeInMinutes < pulsar.MinRetentionTimeInMinutes && p.RetentionPolicies.RetentionTimeInMinutes != NotSet {
			return fmt.Errorf("retentionTimeInMinutes is invalid: %d", p.RetentionPolicies.RetentionTimeInMinutes)
		}

		if p.RetentionPolicies.RetentionSizeInMB < pulsar.MinRetentionSizeInMB && p.RetentionPolicies.RetentionTimeInMinutes != NotSet {
			return fmt.Errorf("retentionTimeInMinutes is invalid: %d", p.RetentionPolicies.RetentionSizeInMB)
		}

		if p.Bundles.NumBundles <= 0 && p.Bundles.NumBundles != NotSet {
			return fmt.Errorf("numBundles is invalid: %d", p.Bundles.NumBundles)
		}

		if p.BacklogQuota != nil {
			var backlogQuota = *p.BacklogQuota
			switch backlogQuota["destination_storage"].Policy {
			case producerRequestHold:
			case consumerBacklogEviction:
			case producerException:
				break
			default:
				if backlogQuota["destination_storage"].Policy != NotSetString {
					return fmt.Errorf("backlogQuota policy is invalid: %s", backlogQuota["destination_storage"].Policy)
				}
			}
		}
	}
	return nil

}
func (a *Topicgroup) Validate() error {
	for k, v := range map[string]string{
		"name": a.Name,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}

	p := a.Policies
	//TODO 参数校验待验证
	if err := p.Validate(); err != nil {
		return err
	}
	a.ID = names.NewID()
	return nil
}
