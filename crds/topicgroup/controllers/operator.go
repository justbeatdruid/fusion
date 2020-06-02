package controllers

import (
	"errors"
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
	"net/http"
	"strconv"
	"strings"
)

type Operator struct {
	Host           string
	Port           int
	AuthEnable     bool
	SuperUserToken string
}

type TenantCreateRequest struct {
	AllowedClusters []string `json:"allowedClusters"`
}

const namespaceUrl, protocol = "/admin/v2/namespaces/%s/%s", "http"
const (
	post                              = "post"
	put                               = "put"
	backlogUrlSuffix                  = "/backlogQuota?backlogQuotaType=destination_storage"
	messageTTLSuffix                  = "/messageTTL"
	retentionSuffix                   = "/retention"
	deduplicationSuffix               = "/deduplication"           //Enable or disable broker side deduplication for all topics in a namespace
	isAllowAutoUpdateSchemaSuffix     = "/isAllowAutoUpdateSchema" //UpdateFailed flag of whether allow auto update schema
	schemaValidationEnforcedSuffix    = "/schemaValidationEnforced"
	maxConsumersPerSubscriptionSuffix = "/maxConsumersPerSubscription"
	maxConsumersPerTopicSuffix        = "/maxConsumersPerTopic"
	maxProducersPerTopicSuffix        = "/maxProducersPerTopic"
	offloadDeletionLagMsSuffix        = "/offloadDeletionLagMs" //Set number of milliseconds to wait before deleting a ledger segment which has been offloaded from the Pulsar cluster's local storage (i.e. BookKeeper)
	offloadThresholdSuffix            = "/offloadThreshold"     //Set maximum number of bytes stored on the pulsar cluster for a topic, before the broker will start offloading to longterm storage
	CompactionThresholdSuffix         = "/compactionThreshold"  //Set maximum number of uncompacted bytes in a topic before compaction is triggered.
	PersistenceSuffix                 = "/persistence"
	dispatchRateSuffix                = "/dispatchRate"       //Set dispatch-rate throttling for all topics of the namespace
	encryptionRequiredSuffix          = "/encryptionRequired" //Message encryption is required or not for all topics in a namespace
	schemaCompatibilityStrategySuffix = "/schemaCompatibilityStrategy"
	subscribeRateSuffix               = "/subscribeRate"
	subscriptionAuthModeSuffix        = "/subscriptionAuthMode"
	subscriptionDispatchRateSuffix    = "/subscriptionDispatchRate"
	runtimeConfigurationUrl           = "/admin/v2/brokers/runtime/configuration"

	clusters     = "/admin/v2/clusters"
	tenants      = "/admin/v2/tenants"
	singleTenant = "/admin/v2/tenants/%s"
)

type Policies struct {
	RetentionPolicies           *v1.RetentionPolicies        `json:"retention_policies,omitempty"` //消息保留策略
	MessageTtlInSeconds         *int                         `json:"message_ttl_in_seconds"`       //未确认消息的最长保留时长
	BacklogQuota                *map[string]v1.BacklogQuota  `json:"backlog_quota_map"`
	Bundles                     *v1.Bundles                  `json:"bundles"` //key:destination_storage
	TopicDispatchRate           *map[string]v1.DispatchRate  `json:"topicDispatchRate"`
	SubscriptionDispatchRate    *map[string]v1.DispatchRate  `json:"subscriptionDispatchRate"`
	ClusterSubscribeRate        *map[string]v1.SubscribeRate `json:"clusterSubscribeRate"`
	Persistence                 *PersistencePolicies         `json:"persistence"` //Configuration of bookkeeper persistence policies.
	DeduplicationEnabled        *bool                        `json:"deduplicationEnabled"`
	EncryptionRequired          *bool                        `json:"encryption_required"`
	SubscriptionAuthMode        *string                      `json:"subscription_auth_mode"` //None/Prefix
	MaxProducersPerTopic        *int                         `json:"max_producers_per_topic"`
	MaxConsumersPerTopic        *int                         `json:"max_consumers_per_topic"`
	MaxConsumersPerSubscription *int                         `json:"max_consumers_per_subscription"`
	CompactionThreshold         *int64                       `json:"compaction_threshold"`
	OffloadThreshold            *int64                       `json:"offload_threshold"`
	OffloadDeletionLagMs        *int64                       `json:"offload_deletion_lag_ms"`
	IsAllowAutoUpdateSchema     *bool                        `json:"is_allow_auto_update_schema"`
	SchemaValidationEnforced    *bool                        `json:"schema_validation_enforced"`
	SchemaCompatibilityStrategy *string                      `json:"schema_compatibility_strategy"`
}
type PersistencePolicies struct {
	BookkeeperEnsemble             int     `json:"bookkeeperEnsemble,omitempty"`
	BookkeeperWriteQuorum          int     `json:"bookkeeperWriteQuorum,omitempty"`
	BookkeeperAckQuorum            int     `json:"bookkeeperAckQuorum,omitempty"`
	ManagedLedgerMaxMarkDeleteRate float64 `json:"managedLedgerMaxMarkDeleteRate,omitempty"`
}
type requestLogger struct {
	prefix string
}

