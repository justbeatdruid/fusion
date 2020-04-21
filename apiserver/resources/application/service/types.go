package service

import (
	"encoding/json"
	"fmt"
	"strings"

	v1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"
)

type Application struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Namespace string     `json:"namespace"`
	Users     user.Users `json:"users"`

	Description     string          `json:"description"`
	AccessKey       string          `json:"accessKey"`
	AccessSecretKey string          `json:"accessSecretKey"`
	APIs            []v1.Api        `json:"apis"`
	ConsumerInfo    v1.ConsumerInfo `json:"consumerInfo"`

	Status    v1.Status `json:"status"`
	UserCount int       `json:"userCount"`
	APICount  int       `json:"apiCount"`

	Writable bool `json:"writable"`

	CreatedAt util.Time `json:"createdAt"`

	Group     string `json:"group"`
	GroupName string `json:"groupName"`
	//sunyu+
	Result        v1.Result    `json:"result"`
	DisplayStatus v1.DisStatus `json:"disStatus"`
}

type Statistics struct {
	Total      int    `json:"total"`
	Increment  int    `json:"increment"`
	Percentage string `json:"percentage"`
}

// only used in creation options
func ToAPI(app *Application) *v1.Application {
	crd := &v1.Application{}
	crd.TypeMeta.Kind = "Application"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = app.Namespace
	crd.Spec = v1.ApplicationSpec{
		Name:            app.Name,
		Description:     app.Description,
		AccessKey:       app.AccessKey,
		AccessSecretKey: app.AccessSecretKey,
		APIs:            []v1.Api{},
		ConsumerInfo:    app.ConsumerInfo,
		Result:          app.Result,
		DisplayStatus:   app.DisplayStatus,
	}
	crd.Status = v1.ApplicationStatus{
		Status: v1.Init,
	}
	if len(app.Group) > 0 {
		if crd.ObjectMeta.Labels == nil {
			crd.ObjectMeta.Labels = make(map[string]string)
		}
		crd.ObjectMeta.Labels[v1.GroupLabel] = app.Group
	}
	// add user labels
	crd.ObjectMeta.Labels = user.AddUsersLabels(app.Users, crd.ObjectMeta.Labels)
	return crd
}

func ToModel(obj *v1.Application, opts ...util.OpOption) *Application {
	switch obj.Spec.Result {
	case v1.CREATING:
		(*obj).Spec.DisplayStatus = v1.SuCreating
	case v1.CREATESUCCESS:
		(*obj).Spec.DisplayStatus = v1.CreateSuccess
	case v1.CREATEFAILED:
		(*obj).Spec.DisplayStatus = v1.CreateFailed
	case v1.UPDATESUCCESS:
		(*obj).Spec.DisplayStatus = v1.UpdateSuccess
	case v1.UPDATEFAILED:
		(*obj).Spec.DisplayStatus = v1.UpdateFailed
	case v1.DELETEFAILED:
		(*obj).Spec.DisplayStatus = v1.DeleteFailed
	}
	app := &Application{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,

		Description:     obj.Spec.Description,
		AccessKey:       obj.Spec.AccessKey,
		AccessSecretKey: obj.Spec.AccessSecretKey,
		APIs:            obj.Spec.APIs,
		ConsumerInfo:    obj.Spec.ConsumerInfo,

		Status:   obj.Status.Status,
		APICount: len(obj.Spec.APIs),

		CreatedAt:     util.NewTime(obj.ObjectMeta.CreationTimestamp.Time),
		Result:        obj.Spec.Result,
		DisplayStatus: obj.Spec.DisplayStatus,
	}
	app.Users = user.GetUsersFromLabels(obj.ObjectMeta.Labels)
	app.UserCount = user.GetUserCountFromLabels(obj.ObjectMeta.Labels)
	if group, ok := obj.ObjectMeta.Labels[v1.GroupLabel]; ok {
		app.Group = group
		app.GroupName = obj.Spec.Group.Name
	}
	u := util.OpList(opts...).User()
	if len(u) > 0 {
		app.Writable = user.WritePermitted(u, obj.ObjectMeta.Labels)
	}
	return app
}

func ToListModel(items *v1.ApplicationList, groups map[string]string, opts ...util.OpOption) []*Application {
	if len(opts) > 0 {
		nameLike := util.OpList(opts...).NameLike()
		if len(nameLike) > 0 {
			var apps []*Application = make([]*Application, 0)
			for _, item := range items.Items {
				if !strings.Contains(item.Spec.Name, nameLike) {
					continue
				}
				if gid, ok := item.ObjectMeta.Labels[v1.GroupLabel]; ok {
					item.Spec.Group.ID = gid
				}
				if gname, ok := groups[item.Spec.Group.ID]; ok {
					item.Spec.Group.Name = gname
				}
				app := ToModel(&item, opts...)
				apps = append(apps, app)
			}
			return apps
		}
	}
	var apps []*Application = make([]*Application, len(items.Items))
	for i, item := range items.Items {
		if gid, ok := item.ObjectMeta.Labels[v1.GroupLabel]; ok {
			item.Spec.Group.ID = gid
		}
		if gname, ok := groups[item.Spec.Group.ID]; ok {
			item.Spec.Group.Name = gname
		}
		apps[i] = ToModel(&item, opts...)
	}
	return apps
}

func (s *Service) Validate(a *Application) error {
	for k, v := range map[string]string{
		"name": a.Name,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	appList, errs := s.List(util.WithNamespace(a.Namespace))
	if errs != nil {
		return fmt.Errorf("cannot list app object: %+v", errs)
	}
	for _, p := range appList.Items {
		if p.Spec.Name == a.Name {
			return errors.NameDuplicatedError("app name duplicated: %+v", errs)
		}
	}

	if len(a.Users.Owner.ID) == 0 {
		return fmt.Errorf("owner not set")
	}
	a.ID = names.NewID()
	var err error
	a.AccessKey, a.AccessSecretKey, err = s.getKeyPairs()
	if err != nil {
		return fmt.Errorf("cannot get key pairs: %+v", err)
	}
	return nil
}

func (s *Service) getKeyPairs() (string, string, error) {
	return "10086", "12345", nil
}

func (s *Service) assignment(target *v1.Application, reqData interface{}) error {
	data, ok := reqData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("reqData type is error,req data: %v", reqData)
	}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err)
	}
	var source Application
	if err = json.Unmarshal(b, &source); err != nil {
		return fmt.Errorf("json.Unmarshal error,: %v", err)
	}
	if _, ok := data["name"]; ok {
		if target.Spec.Name != source.Name {
			appList, errs := s.List(util.WithNamespace(target.ObjectMeta.Namespace))
			if errs != nil {
				return fmt.Errorf("cannot list app object: %+v", errs)
			}
			for _, p := range appList.Items {
				if p.Spec.Name == source.Name {
					return errors.NameDuplicatedError("app name duplicated: %+v", errs)
				}
			}
		}
		target.Spec.Name = source.Name

	}
	if _, ok := data["description"]; ok {
		target.Spec.Description = source.Description
	}
	if _, ok := data["group"]; ok {
		if target.ObjectMeta.Labels == nil {
			target.ObjectMeta.Labels = make(map[string]string)
		}
		target.ObjectMeta.Labels[v1.GroupLabel] = source.Group
	}
	if target.Spec.APIs == nil {
		target.Spec.APIs = make([]v1.Api, 0)
	}
	return nil
}
