package service

import (
	"fmt"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	"github.com/chinamobile/nlpt/crds/application/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Application struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	Users           []apiv1.User `json:"users"`
	AccessKey       string       `json:"accessKey"`
	AccessSecretKey string       `json:"accessSecretKey"`
	APIs            []v1.Api     `json:"apis"`

	Status    v1.Status `json:"status"`
	UserCount int       `json:"userCount"`
	APICount  int       `json:"apiCount"`
}

// only used in creation options
func ToAPI(app *Application) *v1.Application {
	crd := &v1.Application{}
	crd.TypeMeta.Kind = "Application"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.ApplicationSpec{
		Name:            app.Name,
		Users:           app.Users,
		AccessKey:       app.AccessKey,
		AccessSecretKey: app.AccessSecretKey,
		APIs:            []v1.Api{},
	}
	return crd
}

func ToModel(obj *v1.Application) *Application {
	return &Application{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,

		Users:           obj.Spec.Users,
		AccessKey:       obj.Spec.AccessKey,
		AccessSecretKey: obj.Spec.AccessSecretKey,
		APIs:            obj.Spec.APIs,

		Status:    obj.Status.Status,
		UserCount: len(obj.Spec.Users),
		APICount:  len(obj.Spec.APIs),
	}
}

func ToListModel(items *v1.ApplicationList) []*Application {
	var app []*Application = make([]*Application, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (s *Service) Validate(a *Application) error {
	for k, v := range map[string]string{
		"name": a.Name,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	a.ID = names.NewID()
	var err error
	a.AccessKey, a.AccessSecretKey, err = s.getKeyPairs()
	if err != nil {
		return fmt.Errorf("cannot get key pairs: %+v", err)
	}
	return nil
}

func (s *Service) getKeyPairs() (string, string, error) {
	return "10086", "12345", nil
}
