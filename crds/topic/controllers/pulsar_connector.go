package controllers

import (
	"errors"
	"fmt"
	nlptv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
	"net/http"
	"strings"
	"time"
)

const (
	persistentTopicUrl         = "/admin/v2/persistent/%s/%s/%s"
	nonPersistentTopicUrl      = "/admin/v2/non-persistent/%s/%s/%s"
	protocol                   = "http"
	persistentPermissionUrl    = "/admin/v2/persistent/%s/%s/%s/permissions/%s"
	nonPersistentPermissionUrl = "/admin/v2/non-persistent/%s/%s/%s/permissions/%s"
	statsUrl                   = "/stats"
	partitionedStatsUrl        = "/partitioned-stats"
	namespaceUrl               = "/admin/v2/namespaces/%s/%s"
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

//Connector 定义连接Pulsar所需要的参数
type Connector struct {
	Host           string
	Port           int
	AuthEnable     bool
	SuperUserToken string
}

type Stats struct {
	MsgRateIn           float64                     `json:"msgRateIn"`
	MsgRateOut          float64                     `json:"msgRateOut"`
	MsgThroughputIn     float64                     `json:"msgThroughputIn"`
	MsgThroughputOut    float64                     `json:"msgThroughputOut"`
	MsgInCounter        int64                       `json:"MsgInCounter"`
	AverageMsgSize      float64                     `json:"averageMsgSize"`
	BytesInCounter      int64                       `json:"bytesInCounter"`
	StorageSize         int64                       `json:"storageSize"`
	BacklogSize         int64                       `json:"backlogSize"`
	DeduplicationStatus string                      `json:"deduplicationStatus"`
	Subscriptions       map[string]SubscriptionStat `json:"subscriptions"`
	Publishers          []Publisher                 `json:"publishers"`
	Partitions          map[string]Stats            `json:"partitions"`
}
type Publisher struct {
	MsgRateIn       float64 `json:"msgRateIn"`
	MsgThroughputIn float64 `json:"msgThroughputIn"`
	AverageMsgSize  float64 `json:"averageMsgSize"`
	ProducerId      int64   `json:"producerId"`
	ProducerName    string  `json:"producerName"`
	Address         string  `json:"address"`
	ConnectedSince  string  `json:"connectedSince"`
}
type SubscriptionStat struct {
	MsgRateOut                       float64        `json:"msgRateOut"`
	MsgThroughputOut                 float64        `json:"msgThroughputOut"`
	MsgRateRedeliver                 float64        `json:"msgRateRedeliver"`
	MsgBacklog                       int64          `json:"msgBacklog"`
	BlockedSubscriptionOnUnackedMsgs bool           `json:"blockedSubscriptionOnUnackedMsgs"`
	MsgDelayed                       int64          `json:"msgDelayed"`
	UnackedMessages                  int64          `json:"unackedMessages"`
	Type                             string         `json:"type"`
	MsgRateExpired                   float64        `json:"msgRateExpired"`
	LastExpireTimestamp              int64          `json:"lastExpireTimestamp"`
	LastConsumedFlowTimestamp        int64          `json:"lastConsumedFlowTimestamp"`
	LastConsumedTimestamp            int64          `json:"lastConsumedTimestamp"`
	LastAckedTimestamp               int64          `json:"lastAckedTimestamp"`
	Consumers                        []ConsumerStat `json:"consumers"`
	IsReplicated                     bool           `json:"isReplicated"`
}

type ConsumerStat struct {
	MsgRateOut            float64 `json:"msgRateOut"`
	MsgThroughputOut      float64 `json:"msgThroughputOut"`
	ConsumerName          string  `json:"consumerName"`
	AvailablePermits      int     `json:"availablePermits"`
	UnackedMessages       int     `json:"unackedMessages"`
	LastAckedTimestamp    int64   `json:"lastAckedTimestamp"`
	LastConsumedTimestamp int64   `json:"lastConsumedTimestamp"`
	ConnectedSince        string  `json:"connectedSince"`
	Address               string  `json:"address"`
}

//CreateTopic 调用Pulsar的Restful Admin API，创建Topic
func (r *Connector) CreateTopic(topic *nlptv1.Topic) (err error) {
	if topic.Spec.Partitioned {
		return r.CreatePartitionedTopic(topic)
	}

	request := r.GetHttpRequest()
	//klog.Infof("Param: tenant:%s, namespace:%s, topicName:%s", topic.Namespace, topic.Spec.TopicGroup, topic.Spec.Name)
	topicUrl := r.getUrl(topic)
	response, _, errs := request.Put(topicUrl).Send("").EndStruct("")
	if response.StatusCode == 204 {
		stats, err := r.GetStats(*topic)
		if err != nil {
			klog.Errorf("%+v", err)
			return nil
		}
		topic.Spec.Stats = *stats
		return nil
	} else {
		errMsg := fmt.Sprintf("Create topic error, url: %s, Error code: %d, Error Message: %+v", topicUrl, response.StatusCode, errs)
		klog.Error(errMsg)
		return errors.New(errMsg)
	}
}

func (r *Connector) CreatePartitionedTopic(topic *nlptv1.Topic) (err error) {
	request := r.GetHttpRequest()
	klog.Infof("CreatePartitionedTopic Param: tenant:%s, namespace:%s, topicName:%s", topic.Namespace, topic.Spec.TopicGroup, topic.Spec.Name)
	topicUrl := r.getUrl(topic)

	response, _, errs := request.Put(topicUrl).Send(topic.Spec.PartitionNum).EndStruct("")
	if response.StatusCode == 204 {
		return nil
	} else {
		errMsg := fmt.Sprintf("Create topic error, url: %s, Error code: %d, Error Message: %+v, body: %+v", topicUrl, response.StatusCode, errs, response)
		klog.Error(errMsg)
		return errors.New(errMsg)
	}
}

//DeleteTopic 调用Pulsar的Restful Admin API，删除Topic
func (r *Connector) DeleteTopic(topic *nlptv1.Topic, force bool) (err error) {
	request := r.GetHttpRequest()
	topicUrl := r.getUrl(topic)
	if force {
		topicUrl = fmt.Sprintf("%s?force=true", topicUrl)
	}
	response, body, errs := request.Delete(topicUrl).Retry(3, 5*time.Second, http.StatusBadRequest, http.StatusInternalServerError).End()
	if err != nil {
		return fmt.Errorf("delete topic(%+v) error: %+v", topicUrl, errs)
	}
	if response.StatusCode == http.StatusNoContent {
		return nil
	} else if strings.Contains(body, "Topic not found") || strings.Contains(body, "Partitioned topic does not exist") || strings.Contains(body, "Policies not found for") {
		return nil
	} else {
		errMsg := fmt.Sprintf("delete topic error, url: %s, Error code: %d, Error Message: %s", topicUrl, response.StatusCode, body)
		klog.Error(errMsg)
		return errors.New(errMsg)
	}
}

func (r *Connector) GrantPermission(topic *nlptv1.Topic, permission *nlptv1.Permission) (err error) {
	request := r.GetHttpRequest()
	var url string
	if topic.Spec.Persistent {
		url = persistentPermissionUrl
	} else {
		url = nonPersistentPermissionUrl
	}
	url = fmt.Sprintf(url, topic.Namespace, topic.Spec.TopicGroup, topic.Spec.Name, permission.AuthUserName)
	url = fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, url)
	response, body, errs := request.Post(url).Send(permission.Actions).End()
	if err != nil {
		return fmt.Errorf("grant permission error,  topic(%+v) error: %+v", topic.Spec.Url, errs)
	}
	klog.Infof("grant permission result, url: %+v, response: %+v, body: %+v, err:%+v", url, response, body, errs)
	if response.StatusCode == 204 {
		return nil
	}

	return fmt.Errorf("grant permission error: %+v", errs)

}

