package service

import (
	"encoding/json"
	"fmt"
	"github.com/chinamobile/nlpt/pkg/names"
	"strings"

	"github.com/chinamobile/nlpt/crds/api/api/v1"
	dwv1 "github.com/chinamobile/nlpt/crds/api/datawarehouse/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/klog"
)

type Api struct {
	ID          string `json:"id"`
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Description string `json:"description"`

	Serviceunit    v1.Serviceunit   `json:"serviceunit"`
	Applications   []v1.Application `json:"applications"`
	Users          user.Users       `json:"users"`
	Frequency      int              `json:"frequency"`
	ApiType        v1.ApiType       `json:"apiType"`
	AuthType       v1.AuthType      `json:"authType"`
	Tags           string           `json:"tags"`
	ApiBackendType string           `json:"apiBackendType"`
	//data api
	Method                v1.Method                `json:"method"`
	Protocol              v1.Protocol              `json:"protocol"`
	ReturnType            v1.ReturnType            `json:"returnType"`
	RDBQuery              *v1.RDBQuery             `json:"rdbQuery,omitempty"`
	DataWarehouseQuery    *dwv1.DataWarehouseQuery `json:"datawarehouseQuery,omitempty"`
	ApiRequestParameters  []v1.ApiParameter        `json:"apiRequestParameters"`
	ApiResponseParameters []v1.ApiParameter        `json:"apiResponseParameters"`
	ApiPublicParameters   []v1.ApiParameter        `json:"apiPublicParameters"`
	//web api
	ApiDefineInfo v1.ApiDefineInfo `json:"apiDefineInfo"`
	KongApi       v1.KongApiInfo   `json:"kongApi"`
	ApiQueryInfo  v1.ApiQueryInfo  `json:"apiQueryInfo"`
	ApiReturnInfo v1.ApiReturnInfo `json:"apiReturnInfo"`

	Traffic     v1.Traffic     `json:"traffic"`
	Restriction v1.Restriction `json:"restriction"`

	Status           v1.Status        `json:"status"`
	Action           v1.Action        `json:"action"`
	PublishStatus    v1.PublishStatus `json:"publishStatus"`
	AccessLink       v1.AccessLink    `json:"access"`
	UpdatedAt        util.Time        `json:"updatedAt"`
	ReleasedAt       util.Time        `json:"releasedAt"`
	ApplicationCount int              `json:"applicationCount"`
	CalledCount      int              `json:"calledCount"`
	FailedCount      int              `json:"failedCount"`
	LatencyCount     int              `json:"latencyCount"`
	CallFrequency    int              `json:"callFrequency"`

	PublishInfo v1.PublishInfo

	ApplicationBindStatus *v1.ApiApplicationStatus `json:"applicationBindStatus"`
}

type Statistics struct {
	Total       int `json:"total"`
	Increment   int `json:"increment"`
	TotalCalled int `json:"totalCalled"`
}

// only used in creation
func ToAPI(api *Api) *v1.Api {
	crd := &v1.Api{}
	crd.TypeMeta.Kind = "Api"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = api.ID
	crd.ObjectMeta.Namespace = api.Namespace
	crd.ObjectMeta.Labels = make(map[string]string)
	crd.ObjectMeta.Labels[v1.ServiceunitLabel] = api.Serviceunit.ID
	crd.Spec = v1.ApiSpec{
		Name:               api.Name,
		Description:        api.Description,
		Serviceunit:        api.Serviceunit,
		Applications:       api.Applications,
		Frequency:          api.Frequency,
		ApiType:            api.ApiType,
		AuthType:           api.AuthType,
		Tags:               api.Tags,
		ApiBackendType:     api.Serviceunit.Type,
		Method:             api.Method,
		Protocol:           api.Protocol,
		ReturnType:         api.ReturnType,
		RDBQuery:           api.RDBQuery,
		DataWarehouseQuery: api.DataWarehouseQuery,
		ApiDefineInfo:      api.ApiDefineInfo,
		KongApi:            api.KongApi,
		ApiQueryInfo:       api.ApiQueryInfo,
		ApiReturnInfo:      api.ApiReturnInfo,
		Traffic:            api.Traffic,
		Restriction:        api.Restriction,
		PublishInfo:        api.PublishInfo,
	}

	crd.Status = v1.ApiStatus{
		Status: v1.Init,
		Action: v1.Create,
		//create api update status to unreleased
		PublishStatus:    v1.UnRelease,
		AccessLink:       api.AccessLink,
		UpdatedAt:        metav1.Now(),
		ReleasedAt:       metav1.Now(),
		ApplicationCount: api.ApplicationCount,
		CalledCount:      api.CalledCount,
		FailedCount:      api.FailedCount,
		LatencyCount:     api.LatencyCount,
		CallFrequency:    api.CallFrequency,
	}
	// add user labels
	crd.ObjectMeta.Labels = user.AddUsersLabels(api.Users, crd.ObjectMeta.Labels)
	return crd
}

