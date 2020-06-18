package model

import (
	"encoding/json"
	"fmt"

	"github.com/chinamobile/nlpt/crds/api/api/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Api struct {
	Id        string `orm:"pk;unique"`
	Namespace string
	Name      string
	Status    string
	Raw       string `orm:"type(text)"`
}

func (*Api) TableName() string {
	return "api"
}

func (*Api) ResourceType() string {
	return apiType
}

func (a *Api) ResourceId() string {
	return a.Id
}

const apiType = "api"

func ApiFromApi(api *v1.Api) (Api, []UserRelation, error) {
	raw, err := json.Marshal(api)
	if err != nil {
		return Api{}, nil, fmt.Errorf("marshal crd v1.api error: %+v", err)
	}
	if api.ObjectMeta.Labels == nil {
		return Api{}, nil, fmt.Errorf("api labels is null")
	}
	rls := FromUser(apiType, api.ObjectMeta.Name, api.ObjectMeta.Labels)
	return Api{
		Id:        api.ObjectMeta.Name,
		Namespace: api.ObjectMeta.Namespace,
		Name:      api.Spec.Name,
		Status:    string(api.Status.Status),

		Raw: string(raw),
	}, rls, nil
}

func ApiToApi(a Api) (*v1.Api, error) {
	api := &v1.Api{}
	err := json.Unmarshal([]byte(a.Raw), api)
	if err != nil {
		return nil, fmt.Errorf("unmarshal crd v1.api error: %+v", err)
	}
	return api, nil
}

func ApiGetFromObject(obj interface{}) (Api, []UserRelation, error) {
	un, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return Api{}, nil, fmt.Errorf("cannot cast obj %+v to unstructured", obj)
	}
	api := &v1.Api{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), api); err != nil {
		return Api{}, nil, fmt.Errorf("cannot convert from unstructured: %+v", err)
	}
	return ApiFromApi(api)
}
