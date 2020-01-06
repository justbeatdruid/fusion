package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/crds/applicationgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type ApplicationGroup struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// only used in creation options
func ToAPI(app *ApplicationGroup) *v1.ApplicationGroup {
	crd := &v1.ApplicationGroup{}
	crd.TypeMeta.Kind = "ApplicationGroup"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.ApplicationGroupSpec{
		Name: app.Name,
	}
	return crd
}

func ToModel(obj *v1.ApplicationGroup) *ApplicationGroup {
	return &ApplicationGroup{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,
	}
}

func ToListModel(items *v1.ApplicationGroupList) []*ApplicationGroup {
	var app []*ApplicationGroup = make([]*ApplicationGroup, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (a *ApplicationGroup) Validate() error {
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
