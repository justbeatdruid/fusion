package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/application/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Application struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// only used in creation options
func ToAPI(app *Application) *v1.Application {
	crd := &v1.Application{}
	crd.TypeMeta.Kind = "Application"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.ApplicationSpec{
		Name: app.Name,
	}
	return crd
}

func ToModel(obj *v1.Application) *Application {
	return &Application{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,
	}
}

func ToListModel(items *v1.ApplicationList) []*Application {
	var app []*Application = make([]*Application, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (a *Application) Validate() error {
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
