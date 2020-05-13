package service

import (
	"fmt"
	pulsar "github.com/chinamobile/nlpt/apiserver/resources/topicgroup/pulsar"
	"github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/names"
	"regexp"
	"strconv"
)

const (
	producerRequestHold     = "producer_request_hold"
	producerException       = "producer_exception"
	consumerBacklogEviction = "consumer_backlog_eviction"
	NotSet                  = -12323344
	NotSetString            = "NotSet"
	destinationStorage      = "destination_storage"
	UNDEFINED               = "UNDEFINED"
	ALWAYS_INCOMPATIBLE     = "ALWAYS_INCOMPATIBLE"
	ALWAYS_COMPATIBLE       = "ALWAYS_COMPATIBLE"
	BACKWARD                = "BACKWARD"
	FORWARD                 = "FORWARD"
	FULL                    = "FULL"
	BACKWARD_TRANSITIVE     = "BACKWARD_TRANSITIVE"
	FORWARD_TRANSITIVE      = "FORWARD_TRANSITIVE"
	FULL_TRANSITIVE         = "FULL_TRANSITIVE"
	None                    = "None"
	Prefix                  = "Prefix"
	NameReg                 = "^[-=:.\\w]{100}$"
)

type Topicgroup struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"` //Topic分组名称
	Namespace   string     `json:"namespace"`
	Description string     `json:"description"`        //描述
	Policies    *Policies  `json:"policies,omitempty"` //Topic分组的策略
	CreatedAt   int64      `json:"createdAt"`          //创建时间
	Users       user.Users `json:"users"`
	Status      v1.Status  `json:"status"`
	Message     string     `json:"message"`
	Available   bool       `json:"available"` //是否可用
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
		Name:        app.Name,
		Policies:    ToPolicesApi(app.Policies),
		Description: app.Description,
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
		ID:          obj.ObjectMeta.Name,
		Name:        obj.Spec.Name,
		Namespace:   obj.ObjectMeta.Namespace,
		Status:      obj.Status.Status,
		Message:     obj.Status.Message,
		CreatedAt:   obj.ObjectMeta.CreationTimestamp.Unix(),
		Policies:    ToPolicesModel(obj.Spec.Policies),
		Users:       user.GetUsersFromLabels(obj.ObjectMeta.Labels),
		Available:   obj.Spec.Available,
		Description: obj.Spec.Description,
	}
}

func ToPolicesModel(obj *v1.Policies) *Policies {
	cRate, sRate, tRate := ToDispatchRateModel(obj)
	bMap := ToBacklogQuotaModel(obj)
	managedLedgerMaxMarkDeleteRate := ToManagedLedgerMaxMarkDeleteRateModel(obj)
	retentionPolicies := ToRetentionPolicesModel(obj)
	bundles := ToBundlesModel(obj)

	var persistent *PersistencePolicies
	if obj.Persistence != nil {
		persistent = &PersistencePolicies{
			BookkeeperEnsemble:             obj.Persistence.BookkeeperEnsemble,
			BookkeeperWriteQuorum:          obj.Persistence.BookkeeperWriteQuorum,
			BookkeeperAckQuorum:            obj.Persistence.BookkeeperAckQuorum,
			ManagedLedgerMaxMarkDeleteRate: managedLedgerMaxMarkDeleteRate,
		}
	}
	return &Policies{
		RetentionPolicies:           retentionPolicies,
		MessageTtlInSeconds:         obj.MessageTtlInSeconds,
		BacklogQuota:                &bMap,
		Bundles:                     bundles,
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
		Persistence:                 persistent,
		SubscriptionDispatchRate:    &sRate,
		ClusterSubscribeRate:        &cRate,
		TopicDispatchRate:           &tRate,
		DeduplicationEnabled:        obj.DeduplicationEnabled,
	}
}

func ToBundlesModel(obj *v1.Policies) *Bundles {
	var bundles *Bundles
	if obj.Bundles != nil {
		bundles = &Bundles{
			Boundaries: obj.Bundles.Boundaries,
			NumBundles: obj.Bundles.NumBundles,
		}
	}
	return bundles
}

