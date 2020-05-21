package model

import (
	"encoding/json"

	"github.com/chinamobile/nlpt/crds/application/api/v1"

	"k8s.io/klog"
)

type Application struct {
	Id        string `orm:"pk;unique"`
	Namespace string
	Name      string
	Group     string
	Status    string
	Raw       string `orm:"type(text)"`
}

const applicationType = "application"

func FromApi(api *v1.Application) (Application, []UserRelation) {
	raw, err := json.Marshal(api)
	if err != nil {
		klog.Errorf("marshal crd v1.application error: %+v", err)
		return Application{}, nil
	}
	if api.ObjectMeta.Labels == nil {
		klog.Errorf("application labels is null")
		return Application{}, nil
	}
	rls := FromUser(applicationType, api.ObjectMeta.Name, api.ObjectMeta.Labels)
	return Application{
		Id:        api.ObjectMeta.Name,
		Namespace: api.ObjectMeta.Namespace,
		Name:      api.Spec.Name,
		Group:     api.Spec.Group.ID,
		Status:    string(api.Status.Status),

		Raw: string(raw),
	}, rls
}

func ToApi(a Application) *v1.Application {
	api := &v1.Application{}
	err := json.Unmarshal([]byte(a.Raw), api)
	if err != nil {
		klog.Errorf("unmarshal crd v1.application error: %+v", err)
	}
	return api
}
