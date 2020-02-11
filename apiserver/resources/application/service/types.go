package service

import (
	"fmt"
	"time"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	"github.com/chinamobile/nlpt/crds/application/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Application struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	Description     string          `json:"description"`
	Users           []apiv1.User    `json:"users"`
	AccessKey       string          `json:"accessKey"`
	AccessSecretKey string          `json:"accessSecretKey"`
	APIs            []v1.Api        `json:"apis"`
	ConsumerInfo    v1.ConsumerInfo `json:"consumerInfo"`

	Status    v1.Status `json:"status"`
	UserCount int       `json:"userCount"`
	APICount  int       `json:"apiCount"`

	CreatedAt time.Time `json:"createdAt"`

	Group string `json:"group"`
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
		Description:     app.Description,
		AccessKey:       app.AccessKey,
		AccessSecretKey: app.AccessSecretKey,
		APIs:            []v1.Api{},
		ConsumerInfo:    app.ConsumerInfo,
	}
	crd.Status = v1.ApplicationStatus{
		Status: v1.Init,
	}
	if len(app.Group) > 0 {
		crd.ObjectMeta.Labels[v1.GroupLabel] = app.Group
	}
	return crd
}

func ToModel(obj *v1.Application) *Application {
	app := &Application{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,

		Description:     obj.Spec.Description,
		Users:           obj.Spec.Users,
		AccessKey:       obj.Spec.AccessKey,
		AccessSecretKey: obj.Spec.AccessSecretKey,
		APIs:            obj.Spec.APIs,
		ConsumerInfo:    obj.Spec.ConsumerInfo,

		Status:    obj.Status.Status,
		UserCount: len(obj.Spec.Users),
		APICount:  len(obj.Spec.APIs),

		CreatedAt: obj.ObjectMeta.CreationTimestamp.Time,
	}
	if group, ok := obj.ObjectMeta.Labels[v1.GroupLabel]; ok {
		app.Group = group
	}
	return app
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