var logger = &requestLogger{}

func (r *requestLogger) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *requestLogger) Printf(format string, v ...interface{}) {
	klog.V(4).Infof(format, v...)
}

func (r *requestLogger) Println(v ...interface{}) {
	klog.V(4).Infof("%+v", v)
}

//在pulsar中创建命名空间
//403：无权限；404：租户或者集群不存在；409：命名空间已存在；412：名称非法
func (r *Operator) CreateNamespace(namespace *v1.Topicgroup) error {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace)

	response, _, err := request.Put(url).Send("").End()
	if response.StatusCode == http.StatusNoContent {
		return nil
	} else {
		klog.Errorf("Create Topicgroup error, error: %+v, response code: %+v", err, response.StatusCode)
		return errors.New(fmt.Sprintf("%s:%d", "Create Topicgroup error, response code", response.StatusCode))
	}
}

func (r *Operator) DeleteNamespace(namespace *v1.Topicgroup) error {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace)
	response, body, err := request.Delete(url).Send("").End()
	if response.StatusCode == http.StatusNoContent {
		return nil
	} else if strings.Contains(body, "does not exist") {
		return nil
	} else {
		//TODO 报错404：{"reason":"Namespace public/test1 does not exist."}的时候应该如何处理
		klog.Errorf("DeleteFailed Topicgroup error, error: %+v, response code: %+v", err, response.StatusCode)
		return errors.New(fmt.Sprintf("%s:%d%s", "DeleteFailed Topicgroup error, response code", response.StatusCode, body))
	}
}

func (r *Operator) isNamespacesExist(namespace *v1.Topicgroup) (bool, error) {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace)

	response, body, errs := request.Get(url).Send("").End()
	if errs != nil {
		klog.Errorf("get namespace policy finished, url: %+v, response: %+v, errs: %+v", url, response, errs)
		return true, fmt.Errorf("get namespace policy error: %+v ", errs)
	}

	if response.StatusCode == http.StatusOK {
		return true, nil
	}

	if response.StatusCode == http.StatusNotFound && (strings.Contains(body, `Tenant does not exist`) || strings.Contains(body, `Namespace does not exist`)) {
		return false, nil

	}
	return true, nil
}

func (r *Operator) GetNamespacePolicies(namespace *v1.Topicgroup) (*v1.Policies, error) {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace)

	polices := &Policies{}
	response, _, errs := request.Get(url).Send("").EndStruct(polices)
	if errs != nil {
		klog.Errorf("get namespace policy finished, url: %+v, response: %+v, errs: %+v", url, response, errs)
		return nil, fmt.Errorf("get namespace policy error: %+v ", errs)
	}

	if response.StatusCode != http.StatusOK {
		klog.Errorf("get namespace policy finished, url: %+v, response: %+v, errs: %+v", url, response, errs)
		return nil, fmt.Errorf("get namespace policy http code is not 200: %+v ", response.StatusCode)
	}
	persistence, err := r.GetPersistence(namespace)
	if err != nil {
		return nil, fmt.Errorf("get namespace policy error: %+v or http code is not success: %+v", errs, response.StatusCode)
	}

	polices.Persistence = persistence

	retention, err := r.GetRetention(namespace)
	if err != nil {
		return nil, fmt.Errorf("get namespace policy error: %+v or http code is not success: %+v", errs, response.StatusCode)
	}

	polices.RetentionPolicies = retention
	return toCrdModel(polices), nil

}

func (r *Operator) GetPersistence(namespace *v1.Topicgroup) (*PersistencePolicies, error) {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace) + PersistenceSuffix
	polices := &PersistencePolicies{}
	response, _, errs := request.Get(url).Send("").EndStruct(polices)
	if response.StatusCode != http.StatusOK || errs != nil {
		klog.Errorf("get namespace persistence policy finished, url: %+v, response: %+v, errs: %+v", url, response, errs)
		return nil, fmt.Errorf("get namespace persistence policy error: %+v or http code is not success: %+v", errs, response.StatusCode)
	}

	return polices, nil
}

