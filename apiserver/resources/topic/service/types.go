package service

import (
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	topicerr "github.com/chinamobile/nlpt/apiserver/resources/topic/error"
	"github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"
	"regexp"
	"strconv"
	"strings"
)

const (
	Separator         = "/"
	NameReg           = "^[-=:.a-z0-9]{1,100}$"
	MaxDescriptionLen = 1024
)

type Topic struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"` //topic名称
	Namespace    string       `json:"namespace"`
	TopicGroup   string       `json:"topicGroup"`   //topic所属分组ID
	PartitionNum int          `json:"partitionNum"` //topic的分区数量，partitioned为true时，需要指定。默认为1
	Partitioned  *bool        `json:"partitioned"`  //是否多分区，默认为false。true：代表多分区Topic
	Persistent   *bool        `json:"persistent"`   //是否持久化，默认为true，非必填
	URL          string       `json:"url"`          //URL
	CreatedAt    int64        `json:"createdAt"`    //创建Topic的时间戳
	Status       v1.Status    `json:"status"`
	Message      string       `json:"message"`
	Permissions  []Permission `json:"permissions"`
	Users        user.Users   `json:"users"`
	Stats        *Stats       `json:"stats"` //Topic的统计数据

	Description         string           `json:"description"`         //描述
	ShowStatus          v1.ShowStatus    `json:"displayStatus"`       //页面显示状态
	AuthorizationStatus string           `json:"authorizationStatus"` //用户授权状态
	Applications        []v1.Application `json:"applications"`        //绑定的应用列表
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

/**
  "msgRateOut": 0.0,
            "msgThroughputOut": 0.0,
            "msgRateRedeliver": 0.0,
            "consumerName": "lxj",
            "availablePermits": 989,
            "unackedMessages": 11,
            "blockedConsumerOnUnackedMsgs": false,
            "lastAckedTimestamp": 1591006348714,
            "lastConsumedTimestamp": 1591006348527,
            "metadata": {},
            "connectedSince": "2020-06-01T18:11:56.981+08:00",
            "address": "/10.233.91.230:55424"
*/
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

type PartitionedSubscriptionsInfo struct {
	PartitionNo       string            `json:"partitionNo"`
	SubscriptionsInfo SubscriptionsInfo `json:"subscriptionsInfo"`
}

type PartitionedSubscriptionsInfos []PartitionedSubscriptionsInfo
type SubscriptionsInfo struct {
	AverageMsgSize float64        `json:"averageMsgSize"`
	StorageSize    int64          `json:"storageSize"`
	BacklogSize    int64          `json:"backlogSize"`
	Subscriptions  []Subscription `json:"subscriptions"`
}