func (r *Connector) getUrl(topic *nlptv1.Topic) string {
	url := persistentTopicUrl
	if !topic.Spec.Persistent {
		url = nonPersistentTopicUrl
	}

	if topic.Spec.Partitioned {
		url += "/partitions"
	}
	topicUrl := fmt.Sprintf(url, topic.Namespace, topic.Spec.TopicGroup, topic.Spec.Name)

	return fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, topicUrl)
}

func (r *Connector) getNamespaceUrl(topic *nlptv1.Topic) string {
	url := namespaceUrl
	url = fmt.Sprintf(url, topic.Namespace, topic.Spec.TopicGroup)
	return fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, url)
}

//删除授权
func (r *Connector) DeletePer(topic *nlptv1.Topic, P *nlptv1.Permission) (err error) {
	request := r.GetHttpRequest()
	url := persistentTopicUrl
	if !topic.Spec.Persistent {
		url = nonPersistentTopicUrl
	}

	topicUrl := fmt.Sprintf(url, topic.Namespace, topic.Spec.TopicGroup, topic.Spec.Name)
	topicUrl = fmt.Sprintf("%s://%s:%d%s/%s/%s", protocol, r.Host, r.Port, topicUrl, "permissions", P.AuthUserName)
	response, _, errs := request.Delete(topicUrl).Retry(3, 5*time.Second).Send("").EndStruct("")

	if err != nil {
		return fmt.Errorf("revoke permission error,  topic(%+v) error: %+v", topic.Spec.Url, errs)
	}
	if response.StatusCode == http.StatusNoContent {
		return nil
	}
	errMsg := fmt.Sprintf("revoke permission error, url: %s, Error code: %d, Error Message: %+v", topicUrl, response.StatusCode, errs)
	klog.Error(errMsg)
	return errors.New(errMsg)
}