func (r *Operator) SetMessageTTL(namespace *v1.Topicgroup) error {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace) + messageTTLSuffix
	response, body, errs := request.Post(url).Send(namespace.Spec.Policies.MessageTtlInSeconds).End()

	klog.Infof("set messageTTLInSeconds finished, url: %+v, response: %+v, body: %+v, errs: %+v", url, response, body, errs)
	if response.StatusCode != http.StatusNoContent || errs != nil {
		return fmt.Errorf("set messageTTLInSeconds error: %+v or http code is not success: %+v", errs, response.StatusCode)
	}
	return nil
}

func (r *Operator) SetDeduplication(namespace *v1.Topicgroup) error {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace) + deduplicationSuffix
	response, body, errs := request.Post(url).Send(namespace.Spec.Policies.DeduplicationEnabled).End()

	klog.Infof("set deduplication finished, url: %+v, response: %+v, body: %+v, errs: %+v", url, response, body, errs)
	if response.StatusCode != http.StatusNoContent || errs != nil {
		return fmt.Errorf("set deduplication error: %+v or http code is not success: %+v", errs, response.StatusCode)
	}
	return nil
}

func (r *Operator) SetPolicy(suffix string, namespace *v1.Topicgroup, content interface{}, method string) error {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace) + suffix
	var response gorequest.Response
	var body string
	var errs = make([]error, 0)
	if method == "post" {
		response, body, errs = request.Post(url).Send(content).End()
	} else {
		response, body, errs = request.Put(url).Send(content).End()
	}

	klog.Infof("set compactionThreshold finished, url: %+v, response: %+v, body: %+v, errs: %+v", url, response, body, errs)
	if response.StatusCode != http.StatusNoContent || errs != nil {
		return fmt.Errorf("set compactionThreshold error: %+v or http code is not success: %+v", errs, response.StatusCode)
	}
	return nil

}

func (r *Operator) SetRetention(namespace *v1.Topicgroup) error {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace) + retentionSuffix
	response, body, errs := request.Post(url).Send(&namespace.Spec.Policies.RetentionPolicies).End()

	klog.Infof("set retention finished, url: %+v, response: %+v, body: %+v, errs: %+v", url, response, body, errs)
	if response.StatusCode != http.StatusNoContent || errs != nil {
		return fmt.Errorf("set retention error: %+v or http code is not success: %+v", errs, response.StatusCode)
	}
	return nil
}

func (r *Operator) GetRetention(namespace *v1.Topicgroup) (*v1.RetentionPolicies, error) {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace) + retentionSuffix

	rentention := &v1.RetentionPolicies{}
	response, _, errs := request.Get(url).Send("").EndStruct(rentention)
	if response.StatusCode != http.StatusOK || errs != nil {
		return nil, fmt.Errorf("get retention error: %+v or http code is not success: %+v", errs, response.StatusCode)
	}
	return rentention, nil
}
func (r *Operator) SetBacklogQuota(namespace *v1.Topicgroup) error {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace) + backlogUrlSuffix

	response, body, errs := request.Post(url).Send(&namespace.Spec.Policies.BacklogQuota).End()
	klog.Infof("set backlog quota finished, url: %+v, response: %+v, body: %+v, errs: %+v", url, response, body, errs)
	if response.StatusCode != http.StatusNoContent || errs != nil {
		return fmt.Errorf("set backlog quota error: %+v or http code is not success: %+v", errs, response.StatusCode)
	}

	return nil
}
func (r *Operator) getUrl(namespace *v1.Topicgroup) string {
	url := fmt.Sprintf(namespaceUrl, namespace.ObjectMeta.Namespace, namespace.Spec.Name)
	return fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, url)
}

func (r *Operator) AddTokenToHeader(request *gorequest.SuperAgent) *gorequest.SuperAgent {
	if r.AuthEnable {
		request.Header.Set("Authorization", "Bearer "+r.SuperUserToken)
	}
	return request
}

func (r *Operator) GetHttpRequest() *gorequest.SuperAgent {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)
	//request.Header.Add("Content-Type", "application/json")
	return r.AddTokenToHeader(request)

}

