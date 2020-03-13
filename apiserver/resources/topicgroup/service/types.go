package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

const (
	MIN_MESSAGE_TTL = -1
	producer_request_hold = "producer_request_hold"
	producer_exception = "producer_exception"
	consumer_backlog_eviction = "consumer_backlog_eviction"
)
type Topicgroup struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"` //namespace名称
	Namespace string    `json:"namespace"`
	Tenant    string    `json:"tenant"`             //namespace的所属租户名称
	Policies  Policies  `json:"policies,omitempty"` //namespace的策略
	CreatedAt  int64     `json:createdAt`  //创建时间
	Status    v1.Status `json:"status"`
	Message   string    `json:"message"`
}

type Policies struct {
	RetentionPolicies   RetentionPolicies `json:"retentionPolicies,omitempty"` //消息保留策略
	MessageTtlInSeconds int               `json:"messageTtlInSeconds"`         //未确认消息的最长保留时长
	BacklogQuota        BacklogQuota      `json:"backlogQuota,omitempty"`
	NumBundles          int               `json:"numBundles"`
}

type RetentionPolicies struct {
	RetentionTimeInMinutes int `json:"retentionTimeInMinutes"`
	RetentionSizeInMB      int `json:"retentionSizeInMB"`
}

type BacklogQuota struct {
	Limit  int64  `json:"limit"`  //未确认消息的积压大小
	Policy string `json:"policy"` //producer_request_hold,producer_exception,consumer_backlog_eviction

}

// only used in creation options
func ToAPI(app *Topicgroup) *v1.Topicgroup {
	crd := &v1.Topicgroup{}
	crd.TypeMeta.Kind = "Topicgroup"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace

	crd.Spec = v1.TopicgroupSpec{
		Name:      app.Name,
		Tenant:    app.Tenant,
		Namespace: app.Namespace,
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Init
	}
	crd.Status = v1.TopicgroupStatus{
		Status:  status,
		Message: app.Message,
	}
	return crd
}

func ToModel(obj *v1.Topicgroup) *Topicgroup {
	return &Topicgroup{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,
		Tenant:    obj.Spec.Tenant,
		Status:    obj.Status.Status,
		Message:   obj.Status.Message,
		CreatedAt: obj.ObjectMeta.CreationTimestamp.Unix(),
	}
}

func ToListModel(items *v1.TopicgroupList) []*Topicgroup {
	var app []*Topicgroup = make([]*Topicgroup, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
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
	if p != (Policies{}) {
		if p.MessageTtlInSeconds < MIN_MESSAGE_TTL {
			return fmt.Errorf("messageTtlInSeconds is invalid: %d", p.MessageTtlInSeconds)
		}

		if p.RetentionPolicies.RetentionTimeInMinutes < -1 {
			return fmt.Errorf("retentionTimeInMinutes is invalid: %d", p.RetentionPolicies.RetentionTimeInMinutes)
		}

		if p.RetentionPolicies.RetentionSizeInMB < -1 {
			return fmt.Errorf("retentionTimeInMinutes is invalid: %d", p.RetentionPolicies.RetentionSizeInMB)
		}


		if p.NumBundles <= 0 {
			return fmt.Errorf("numBundles is invalid: %d", p.NumBundles)
		}

		if p.BacklogQuota != (BacklogQuota{}) {
			switch p.BacklogQuota.Policy {
			case producer_request_hold:
			case consumer_backlog_eviction:
			case producer_exception:
				break;
			default:
				return fmt.Errorf("backlogQuota policy is invalid: %s", p.BacklogQuota.Policy)
			}
		}
	} else {
		//如果此参数未填，则返回默认值
		a.Policies = Policies{
			RetentionPolicies:   RetentionPolicies{
				RetentionTimeInMinutes:0,
				RetentionSizeInMB:0,
			},
			MessageTtlInSeconds: -1,
			BacklogQuota:        BacklogQuota{
				Limit:-1,
				Policy:producer_request_hold,
			},
			NumBundles:          4,
		}

	}

	a.ID = names.NewID()
	return nil
}