func ToBacklogQuotaModel(obj *v1.Policies) map[string]BacklogQuota {
	bMap := make(map[string]BacklogQuota)
	if obj.BacklogQuota != nil {
		var backlogQuota BacklogQuota
		var objBacklogQuota = *obj.BacklogQuota
		backlogQuota.Limit = objBacklogQuota["destination_storage"].Limit
		backlogQuota.Policy = objBacklogQuota["destination_storage"].Policy
		bMap["destination_storage"] = backlogQuota
	}
	return bMap
}

func ToDispatchRateModel(obj *v1.Policies) (map[string]SubscribeRate, map[string]DispatchRate, map[string]DispatchRate) {
	cRate := make(map[string]SubscribeRate)
	if obj.ClusterSubscribeRate != nil {
		for k, v := range *obj.ClusterSubscribeRate {
			cRate[k] = SubscribeRate{
				SubscribeThrottlingRatePerConsumer: v.SubscribeThrottlingRatePerConsumer,
				RatePeriodInSecond:                 v.RatePeriodInSecond,
			}
		}
	}

	sRate := make(map[string]DispatchRate)
	if obj.SubscriptionDispatchRate != nil {
		for k, v := range *obj.SubscriptionDispatchRate {
			sRate[k] = DispatchRate{
				DispatchThrottlingRateInMsg:  v.DispatchThrottlingRateInMsg,
				DispatchThrottlingRateInByte: v.DispatchThrottlingRateInByte,
				RelativeToPublishRate:        v.RelativeToPublishRate,
				RatePeriodInSecond:           v.RatePeriodInSecond,
			}
		}
	}

	tRate := make(map[string]DispatchRate)
	if obj.TopicDispatchRate != nil {
		for k, v := range *obj.TopicDispatchRate {
			sRate[k] = DispatchRate{
				DispatchThrottlingRateInMsg:  v.DispatchThrottlingRateInMsg,
				DispatchThrottlingRateInByte: v.DispatchThrottlingRateInByte,
				RelativeToPublishRate:        v.RelativeToPublishRate,
				RatePeriodInSecond:           v.RatePeriodInSecond,
			}
		}
	}
	return cRate, sRate, tRate
}

func ToManagedLedgerMaxMarkDeleteRateModel(obj *v1.Policies) float64 {
	var managedLedgerMaxMarkDeleteRate float64
	if obj.Persistence != nil {
		//kubernetes不支持float64类型，因此存储的时候转成string，展示的时候再转成float64
		managedLedgerMaxMarkDeleteRate, _ = strconv.ParseFloat(obj.Persistence.ManagedLedgerMaxMarkDeleteRate, 64)
	}
	return managedLedgerMaxMarkDeleteRate
}

func ToRetentionPolicesModel(obj *v1.Policies) *RetentionPolicies {
	var retentionPolicies *RetentionPolicies
	if obj.RetentionPolicies != nil {
		retentionPolicies = &RetentionPolicies{
			RetentionSizeInMB:      obj.RetentionPolicies.RetentionSizeInMB,
			RetentionTimeInMinutes: obj.RetentionPolicies.RetentionTimeInMinutes,
		}
	}
	return retentionPolicies
}
func ToPolicesApi(policies *Policies) *v1.Policies {
	if policies == nil {
		return nil
	}

	cRate, sRate, tRate := ToDispatchRate(policies)

	bMap := make(map[string]v1.BacklogQuota)
	if policies.BacklogQuota != nil {
		var backlogQuota v1.BacklogQuota
		backlogQuota.Limit = (*policies.BacklogQuota)["destination_storage"].Limit
		backlogQuota.Policy = (*policies.BacklogQuota)["destination_storage"].Policy
		bMap["destination_storage"] = backlogQuota
	}

	crd := v1.Policies{
		Bundles: &v1.Bundles{
			Boundaries: policies.Bundles.Boundaries,
			NumBundles: policies.Bundles.NumBundles,
		},
		MessageTtlInSeconds:         policies.MessageTtlInSeconds,
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

		SubscriptionDispatchRate: &sRate,
		ClusterSubscribeRate:     &cRate,
		TopicDispatchRate:        &tRate,
		DeduplicationEnabled:     policies.DeduplicationEnabled,
	}

	if policies.Persistence != nil {
		persistence := &v1.PersistencePolicies{
			BookkeeperEnsemble:             policies.Persistence.BookkeeperEnsemble,
			BookkeeperWriteQuorum:          policies.Persistence.BookkeeperWriteQuorum,
			BookkeeperAckQuorum:            policies.Persistence.BookkeeperAckQuorum,
			ManagedLedgerMaxMarkDeleteRate: strconv.FormatFloat(policies.Persistence.ManagedLedgerMaxMarkDeleteRate, 'f', -1, 64)}
		crd.Persistence = persistence
	}

	if policies.RetentionPolicies != nil {
		retentionPolicies := &v1.RetentionPolicies{
			RetentionTimeInMinutes: policies.RetentionPolicies.RetentionTimeInMinutes,
			RetentionSizeInMB:      policies.RetentionPolicies.RetentionSizeInMB,
		}
		crd.RetentionPolicies = retentionPolicies
	}
	return &crd
}

