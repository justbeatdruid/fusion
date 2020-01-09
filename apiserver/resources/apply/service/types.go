package service

import (
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/crds/apply/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Apply struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`

	TargetType v1.TargetType `json:"targetType"`
	TargetID   string        `json:"targetID"`
	TargetName string        `json:"targetName"`
	AppID      string        `json:"appID"`
	AppName    string        `json:"appName"`
	ExpireAt   time.Time     `json:"expireAt"`
}

// only used in creation options
func ToAPI(app *Apply) *v1.Apply {
	crd := &v1.Apply{}
	crd.TypeMeta.Kind = "Apply"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.ApplySpec{
		Name:       app.Name,
		TargetType: app.TargetType,
		TargetID:   app.TargetID,
		TargetName: app.TargetName,
		AppID:      app.AppID,
		AppName:    app.AppName,
		ExpireAt:   app.ExpireAt,
	}
	return crd
}

func ToModel(obj *v1.Apply) *Apply {
	return &Apply{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.ObjectMeta.Namespace,

		TargetType: obj.Spec.TargetType,
		TargetID:   obj.Spec.TargetID,
		TargetName: obj.Spec.TargetName,
		AppID:      obj.Spec.AppID,
		AppName:    obj.Spec.AppName,
		ExpireAt:   obj.Spec.ExpireAt,
	}
}

func ToListModel(items *v1.ApplyList) []*Apply {
	var app []*Apply = make([]*Apply, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (a *Apply) Validate() error {
	for k, v := range map[string]string{
		"name":        a.Name,
		"target type": string(a.TargetType),
		"target ID":   a.TargetID,
		"target name": a.TargetName,
		"app ID":      a.AppID,
		"app name":    a.AppName,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	if a.ExpireAt.IsZero() {
		return fmt.Errorf("expire time not set")
	}
	a.ID = names.NewID()
	return nil
}