func ToModel(obj *v1.Api) *Api {
	klog.V(5).Infof("obj to model: %+v", obj)
	model := &Api{
		ID:        obj.ObjectMeta.Name,
		Namespace: obj.ObjectMeta.Namespace,

		Name:           obj.Spec.Name,
		Description:    obj.Spec.Description,
		Serviceunit:    obj.Spec.Serviceunit,
		Applications:   obj.Spec.Applications,
		Frequency:      obj.Spec.Frequency,
		ApiType:        obj.Spec.ApiType,
		AuthType:       obj.Spec.AuthType,
		Tags:           obj.Spec.Tags,
		ApiBackendType: obj.Spec.Serviceunit.Type,
		Method:         obj.Spec.ApiDefineInfo.Method,
		Protocol:       obj.Spec.ApiDefineInfo.Protocol,
		ReturnType:     obj.Spec.ReturnType,
		ApiDefineInfo:  obj.Spec.ApiDefineInfo,
		KongApi:        obj.Spec.KongApi,
		ApiQueryInfo:   obj.Spec.ApiQueryInfo,
		ApiReturnInfo:  obj.Spec.ApiReturnInfo,
		Traffic:        obj.Spec.Traffic,
		Restriction:    obj.Spec.Restriction,
		PublishInfo:    obj.Spec.PublishInfo,

		Status:           obj.Status.Status,
		Action:           obj.Status.Action,
		PublishStatus:    obj.Status.PublishStatus,
		AccessLink:       obj.Status.AccessLink,
		UpdatedAt:        util.NewTime(obj.Status.UpdatedAt.Time),
		ReleasedAt:       util.NewTime(obj.Status.ReleasedAt.Time),
		ApplicationCount: 0,
		CalledCount:      obj.Status.CalledCount,
		FailedCount:      obj.Status.FailedCount,
		LatencyCount:     obj.Status.LatencyCount,
		CallFrequency:    obj.Status.CallFrequency,
	}

	if len(model.Method) == 0 {
		model.Method = obj.Spec.Method
	}
	if len(model.Protocol) == 0 {
		model.Protocol = obj.Spec.Protocol
	}

	if model.Applications == nil {
		model.Applications = []v1.Application{}
	}

	// for data service (rdb)
	if obj.Spec.RDBQuery != nil {
		model.ApiRequestParameters = model.ApiQueryInfo.WebParams

		q := []v1.ApiParameter{}
		for _, f := range obj.Spec.RDBQuery.QueryFields {
			q = append(q, v1.RDBParameterFromQuery(f))
		}
		model.ApiResponseParameters = q

		model.ApiPublicParameters = publicParameters(model.ApiDefineInfo.Method == v1.POST)
	}

	// web params
	if model.ApiQueryInfo.WebParams == nil {
		model.ApiQueryInfo.WebParams = []v1.WebParams{}
	}

	// for data service (datawarehouse api)
	if obj.Spec.DataWarehouseQuery != nil {
		model.ApiRequestParameters = model.ApiQueryInfo.WebParams

		q := []v1.ApiParameter{}
		for _, f := range obj.Spec.DataWarehouseQuery.Properties {
			klog.V(5).Infof("build resp params from field %+v", f)
			q = append(q, v1.ParameterFromDataWarehouseQuery(f))
		}
		model.ApiResponseParameters = q

		model.ApiPublicParameters = publicParameters(model.ApiDefineInfo.Method == v1.POST)
	}

	for l := range obj.ObjectMeta.Labels {
		if v1.IsApplicationLabel(l) {
			model.ApplicationCount = model.ApplicationCount + 1
		}
	}
	model.Users = user.GetUsersFromLabels(obj.ObjectMeta.Labels)
	return model
}

