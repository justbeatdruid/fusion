package service

import (
	"encoding/json"
	"fmt"
	"time"
	"github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Trafficcontrol struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace"`
	Type        v1.LimitType  `json:"type"`
	Config      v1.ConfigInfo `json:"config"`
	Apis        []v1.Api      `json:"apis"`
	Description string        `json:"description"`
	User        string        `json:"user"`

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
	crd.ObjectMeta.Labels = make(map[string]string)
	crd.ObjectMeta.Labels[app.ID] = app.ID
	crd.Spec = v1.TrafficcontrolSpec{
		Name:        app.Name,
		Type:        app.Type,
		Config:      app.Config,
		Apis:        app.Apis,
		User:        app.User,
		Description: app.Description,
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
		Name:        app.Name,
		Type:        app.Type,
		Config:      app.Config,
		Description: app.Description,
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
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,
		Type:      obj.Spec.Type,
		Config:    obj.Spec.Config,
		Apis:      obj.Spec.Apis,
		User:      obj.Spec.User,

		Description: obj.Spec.Description,

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

	if len(a.Type) == 0 {
		return fmt.Errorf("type is null")
	}

	switch a.Type {
	case v1.APIC, v1.APPC, v1.IPC, v1.USERC, v1.SPECAPPC:
	default:
		return fmt.Errorf("wrong type: %s.", a.Type)
	}

	switch a.Type {
	case v1.APIC, v1.APPC, v1.IPC, v1.USERC:
		if (a.Config.Year + a.Config.Month + a.Config.Day + a.Config.Hour + a.Config.Minute + a.Config.Second) == 0 {
			return fmt.Errorf("at least one limit config must exist.")
		}
	case v1.SPECAPPC:
		if len(a.Config.Special) == 0 {
			return fmt.Errorf("at least one special config must exist.")
		}
		if len(a.Config.Special) > v1.MAXNUM {
			return fmt.Errorf("special config maxinum limit exceeded.")
		}
	default:
		return fmt.Errorf("wrong type: %s.", a.Type)
	}

	a.ID = names.NewID()
	return nil
}

func (s *Service) assignment(target *v1.Trafficcontrol, reqData interface{}) error {
	data, ok := reqData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("reqData type is error,req data: %v", reqData)
	}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err)
	}
	var source Trafficcontrol
	if err = json.Unmarshal(b, &source); err != nil {
		return fmt.Errorf("json.Unmarshal error,: %v", err)
	}
	if _, ok = data["name"]; ok {
		target.Spec.Name = source.Name
	}
	if _, ok = data["namespace"]; ok {
		target.ObjectMeta.Namespace = source.Namespace
	}
	if _, ok = data["type"]; ok {
		target.Spec.Type = source.Type
	}
	if _, ok = data["user"]; ok {
		target.Spec.User = source.User
	}
	if _, ok = data["apiCount"]; ok {
		target.Status.APICount = source.APICount
	}
	if _, ok = data["description"]; ok {
		target.Spec.Description = source.Description
	}
	if _, ok := data["apis"]; ok {
		target.Spec.Apis = source.Apis
	}
	if reqConfig, ok := data["config"]; ok {
		if config, ok := reqConfig.(map[string]interface{}); ok {
			if _, ok = config["year"]; ok {
				target.Spec.Config.Year = source.Config.Year
			}
			if _, ok = config["month"]; ok {
				target.Spec.Config.Month = source.Config.Month
			}
			if _, ok = config["day"]; ok {
				target.Spec.Config.Day = source.Config.Day
			}
			if _, ok = config["hour"]; ok {
				target.Spec.Config.Hour = source.Config.Hour
			}
			if _, ok = config["minute"]; ok {
				target.Spec.Config.Minute = source.Config.Minute
			}
			if _, ok = config["second"]; ok {
				target.Spec.Config.Second = source.Config.Second
			}
		}
	}
	target.Status.UpdatedAt = time.Now()
	return nil
}