func (r *Connector) AddTokenToHeader(request *gorequest.SuperAgent) *gorequest.SuperAgent {
	if r.AuthEnable {
		request.Header.Set("Authorization", "Bearer "+r.SuperUserToken)
	}
	return request
}

func (r *Connector) GetHttpRequest() *gorequest.SuperAgent {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)
	return r.AddTokenToHeader(request)

}

func (r *Connector) GetStats(topic nlptv1.Topic) (*nlptv1.Stats, error) {
	request := r.GetHttpRequest()
	url := persistentTopicUrl
	if !topic.Spec.Persistent {
		url = nonPersistentTopicUrl
	}

	if topic.Spec.Partitioned {
		url += partitionedStatsUrl
	} else {
		url += statsUrl
	}
	url = fmt.Sprintf(url, topic.Namespace, topic.Spec.TopicGroup, topic.Spec.Name)
	url = fmt.Sprintf("%s://%s:%d%s", protocol, r.Host, r.Port, url)
	stats := &Stats{}
	response, _, errs := request.Get(url).Retry(3, 5*time.Second, http.StatusNotFound, http.StatusBadGateway).EndStruct(stats)
	if errs != nil {
		klog.Errorf("get stats error, url: %+v, response: %+v, errs: %+v", url, response, errs)
		return nil, fmt.Errorf("get stats error: %+v", errs)
	}

	if response.StatusCode == http.StatusOK {
		//处理float64的精度，float类型自动省略.00的问题
		//直接将无小数点的float类型数据保存在crd中，反序列化时会当成int64类型，导致类型出错
		return r.FormatStats(stats), nil
	}
	//klog.Errorf("get stats error, url: %+v, status code: %+v, errs: %+v", url, response.StatusCode, errs)
	return nil, fmt.Errorf("get stats error: %+v", errs)

}

