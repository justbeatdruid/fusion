package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/crds/api/api/v1"
	dwv1 "github.com/chinamobile/nlpt/crds/api/datawarehouse/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Api struct {
	ID          string `json:"id"`
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Serviceunit   v1.Serviceunit    `json:"serviceunit"`
	Applications  []v1.Application  `json:"applications"`
	Users         []v1.User         `json:"users"`
	Frequency     int               `json:"frequency"`
	Method        v1.Method         `json:"method"`
	Protocol      v1.Protocol       `json:"protocol"`
	ReturnType    v1.ReturnType     `json:"returnType"`
	ApiFields     []v1.Field        `json:"apiFields"`
	Query         *dwv1.Query       `json:"dataserviceQuery,omitempty"`
	ApiParameters []v1.ApiParameter `json:"apiParameters"`
	WebParams     []v1.WebParams    `json:"webParams"`
	ApiType       v1.ApiType        `json:"apiType"`
	AuthType      v1.AuthType       `json:"authType"`
	Traffic       v1.Traffic        `json:"traffic"`
	KongApi       v1.KongApiInfo    `json:"KongApi"`

	Status v1.Status `json:"status"`
	//Publish          v1.Publish    `json:"publish"`
	AccessLink       v1.AccessLink `json:"access"`
	UpdatedAt        time.Time     `json:"updatedAt"`
	ReleasedAt       time.Time     `json:"releasedAt"`
	ApplicationCount int           `json:"applicationCount"`
	CalledCount      int           `json:"calledCount"`
	PublishInfo      v1.PublishInfo

	ApplicationBindStatus *v1.ApiApplicationStatus `json:"applicationBindStatus"`
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
		Description:  api.Description,
		Serviceunit:  api.Serviceunit,
		Applications: api.Applications,
		Users:        api.Users,
		Frequency:    api.Frequency,
		Method:       api.Method,
		Protocol:     api.Protocol,
		ReturnType:   api.ReturnType,
		ApiFields:    api.ApiFields,
		Query:        api.Query,
		WebParams:    api.WebParams,
		KongApi:      api.KongApi,
		ApiType:      api.ApiType,
		AuthType:     api.AuthType,
		Traffic:      api.Traffic,
	}
	crd.Status = v1.ApiStatus{
		Status:           v1.Init,
		AccessLink:       api.AccessLink,
		UpdatedAt:        metav1.Now(),
		ReleasedAt:       metav1.Now(),
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
		Description:  obj.Spec.Description,
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
		Traffic:      obj.Spec.Traffic,

		Status:           obj.Status.Status,
		AccessLink:       obj.Status.AccessLink,
		UpdatedAt:        obj.Status.UpdatedAt.Time,
		ReleasedAt:       obj.Status.ReleasedAt.Time,
		ApplicationCount: 0,
		CalledCount:      obj.Status.CalledCount,
	}
	if model.Applications == nil {
		model.Applications = []v1.Application{}
	}
	if model.ApiFields == nil {
		model.ApiFields = []v1.Field{}
	}

	// for data service (rdb)
	p := []v1.ApiParameter{}
	for _, f := range model.ApiFields {
		if f.ParameterInfo != nil {
			p = append(p, *f.ParameterInfo)
		}
	}
	model.ApiParameters = p

	// web params
	if model.WebParams == nil {
		model.WebParams = []v1.WebParams{}
	}

	// for data service (datawarehouse api)
	p = []v1.ApiParameter{}
	if obj.Spec.Query != nil {
		for _, w := range obj.Spec.Query.WhereFieldInfo {
			if w.ParameterEnabled {
				p = append(p, v1.ParameterFromQuery(w))
			}
		}
	}
	model.ApiParameters = p

	for l := range obj.ObjectMeta.Labels {
		if v1.IsApplicationLabel(l) {
			model.ApplicationCount = model.ApplicationCount + 1
		}
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
		if a.Query != nil {
			if err = a.Query.Validate(); err != nil {
				return fmt.Errorf("query field validate error: %+v")
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

func (s *Service) assignment(target *v1.Api, reqData interface{}) error {
	data, ok := reqData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("reqData type is error,req data: %v", reqData)
	}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err)
	}
	var source Api
	if err = json.Unmarshal(b, &source); err != nil {
		return fmt.Errorf("json.Unmarshal error,: %v", err)
	}
	if _, ok := data["name"]; ok {
		target.Spec.Name = source.Name
	}
	if _, ok := data["namespace"]; ok {
		target.ObjectMeta.Namespace = source.Namespace
	}
	if _, ok := data["applications"]; ok {
		target.Spec.Applications = source.Applications
	}
	if _, ok = data["users"]; ok {
		target.Spec.Users = source.Users
	}
	if _, ok := data["frequency"]; ok {
		target.Spec.Frequency = source.Frequency
	}
	if _, ok := data["method"]; ok {
		target.Spec.Method = source.Method
	}
	if _, ok := data["protocol"]; ok {
		target.Spec.Protocol = source.Protocol
	}
	if _, ok := data["apiFields"]; ok {
		target.Spec.ApiFields = source.ApiFields
	}
	if _, ok = data["returnType"]; ok {
		target.Spec.ReturnType = source.ReturnType
	}
	if _, ok = data["webParams"]; ok {
		target.Spec.WebParams = source.WebParams
	}
	if _, ok = data["apiType"]; ok {
		target.Spec.ApiType = source.ApiType
	}
	if _, ok = data["authType"]; ok {
		target.Spec.AuthType = source.AuthType
	}
	if _, ok = data["traffic"]; ok {
		target.Spec.Traffic = source.Traffic
	}
	if _, ok = data["KongApi"]; ok {
		target.Spec.KongApi = source.KongApi
	}
	target.Status.UpdatedAt = metav1.Now()
	return nil
}
