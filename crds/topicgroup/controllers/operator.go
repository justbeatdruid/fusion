package controllers

import (
	"errors"
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
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
	Host string
	Port int
}

const namespaceUrl, protocol, success204 = "/admin/v2/namespaces/%s/%s", "http", 204

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

func (r *Operator) CreateNamespace(namespace *v1.Topicgroup) error {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
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
	response, _, err := request.Put(url).Send(createRequest).EndStruct("")
	if response.StatusCode == success204 {
		return nil
	} else {
		klog.Errorf("Create Topicgroup error, error: %+v, response code: %+v", err, response.StatusCode)
		return errors.New(fmt.Sprintf("%s:%d", "Create Topicgroup error, response code", response.StatusCode))
	}
}

func (r *Operator) DeleteNamespace(namespace *v1.Topicgroup) error {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	url := r.getUrl(namespace)
	response, _, err := request.Delete(url).Send("").EndStruct("")
	if response.StatusCode == success204 {
		return nil
	} else {
		//TODO 报错404：{"reason":"Namespace public/test1 does not exist."}的时候应该如何处理
		klog.Errorf("Delete Topicgroup error, error: %+v, response code: %+v", err, response.StatusCode)
		return errors.New(fmt.Sprintf("%s:%d", "Delete Topicgroup error, response code", response.StatusCode))
	}
}

func (r *Operator) getUrl(namespace *v1.Topicgroup) string {
	url := fmt.Sprintf(namespaceUrl, namespace.Spec.Tenant, namespace.Spec.Name)
	return fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, url)
}
