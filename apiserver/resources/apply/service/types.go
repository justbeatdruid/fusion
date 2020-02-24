package service

import (
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/crds/apply/api/v1"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/names"

	//"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Apply struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`

	Target     Resource  `json:"target"`
	Source     Resource  `json:"source"`
	Action     v1.Action `json:"action"`
	Message    string    `json:"message"`
	ExpireAt   time.Time `json:"expireAt"`
	AppliedBy  string    `json:"appliedBy"`
	ApprovedBy string    `json:"approvedBy"`

	Status     v1.Status `json:"status"`
	Reason     string    `json:"reason"`
	AppliedAt  time.Time `json:"appliedAt"`
	ApprovedAt time.Time `json:"approvedAt"`
}

type Resource struct {
	Type      v1.Type   `json:"type"`
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Owner     string    `json:"owner"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// only used in creation options
func ToAPI(app *Apply) *v1.Apply {
	app = ApiBindingApply(app)
	crd := &v1.Apply{}
	crd.TypeMeta.Kind = "Apply"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.ApplySpec{
		TargetType: app.Target.Type,
		TargetID:   app.Target.ID,
		SourceType: app.Source.Type,
		SourceID:   app.Source.ID,
		Action:     app.Action,
		ExpireAt:   metav1.NewTime(app.ExpireAt),
		AppliedBy:  app.AppliedBy,
		ApprovedBy: app.ApprovedBy,
		Message:    app.Message,
	}
	crd.Status = v1.ApplyStatus{
		Status:     v1.Waiting,
		AppliedAt:  metav1.Now(),
		ApprovedAt: metav1.Unix(0, 0),
	}
	return crd
}

func ApiBindingApply(a *Apply) *Apply {
	a.Target.Type = v1.Api
	a.Source.Type = v1.Application
	a.Action = v1.Bind

	return a
}

func (s *Service) ToModel(obj *v1.Apply) (*Apply, error) {
	a := &Apply{
		ID:        obj.ObjectMeta.Name,
		Namespace: obj.ObjectMeta.Namespace,

		Target: Resource{
			Type: obj.Spec.TargetType,
			ID:   obj.Spec.TargetID,
			Name: obj.Spec.TargetName,
		},
		Source: Resource{
			Type: obj.Spec.SourceType,
			ID:   obj.Spec.SourceID,
			Name: obj.Spec.SourceName,
		},
		Action:   obj.Spec.Action,
		Message:  obj.Spec.Message,
		ExpireAt: obj.Spec.ExpireAt.Time,

		Status:     obj.Status.Status,
		Reason:     obj.Status.Reason,
		AppliedAt:  obj.Status.AppliedAt.Time,
		ApprovedAt: obj.Status.ApprovedAt.Time,
	}
	var err error
	a.Source, err = s.Completion(a.Source)
	if err != nil {
		return nil, err
	}
	a.Target, err = s.Completion(a.Target)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Service) Completion(r Resource) (Resource, error) {
	switch r.Type {
	case v1.Api:
		api, err := s.getApi(r.ID)
		if err != nil {
			if errors.IsNotFound(err) {
				r.Status = "missing"
				return r, nil
			}
			return r, fmt.Errorf("get api error: %+v", err)
		}
		r.Owner = "TODO"
		r.Name = api.Spec.Name
		r.CreatedAt = api.ObjectMeta.CreationTimestamp.Time
		r.Status = string(api.Status.Status)
	case v1.Application:
		app, err := s.getApplication(r.ID)
		if err != nil {
			if errors.IsNotFound(err) {
				r.Status = "missing"
				return r, nil
			}
			return r, fmt.Errorf("get application error: %+v", err)
		}
		r.Owner = "TODO"
		r.Name = app.Spec.Name
		r.CreatedAt = app.ObjectMeta.CreationTimestamp.Time
		r.Status = string(app.Status.Status)
	}
	return r, nil
}

func (s *Service) ToListModel(items *v1.ApplyList) ([]*Apply, error) {
	var app []*Apply = make([]*Apply, len(items.Items))
	for i := range items.Items {
		var err error
		app[i], err = s.ToModel(&items.Items[i])
		if err != nil {
			return nil, err
		}
	}
	return app, nil
}

func (a *Apply) Validate() error {
	for k, v := range map[string]string{
		//"name":        a.Name,
		//"target type": a.Target.Type.String(),
		"target ID": a.Target.ID,
		//"target name": a.Target.Name,
		//"source type": a.Source.Type.String(),
		"source ID": a.Source.ID,
		//"source name": a.Source.Name,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	switch a.Action {
	case v1.Bind, v1.Release:
	default:
		return fmt.Errorf("wrong action: %s", a.Action)
	}
	if a.ExpireAt.IsZero() {
		return fmt.Errorf("expire time not set")
	}
	a.ID = names.NewID()
	return nil
}
