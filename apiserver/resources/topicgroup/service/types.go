package service

import (
	"fmt"
	pulsar "github.com/chinamobile/nlpt/apiserver/resources/topicgroup/pulsar"
	"github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/names"
)

const (
	producer_request_hold     = "producer_request_hold"
	producer_exception        = "producer_exception"
	consumer_backlog_eviction = "consumer_backlog_eviction"
	NotSet                    = -12323344
	NotSetString              = "NotSet"
	DefaultTenant             = "public" //暂时默认为public租户
)

type Topicgroup struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"` //namespace名称
	Namespace string     `json:"namespace"`
	Tenant    string     `json:"tenant"`             //namespace的所属租户名称
	Policies  Policies   `json:"policies,omitempty"` //namespace的策略
	CreatedAt int64      `json:"createdAt"`          //创建时间
	Users     user.Users `json:"users"`
	Status    v1.Status  `json:"status"`
	Message   string     `json:"message"`
	Available bool       `json:"available"` //是否可用
}

type Policies struct {
	RetentionPolicies   RetentionPolicies `json:"retentionPolicies,omitempty"` //消息保留策略
	MessageTtlInSeconds int               `json:"messageTtlInSeconds"`         //未确认消息的最长保留时长
	BacklogQuota        BacklogQuota      `json:"backlogQuota,omitempty"`
	NumBundles          int               `json:"numBundles"`
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
	if fillDefaultValue {
		//如果此参数未填，则返回默认值
		return &Policies{
			RetentionPolicies: RetentionPolicies{
				RetentionTimeInMinutes: pulsar.DefaultRetentionTimeInMinutes,
				RetentionSizeInMB:      pulsar.DefaultRetentionSizeInMB,
			},
			MessageTtlInSeconds: pulsar.DefaultMessageTTlInSeconds,
			BacklogQuota: BacklogQuota{
				Limit:  -1,
				Policy: producer_request_hold,
			},
			NumBundles: pulsar.DefaultNumberOfNamespaceBundles,
		}
	} else {
		return &Policies{
			RetentionPolicies: RetentionPolicies{
				RetentionSizeInMB:      NotSet,
				RetentionTimeInMinutes: NotSet,
			},
			MessageTtlInSeconds: NotSet,
			BacklogQuota: BacklogQuota{
				Limit:  NotSet,
				Policy: NotSetString,
			},
			NumBundles: NotSet,
		}
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
		Name:      app.Name,
		Tenant:    app.Tenant,
		Policies:  ToPolicesApi(&app.Policies),
	}

	if len(crd.Spec.Tenant) == 0 {
		crd.Spec.Tenant = DefaultTenant
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
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,
		Tenant:    obj.Spec.Tenant,
		Status:    obj.Status.Status,
		Message:   obj.Status.Message,
		CreatedAt: obj.ObjectMeta.CreationTimestamp.Unix(),
		Policies:  ToPolicesModel(&obj.Spec.Policies),
		Users:     user.GetUsersFromLabels(obj.ObjectMeta.Labels),
		Available: obj.Spec.Available,
	}
}

func ToPolicesModel(obj *v1.Policies) Policies {
	return Policies{
		RetentionPolicies: RetentionPolicies{
			RetentionSizeInMB:      obj.RetentionPolicies.RetentionSizeInMB,
			RetentionTimeInMinutes: obj.RetentionPolicies.RetentionTimeInMinutes,
		},
		MessageTtlInSeconds: obj.MessageTtlInSeconds,
		BacklogQuota: BacklogQuota{
			Limit:  obj.BacklogQuota.Limit,
			Policy: obj.BacklogQuota.Policy,
		},
		NumBundles: obj.NumBundles,
	}
}
func ToPolicesApi(policies *Policies) v1.Policies {
	if policies == nil {
		return v1.Policies{}
	}
	crd := v1.Policies{
		NumBundles:          policies.NumBundles,
		MessageTtlInSeconds: policies.MessageTtlInSeconds,
		RetentionPolicies: v1.RetentionPolicies{
			RetentionTimeInMinutes: policies.RetentionPolicies.RetentionTimeInMinutes,
			RetentionSizeInMB:      policies.RetentionPolicies.RetentionSizeInMB,
		},
		BacklogQuota: v1.BacklogQuota{
			Limit:  policies.BacklogQuota.Limit,
			Policy: policies.BacklogQuota.Policy,
		},
	}

	if policies.NumBundles <= 0 {
		crd.NumBundles = pulsar.DefaultNumberOfNamespaceBundles
	}

	if policies.MessageTtlInSeconds < 0 {
		crd.MessageTtlInSeconds = pulsar.DefaultMessageTTlInSeconds
	}

	if len(policies.BacklogQuota.Policy) == 0 {
		crd.BacklogQuota.Policy = pulsar.DefaultBacklogPolicy
	}

	if policies.RetentionPolicies.RetentionTimeInMinutes < 0 {
		crd.RetentionPolicies.RetentionTimeInMinutes = pulsar.DefaultRetentionTimeInMinutes
	}

	if policies.RetentionPolicies.RetentionSizeInMB < 0 {
		crd.RetentionPolicies.RetentionSizeInMB = pulsar.DefaultRetentionSizeInMB
	}
	return crd
}

func ToListModel(items *v1.TopicgroupList) []*Topicgroup {
	var app []*Topicgroup = make([]*Topicgroup, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (p Policies) Validate() error {
	//TODO 参数校验待验证
	if p != (Policies{}) {
		if p.MessageTtlInSeconds < pulsar.DefaultMessageTTlInSeconds && p.MessageTtlInSeconds != NotSet {
			return fmt.Errorf("messageTtlInSeconds is invalid: %d", p.MessageTtlInSeconds)
		}

		if p.RetentionPolicies.RetentionTimeInMinutes < pulsar.MinRetentionTimeInMinutes && p.RetentionPolicies.RetentionTimeInMinutes != NotSet {
			return fmt.Errorf("retentionTimeInMinutes is invalid: %d", p.RetentionPolicies.RetentionTimeInMinutes)
		}

		if p.RetentionPolicies.RetentionSizeInMB < pulsar.MinRetentionSizeInMB && p.RetentionPolicies.RetentionTimeInMinutes != NotSet {
			return fmt.Errorf("retentionTimeInMinutes is invalid: %d", p.RetentionPolicies.RetentionSizeInMB)
		}

		if p.NumBundles <= 0 && p.NumBundles != NotSet {
			return fmt.Errorf("numBundles is invalid: %d", p.NumBundles)
		}

		if p.BacklogQuota != (BacklogQuota{}) {
			switch p.BacklogQuota.Policy {
			case producer_request_hold:
			case consumer_backlog_eviction:
			case producer_exception:
				break
			default:
				if p.BacklogQuota.Policy != NotSetString {
					return fmt.Errorf("backlogQuota policy is invalid: %s", p.BacklogQuota.Policy)
				}
			}
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
		}
	}

	p := a.Policies
	//TODO 参数校验待验证
	if err := p.Validate(); err != nil {
		return err
	}
	a.ID = names.NewID()
	return nil
}
