package service

import (
	"github.com/chinamobile/nlpt/application-controller/api/v1"
)

type Application struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

// only used in creation options
func ToAPI(app *Application) *v1.Application {
	crd := &v1.Application{}
	crd.TypeMeta.Kind = "Application"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.Name
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.ApplicationSpec{
		Type: app.Type,
	}
	return crd
}

func ToModel(obj *v1.Application) *Application {
	return &Application{
		Name:      obj.ObjectMeta.Name,
		Namespace: obj.ObjectMeta.Namespace,

		Type: obj.Spec.Type,

		Status: obj.Status.Status,
	}
}

func ToListModel(items *v1.ApplicationList) []*Application {
	var app []*Application = make([]*Application, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (*Application) Validate() error {
	return nil
}
