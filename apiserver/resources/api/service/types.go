package service

import (
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/crds/api/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
)

type Api struct {
	ID        string `json:"id"`
	Namespace string `json:"namespace"`

	Name        string         `json:"name"`
	Serviceunit v1.Serviceunit `json:"serviceunit"`
	Users       []v1.User      `json:"users"`
	Frequency   int            `json:"frequency"`
	Method      v1.Method      `json:"method"`
	Protocol    v1.Protocol    `json:"protocol"`
	ReturnType  v1.ReturnType  `json:"returnType"`
	Parameters  []v1.Parameter `json:"parameter"`

	UpdatedAt        time.Time `json:"updatedAt"`
	ReleasedAt       time.Time `json:"releasedAt"`
	ApplicationCount int       `json:"applicationCount"`
	CalledCount      int       `json:"calledCount"`
}

func ToAPI(api *Api) *v1.Api {
	crd := &v1.Api{}
	crd.TypeMeta.Kind = "Api"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = api.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.ApiSpec{
		Name:        api.Name,
		Serviceunit: api.Serviceunit,
		Users:       api.Users,
		Frequency:   api.Frequency,
		Method:      api.Method,
		Protocol:    api.Protocol,
		ReturnType:  api.ReturnType,
		Parameters:  api.Parameters,
	}
	crd.Status = v1.ApiStatus{
		UpdatedAt:        api.UpdatedAt,
		ReleasedAt:       api.ReleasedAt,
		ApplicationCount: api.ApplicationCount,
		CalledCount:      api.CalledCount,
	}
	return crd
}

func ToModel(obj *v1.Api) *Api {
	return &Api{
		ID:        obj.ObjectMeta.Name,
		Namespace: obj.ObjectMeta.Namespace,

		Name:        obj.Spec.Name,
		Serviceunit: obj.Spec.Serviceunit,
		Users:       obj.Spec.Users,
		Frequency:   obj.Spec.Frequency,
		Method:      obj.Spec.Method,
		Protocol:    obj.Spec.Protocol,
		ReturnType:  obj.Spec.ReturnType,
		Parameters:  obj.Spec.Parameters,

		UpdatedAt:        obj.Status.UpdatedAt,
		ReleasedAt:       obj.Status.ReleasedAt,
		ApplicationCount: obj.Status.ApplicationCount,
		CalledCount:      obj.Status.CalledCount,
	}
}

func ToListModel(items *v1.ApiList) []*Api {
	var api []*Api = make([]*Api, len(items.Items))
	for i := range items.Items {
		api[i] = ToModel(&items.Items[i])
	}
	return api
}

func (a *Api) Validate() error {
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
