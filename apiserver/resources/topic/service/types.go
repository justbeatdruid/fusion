package service

import (
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	topicerr "github.com/chinamobile/nlpt/apiserver/resources/topic/error"
	"github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"
	"strings"
)

const (
	DefaultTenant    = "public"
	DefaultNamespace = "default"
	Separator        = "/"
)

type Topic struct {
	ID              string       `json:"id"`
	Name            string       `json:"name"` //topic名称
	Namespace       string       `json:"namespace"`
	Tenant          string       `json:"tenant"`          //topic的所属租户名称
	TopicGroup      string       `json:"topicGroup"`      //topic所属分组ID
	Partition       int          `json:"partition"`       //topic的分区数量，不指定时默认为1，指定partition大于1，则该topic的消息会被多个broker处理
	IsNonPersistent bool         `json:"isNonPersistent"` //非持久化，默认为false，非必填topic
	URL             string       `json:"url"`             //URL
	CreatedAt       int64        `json:"createdAt"`       //创建Topic的时间戳
	Status          v1.Status    `json:"status"`
	Message         string       `json:"message"`
	Permissions     []Permission `json:"permissions"`
	Users           user.Users   `json:"users"`
	MessageSize     float64      `json:"messageSize"` //消息总量
}

type Message struct {
	TopicName string           `json:"topicName"`
	ID        pulsar.MessageID `json:"id"`
	Time      util.Time        `json:"time"`
	Messages  string           `json:"messages"`
}

type Actions []string

type Permission struct {
	AuthUserID   string  `json:"authUserId"`   //对应clientauth的ID
	AuthUserName string  `json:"authUserName"` //对应clientauth的NAME
	Actions      Actions `json:"actions"`      //授权的操作：发布、订阅或者发布+订阅
	Status       string  `json:"status"`       //用户的授权状态，已授权、待删除、待授权
}

type Statistics struct {
	Total        int `json:"total"`
	Increment    int `json:"increment"`
	TotalMessage int `json:"totalMessage"`
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
	crd.ObjectMeta.Namespace = crdNamespace

	crd.Spec = v1.TopicSpec{
		Name:            app.Name,
		Tenant:          app.Tenant,
		Namespace:       app.Namespace,
		TopicGroup:      app.TopicGroup,
		Partition:       app.Partition,
		IsNonPersistent: app.IsNonPersistent,
	}

	if crd.Spec.Partition <= 0 {
		crd.Spec.Partition = 1
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
		}
		ps = append(ps, p)
	}

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
		CreatedAt:       obj.CreationTimestamp.Unix(),
		Permissions:     ps,
		Users:           user.GetUsersFromLabels(obj.ObjectMeta.Labels),
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
		"name": a.Name,
	} {
		if len(v) == 0 {
			return topicerr.TopicError{
				Err:       fmt.Errorf("%s is null", k),
				ErrorCode: topicerr.ErrorBadRequest,
			}
		}
	}
	a.ID = names.NewID()
	return topicerr.TopicError{
		Err: nil,
	}
}

func (a *Topic) GetUrl() (url string) {

	var build strings.Builder
	if a.IsNonPersistent {
		build.WriteString("non-persistent://")
	} else {
		build.WriteString("persistent://")
	}

	build.WriteString(a.Tenant)
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
