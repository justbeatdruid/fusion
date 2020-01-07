package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Datasource struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// only used in creation options
func ToAPI(ds *Datasource) *v1.Datasource {
	crd := &v1.Datasource{}
	crd.TypeMeta.Kind = "Datasource"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = ds.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.DatasourceSpec{
		Name: ds.Name,
	}
	return crd
}

func ToModel(obj *v1.Datasource) *Datasource {
	return &Datasource{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,
	}
}

func ToListModel(items *v1.DatasourceList) []*Datasource {
	var ds []*Datasource = make([]*Datasource, len(items.Items))
	for i := range items.Items {
		ds[i] = ToModel(&items.Items[i])
	}
	return ds
}

func (a *Datasource) Validate() error {
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
