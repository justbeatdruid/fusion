package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/crds/serviceunitgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type ServiceunitGroup struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// only used in creation options
func ToAPI(app *ServiceunitGroup) *v1.ServiceunitGroup {
	crd := &v1.ServiceunitGroup{}
	crd.TypeMeta.Kind = "ServiceunitGroup"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.ServiceunitGroupSpec{
		Name: app.Name,
	}
	return crd
}

func ToModel(obj *v1.ServiceunitGroup) *ServiceunitGroup {
	return &ServiceunitGroup{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,
	}
}

func ToListModel(items *v1.ServiceunitGroupList) []*ServiceunitGroup {
	var app []*ServiceunitGroup = make([]*ServiceunitGroup, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (a *ServiceunitGroup) Validate() error {
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
