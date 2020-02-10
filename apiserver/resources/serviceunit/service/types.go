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
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Namespace    string                `json:"namespace"`
	Type         v1.ServiceType        `json:"type"`
	DatasourceID v1.Datasource         `json:"datasources,omitempty"`
	Datasource   datav1.DatasourceSpec `json:"-"`
	KongSevice   v1.KongServiceInfo    `json:"kongService"`
	Users        []apiv1.User          `json:"users"`
	Description  string                `json:"description"`

	Status    v1.Status `json:"status"`
	UpdatedAt time.Time `json:"time"`
	APICount  int       `json:"apiCount"`
	Published bool      `json:"published"`

	Group string `json:"group"`
}

// only used in creation options
func ToAPI(app *Serviceunit) *v1.Serviceunit {
	crd := &v1.Serviceunit{}
	crd.TypeMeta.Kind = "Serviceunit"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.ServiceunitSpec{
		Name:         app.Name,
		Type:         app.Type,
		DatasourceID: app.DatasourceID,
		Datasource:   app.Datasource,
		KongService:  app.KongSevice,
		Users:        app.Users,
		Description:  app.Description,
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
	if len(app.Group) > 0 {
		crd.ObjectMeta.Labels[v1.GroupLabel] = app.Group
	}
	return crd
}

// +update_sunyu
func ToAPIUpdate(su *Serviceunit, crd *v1.Serviceunit) *v1.Serviceunit {
	id := crd.Spec.KongService.ID
	crd.Spec = v1.ServiceunitSpec{
		Name:         su.Name,
		Type:         su.Type,
		DatasourceID: su.DatasourceID,
		Datasource:   su.Datasource,
		KongService:  su.KongSevice,
		Users:        su.Users,
		Description:  su.Description,
	}
	crd.Spec.KongService.ID = id
	status := su.Status
	if len(status) == 0 {
		status = v1.Update
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
	su := &Serviceunit{
		ID:           obj.ObjectMeta.Name,
		Name:         obj.Spec.Name,
		Namespace:    obj.ObjectMeta.Namespace,
		Type:         obj.Spec.Type,
		DatasourceID: obj.Spec.DatasourceID,
		KongSevice:   obj.Spec.KongService,
		Users:        obj.Spec.Users,
		Description:  obj.Spec.Description,

		Status:    obj.Status.Status,
		UpdatedAt: obj.Status.UpdatedAt,
		APICount:  obj.Status.APICount,
		Published: obj.Status.Published,
	}
	if group, ok := obj.ObjectMeta.Labels[v1.GroupLabel]; ok {
		su.Group = group
	}
	return su
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
	case v1.DataService:
		if len(a.DatasourceID.ID) == 0 {
			return fmt.Errorf("datasource is null")
		} else {
			if ds, err := s.checkDatasource(&a.DatasourceID); err != nil {
				return fmt.Errorf("error datasource: %+v", err)
			} else {
				a.Datasource = *ds
			}
		}
	case v1.WebService:
		if len(a.KongSevice.Host) == 0 || len(a.KongSevice.Protocol) == 0 {
			return fmt.Errorf("webservice is null")
		}
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

	for k, v := range map[string]string{
		"name": ds.Name,
		"type": string(ds.Type),
	} {
		if len(v) == 0 {
			return nil, fmt.Errorf("%s is null", k)
		}
	}

	if ds.Type == datav1.RDBType {
		if ds.RDB == nil {
			return nil, fmt.Errorf("cannot find rdb info")
		}
		if ds.RDB.Connect.Port < 1 || ds.RDB.Connect.Port > 65535 {
			return nil, fmt.Errorf("invalid port: %d", ds.RDB.Connect.Port)
		}

		// TODO move to api
		/*
			for i, field := range ds.Fields {
				if err := field.Validate(); err != nil {
					return nil, fmt.Errorf("%dth field invalide: %+v", i, err)
				}
			}
		*/
	}
	return ds, nil
}
