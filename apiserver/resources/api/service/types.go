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

	Name          string            `json:"name"`
	Serviceunit   v1.Serviceunit    `json:"serviceunit"`
	Applications  []v1.Application  `json:"applications"`
	Users         []v1.User         `json:"users"`
	Frequency     int               `json:"frequency"`
	Method        v1.Method         `json:"method"`
	Protocol      v1.Protocol       `json:"protocol"`
	ReturnType    v1.ReturnType     `json:"returnType"`
	ApiFields     []v1.Field        `json:"apiFields"`
	ApiParameters []v1.ApiParameter `json:"apiParameters"`
	WebParams     []v1.WebParams    `json:"webParams"`
	ApiType       v1.ApiType        `json:"apiType"`
	AuthType      v1.AuthType       `json:"authType"`
	KongApi       v1.KongApiInfo    `json:"KongApi"`

	Status v1.Status `json:"status"`
	//Publish          v1.Publish    `json:"publish"`
	AccessLink       v1.AccessLink `json:"access"`
	UpdatedAt        time.Time     `json:"updatedAt"`
	ReleasedAt       time.Time     `json:"releasedAt"`
	ApplicationCount int           `json:"applicationCount"`
	CalledCount      int           `json:"calledCount"`
	PublishInfo      v1.PublishInfo

	ApplicationBindStatus *v1.ApiApplicationStatus `json:"applicationBindSucceeded"`
}

// only used in creation
func ToAPI(api *Api) *v1.Api {
	crd := &v1.Api{}
	crd.TypeMeta.Kind = "Api"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = api.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.ObjectMeta.Labels = make(map[string]string)
	crd.ObjectMeta.Labels[v1.ServiceunitLabel] = api.Serviceunit.ID
	crd.Spec = v1.ApiSpec{
		Name:         api.Name,
		Serviceunit:  api.Serviceunit,
		Applications: api.Applications,
		Users:        api.Users,
		Frequency:    api.Frequency,
		Method:       api.Method,
		Protocol:     api.Protocol,
		ReturnType:   api.ReturnType,
		ApiFields:    api.ApiFields,
		WebParams:    api.WebParams,
		KongApi:      api.KongApi,
		ApiType:      api.ApiType,
		AuthType:     api.AuthType,
	}
	crd.Status = v1.ApiStatus{
		Status:           v1.Init,
		AccessLink:       api.AccessLink,
		UpdatedAt:        time.Now(),
		ReleasedAt:       time.Now(),
		ApplicationCount: api.ApplicationCount,
		CalledCount:      api.CalledCount,
	}
	return crd
}

func ToModel(obj *v1.Api) *Api {
	model := &Api{
		ID:        obj.ObjectMeta.Name,
		Namespace: obj.ObjectMeta.Namespace,

		Name:         obj.Spec.Name,
		Serviceunit:  obj.Spec.Serviceunit,
		Applications: obj.Spec.Applications,
		Users:        obj.Spec.Users,
		Frequency:    obj.Spec.Frequency,
		Method:       obj.Spec.Method,
		Protocol:     obj.Spec.Protocol,
		ReturnType:   obj.Spec.ReturnType,
		ApiFields:    obj.Spec.ApiFields,
		WebParams:    obj.Spec.WebParams,
		KongApi:      obj.Spec.KongApi,
		ApiType:      obj.Spec.ApiType,
		AuthType:     obj.Spec.AuthType,

		Status:           obj.Status.Status,
		AccessLink:       obj.Status.AccessLink,
		UpdatedAt:        obj.Status.UpdatedAt,
		ReleasedAt:       obj.Status.ReleasedAt,
		ApplicationCount: obj.Status.ApplicationCount,
		CalledCount:      obj.Status.CalledCount,
	}
	if model.Applications == nil {
		model.Applications = []v1.Application{}
	}
	if model.ApiFields == nil {
		model.ApiFields = []v1.Field{}
	}
	p := []v1.ApiParameter{}
	for _, f := range model.ApiFields {
		if f.ParameterInfo != nil {
			p = append(p, *f.ParameterInfo)
		}
	}
	model.ApiParameters = p
	if model.WebParams == nil {
		model.WebParams = []v1.WebParams{}
	}
	return model
}

func ToListModel(items *v1.ApiList) []*Api {
	var api []*Api = make([]*Api, len(items.Items))
	for i := range items.Items {
		api[i] = ToModel(&items.Items[i])
	}
	return api
}

func (s *Service) Validate(a *Api) error {
	if len(a.Name) == 0 {
		return fmt.Errorf("name is null")
	}
	if len(a.Serviceunit.ID) == 0 {
		return fmt.Errorf("serviceunit id is null")
	}
	if len(a.ApiType) == 0 {
		return fmt.Errorf("api type is null")
	}
	if len(a.AuthType) == 0 {
		return fmt.Errorf("api auth type is null")
	}
	a.Applications = []v1.Application{}

	if a.Frequency == 0 {
		return fmt.Errorf("frequency is null")
	}
	if a.Protocol != v1.HTTPS {
		a.Protocol = v1.HTTP
	}
	a.ReturnType = v1.Json

	su, err := s.getServiceunit(a.Serviceunit.ID)
	if err != nil {
		return fmt.Errorf("cannot get serviceunit: %+v", err)
	}
	if su.Spec.Type == "data" {
		switch a.Method {
		case v1.GET, v1.LIST:
		default:
			return fmt.Errorf("wrong method type: %s. only %s and %s are allowed", a.Method, v1.GET, v1.LIST)
		}
		for i, f := range a.ApiFields {
			p := f.ParameterInfo
			if p == nil {
				continue
			}
			if len(p.Name) == 0 {
				return fmt.Errorf("%dth parameter name is null", i)
			}
			if len(p.Type) == 0 {
				p.Type = v1.ParameterType("null")
			}
			switch p.Type {
			case v1.String, v1.Int, v1.Bool, v1.Float:
			default:
				return fmt.Errorf("%dth parameter type is wrong: %s", i, p.Type)
			}
			p.Operator = v1.Equal
			if len(p.Example) == 0 {
				return fmt.Errorf("%dth parameter example is null", i)
			}
			if len(p.Description) == 0 {
				return fmt.Errorf("%dth parameter description is null", i)
			}
		}
	}
	//参数校验
	if su.Spec.Type == "web" {
		switch a.Method {
		case v1.GET, v1.POST, v1.PUT, v1.DELETE, v1.PATCH:
		default:
			return fmt.Errorf("wrong method type: %s. ", a.Method)
		}
		for i, p := range a.WebParams {
			if len(p.Name) == 0 {
				return fmt.Errorf("%dth parameter name is null", i)
			}
			if len(p.Type) == 0 {
				p.Type = v1.ParameterType("null")
			}
			switch p.Type {
			case v1.String, v1.Int:
			default:
				return fmt.Errorf("%dth parameter type is wrong: %s", i, p.Type)
			}
			switch p.Location {
			case v1.Path, v1.Header, v1.Query:
			default:
				return fmt.Errorf("%dth parameter location is wrong: %s", i, p.Location)
			}
		}
	}
	a.UpdatedAt = time.Now()

	if !su.Status.Published {
		return fmt.Errorf("serviceunit %s is unpublished", a.Serviceunit.ID)
	}
	a.Serviceunit = v1.Serviceunit{
		ID:    su.ObjectMeta.Name,
		Name:  su.Spec.Name,
		Group: su.Spec.Group.Name,
	}
	a.ID = names.NewID()
	return nil
}
