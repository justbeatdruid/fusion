package service

import (
	"fmt"
	"github.com/apache/pulsar/pulsar-client-go/pulsar"
	"github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
	"strings"
	"time"
)

const (
	DefaultTenant    = "public"
	DefaultNamespace = "default"
)

type Topic struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"` //topic名称
	Namespace       string    `json:"namespace"`
	Tenant          string    `json:"tenant"` //topic的所属租户名称
	TopicGroup      string    `json:"topicGroup"`
	Partition       int       `json:"partition"`       //topic的分区数量，不指定时默认为1，指定partition大于1，则该topic的消息会被多个broker处理
	IsNonPersistent bool      `json:"isNonPersistent"` //非持久化，默认为false，非必填topic
	URL             string    `json:"url"`             //URL
	CreateTime      int64     `json:"createTime""`     //创建Topic的时间戳
	Status          v1.Status `json:"status"`
	Message         string    `json:"message"`
}

type Message struct {
	TopicName string           `json:"topicName"`
	ID        pulsar.MessageID `json:"id"`
	Time      time.Time        `json:"time"`
	Messages  string           `json:"messages"`
}

// only used in creation options
func ToAPI(app *Topic) *v1.Topic {
	crd := &v1.Topic{}
	crd.TypeMeta.Kind = "Topic"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace

	crd.Spec = v1.TopicSpec{
		Name:            app.Name,
		Tenant:          app.Tenant,
		Namespace:       app.Namespace,
		TopicGroup:      app.TopicGroup,
		Partition:       app.Partition,
		IsNonPersistent: app.IsNonPersistent,
		CreatTime:       app.CreateTime,
	}

	if len(crd.Spec.Tenant) == 0 {
		crd.Spec.Tenant = DefaultTenant
	}
	if len(crd.Spec.TopicGroup) == 0 {
		crd.Spec.TopicGroup = DefaultNamespace
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Init
	}

	crd.Status = v1.TopicStatus{
		Status:  status,
		Message: app.Message,
	}
	return crd
}

func ToModel(obj *v1.Topic) *Topic {
	return &Topic{
		ID:              obj.ObjectMeta.Name,
		Name:            obj.Spec.Name,
		Namespace:       obj.ObjectMeta.Namespace,
		Tenant:          obj.Spec.Tenant,
		TopicGroup:      obj.Spec.TopicGroup,
		IsNonPersistent: obj.Spec.IsNonPersistent,
		Partition:       obj.Spec.Partition,
		Status:          obj.Status.Status,
		Message:         obj.Status.Message,
		URL:             obj.Spec.Url,
	}
}

func ToListModel(items *v1.TopicList) []*Topic {
	var app []*Topic = make([]*Topic, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (a *Topic) Validate() error {
	for k, v := range map[string]string{
		"name": a.Name,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}

	a.ID = names.NewID()
	return nil
}

func (a *Topic) GetUrl() (url string) {

	var build strings.Builder
	if a.IsNonPersistent {
		build.WriteString("non-persistent://")
	} else {
		build.WriteString("persistent://")
	}

	build.WriteString(a.Tenant)
	build.WriteString("/")
	build.WriteString(a.TopicGroup)
	build.WriteString("/")
	build.WriteString(a.Name)

	return build.String()
}
