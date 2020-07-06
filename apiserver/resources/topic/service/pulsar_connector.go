package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	v1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
	"net/http"
	"time"
)

const (
	skipAllUrl          = "/admin/v2/%s/%s/%s/%s/subscription/%s/skip_all"
	skillNMessage       = "/admin/v2/%s/%s/%s/%s/subscription/%s/skip/%d"
	getStatsUrl         = "/admin/v2/%s/%s/%s/%s/stats"
	partitionedStatsUrl = "/admin/v2/%s/%s/%s/%s/partitioned-stats"
)

//Connector 定义连接Pulsar所需要的参数
type Connector struct {
	Host           string
	Port           int
	AuthEnable     bool
	SuperUserToken string
}

func NewConnector(tpConfig *config.TopicConfig) *Connector {
	return &Connector{
		Host:           tpConfig.Host,
		Port:           tpConfig.HttpPort,
		AuthEnable:     tpConfig.AuthEnable,
		SuperUserToken: tpConfig.AdminToken,
	}
}
func (r *Connector) getHttpRequest() *gorequest.SuperAgent {
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)
	return r.addTokenToHeader(request)

}

func (r *Connector) addTokenToHeader(request *gorequest.SuperAgent) *gorequest.SuperAgent {
	if r.AuthEnable {
		request.Header.Set("Authorization", "Bearer "+r.SuperUserToken)
	}

	klog.Errorf("%+v", request)
	return request
}

func (r *Connector) SkipAllMessages(tp *v1.Topic, subscriptionName string) error {

	var domain = "persistent"
	if !tp.Spec.Persistent {
		domain = "non-persistent"
	}

	url := fmt.Sprintf(skipAllUrl, domain, tp.Namespace, tp.Spec.TopicGroup, tp.Spec.Name, subscriptionName)
	url = fmt.Sprintf("%s://%s:%d%s", "http", r.Host, r.Port, url)

	response, body, errs := r.getHttpRequest().Post(url).End()
	if errs != nil {
		return fmt.Errorf("request(%s) failed, error: %+v", url, errs)
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("request(%s) failed, response: %+v, body: %+v, error: %+v", url, response, body, errs)
	}

	return nil
}

func (r *Connector) SkipMessages(tp *v1.Topic, subscriptionName string, numMessage int64) error {

	var domain = "persistent"
	if !tp.Spec.Persistent {
		domain = "non-persistent"
	}

	url := fmt.Sprintf(skillNMessage, domain, tp.Namespace, tp.Spec.TopicGroup, tp.Spec.Name, subscriptionName, numMessage)
	url = fmt.Sprintf("%s://%s:%d%s", "http", r.Host, r.Port, url)

	response, body, errs := r.getHttpRequest().Post(url).End()
	if errs != nil {
		return fmt.Errorf("request(%s) failed, error: %+v", url, errs)
	}

	if response.StatusCode != http.StatusNoContent {
		return fmt.Errorf("request(%s) failed, response: %+v, body: %+v, error: %+v", url, response, body, errs)
	}

	return nil
}

func (r *Connector) GetStats(tp *v1.Topic) (*v1.Stats, error) {
	request := r.getHttpRequest()
	var domain = "persistent"
	if !tp.Spec.Persistent {
		domain = "non-persistent"
	}

	var url string
	if tp.Spec.Partitioned {
		url = partitionedStatsUrl
	} else {
		url = getStatsUrl
	}
	url = fmt.Sprintf(url, domain, tp.Namespace, tp.Spec.TopicGroup, tp.Spec.Name)
	url = fmt.Sprintf("%s://%s:%d%s", "http", r.Host, r.Port, url)
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

func (r *Connector) FormatStats(stats *Stats) *v1.Stats {
	parsedStats := &v1.Stats{
		DeduplicationStatus: stats.DeduplicationStatus,
		MsgInCounter:        stats.MsgInCounter,
		BytesInCounter:      stats.BytesInCounter,
		StorageSize:         stats.StorageSize,
		BacklogSize:         stats.BacklogSize,
	}

	if stats.Partitions != nil {
		parsedStats.Partitions = make(map[string]v1.Stats, 0)
		for k, v := range stats.Partitions {
			parsedStats.Partitions[k] = *r.FormatStats(&v)
		}
	}

	parsedStats.MsgRateIn = fmt.Sprintf("%.3f", stats.MsgRateIn)
	parsedStats.MsgRateOut = fmt.Sprintf("%.3f", stats.MsgRateOut)
	parsedStats.MsgThroughputOut = fmt.Sprintf("%.3f", stats.MsgThroughputOut)
	parsedStats.AverageMsgSize = fmt.Sprintf("%.3f", stats.AverageMsgSize)
	parsedStats.MsgThroughputIn = fmt.Sprintf("%.3f", stats.MsgThroughputIn)

	var subscriptions = make(map[string]v1.SubscriptionStat)
	if len(stats.Subscriptions) > 0 {
		for k, v := range stats.Subscriptions {
			var subscription v1.SubscriptionStat
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

			var consumers = make([]v1.ConsumerStat, 0)
			for _, consumer := range v.Consumers {
				var c v1.ConsumerStat
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