func (r *Connector) FormatStats(stats *Stats) *nlptv1.Stats {
	parsedStats := &nlptv1.Stats{
		DeduplicationStatus: stats.DeduplicationStatus,
		MsgInCounter:        stats.MsgInCounter,
		BytesInCounter:      stats.BytesInCounter,
		StorageSize:         stats.StorageSize,
		BacklogSize:         stats.BacklogSize,
	}

	if stats.Partitions != nil {
		parsedStats.Partitions = make(map[string]nlptv1.Stats, 0)
		for k, v := range stats.Partitions {
			parsedStats.Partitions[k] = *r.FormatStats(&v)
		}
	}

	parsedStats.MsgRateIn = fmt.Sprintf("%.3f", stats.MsgRateIn)
	parsedStats.MsgRateOut = fmt.Sprintf("%.3f", stats.MsgRateOut)
	parsedStats.MsgThroughputOut = fmt.Sprintf("%.3f", stats.MsgThroughputOut)
	parsedStats.AverageMsgSize = fmt.Sprintf("%.3f", stats.AverageMsgSize)
	parsedStats.MsgThroughputIn = fmt.Sprintf("%.3f", stats.MsgThroughputIn)

	var subscriptions = make(map[string]nlptv1.SubscriptionStat)
	if len(stats.Subscriptions) > 0 {
		for k, v := range stats.Subscriptions {
			var subscription nlptv1.SubscriptionStat
			subscription.IsReplicated = v.IsReplicated
			subscription.LastAckedTimestamp = v.LastAckedTimestamp
			subscription.LastConsumedTimestamp = v.LastConsumedTimestamp
			subscription.Type = v.Type
			subscription.LastExpireTimestamp = v.LastExpireTimestamp
			subscription.UnackedMessages = v.UnackedMessages
			subscription.LastConsumedFlowTimestamp = v.LastConsumedFlowTimestamp
			subscription.BlockedSubscriptionOnUnackedMsgs = v.BlockedSubscriptionOnUnackedMsgs
			subscription.UnackedMessages = v.UnackedMessages
			subscription.MsgThroughputOut = fmt.Sprintf("%.3f", v.MsgThroughputOut)
			subscription.MsgRateOut = fmt.Sprintf("%.3f", v.MsgRateOut)
			subscription.MsgRateExpired = fmt.Sprintf("%.3f", v.MsgRateExpired)
			subscription.MsgRateRedeliver = fmt.Sprintf("%.3f", v.MsgRateRedeliver)

			var consumers = make([]nlptv1.ConsumerStat, 0)
			for _, consumer := range v.Consumers {
				var c nlptv1.ConsumerStat
				c.MsgRateOut = fmt.Sprintf("%.8f", consumer.MsgRateOut)
				c.MsgThroughputOut = fmt.Sprintf("%.8f", consumer.MsgThroughputOut)
				c.Address = consumer.Address
				c.ConsumerName = consumer.ConsumerName
				c.ConnectedSince = consumer.ConnectedSince
				c.LastConsumedTimestamp = consumer.LastConsumedTimestamp
				c.LastAckedTimestamp = consumer.LastAckedTimestamp
				c.UnackedMessages = consumer.UnackedMessages
				c.AvailablePermits = consumer.AvailablePermits
				consumers = append(consumers, c)
			}
			subscription.Consumers = consumers
			subscriptions[k] = subscription
		}
	}
	parsedStats.Subscriptions = subscriptions
	return parsedStats
}

func (r *Connector) AddPartitionsOfTopic(topic *nlptv1.Topic) error {
	request := r.GetHttpRequest()
	url := r.getUrl(topic)
	response, body, errs := request.Post(url).Send(topic.Spec.PartitionNum).Retry(3, 5*time.Second).End()
	if response.StatusCode == 204 {
		return nil
	}
	return fmt.Errorf("Increment partitions error: %+v; %+v ", body, errs)
}

func (r *Connector) isNamespacesExist(topic *nlptv1.Topic) (bool, error) {
	request := r.GetHttpRequest()
	url := r.getNamespaceUrl(topic)

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
