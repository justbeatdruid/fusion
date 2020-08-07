package model

import (
	"encoding/json"
	"fmt"

	"github.com/chinamobile/nlpt/crds/datasource/api/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Datasource struct {
	Id        string `orm:"pk;unique"`
	Namespace string
	Name      string
	User      string
	Status    string
	Type      string
	Raw       string `orm:"type(text)"`
}

func (*Datasource) TableName() string {
	return "datasource"
}

func (*Datasource) ResourceType() string {
	return datasourceType
}

func (a *Datasource) ResourceId() string {
	return a.Id
}

const datasourceType = "datasource"

func DatasourceFromApi(api *v1.Datasource) (Datasource, error) {
	raw, err := json.Marshal(api)
	if err != nil {
		return Datasource{}, fmt.Errorf("marshal crd v1.datasource error: %+v", err)
	}
	if api.ObjectMeta.Labels == nil {
		return Datasource{}, fmt.Errorf("datasource labels is null")
	}
	dbType := string(api.Spec.Type)
	if api.Spec.Type == v1.RDBType && api.Spec.RDB != nil {
		dbType = api.Spec.RDB.Type
	}
	return Datasource{
		Id:        api.ObjectMeta.Name,
		Namespace: api.ObjectMeta.Namespace,
		Name:      api.Spec.Name,
		User:      GetOwner(api.ObjectMeta.Labels),
		Type:      dbType,
		Status:    string(api.Status.Status),
		Raw:       string(raw),
	}, nil
}

func DatasourceToApi(a Datasource) (*v1.Datasource, error) {
	api := &v1.Datasource{}
	err := json.Unmarshal([]byte(a.Raw), api)
	if err != nil {
		return nil, fmt.Errorf("unmarshal crd v1.application error: %+v", err)
	}
	return api, nil
}

func DatasourceGetFromObject(obj interface{}) (Datasource, error) {
	un, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return Datasource{}, fmt.Errorf("cannot cast obj %+v to unstructured", obj)
	}
	api := &v1.Datasource{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), api); err != nil {
		return Datasource{}, fmt.Errorf("cannot convert from unstructured: %+v", err)
	}
	return DatasourceFromApi(api)
}