//Get the list of all the Pulsar clusters
func (r *Operator) GetAllClusters() ([]string, error) {
	request := r.GetHttpRequest()
	url := fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, clusters)

	var clusters = make([]string, 1)
	response, _, errs := request.Get(url).EndStruct(&clusters)
	if errs != nil {
		klog.Errorf("get all clusters error: %+v", errs)
		return nil, fmt.Errorf("get all clusters error: %+v", errs)
	}
	if response.StatusCode == http.StatusOK {
		return clusters, nil
	}

	return nil, fmt.Errorf("get all clusters error, code: %+v, response: %+v", response.StatusCode, response)
}

//Get the list of existing tenants
func (r *Operator) GetAllTenants() ([]string, error) {
	request := r.GetHttpRequest()
	url := fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, tenants)

	var tenants = make([]string, 1)
	response, _, errs := request.Get(url).EndStruct(&tenants)
	if errs != nil {
		klog.Errorf("get all tenants error: %+v", errs)
		return nil, fmt.Errorf("get all tenants error: %+v", errs)
	}

	if response.StatusCode == http.StatusOK {
		return tenants, nil
	}

	return nil, fmt.Errorf("get all tenants error, code: %+v, response: %+v", response.StatusCode, response)
}

//Create tenant if not exist
func (r *Operator) CreateTenantIfNotExist(tenant string) error {
	tenants, err := r.GetAllTenants()
	if err != nil {
		return err
	}

	//先判断租户是否已存在
	for _, t := range tenants {
		if t == tenant {
			//租户已存在，无需创建
			return nil
		}
	}
	//先查询集群信息
	clusters, err := r.GetAllClusters()
	if err != nil {
		return fmt.Errorf("unable to create tenant, tenant: %+v, err: %+v", tenant, err)
	}

	//创建租户
	if err = r.CreateTenant(tenant, clusters); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

//Create a new tenant
func (r *Operator) CreateTenant(tenant string, clusters []string) error {
	request := r.GetHttpRequest()
	url := fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, fmt.Sprintf(singleTenant, tenant))

	requestBody := TenantCreateRequest{AllowedClusters: clusters}
	response, _, errs := request.Put(url).Send(requestBody).End()
	if errs != nil {
		return fmt.Errorf("unable to create tenant, url: %+v, error: %+v", url, errs)
	}

	if response.StatusCode == http.StatusNoContent || response.StatusCode == http.StatusConflict {
		return nil
	}

	return fmt.Errorf("unable to create tenant, url: %+v, response: %+v", url, response)
}

//TODO 待补充
func (r *Operator) GetRuntimeConfiguration() (*v1.RuntimeConfiguration, error) {
	return nil, nil
}

func toCrdModel(policies *Policies) *v1.Policies {
	crd := &v1.Policies{
		DeduplicationEnabled:        policies.DeduplicationEnabled,
		IsAllowAutoUpdateSchema:     policies.IsAllowAutoUpdateSchema,
		SubscriptionAuthMode:        policies.SubscriptionAuthMode,
		SchemaCompatibilityStrategy: policies.SchemaCompatibilityStrategy,
		SchemaValidationEnforced:    policies.SchemaValidationEnforced,
		MessageTtlInSeconds:         policies.MessageTtlInSeconds,
		MaxProducersPerTopic:        policies.MaxProducersPerTopic,
		MaxConsumersPerTopic:        policies.MaxConsumersPerTopic,
		MaxConsumersPerSubscription: policies.MaxConsumersPerSubscription,
		CompactionThreshold:         policies.CompactionThreshold,
		OffloadThreshold:            policies.OffloadThreshold,
		EncryptionRequired:          policies.EncryptionRequired,
		RetentionPolicies:           policies.RetentionPolicies,
		Bundles:                     policies.Bundles,
		BacklogQuota:                policies.BacklogQuota,
		TopicDispatchRate:           policies.TopicDispatchRate,
		SubscriptionDispatchRate:    policies.SubscriptionDispatchRate,
		ClusterSubscribeRate:        policies.ClusterSubscribeRate,
	}
	if crd.DeduplicationEnabled == nil {
		var deduplication = false
		crd.DeduplicationEnabled = &deduplication
	}

	if policies.Persistence != nil {
		persistence := &v1.PersistencePolicies{
			BookkeeperEnsemble:             policies.Persistence.BookkeeperEnsemble,
			BookkeeperWriteQuorum:          policies.Persistence.BookkeeperWriteQuorum,
			BookkeeperAckQuorum:            policies.Persistence.BookkeeperAckQuorum,
			ManagedLedgerMaxMarkDeleteRate: strconv.FormatFloat(policies.Persistence.ManagedLedgerMaxMarkDeleteRate, 'f', -1, 64),
		}
		crd.Persistence = persistence
	}

	return crd
}
