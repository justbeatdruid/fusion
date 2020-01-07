package service

import (
	"fmt"
	"time"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	datav1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Serviceunit struct {
	ID                 string                  `json:"id"`
	Name               string                  `json:"name"`
	Namespace          string                  `json:"namespace"`
	Type               v1.ServiceType          `json:"type"`
	SingleDatasourceID *v1.Datasource          `json:"singleDatasource"`
	MultiDatasourceID  []v1.Datasource         `json:"multiDatasource"`
	SingleDatasource   *datav1.DatasourceSpec  `json:"-"`
	MultiDatasource    []datav1.DatasourceSpec `json:"-"`
	Users              []apiv1.User            `json:"users"`
	Description        string                  `json:"description"`

	Status    v1.Status `json:"status"`
	UpdatedAt time.Time `json:"time.Time"`
	APICount  int       `json:"apiCount"`
	Published bool      `json:"published"`
}

// only used in creation options
func ToAPI(app *Serviceunit) *v1.Serviceunit {
	crd := &v1.Serviceunit{}
	crd.TypeMeta.Kind = "Serviceunit"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.ServiceunitSpec{
		Name:               app.Name,
		Type:               app.Type,
		SingleDatasourceID: app.SingleDatasourceID,
		MultiDatasourceID:  app.MultiDatasourceID,
		SingleDatasource:   app.SingleDatasource,
		MultiDatasource:    app.MultiDatasource,
		Users:              app.Users,
		Description:        app.Description,
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Init
	}
	crd.Status = v1.ServiceunitStatus{
		Status:    status,
		UpdatedAt: time.Now(),
		APICount:  0,
		Published: false,
	}
	return crd
}

func ToModel(obj *v1.Serviceunit) *Serviceunit {
	return &Serviceunit{
		ID:                 obj.ObjectMeta.Name,
		Name:               obj.Spec.Name,
		Namespace:          obj.ObjectMeta.Namespace,
		Type:               obj.Spec.Type,
		SingleDatasourceID: obj.Spec.SingleDatasourceID,
		MultiDatasourceID:  obj.Spec.MultiDatasourceID,
		Users:              obj.Spec.Users,
		Description:        obj.Spec.Description,

		Status:    obj.Status.Status,
		UpdatedAt: obj.Status.UpdatedAt,
		APICount:  obj.Status.APICount,
		Published: obj.Status.Published,
	}
}

func ToListModel(items *v1.ServiceunitList) []*Serviceunit {
	var app []*Serviceunit = make([]*Serviceunit, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

// check create parameters
func (s *Service) Validate(a *Serviceunit) error {
	for k, v := range map[string]string{
		"name":        a.Name,
		"description": a.Description,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}

	switch a.Type {
	case v1.Single:
		if ds, err := s.checkDatasource(a.SingleDatasourceID); err != nil {
			return fmt.Errorf("datasource error: %+v", err)
		} else {
			a.SingleDatasource = ds
		}
	case v1.Multi:
		if len(a.MultiDatasourceID) == 0 {
			return fmt.Errorf("datasources is length is 0")
		}
		a.MultiDatasource = make([]datav1.DatasourceSpec, len(a.MultiDatasourceID))
		for i, dsid := range a.MultiDatasourceID {
			if ds, err := s.checkDatasource(&dsid); err != nil {
				return fmt.Errorf("%dth datasource error: %+v", i, err)
			} else {
				a.MultiDatasource[i] = *ds
			}
		}
	default:
		return fmt.Errorf("unknown datasource type: %s", a.Type)
	}

	a.ID = names.NewID()
	return nil
}

func (s *Service) checkDatasource(d *v1.Datasource) (*datav1.DatasourceSpec, error) {
	if d == nil {
		return nil, fmt.Errorf("datasource is null")
	}
	ds, err := s.getDatasource(d.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot get datasource: %+v", err)
	}

	//TODO check fileds match
	ds.Fields = d.Fields

	for k, v := range map[string]string{
		"name":     ds.Name,
		"type":     ds.Type,
		"database": ds.Database,
		"table":    ds.Table,

		"host":     ds.Connect.Host,
		"username": ds.Connect.Username,
		"password": ds.Connect.Password,
	} {
		if len(v) == 0 {
			return nil, fmt.Errorf("%s is null", k)
		}
	}
	if ds.Connect.Port < 1 || ds.Connect.Port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", ds.Connect.Port)
	}

	if len(ds.Fields) == 0 {
		return nil, fmt.Errorf("filed length is 0")
	}
	for i, field := range ds.Fields {
		if err := field.Validate(); err != nil {
			return nil, fmt.Errorf("%dth field invalide: %+v", i, err)
		}
	}
	return ds, nil
}