type Subscription struct {
	Name                             string         `json:"name"`
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
type Messages struct {
	ProducerName interface{} `json:"producerName"`
	ID           interface{} `json:"id"`
	Time         interface{} `json:"time"`
	Message      interface{} `json:"message"`
	Size         interface{} `json:"size"`
	Partition    interface{} `json:"partition"`
	Key          interface{} `json:"key"`
	Total        interface{} `json:"total"`
}
type Message struct {
	TopicName string           `json:"topicName"`
	ID        pulsar.MessageID `json:"id"`
	Time      util.Time        `json:"time"`
	Messages  string           `json:"messages"`
	Size      int              `json:"size"`
}

type Permission struct {
	AuthUserID   string        `json:"authUserId"`   //对应clientauth的ID
	AuthUserName string        `json:"authUserName"` //对应clientauth的NAME
	Actions      v1.Actions    `json:"actions"`      //授权的操作：发布、订阅或者发布+订阅
	Status       v1.Status     `json:"status"`       //用户的授权状态，已授权、待删除、待授权
	Token        string        `json:"token"`        //Token
	Effective    bool          `json:"effective"`
	IssuedAt     int64         `json:"issuedAt"`
	ExpireAt     int64         `json:"expireAt"`
	ShowStatus   v1.ShowStatus `json:"showStatus"`
	IsPermanent   bool           `json:"isPermanent"` //token是否永久有效
}

type Statistics struct {
	Total       int    `json:"total"`
	Increment   int    `json:"increment"`
	MessageSize string `json:"MessageSize"`
}

type BindInfo struct {
	ID      string     `json:"id"`
	Actions v1.Actions `json:"actions"`
}
type SendMessages struct {
	ID          string `json:"id"`
	Key         string `json:"tag"`
	MessageBody string `json:"messageBody"`
}
type ResetPosition struct {
	ID        string `json:"id"`      //topicId
	SubName   string `json:"subName"` //订阅者的名称
	MessageId string `json:"messageId"`
	Timestamp int64  `json:"timestamp"` //以ms为单位
}
type GrantPermissions struct {
	ID      string     `json:"id"`      //认证用户的id
	Actions v1.Actions `json:"actions"` //认证用户的权限
}

const (
	Consume = "consume"
	Produce = "produce"
)

// only used in creation options
func ToAPI(app *Topic) *v1.Topic {
	crd := &v1.Topic{}
	crd.TypeMeta.Kind = "Topic"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = app.Namespace

	crd.Spec = v1.TopicSpec{
		Name:             app.Name,
		TopicGroup:       app.TopicGroup,
		PartitionNum:     app.PartitionNum,
		Description:      app.Description,
		Permissions:      make([]v1.Permission, 0),
		PartitionedStats: make([]v1.PartitionedStats, 0),
		Applications:     make(map[string]v1.Application),
		Stats:            v1.Stats{},
	}

	if app.Persistent != nil {
		crd.Spec.Persistent = *app.Persistent
	}

	if app.Partitioned != nil {
		crd.Spec.Partitioned = *app.Partitioned
	}

	status := app.Status
	if len(status) == 0 {
		status = v1.Creating
	}

	crd.Status = v1.TopicStatus{
		Status:  status,
		Message: app.Message,
	}

	crd.ObjectMeta.Labels = user.AddUsersLabels(app.Users, crd.ObjectMeta.Labels)
	return crd
}

func ToModel(obj *v1.Topic) *Topic {
	var ps []Permission
	for _, psm := range obj.Spec.Permissions {
		var acs []string
		for _, ac := range psm.Actions {
			acs = append(acs, ac)
		}
		p := Permission{
			AuthUserID:   psm.AuthUserID,
			AuthUserName: psm.AuthUserName,
			Actions:      acs,
			Status:       psm.Status.Status,
			ShowStatus:   v1.ShowStatusMap[psm.Status.Status],
		}
		ps = append(ps, p)
	}

	apps := make([]v1.Application, 0)

	if obj.Spec.Applications != nil {
		for _, v := range obj.Spec.Applications {
			v.DisplayStatus = v1.ShowStatusMap[v.Status]
			apps = append(apps, v)
		}
	}

	return &Topic{
		ID:           obj.ObjectMeta.Name,
		Name:         obj.Spec.Name,
		Namespace:    obj.ObjectMeta.Namespace,
		TopicGroup:   obj.Spec.TopicGroup,
		Persistent:   &obj.Spec.Persistent,
		Partitioned:  &obj.Spec.Partitioned,
		PartitionNum: obj.Spec.PartitionNum,
		Status:       obj.Status.Status,
		Message:      obj.Status.Message,
		URL:          obj.Spec.Url,
		CreatedAt:    obj.CreationTimestamp.Unix(),
		Permissions:  ps,
		Users:        user.GetUsersFromLabels(obj.ObjectMeta.Labels),
		Stats:        ToStatsModel(obj.Spec.Stats),
		Description:  obj.Spec.Description,
		ShowStatus:   v1.ShowStatusMap[obj.Status.Status],
		Applications: apps,
	}

}

func ToSubscriptionStatModel(obj v1.SubscriptionStat) SubscriptionStat {
	var consumers []ConsumerStat
	if obj.Consumers != nil {
		consumers = make([]ConsumerStat, 0)
		for _, c := range obj.Consumers {
			consumers = append(consumers, ToConsumersModel(c))
		}
	}

	return SubscriptionStat{
		MsgRateOut:                       ParseFloat(obj.MsgRateOut),
		MsgThroughputOut:                 ParseFloat(obj.MsgThroughputOut),
		MsgRateRedeliver:                 ParseFloat(obj.MsgRateRedeliver),
		MsgBacklog:                       obj.MsgBacklog,
		BlockedSubscriptionOnUnackedMsgs: obj.BlockedSubscriptionOnUnackedMsgs,
		MsgDelayed:                       obj.MsgDelayed,
		UnackedMessages:                  obj.UnackedMessages,
		Type:                             obj.Type,
		MsgRateExpired:                   ParseFloat(obj.MsgRateExpired),
		LastExpireTimestamp:              obj.LastExpireTimestamp,
		LastConsumedFlowTimestamp:        obj.LastConsumedFlowTimestamp,
		LastConsumedTimestamp:            obj.LastConsumedTimestamp,
		LastAckedTimestamp:               obj.LastAckedTimestamp,
		Consumers:                        consumers,
		IsReplicated:                     obj.IsReplicated,
	}
}

func ParseFloat(s string) float64 {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return value
}

func ToConsumersModel(obj v1.ConsumerStat) ConsumerStat {
	return ConsumerStat{
		MsgRateOut:            ParseFloat(obj.MsgRateOut),
		MsgThroughputOut:      ParseFloat(obj.MsgThroughputOut),
		ConnectedSince:        obj.ConnectedSince,
		ConsumerName:          obj.ConsumerName,
		AvailablePermits:      obj.AvailablePermits,
		Address:               obj.Address,
		UnackedMessages:       obj.UnackedMessages,
		LastAckedTimestamp:    obj.LastAckedTimestamp,
		LastConsumedTimestamp: obj.LastConsumedTimestamp,
	}
}

func ToStatsModel(obj v1.Stats) *Stats {
	var subscriptions = make(map[string]SubscriptionStat)

	if obj.Subscriptions != nil {
		for k, v := range obj.Subscriptions {
			subscriptions[k] = ToSubscriptionStatModel(v)
		}
	}

	pStats := make(map[string]Stats, 0)
	if obj.Partitions != nil {
		for k, v := range obj.Partitions {
			pStats[k] = *ToStatsModel(v)
		}
	}

	return &Stats{
		MsgRateIn:           ParseFloat(obj.MsgRateIn),
		MsgRateOut:          ParseFloat(obj.MsgRateOut),
		MsgThroughputIn:     ParseFloat(obj.MsgThroughputIn),
		MsgThroughputOut:    ParseFloat(obj.MsgThroughputOut),
		MsgInCounter:        obj.MsgInCounter,
		BytesInCounter:      obj.BytesInCounter,
		StorageSize:         obj.StorageSize,
		BacklogSize:         obj.BacklogSize,
		DeduplicationStatus: obj.DeduplicationStatus,
		Subscriptions:       subscriptions,
		Partitions:          pStats,
	}
}
func ToListModel(items *v1.TopicList) []*Topic {
	var app []*Topic = make([]*Topic, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (a *Topic) Validate() topicerr.TopicError {
	for k, v := range map[string]string{
		"name":       a.Name,
		"topicGroup": a.TopicGroup,
	} {
		if len(v) == 0 {
			return topicerr.TopicError{
				Err:       fmt.Errorf("%s is null", k),
				ErrorCode: topicerr.ErrorBadRequest,
			}
		}

	}

	if ok, _ := regexp.MatchString(NameReg, a.Name); !ok {
		return topicerr.TopicError{
			Err:       fmt.Errorf("name is illegal: %v ", a.Name),
			ErrorCode: topicerr.ErrorCreateTopic,
		}
	}

	if a.Persistent == nil {
		return topicerr.TopicError{
			Err:       fmt.Errorf("%s is null", "persistent"),
			ErrorCode: topicerr.ErrorBadRequest,
		}
	}

	if a.Partitioned == nil {
		return topicerr.TopicError{
			Err:       fmt.Errorf("%s is null", "partitioned"),
			ErrorCode: topicerr.ErrorBadRequest,
		}
	}

	if *a.Partitioned {
		if a.PartitionNum <= 0 || a.PartitionNum > 20 {
			return topicerr.TopicError{
				Err:       fmt.Errorf("parition number of partitioned topic must be greater than 0 and less than 20"),
				ErrorCode: topicerr.ErrorPartitionTopicPartitionEqualZero,
			}
		}
	}

	if len([]rune(a.Description)) > MaxDescriptionLen {
		return topicerr.TopicError{
			Err:       fmt.Errorf("description is not valid"),
			ErrorCode: topicerr.ErrorBadRequest,
		}
	}
	a.ID = names.NewID()
	return topicerr.TopicError{
		Err: nil,
	}
}

func (a *Topic) GetUrl() (url string) {

	var build strings.Builder
	if *a.Persistent {
		build.WriteString("non-persistent://")
	} else {
		build.WriteString("persistent://")
	}

	build.WriteString(a.Namespace)
	build.WriteString(Separator)
	build.WriteString(a.TopicGroup)
	build.WriteString(Separator)
	build.WriteString(a.Name)

	return build.String()
}

func (p *Permission) Validate() error {
	for k, v := range map[string]string{
		"authUserId": p.AuthUserID,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}

	for _, a := range p.Actions {
		if a != Consume && a != Produce {
			return fmt.Errorf("action:%s is invalid", a)
		}
	}

	return nil
}

func (a *Topic) ToPartitionedSubscriptionsModel() PartitionedSubscriptionsInfos {
	if a.Stats.Partitions == nil {
		return nil
	}

	pSubInfos := make([]PartitionedSubscriptionsInfo, 0)
	for k, v := range a.Stats.Partitions {
		partition := PartitionedSubscriptionsInfo{}
		startIndex := strings.LastIndex(k, "-")
		partition.PartitionNo = k[startIndex+1 : len(k)]
		partition.SubscriptionsInfo = *toSubscription(&v)
		pSubInfos = append(pSubInfos, partition)
	}
	return pSubInfos
}
func (a *Topic) ToSubscriptionsModel() *SubscriptionsInfo {
	if a.Stats == nil {
		return nil
	}

	return toSubscription(a.Stats)
}

func toSubscription(stats *Stats) *SubscriptionsInfo {
	subs := &SubscriptionsInfo{
		AverageMsgSize: stats.AverageMsgSize,
		StorageSize:    stats.StorageSize,
		BacklogSize:    stats.BacklogSize,
	}

	if stats.Subscriptions != nil {
		subs.Subscriptions = make([]Subscription, 0)
		for k, v := range stats.Subscriptions {
			sub := Subscription{}
			sub.Name = k
			sub.LastConsumedTimestamp = v.LastConsumedTimestamp
			sub.LastAckedTimestamp = v.LastAckedTimestamp
			sub.UnackedMessages = v.UnackedMessages
			sub.MsgThroughputOut = v.MsgThroughputOut
			sub.MsgRateOut = v.MsgRateOut
			sub.Consumers = v.Consumers
			sub.Type = v.Type
			sub.MsgBacklog = v.MsgBacklog
			sub.BlockedSubscriptionOnUnackedMsgs = v.BlockedSubscriptionOnUnackedMsgs
			sub.MsgDelayed = v.MsgDelayed
			sub.MsgRateExpired = v.MsgRateExpired
			sub.MsgRateRedeliver = v.MsgRateRedeliver
			sub.MsgThroughputOut = v.MsgThroughputOut
			sub.IsReplicated = v.IsReplicated
			subs.Subscriptions = append(subs.Subscriptions, sub)
		}
	}
	return subs
}
