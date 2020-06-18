package model

import (
	"encoding/json"
	"fmt"

	"github.com/chinamobile/nlpt/crds/application/api/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Application struct {
	Id        string `orm:"pk;unique"`
	Namespace string
	Name      string
	Group     string
	Status    string
	Raw       string `orm:"type(text)"`
}

func (*Application) TableName() string {
	return "application"
}

func (*Application) ResourceType() string {
	return applicationType
}

func (a *Application) ResourceId() string {
	return a.Id
}

const applicationType = "application"

func ApplicationFromApi(api *v1.Application) (Application, []UserRelation, []Relation, error) {
	raw, err := json.Marshal(api)
	if err != nil {
		return Application{}, nil, nil, fmt.Errorf("marshal crd v1.application error: %+v", err)
	}
	if api.ObjectMeta.Labels == nil {
		return Application{}, nil, nil, fmt.Errorf("application labels is null")
	}
	rls := FromUser(applicationType, api.ObjectMeta.Name, api.ObjectMeta.Labels)
	relations := getApplicationRelation(api)
	return Application{
		Id:        api.ObjectMeta.Name,
		Namespace: api.ObjectMeta.Namespace,
		Name:      api.Spec.Name,
		Group:     api.Spec.Group.ID,
		Status:    string(api.Status.Status),

		Raw: string(raw),
	}, rls, relations, nil
}

func ApplicationToApi(a Application) (*v1.Application, error) {
	api := &v1.Application{}
	err := json.Unmarshal([]byte(a.Raw), api)
	if err != nil {
		return nil, fmt.Errorf("unmarshal crd v1.application error: %+v", err)
	}
	return api, nil
}

func ApplicationGetFromObject(obj interface{}) (Application, []UserRelation, []Relation, error) {
	un, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return Application{}, nil, nil, fmt.Errorf("cannot cast obj %+v to unstructured", obj)
	}
	api := &v1.Application{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), api); err != nil {
		return Application{}, nil, nil, fmt.Errorf("cannot convert from unstructured: %+v", err)
	}
	return ApplicationFromApi(api)
}

func getApplicationRelation(app *v1.Application) []Relation {
	result := make([]Relation, 0)
	for _, api := range app.Spec.APIs {
		result = append(result, Relation{
			SourceType: applicationType,
			SourceId:   app.ObjectMeta.Name,
			TargetType: apiType,
			TargetId:   api.ID,
		})
	}
	return result
}
