package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Topic struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// only used in creation options
func ToAPI(app *Topic) *v1.Topic {
	crd := &v1.Topic{}
	crd.TypeMeta.Kind = "Topic"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.TopicSpec{
		Name: app.Name,
	}
	return crd
}

func ToModel(obj *v1.Topic) *Topic {
	return &Topic{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,
	}
}

func ToListModel(items *v1.TopicList) []*Topic {
	var app []*Topic = make([]*Topic, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (a *Topic) Validate() error {
	for k, v := range map[string]string{
		"name": a.Name,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	a.ID = names.NewID()
	return nil
}