func ToDispatchRate(policies *Policies) (map[string]v1.SubscribeRate, map[string]v1.DispatchRate, map[string]v1.DispatchRate) {
	cRate := make(map[string]v1.SubscribeRate)

	if policies.ClusterSubscribeRate != nil {
		for k, v := range *policies.ClusterSubscribeRate {
			cRate[k] = v1.SubscribeRate{
				SubscribeThrottlingRatePerConsumer: v.SubscribeThrottlingRatePerConsumer,
				RatePeriodInSecond:                 v.RatePeriodInSecond,
			}
		}
	}

	sRate := make(map[string]v1.DispatchRate)
	if policies.SubscriptionDispatchRate != nil {
		for k, v := range *policies.SubscriptionDispatchRate {
			sRate[k] = v1.DispatchRate{
				DispatchThrottlingRateInMsg:  v.DispatchThrottlingRateInMsg,
				DispatchThrottlingRateInByte: v.DispatchThrottlingRateInByte,
				RelativeToPublishRate:        v.RelativeToPublishRate,
				RatePeriodInSecond:           v.RatePeriodInSecond,
			}
		}

	}

	tRate := make(map[string]v1.DispatchRate)
	if policies.TopicDispatchRate != nil {
		for k, v := range *policies.TopicDispatchRate {
			tRate[k] = v1.DispatchRate{
				DispatchThrottlingRateInMsg:  v.DispatchThrottlingRateInMsg,
				DispatchThrottlingRateInByte: v.DispatchThrottlingRateInByte,
				RelativeToPublishRate:        v.RelativeToPublishRate,
				RatePeriodInSecond:           v.RatePeriodInSecond,
			}
		}

	}
	return cRate, sRate, tRate
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
		err := p.checkMessageTtlInSeconds()
		if err != nil {
			return err
		}

		err = p.checkRentionPolicies()
		if err != nil {
			return err
		}

		if err = p.checkBundles(); err != nil {
			return err
		}

		if err = p.checkBacklogQuota(); err != nil {
			return err
		}

		if err = p.checkPersistence(); err != nil {
			return err
		}

		if err = p.checkThrottling(); err != nil {
			return err
		}

		if err = p.checkCompationThreshold(); err != nil {
			return err
		}

		if err = p.checkSchemaCompatibilityStrategy(); err != nil {
			return err

		}

		if err = p.checkSubscriptionAuthMode(); err != nil {
			return err
		}
	}
	return nil

}

func (p *Policies) checkCompationThreshold() error {
	if p.CompactionThreshold != nil {
		if *p.CompactionThreshold < 0 {
			return fmt.Errorf("compactionThreshold cannot be less than 0")
		}
	}
	return nil
}

func (p *Policies) checkSubscriptionAuthMode() error {
	if p.SubscriptionAuthMode != nil {
		switch *p.SubscriptionAuthMode {
		case None:
		case Prefix:
			break
		default:
			return fmt.Errorf("subscriptionAuthMode is invalid: %+v", *p.SubscriptionAuthMode)
		}
	}
	return nil
}

func (p *Policies) checkBundles() error {
	if p.Bundles != nil {
		if p.Bundles.NumBundles <= 0 && p.Bundles.NumBundles != NotSet {
			return fmt.Errorf("numBundles is invalid: %d", p.Bundles.NumBundles)
		}
	}
	return nil
}

