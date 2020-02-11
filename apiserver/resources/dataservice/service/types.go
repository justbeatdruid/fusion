package service

import (
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/crds/dataservice/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Dataservice struct {
	ID        string    `json:"id"`
	Namespace string    `json:"namespace"`
	CreatedAt time.Time `json:"createdAt"`
}

// only used in creation options
func ToAPI(app *Dataservice) *v1.Dataservice {
	crd := &v1.Dataservice{}
	crd.TypeMeta.Kind = "Dataservice"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.DataserviceSpec{}
	return crd
}

func ToModel(obj *v1.Dataservice) *Dataservice {
	return &Dataservice{
		ID:        obj.ObjectMeta.Name,
		Namespace: obj.ObjectMeta.Namespace,
		CreatedAt: obj.ObjectMeta.CreationTimestamp.Time,
	}
}

func ToListModel(items *v1.DataserviceList) []*Dataservice {
	var app []*Dataservice = make([]*Dataservice, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (a *Dataservice) Validate() error {
	for k, v := range map[string]string{} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	a.ID = names.NewID()
	return nil
}