func publicParameters(post bool) []v1.ApiParameter {
	ap := []v1.ApiParameter{
		{
			Name:        "Authorization",
			Type:        "string",
			Example:     "Bearer {application-jwt-token}",
			Description: "请求头参数，请求Token，以应用的token代替示例括号中内容",
			Required:    true,
		},
		{
			Name:        "limit",
			Type:        "int",
			Example:     "5",
			Description: "查询参数，返回条目个数限制。位于Query",
			Required:    false,
		},
	}
	if post {
		ap = append(ap, v1.ApiParameter{
			Name:        "Content-Type",
			Type:        "string",
			Example:     "application/json",
			Description: "请求头参数，请求数据格式，只接受json类型",
			Required:    true,
		})
	}
	return ap
}

func ToListModel(items *v1.ApiList, publishedOnly bool, status string, opts ...util.OpOption) []*Api {
	if len(opts) > 0 || publishedOnly {
		nameLike := util.OpList(opts...).NameLike()
		res := util.OpList(opts...).Restriction()
		traff := util.OpList(opts...).Trafficcontrol()
		authType := util.OpList(opts...).AuthType()
		apiBackendType := util.OpList(opts...).ApiBackendType()
		var apis []*Api = make([]*Api, 0)
		for i, a := range items.Items {
			if len(nameLike) > 0 {
				if !strings.Contains(a.Spec.Name, nameLike) {
					continue
				}
			}
			if publishedOnly {
				if a.Status.PublishStatus != v1.Released {
					continue
				}
			}
			//根据发布状态查询API
			if len(status) > 0 {
				if a.Status.PublishStatus != v1.PublishStatus(status) {
					continue
				}
			}
			if res == "bind" {
				if len(a.Spec.Restriction.ID) == 0 {
					continue
				}
			} else if res == "unbind" {
				if len(a.Spec.Restriction.ID) != 0 {
					continue
				}
			}
			if traff == "bind" {
				if len(a.Spec.Traffic.ID) == 0 && len(a.Spec.Traffic.SpecialID) == 0 {
					continue
				}
			} else if traff == "unbind" {
				if len(a.Spec.Traffic.ID) != 0 || len(a.Spec.Traffic.SpecialID) != 0 {
					continue
				}
			}
			if len(authType) > 0 {
				if string(a.Spec.AuthType) != authType {
					continue
				}
			}
			if len(apiBackendType) > 0 {
				if a.Spec.ApiBackendType != apiBackendType {
					continue
				}
			}
			api := ToModel(&items.Items[i])
			apis = append(apis, api)
		}
		return apis
	}
	var apis []*Api = make([]*Api, len(items.Items))
	for i := range items.Items {
		apis[i] = ToModel(&items.Items[i])
	}
	return apis
}

