package service

import (
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/crds/applicationgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type ApplicationGroup struct {
	ID          string    `json:"id"`
	Namespace   string    `json:"namespace"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
}

// only used in creation options
func ToAPI(app *ApplicationGroup) *v1.ApplicationGroup {
	crd := &v1.ApplicationGroup{}
	crd.TypeMeta.Kind = "ApplicationGroup"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = app.Namespace
	crd.Spec = v1.ApplicationGroupSpec{
		Name:        app.Name,
		Description: app.Description,
	}
	return crd
}

func ToModel(obj *v1.ApplicationGroup) *ApplicationGroup {
	return &ApplicationGroup{
		ID:          obj.ObjectMeta.Name,
		Name:        obj.Spec.Name,
		Namespace:   obj.ObjectMeta.Namespace,
		Description: obj.Spec.Description,
		CreatedAt:   obj.ObjectMeta.CreationTimestamp.Time,
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
