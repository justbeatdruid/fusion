package controllers

import (
	"errors"
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
	"net/http"
	"strings"
)

type CreateRequest struct {
	MessageTtlInSeconds int                     `json:"message_ttl_in_seconds"`
	RetentionPolicies   RetentionPolicies       `json:"retention_policies"`
	Bundles             BundlesData             `json:"bundles"`
	BacklogQuotaMap     map[string]BacklogQuota `json:"backlog_quota_map"` //示例：{"destination_storage":{"limit":-1073741824,"policy":"producer_request_hold"}}
}
type RetentionPolicies struct {
	RetentionTimeInMinutes int   `json:"retentionTimeInMinutes"`
	RetentionSizeInMB      int64 `json:"retentionSizeInMB"`
}
type BundlesData struct {
	NumBundles int `json:"numBundles"`
}
type BacklogQuota struct {
	Limit  int64  `json:"limit"`
	Policy string `json:"policy"` //producer_request_hold, producer_exception,consumer_backlog_eviction
}
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
	backlogUrlSuffix                  = "/backlogQuota?backlogQuotaType=destination_storage"
	messageTTLSuffix                  = "/messageTTL"
	retentionSuffix                   = "/retention"
	deduplicationSuffix               = "/deduplication"           //Enable or disable broker side deduplication for all topics in a namespace
	isAllowAutoUpdateSchemaSuffix     = "/isAllowAutoUpdateSchema" //Update flag of whether allow auto update schema
	schemaValidationEnforcedSuffix    = "/schemaValidationEnforced"
	maxConsumersPerSubscriptionSuffix = "/maxConsumersPerSubscription"
	maxConsumersPerTopicSuffix        = "/maxConsumersPerTopic"
	maxProducersPerTopicSuffix        = "/maxProducersPerTopic"
	offloadDeletionLagMsSuffix        = "/offloadDeletionLagMs" //Set number of milliseconds to wait before deleting a ledger segment which has been offloaded from the Pulsar cluster's local storage (i.e. BookKeeper)
	offloadThresholdSuffix            = "/offloadThreshold"     //Set maximum number of bytes stored on the pulsar cluster for a topic, before the broker will start offloading to longterm storage
	compactionThresholdSuffix         = "/compactionThreshold"  //Set maximum number of uncompacted bytes in a topic before compaction is triggered.
	persistenceSuffix                 = "/persistence"
	dispatchRateSuffix                = "/dispatchRate"       //Set dispatch-rate throttling for all topics of the namespace
	encryptionRequiredSuffix          = "/encryptionRequired" //Message encryption is required or not for all topics in a namespace
	schemaCompatibilityStrategySuffix = "/schemaCompatibilityStrategy"
	subscribeRateSuffix               = "/subscribeRate"
	subscriptionAuthModeSuffix        = "/subscriptionAuthMode"
	subscriptionDispatchRateSuffix    = "/subscriptionDispatchRate"

	clusters     = "/admin/v2/clusters"
	tenants      = "/admin/v2/tenants"
	singleTenant = "/admin/v2/tenants/%s"
)

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

	ps := namespace.Spec.Policies

	bq := &BacklogQuota{
		Limit:  ps.BacklogQuota.Limit,
		Policy: ps.BacklogQuota.Policy,
	}
	bmap := make(map[string]BacklogQuota)
	bmap["destination_storage"] = *bq
	createRequest := &CreateRequest{
		MessageTtlInSeconds: ps.MessageTtlInSeconds,
		RetentionPolicies: RetentionPolicies{
			RetentionSizeInMB:      ps.RetentionPolicies.RetentionSizeInMB,
			RetentionTimeInMinutes: ps.RetentionPolicies.RetentionTimeInMinutes,
		},
		Bundles: BundlesData{
			NumBundles: ps.NumBundles,
		},
		BacklogQuotaMap: bmap,
	}

	response, _, err := request.Put(url).Send(createRequest).End()
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
		klog.Errorf("Delete Topicgroup error, error: %+v, response code: %+v", err, response.StatusCode)
		return errors.New(fmt.Sprintf("%s:%d%s", "Delete Topicgroup error, response code", response.StatusCode, body))
	}
}

func (r *Operator) GetNamespacePolicies(namespace *v1.Topicgroup) (*CreateRequest, error) {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace)

	polices := &CreateRequest{}
	response, _, errs := request.Get(url).Send("").EndStruct(polices)
	if response.StatusCode != http.StatusOK || errs != nil {
		klog.Errorf("get namespace policy finished, url: %+v, response: %+v, errs: %+v", url, response, errs)
		return nil, fmt.Errorf("get namespace policy error: %+v or http code is not success: %+v", errs, response.StatusCode)
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

func (r *Operator) SetRetention(namespace *v1.Topicgroup) error {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace) + retentionSuffix
	requestBody := &RetentionPolicies{
		RetentionTimeInMinutes: namespace.Spec.Policies.RetentionPolicies.RetentionTimeInMinutes,
		RetentionSizeInMB:      namespace.Spec.Policies.RetentionPolicies.RetentionSizeInMB,
	}
	response, body, errs := request.Post(url).Send(requestBody).End()

	klog.Infof("set retention finished, url: %+v, response: %+v, body: %+v, errs: %+v", url, response, body, errs)
	if response.StatusCode != http.StatusNoContent || errs != nil {
		return fmt.Errorf("set retention error: %+v or http code is not success: %+v", errs, response.StatusCode)
	}
	return nil
}
func (r *Operator) SetBacklogQuota(namespace *v1.Topicgroup) error {
	request := r.GetHttpRequest()
	url := r.getUrl(namespace) + backlogUrlSuffix

	requestBody := BacklogQuota{
		Limit:  namespace.Spec.Policies.BacklogQuota.Limit,
		Policy: namespace.Spec.Policies.BacklogQuota.Policy,
	}
	response, body, errs := request.Post(url).Send(requestBody).End()
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

	return nil, fmt.Errorf("get all clusters error, response: %+v", response)
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

	return nil, fmt.Errorf("get all tenants error, response: %+v", response)
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
