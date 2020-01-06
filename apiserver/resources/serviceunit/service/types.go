package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Serviceunit struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// only used in creation options
func ToAPI(app *Serviceunit) *v1.Serviceunit {
	crd := &v1.Serviceunit{}
	crd.TypeMeta.Kind = "Serviceunit"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.ServiceunitSpec{
		Name: app.Name,
	}
	return crd
}

func ToModel(obj *v1.Serviceunit) *Serviceunit {
	return &Serviceunit{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,
	}
}

func ToListModel(items *v1.ServiceunitList) []*Serviceunit {
	var app []*Serviceunit = make([]*Serviceunit, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (a *Serviceunit) Validate() error {
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
