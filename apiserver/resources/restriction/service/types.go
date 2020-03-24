package service

import (
	"encoding/json"
	"fmt"
	"github.com/chinamobile/nlpt/crds/restriction/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"time"
)

type Restriction struct {
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Type      v1.LimitType  `json:"type"`
	Action    v1.Action     `json:"action"`
	Config    v1.ConfigInfo `json:"config"`
	Users     user.Users    `json:"users"`
	Apis      []v1.Api      `json:"apis"`
	CreatedAt time.Time     `json:"createdAt"`

	Status    v1.Status   `json:"status"`
	Published bool        `json:"published"`
	UpdatedAt metav1.Time `json:"time"`
	APICount  int         `json:"apiCount"`
}

// only used in creation options
func ToAPI(app *Restriction) *v1.Restriction {
	crd := &v1.Restriction{}
	crd.TypeMeta.Kind = "Restriction"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version
	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.ObjectMeta.Labels = make(map[string]string)
	crd.ObjectMeta.Labels[app.ID] = app.ID
	crd.Spec = v1.RestrictionSpec{
		Name:   app.Name,
		Type:   app.Type,
		Action: app.Action,
		Config: app.Config,
		Apis:   app.Apis,
	}
	if crd.Spec.Apis == nil {
		crd.Spec.Apis = make([]v1.Api, 0)
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Init
	}
	crd.Status = v1.RestrictionStatus{
		Status:    status,
		UpdatedAt: metav1.Now(),
		APICount:  0,
		Published: false,
	}
	// add user labels
	crd.ObjectMeta.Labels = user.AddUsersLabels(app.Users, crd.ObjectMeta.Labels)
	return crd
}

// +update
func ToAPIUpdate(app *Restriction, crd *v1.Restriction) *v1.Restriction {
	crd.Spec = v1.RestrictionSpec{
		Name: app.Name,
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Update
	}
	crd.Status = v1.RestrictionStatus{
		Status: status,
	}
	return crd
}

func ToModel(obj *v1.Restriction) *Restriction {
	restriction := &Restriction{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,
		Type:      obj.Spec.Type,
		Action:    obj.Spec.Action,
		Config:    obj.Spec.Config,
		Apis:      obj.Spec.Apis,
		CreatedAt: obj.ObjectMeta.CreationTimestamp.Time,

		Status:    obj.Status.Status,
		UpdatedAt: obj.Status.UpdatedAt,
		APICount:  obj.Status.APICount,
		Published: obj.Status.Published,
	}
	restriction.Users = user.GetUsersFromLabels(obj.ObjectMeta.Labels)
	return restriction
}

func ToListModel(items *v1.RestrictionList, opts ...util.OpOption) []*Restriction {
	if len(opts) > 0 {
		var apps []*Restriction = make([]*Restriction, 0)
		nameLike := util.OpList(opts...).NameLike()
		if len(nameLike) > 0 {
			for _, item := range items.Items {
				if !strings.Contains(item.Spec.Name, nameLike) {
					continue
				}
				app := ToModel(&item)
				apps = append(apps, app)
			}
			return apps
		}
	}
	var apps []*Restriction = make([]*Restriction, len(items.Items))
	for i := range items.Items {
		apps[i] = ToModel(&items.Items[i])
	}
	return apps
}

// check create parameters
func (s *Service) Validate(a *Restriction) error {
	for k, v := range map[string]string{
		"name": a.Name,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	switch a.Action {
	case v1.WHITE, v1.BLACK:
	default:
		return fmt.Errorf("wrong action: %s.", a.Type)
	}

	switch a.Type {
	case v1.IP, v1.USER:
	default:
		return fmt.Errorf("wrong type: %s.", a.Type)
	}

	switch a.Type {
	case v1.IP:
		if len(a.Config.Ip) == 0 {
			return fmt.Errorf("at least ip limit config must exist.")
		}
	case v1.USER:
		if len(a.Config.User) == 0 {
			return fmt.Errorf("at least user config must exist.")
		}
	default:
		return fmt.Errorf("wrong config type: %s.", a.Type)
	}
	if len(a.Users.Owner.ID) == 0 {
		return fmt.Errorf("owner not set")
	}

	a.ID = names.NewID()
	return nil
}

func (s *Service) assignment(target *v1.Restriction, reqData interface{}) error {
	data, ok := reqData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("reqData type is error,req data: %v", reqData)
	}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err)
	}
	var source Restriction
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

	if _, ok = data["action"]; ok {
		target.Spec.Action = source.Action
	}

	if _, ok = data["config"]; ok {
		target.Spec.Config = source.Config
	}

	target.Status.UpdatedAt = metav1.Now()
	return nil
}
