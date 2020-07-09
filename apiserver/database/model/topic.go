package model

import (
	"encoding/json"
	"fmt"

	"github.com/chinamobile/nlpt/crds/topic/api/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Topic struct {
	Id         string `orm:"pk;unique"`
	Namespace  string
	Name       string
	TopicGroup string
	Status     string
	Raw        string `orm:"type(text)"`
}

func (*Topic) TableName() string {
	return "topic"
}

func (*Topic) ResourceType() string {
	return topicType
}

func (a *Topic) ResourceId() string {
	return a.Id
}

const topicType = "topic"

func TopicFromTopic(api *v1.Topic) (Topic, []UserRelation, error) {
	raw, err := json.Marshal(api)
	if err != nil {
		return Topic{}, nil, fmt.Errorf("marshal crd v1.topic error: %+v", err)
	}
	if api.ObjectMeta.Labels == nil {
		return Topic{}, nil, fmt.Errorf("topic labels is null")
	}
	rls := FromUser(topicType, api.ObjectMeta.Name, api.ObjectMeta.Labels)
	return Topic{
		Id:         api.ObjectMeta.Name,
		Namespace:  api.ObjectMeta.Namespace,
		Name:       api.Spec.Name,
		TopicGroup: api.Spec.TopicGroup,
		Status:     string(api.Status.Status),

		Raw: string(raw),
	}, rls, nil
}

func TopicToTopic(a Topic) (*v1.Topic, error) {
	api := &v1.Topic{}
	err := json.Unmarshal([]byte(a.Raw), api)
	if err != nil {
		return nil, fmt.Errorf("unmarshal crd v1.application error: %+v", err)
	}
	return api, nil
}

func TopicGetFromObject(obj interface{}) (Topic, []UserRelation, error) {
	un, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return Topic{}, nil, fmt.Errorf("cannot cast obj %+v to unstructured", obj)
	}
	api := &v1.Topic{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), api); err != nil {
		return Topic{}, nil, fmt.Errorf("cannot convert from unstructured: %+v", err)
	}
	return TopicFromTopic(api)
}
