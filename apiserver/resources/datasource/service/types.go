package service

import (
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Datasource struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Namespace string     `json:"namespace"`
	Type      string     `json:"foo"`
	Database  string     `json:"database"`
	Schema    string     `json:"schema,omitempty"`
	Table     string     `json:"table"`
	Status    v1.Status  `json:"status"`
	UpdatedAt time.Time  `json:"UpdatedAt"`
	CreatedAt time.Time  `json:"CreatedAt"`
	Fields    []v1.Field `json:"fields"`

	CreateUser v1.CreateUser `json:"createUser"`
	UpdateUser v1.UpdateUser `json:"updateUser"`

	Connect v1.ConnectInfo `json:"connect"`
}

// only used in creation or update options
func ToAPI(ds *Datasource, dealType string) *v1.Datasource {
	crd := &v1.Datasource{}
	crd.TypeMeta.Kind = "Datasource"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = ds.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.DatasourceSpec{
		Name:       ds.Name,
		Type:       ds.Type,
		Database:   ds.Database,
		Schema:     ds.Schema,
		Table:      ds.Table,
		Fields:     ds.Fields,
		Connect:    ds.Connect,
		CreateUser: ds.CreateUser,
		UpdateUser: ds.UpdateUser,
	}
	status := ds.Status
	if len(status) == 0 {
		status = v1.Init
	}
	if dealType == "create" {
		crd.Status = v1.DatasourceStatus{
			Status:    status,
			CreatedAt: time.Now(),
		}
	} else if dealType == "update" {
		crd.Status = v1.DatasourceStatus{
			Status:    status,
			UpdatedAt: time.Now(),
		}
	}
	return crd
}

func ToModel(obj *v1.Datasource) *Datasource {
	return &Datasource{
		ID:        obj.ObjectMeta.Name,
		Namespace: obj.ObjectMeta.Namespace,
		Name:      obj.Spec.Name,
		Type:      obj.Spec.Type,
		Database:  obj.Spec.Database,
		Schema:    obj.Spec.Schema,
		Table:     obj.Spec.Table,

		Status:    obj.Status.Status,
		UpdatedAt: obj.Status.UpdatedAt,
		CreatedAt: obj.Status.CreatedAt,

		Fields: obj.Spec.Fields,

		Connect: obj.Spec.Connect,

		CreateUser: obj.Spec.CreateUser,
		UpdateUser: obj.Spec.UpdateUser,
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
		"name":           a.Name,
		"type":           a.Type,
		"database":       a.Database,
		"table":          a.Table,
		"createUserId":   a.CreateUser.UserId,
		"createUserName": a.CreateUser.UserName,
		"host":           a.Connect.Host,
		"port":           string(a.Connect.Port),
		"username":       a.Connect.Username,
		"password":       a.Connect.Password,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	a.ID = names.NewID()
	return nil
}