func (s *Service) Validate(a *Api) error {
	if len(a.Name) == 0 {
		return fmt.Errorf("name is null")
	}
	klog.V(5).Infof("validate namespace is : %s", a.Namespace)
	apiList, err := s.ListApis(a.Namespace)
	if err != nil {
		return fmt.Errorf("cannot list api object: %+v", err)
	}
	for _, p := range apiList.Items {
		if p.Spec.Name == a.Name {
			return errors.NameDuplicatedError("api name duplicated: %s", p.Spec.Name)
		}
	}
	for _, p := range apiList.Items {
		for _, path := range p.Spec.KongApi.Paths {
			if path == a.KongApi.Paths[0] && p.Spec.ApiDefineInfo.Method == a.ApiDefineInfo.Method {
				return fmt.Errorf("path duplicated: %s", path)
			} else {
				continue
			}
		}
	}
	if len(a.Users.Owner.ID) == 0 {
		return fmt.Errorf("owner not set")
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

	if a.Protocol != v1.HTTPS {
		a.Protocol = v1.HTTP
	}
	a.ReturnType = v1.Json
	klog.V(5).Infof("validate serviceunit namespace is : %s", a.Namespace)
	su, err := s.getServiceunit(a.Serviceunit.ID, a.Namespace)
	if err != nil {
		return fmt.Errorf("cannot get serviceunit: %+v", err)
	}
	if su.Spec.Type == "data" {
		if len(a.ApiDefineInfo.Method) == 0 {
			a.ApiDefineInfo.Method = a.Method
		}
		//后端协议 若后端未设置协议默认和前端协议保持一致
		if len(a.ApiDefineInfo.Protocol) == 0 {
			a.ApiDefineInfo.Protocol = a.Protocol
		}
		if a.Frequency == 0 {
			return fmt.Errorf("frequency is null")
		}
		switch a.ApiDefineInfo.Method {
		case v1.GET, v1.POST:
		default:
			return fmt.Errorf("wrong method type: %s. only %s and %s are allowed", a.ApiDefineInfo.Method, v1.GET, v1.POST)
		}
		if a.RDBQuery != nil {
			for _, p := range a.RDBQuery.QueryFields {
				if err := p.Validate(); err != nil {
					return fmt.Errorf("rdb query field error: %+v", err)
				}
			}
		}
		if a.DataWarehouseQuery != nil {
			if a.DataWarehouseQuery.Type == "hql" {
				if len(a.DataWarehouseQuery.HQL) == 0 || len(a.DataWarehouseQuery.Database) == 0 {
					return fmt.Errorf("hql field error")
				}
			} else {
				if err = a.DataWarehouseQuery.Validate(); err != nil {
					return fmt.Errorf("query field validate error: %+v", err)
				}
			}
		}
	}
	//参数校验
	if su.Spec.Type == "web" {
		switch a.ApiDefineInfo.Protocol {
		case v1.HTTP, v1.HTTPS:
		default:
			return fmt.Errorf("wrong protocol type: %s. ", a.ApiDefineInfo.Protocol)
		}
		switch a.ApiDefineInfo.Method {
		case v1.GET, v1.POST, v1.PUT, v1.DELETE, v1.PATCH:
		default:
			return fmt.Errorf("wrong method type: %s. ", a.ApiDefineInfo.Method)
		}
		for i, p := range a.ApiQueryInfo.WebParams {
			if len(p.Name) == 0 || len(p.BackendInfo.Name) == 0 {
				return fmt.Errorf("%dth parameter name is null", i)
			}
			if len(p.Type) == 0 {
				p.Type = v1.ParameterType("null")
			}
			switch p.Type {
			case v1.String, v1.Int, v1.Bool:
			default:
				return fmt.Errorf("%dth parameter type is wrong: %s", i, p.Type)
			}
			switch p.Location {
			case v1.Path, v1.Header, v1.Query, v1.Body:
			default:
				return fmt.Errorf("%dth query parameter location is wrong: %s", i, p.Location)
			}
			switch p.BackendInfo.Location {
			case v1.Path, v1.Header, v1.Query, v1.Body:
			default:
				return fmt.Errorf("%dth backend parameter location is wrong: %s", i, p.BackendInfo.Location)
			}
		}
		// kongapi paths  正常返回值
		if len(a.KongApi.Paths) == 0 {
			return fmt.Errorf("api paths is null. ")
		}
		// return example is not required
		//if len(a.ApiReturnInfo.NormalExample) == 0 {
		//	return fmt.Errorf("normal example is null. ")
		//}
	}
	a.UpdatedAt = util.Now()

	//data api need service unit publish
	if su.Spec.Type == "data" && !su.Status.Published && !s.tenantEnabled {
		return errors.UnpublishedError("serviceunit %s is unpublished", a.Serviceunit.ID)
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
		if target.Spec.Name != source.Name {
			apiList, err := s.ListApis(target.ObjectMeta.Namespace)
			if err != nil {
				return fmt.Errorf("cannot list api object: %+v", err)
			}
			for _, p := range apiList.Items {
				if p.Spec.Name == source.Name {
					return errors.NameDuplicatedError("api name duplicated: %s", p.Spec.Name)
				}
			}
		}
		target.Spec.Name = source.Name
	}
	if _, ok := data["namespace"]; ok {
		target.ObjectMeta.Namespace = source.Namespace
	}
	if _, ok := data["applications"]; ok {
		target.Spec.Applications = source.Applications
	}
	/*
		if _, ok = data["users"]; ok {
			target.Spec.Users = source.Users
		}
	*/
	if _, ok := data["frequency"]; ok {
		target.Spec.Frequency = source.Frequency
	}
	if _, ok := data["method"]; ok {
		target.Spec.Method = source.Method
		if target.Spec.Serviceunit.Type == "data" {
			target.Spec.KongApi.Methods = []string{strings.ToUpper(string(target.Spec.Method))}
		}
	}
	//协议是从服务单元继承所以暂不支持更新暂时保留
	if _, ok := data["protocol"]; ok {
		target.Spec.Protocol = source.Protocol
		if target.Spec.Serviceunit.Type == "data" {
			target.Spec.KongApi.Protocols = []string{strings.ToLower(string(target.Spec.Protocol))}
		}
	}
	if _, ok := data["rdbQuery"]; ok {
		target.Spec.RDBQuery = source.RDBQuery
	}
	if _, ok = data["returnType"]; ok {
		target.Spec.ReturnType = source.ReturnType
	}
	if _, ok = data["webParams"]; ok {
		target.Spec.ApiQueryInfo.WebParams = source.ApiQueryInfo.WebParams
	}
	if _, ok = data["apiType"]; ok {
		target.Spec.ApiType = source.ApiType
	}
	if _, ok = data["authType"]; ok {
		target.Spec.AuthType = source.AuthType
	}

	if apiInfo, ok := data["apiDefineInfo"]; ok {
		if config, ok := apiInfo.(map[string]interface{}); ok {
			if _, ok = config["protocol"]; ok {
				target.Spec.ApiDefineInfo.Protocol = source.ApiDefineInfo.Protocol
			}
			if _, ok = config["path"]; ok {
				target.Spec.ApiDefineInfo.Path = source.ApiDefineInfo.Path
			}
			if _, ok = config["matchMode"]; ok {
				target.Spec.ApiDefineInfo.MatchMode = source.ApiDefineInfo.MatchMode
			}
			if _, ok = config["method"]; ok {
				target.Spec.ApiDefineInfo.Method = source.ApiDefineInfo.Method
			}
			if _, ok = config["cors"]; ok {
				target.Spec.ApiDefineInfo.Cors = source.ApiDefineInfo.Cors
			}
		}
	}

	if kongInfo, ok := data["kongApi"]; ok {
		if config, ok := kongInfo.(map[string]interface{}); ok {
			//web 类型直接使用kong传入的参数  path method
			if _, ok = config["paths"]; ok {
				target.Spec.KongApi.Paths = source.KongApi.Paths
			}
			if target.Spec.Serviceunit.Type == "web" {
				target.Spec.KongApi.Methods = []string{strings.ToUpper(string(source.ApiDefineInfo.Method))}
			}
			if _, ok = config["hosts"]; ok {
				target.Spec.KongApi.Hosts = source.KongApi.Hosts
			}
		}
	}

	if apiInfo, ok := data["apiReturnInfo"]; ok {
		if config, ok := apiInfo.(map[string]interface{}); ok {
			if _, ok = config["normalExample"]; ok {
				target.Spec.ApiReturnInfo.NormalExample = source.ApiReturnInfo.NormalExample
			}
			if _, ok = config["failureExample"]; ok {
				target.Spec.ApiReturnInfo.FailureExample = source.ApiReturnInfo.FailureExample
			}
		}
	}

	target.Status.UpdatedAt = metav1.Now()
	return nil
}

type Data struct {
	Headers []string            `json:"headers"`
	Columns map[string]string   `json:"columns"`
	Data    []map[string]string `json:"data"`
}
