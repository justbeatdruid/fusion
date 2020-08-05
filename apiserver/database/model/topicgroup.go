package model

import (
	"encoding/json"
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"strconv"
)

type TopicGroup struct {
	Id         string `orm:"pk;unique"`
	Namespace  string
	Name       string
	Available  string
	Status     string
	Raw        string `orm:"type(text)"`
	TopicsCount int
}

func (*TopicGroup) TableName() string {
	return "topicgroup"
}

func (*TopicGroup) ResourceType() string {
	return topicgroupType
}

func (a *TopicGroup) ResourceId() string {
	return a.Id
}

const topicgroupType = "topicgroup"

func TopicgroupFromApi(api *v1.Topicgroup) (TopicGroup, []UserRelation, []Relation, error) {
	raw, err := json.Marshal(api)
	if err != nil {
		return TopicGroup{}, nil, nil, fmt.Errorf("marshal crd v1.application error: %+v", err)
	}
	if api.ObjectMeta.Labels == nil {
		return TopicGroup{}, nil, nil, fmt.Errorf("application labels is null")
	}
	rls := FromUser(applicationType, api.ObjectMeta.Name, api.ObjectMeta.Labels)
	//TODO Toicgroup和Topic的关系
	//relations := getTopicgroupRelation(api)
	return TopicGroup{
		Id:        api.ObjectMeta.Name,
		Namespace: api.ObjectMeta.Namespace,
		Name:      api.Spec.Name,
		Status:    string(api.Status.Status),
		Available: strconv.FormatBool(api.Spec.Available),
		Raw:       string(raw),
	}, rls, nil, nil
}

func TopicgroupToApi(a TopicGroup) (*v1.Topicgroup, error) {
	api := &v1.Topicgroup{}
	err := json.Unmarshal([]byte(a.Raw), api)
	if err != nil {
		return nil, fmt.Errorf("unmarshal crd v1.application error: %+v", err)
	}
	return api, nil
}

func TopicgroupGetFromObject(obj interface{}) (TopicGroup, []UserRelation, []Relation, error) {
	un, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return TopicGroup{}, nil, nil, fmt.Errorf("cannot cast obj %+v to unstructured", obj)
	}
	api := &v1.Topicgroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), api); err != nil {
		return TopicGroup{}, nil, nil, fmt.Errorf("cannot convert from unstructured: %+v", err)
	}
	return TopicgroupFromApi(api)
}

//func insertRelation(app *v1.Topicgroup) []Relation {
//	result := make([]Relation, 0)
//	for _, api := range app.Spec. {
//		result = append(result, Relation{
//			SourceType: applicationType,
//			SourceId:   app.ObjectMeta.Name,
//			TargetType: apiType,
//			TargetId:   api.ID,
//		})
//	}
//	return result
//}