func (p *Policies) checkSchemaCompatibilityStrategy() error {
	if p.SchemaCompatibilityStrategy != nil {
		switch *p.SchemaCompatibilityStrategy {
		case UNDEFINED:
		case ALWAYS_COMPATIBLE:
		case ALWAYS_INCOMPATIBLE:
		case BACKWARD:
		case BACKWARD_TRANSITIVE:
		case FORWARD:
		case FORWARD_TRANSITIVE:
		case FULL:
		case FULL_TRANSITIVE:
			break
		default:
			return fmt.Errorf("schemaCompatibilityStrategy is invalid, value: %+v", *p.SchemaCompatibilityStrategy)
		}
	}
	return nil
}

func (p *Policies) checkThrottling() error {
	if p.MaxConsumersPerSubscription != nil {
		if *p.MaxConsumersPerSubscription < 0 {
			return fmt.Errorf("maxConsumersPerSubscription cannot be less than 0")
		}
	}
	if p.MaxConsumersPerTopic != nil {
		if *p.MaxConsumersPerTopic < 0 {
			return fmt.Errorf("maxConsumersPerTopic cannot be less than 0")

		}
	}

	if p.MaxProducersPerTopic != nil {
		if *p.MaxProducersPerTopic < 0 {
			return fmt.Errorf("maxProducersPerTopic cannot be less than 0")

		}
	}
	return nil
}

func (p *Policies) checkBacklogQuota() error {
	if p.BacklogQuota != nil {
		var backlogQuota = *p.BacklogQuota
		switch backlogQuota[destinationStorage].Policy {
		case producerRequestHold:
		case consumerBacklogEviction:
		case producerException:
			break
		default:
			if backlogQuota[destinationStorage].Policy != NotSetString {
				return fmt.Errorf("backlogQuota policy is invalid: %+v", backlogQuota["destination_storage"].Policy)
			}
		}
	}
	return nil
}

func (p *Policies) checkPersistence() error {
	if p.Persistence != nil {
		if p.Persistence.BookkeeperEnsemble < 0 || p.Persistence.BookkeeperAckQuorum < 0 || p.Persistence.BookkeeperWriteQuorum < 0 {
			return fmt.Errorf("persistent is invalid, BookkeeperEnsemble or BookkeeperAckQuorum or BookkeeperWriteQuorum < 0")
		}

		if p.Persistence.BookkeeperAckQuorum > p.Persistence.BookkeeperWriteQuorum {
			return fmt.Errorf("persistent is invalid, BookkeeperAckQuorum must less than or equal to BookkeeperWriteQuorum")
		}

		if p.Persistence.BookkeeperWriteQuorum > p.Persistence.BookkeeperEnsemble {
			return fmt.Errorf("persistent is invalid, BookkeeperWriteQuorum must be less than or equal to BookkeeperEnsemble")
		}
	}
	return nil
}

func (p *Policies) checkRentionPolicies() error {
	if p.RetentionPolicies != nil {
		if p.RetentionPolicies.RetentionTimeInMinutes < pulsar.MinRetentionTimeInMinutes && p.RetentionPolicies.RetentionTimeInMinutes != NotSet {
			return fmt.Errorf("retentionTimeInMinutes is invalid: %d", p.RetentionPolicies.RetentionTimeInMinutes)
		}
		if p.RetentionPolicies.RetentionSizeInMB < pulsar.MinRetentionSizeInMB && p.RetentionPolicies.RetentionTimeInMinutes != NotSet {
			return fmt.Errorf("retentionTimeInMinutes is invalid: %d", p.RetentionPolicies.RetentionSizeInMB)
		}

	}
	return nil
}

func (p *Policies) checkMessageTtlInSeconds() error {
	if p.MessageTtlInSeconds != nil {
		if *p.MessageTtlInSeconds < pulsar.DefaultMessageTTlInSeconds && *p.MessageTtlInSeconds != NotSet {
			return fmt.Errorf("messageTtlInSeconds is invalid: %d", p.MessageTtlInSeconds)
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
		} else {
			if ok, err := regexp.MatchString(NameReg, v); !ok {
				return fmt.Errorf("name is illegal: %v ", err)
			}
		}
	}

	p := a.Policies
	if err := p.Validate(); err != nil {
		return err
	}
	a.ID = names.NewID()
	return nil
}

func (a *Topicgroup) ValidateModifyBody() error {
	p := a.Policies
	if err := p.Validate(); err != nil {
		return err
	}
	a.ID = names.NewID()
	return nil
}
