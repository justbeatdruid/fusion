package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
	"time"
)

type Trafficcontrol struct {
	ID           string                `json:"id"`
	Name         string                `json:"name"`
	Namespace    string                `json:"namespace"`
	Type         v1.LimitType          `json:"type"`
	Config       v1.ConfigInfo         `json:"config"`
	Apis         v1.Api                `json:"api"`
	Description  string                `json:"description"`
	User         string                `json:"user"`

	Status    v1.Status `json:"status"`
	UpdatedAt time.Time `json:"time"`
	APICount  int       `json:"apiCount"`
	Published bool      `json:"published"`
}

// only used in creation options
func ToAPI(app *Trafficcontrol) *v1.Trafficcontrol {
	crd := &v1.Trafficcontrol{}
	crd.TypeMeta.Kind = "Trafficcontrol"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.TrafficcontrolSpec{
		Name:         app.Name,
		Type:         app.Type,
		Config:       app.Config,
		User:         app.User,
		Description:  app.Description,
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Init
	}
	crd.Status = v1.TrafficcontrolStatus{
		Status:    status,
		UpdatedAt: time.Now(),
		APICount:  0,
		Published: false,
	}
	return crd
}

// +update
func ToAPIUpdate(app *Trafficcontrol, crd *v1.Trafficcontrol) *v1.Trafficcontrol {
	crd.Spec = v1.TrafficcontrolSpec{
		Name:         app.Name,
		Type:         app.Type,
		Config:       app.Config,
		Description:  app.Description,
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Update
	}
	crd.Status = v1.TrafficcontrolStatus{
		Status:    status,
		UpdatedAt: time.Now(),
		APICount:  0,
		Published: false,
	}
	return crd
}

func ToModel(obj *v1.Trafficcontrol) *Trafficcontrol {
	return &Trafficcontrol{
		ID:           obj.ObjectMeta.Name,
		Name:         obj.Spec.Name,
		Namespace:    obj.ObjectMeta.Namespace,
		Type:         obj.Spec.Type,
		Config:       obj.Spec.Config,

		Description:  obj.Spec.Description,

		Status:    obj.Status.Status,
		UpdatedAt: obj.Status.UpdatedAt,
		APICount:  obj.Status.APICount,
		Published: obj.Status.Published,
	}
}

func ToListModel(items *v1.TrafficcontrolList) []*Trafficcontrol {
	var app []*Trafficcontrol = make([]*Trafficcontrol, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

// check create parameters
func (s *Service) Validate(a *Trafficcontrol) error {
	for k, v := range map[string]string{
		"name":        a.Name,
		"description": a.Description,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}

	if len(a.Type) == 0{
		return fmt.Errorf("type is null")
	}

	switch a.Type {
	case v1.APIC,v1.APPC, v1.IPC, v1.USERC:
	default:
		return fmt.Errorf("wrong type: %s.", a.Type)
	}

	if (a.Config.Year + a.Config.Month + a.Config.Day + a.Config.Hour + a.Config.Minute + a.Config.Second) == 0{
		return fmt.Errorf("at least one limit config must exist.")
	}

	a.ID = names.NewID()
	return nil
}

