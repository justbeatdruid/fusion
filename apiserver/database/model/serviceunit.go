package model

import (
	"encoding/json"
	"fmt"

	"github.com/chinamobile/nlpt/crds/serviceunit/api/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Serviceunit struct {
	Id        string `orm:"pk;unique"`
	Namespace string
	Name      string
	Status    string
	Raw       string `orm:"type(text)"`
}

func (*Serviceunit) TableName() string {
	return "serviceunit"
}

func (*Serviceunit) ResourceType() string {
	return serviceunitType
}

func (a *Serviceunit) ResourceId() string {
	return a.Id
}

const serviceunitType = "serviceunit"

func ServiceunitFromServiceunit(api *v1.Serviceunit) (Serviceunit, []UserRelation, error) {
	raw, err := json.Marshal(api)
	if err != nil {
		return Serviceunit{}, nil, fmt.Errorf("marshal crd v1.serviceunit error: %+v", err)
	}
	if api.ObjectMeta.Labels == nil {
		return Serviceunit{}, nil, fmt.Errorf("serviceunit labels is null")
	}
	rls := FromUser(serviceunitType, api.ObjectMeta.Name, api.ObjectMeta.Labels)
	return Serviceunit{
		Id:        api.ObjectMeta.Name,
		Namespace: api.ObjectMeta.Namespace,
		Name:      api.Spec.Name,
		Status:    string(api.Status.Status),

		Raw: string(raw),
	}, rls, nil
}

func ServiceunitToServiceunit(a Serviceunit) (*v1.Serviceunit, error) {
	api := &v1.Serviceunit{}
	err := json.Unmarshal([]byte(a.Raw), api)
	if err != nil {
		return nil, fmt.Errorf("unmarshal crd v1.application error: %+v", err)
	}
	return api, nil
}

func ServiceunitGetFromObject(obj interface{}) (Serviceunit, []UserRelation, error) {
	un, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return Serviceunit{}, nil, fmt.Errorf("cannot cast obj %+v to unstructured", obj)
	}
	api := &v1.Serviceunit{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), api); err != nil {
		return Serviceunit{}, nil, fmt.Errorf("cannot convert from unstructured: %+v", err)
	}
	return ServiceunitFromServiceunit(api)
}
