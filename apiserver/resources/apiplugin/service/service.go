package service

import (
	"encoding/json"
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/util"
	"github.com/parnurzeal/gorequest"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"time"

	"github.com/chinamobile/nlpt/apiserver/database"
	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
	// "k8s.io/klog"
)

var crdNamespace = "default"

type Service struct {
	tenantEnabled bool
	db            *database.DatabaseConnection
	apiClient     dynamic.NamespaceableResourceInterface
	suClient      dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface, tenantEnabled bool, db *database.DatabaseConnection) *Service {
	return &Service{
		tenantEnabled: tenantEnabled,
		db:            db,
		apiClient:     client.Resource(apiv1.GetOOFSGVR()),
		suClient:      client.Resource(suv1.GetOOFSGVR()),
	}
}

func (s *Service) CreateApiPlugin(model *ApiPlugin) (*ApiPlugin, error) {
	model.Status = "unpublished"
	p, ss, err := ToModel(*model)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	if err = s.db.AddApiPlugin(p, ss); err != nil {
		return nil, fmt.Errorf("cannot write database: %+v", err)
	}
	klog.Infof("create apiplugin success :%+v", model)
	return model, nil
}

func (s *Service) ListApiPlugin(p ApiPlugin) ([]*ApiPlugin, error) {
	condition, _, err := ToModel(p)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	apiplugins, err := s.db.QueryApiPlugin(condition)
	if err != nil {
		return nil, fmt.Errorf("cannot read database: %+v", err)
	}
	result := make([]*ApiPlugin, len(apiplugins))
	for i := range apiplugins {
		apiplugin, err := FromModel(apiplugins[i], nil)
		if err != nil {
			return nil, fmt.Errorf("cannot get model: %+v", err)
		}
		result[i] = &apiplugin
	}
	return result, nil
}

func (s *Service) GetApiPlugin(id string) (*ApiPlugin, error) {
	p, ss, err := s.db.GetApiPlugin(id)
	if err != nil {
		return nil, fmt.Errorf("cannot query database: %+v", err)
	}
	product, err := FromModel(p, ss)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	return &product, nil
}

func (s *Service) DeleteApiPlugin(id string) error {
	if err := s.db.RemoveApiPlugin(id); err != nil {
		return fmt.Errorf("cannot write database: %+v", err)
	}
	return nil
}

func (s *Service) UpdateApiPlugin(model *ApiPlugin, id string) (*ApiPlugin, error) {
	existed, err := s.GetApiPlugin(id)
	if err != nil {
		return nil, fmt.Errorf("cannot find apiplugins with id %s: %+v", id, err)
	}
	if existed.User != model.User {
		return nil, fmt.Errorf("permission denied: wrong user")
	}
	if existed.Namespace != model.Namespace {
		return nil, fmt.Errorf("permission denied: wrong tenant")
	}
	//当前支持更新名称、描述信息、应用id和插件配置
	if len(model.Name) != 0 {
		existed.Name = model.Name
	}
	if len(model.Description) != 0 {
		existed.Description = model.Description
	}
	if len(model.ConsumerId) != 0 {
		existed.ConsumerId = model.ConsumerId
	}
	klog.Infof("get existed.Config config %+v", existed.Config)
	existed.Config = model.Config
	klog.Infof("update existed.Config config %+v", existed.Config)
	//


	var kongPluginId string
	for _, value := range existed.ApiRelation {
		kongPluginId = value.KongPluginId
	}
	switch model.Type {
	case "response-transformer":
		if err := s.UpdateResponseTransformerByKong(kongPluginId,existed); err != nil {
			klog.Errorf("patch response transformer by kong err : %+v", err)
		}
	case "request-transformer":
		if err := s.UpdateRequestTransformerByKong(kongPluginId,existed); err != nil {
			klog.Errorf("patch response transformer by kong err : %+v", err)
		}
	}

	//p, _, err := ToModel(*model)
	p, _, err := ToModel(*existed)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	//更新时只能更新名称描述信息，关联关系通过其他接口更新
	if err := s.db.UpdateApiPlugin(p, nil); err != nil {
		return nil, fmt.Errorf("cannot write database: %+v", err)
	}
	return model, nil
}

func (s *Service) UpdateApiPluginStatus(model *ApiPlugin) (*ApiPlugin, error) {
	existedProduct, err := s.GetApiPlugin(model.Id)
	if err != nil {
		return nil, fmt.Errorf("cannot find product with id %s: %+v", model.Id, err)
	}
	if existedProduct.User != model.User {
		return nil, fmt.Errorf("permission denied: wrong user")
	}
	if existedProduct.Namespace != model.Namespace {
		return nil, fmt.Errorf("permission denied: wrong tenant")
	}
	switch model.Status {
	case "online", "offline":
	default:
		return nil, fmt.Errorf("wrong status: %s", model.Status)
	}
	existedProduct.Status = model.Status
	p, _, err := ToModel(*existedProduct)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	if err := s.db.UpdateApiPlugin(p, nil); err != nil {
		return nil, fmt.Errorf("cannot write database: %+v", err)
	}
	return model, nil
}

func (s *Service) getAPi(id string, crdNamespace string) (*apiv1.Api, error) {
	crd, err := s.apiClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	su := &apiv1.Api{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.api info: %+v", su)
	return su, nil
}

func (s *Service) getServiceUnit(id string, crdNamespace string) (*suv1.Serviceunit, error) {
	crd, err := s.suClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	su := &suv1.Serviceunit{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunit info: %+v", su)
	return su, nil
}

func (s *Service) BatchBindOrRelease(groupId string, data BindReq, opts ...util.OpOption) error {
	switch data.Operation {
	case "bind":
		return s.BatchBindApi(groupId, data, opts...)
	case "unbind":
		return s.BatchUnbindApi(groupId, data, opts...)
	default:
		return fmt.Errorf("unknown operation %s, expect bind or unbind", data.Operation)
	}
}

func (s *Service) BatchBindApi(groupId string, body BindReq, opts ...util.OpOption) error {
	if len(body.Apis) == 0 {
		return fmt.Errorf("at least one api must select to bind")
	}

	existed, err := s.GetApiPlugin(groupId)
	if err != nil {
		return fmt.Errorf("cannot find apiplugin with id %s: %+v", groupId, err)
	}
	crdNamespace := util.OpList(opts...).Namespace()
	//user := util.OpList(opts...).User()

	if existed.Namespace != crdNamespace {
		return fmt.Errorf("permission denied: wrong tenant")
	}

	//先校验是否所有API满足绑定条件，有一个不满足直接返回错误
	status := make([]bool, 0)
	for _, api := range body.Apis {
		switch body.Type {
		case ApiType:
			apiSource, err := s.getAPi(api.ID, crdNamespace)
			if err != nil {
				return fmt.Errorf("cannot get api: %+v", err)
			}
			if apiSource.Status.PublishStatus != apiv1.Released {
				return fmt.Errorf("api not released: %s", apiSource.Spec.Name)
			}
		case ServiceunitType:
			_, err := s.getServiceUnit(api.ID, crdNamespace)
			if err != nil {
				return fmt.Errorf("cannot get serviceunit: %+v", err)
			}
		}
		isBind := false
		for _, relation := range existed.ApiRelation {
			if relation.TargetId == api.ID {
				isBind = true
				klog.Infof("api %s has bind apiplugin %s ", api.ID, groupId)
				break
			}
		}
		if !isBind {
			klog.Infof("api %s has no bind apiplugin %s ", api.ID, groupId)
		}
		status = append(status, isBind)
	}
	result := make([]model.ApiPluginRelation, 0)
	for index, value := range body.Apis {
		//检测是否已经绑定，已经绑定的api跳过
		if !status[index] {
			klog.Infof("apiplugins no bound to target and need bind %+v", value)
			kongPluginId, err := s.AddTransformerByKong(value.ID, existed)
			if err != nil {
				klog.Errorf("add rsp transformer by kong err : %+v", err)
			}
			result = append(result, model.ApiPluginRelation{
				ApiPluginId: groupId,
				TargetId:    value.ID,
				TargetType:  body.Type,
				KongPluginId: kongPluginId,
			})
		}
	}
	if err = s.db.AddApiPluginRelation(result); err != nil {
		return fmt.Errorf("cannot write database api relation: %+v", err)
	}

	klog.Infof("bind apis success :%+v", result)
	return nil
}

func (s *Service) BatchUnbindApi(groupId string, body BindReq, opts ...util.OpOption) error {
	if len(body.Apis) == 0 {
		return fmt.Errorf("at least one api must select to unbind")
	}

	existed, err := s.GetApiPlugin(groupId)
	if err != nil {
		return fmt.Errorf("cannot find apiplugin with id %s: %+v", groupId, err)
	}
	crdNamespace := util.OpList(opts...).Namespace()
	//user := util.OpList(opts...).User()

	if existed.Namespace != crdNamespace {
		return fmt.Errorf("permission denied: wrong tenant")
	}

	status := make([]bool, 0)
	relationIds := make([]int, 0)
	//先校验是否所有API满足绑定条件，有一个不满足直接返回错误
	for _, api := range body.Apis {
		switch body.Type {
		case ApiType:
			_, err := s.getAPi(api.ID, crdNamespace)
			if err != nil {
				return fmt.Errorf("cannot get api: %+v", err)
			}
		case ServiceunitType:
			_, err := s.getServiceUnit(api.ID, crdNamespace)
			if err != nil {
				return fmt.Errorf("cannot get serviceunit: %+v", err)
			}
		}
		isBind := false
		for _, relation := range existed.ApiRelation {
			if relation.TargetId == api.ID {
				isBind = true
				relationIds = append(relationIds, relation.Id)
				klog.Infof("api %s has bind apiplugin %s ", api.ID, groupId)
				break
			}
		}
		if !isBind {
			klog.Infof("api %s has no bind apiplugin %s ", api.ID, groupId)
			//return fmt.Errorf("apiplugin not bound to api")
		}
		status = append(status, isBind)
	}
	result := make([]model.ApiPluginRelation, 0)
	for index, value := range body.Apis {
		//已经检测是否绑定 只有都绑定才需要解绑
		if status[index] {
			klog.Infof("apiplugins has bound to api and need unbind %+v", value)
			var kongPluginId string
			for _, plugin := range existed.ApiRelation {
				if value.ID == plugin.TargetId {
					kongPluginId = plugin.KongPluginId
				}
			}
			klog.Infof("apiplugins id %+v", kongPluginId)
			if err := s.DeleteTransformerByKong(value.ID, kongPluginId); err != nil {
				klog.Errorf("delete rsp transformer by kong err : %+v", err)
			}
			result = append(result, model.ApiPluginRelation{
				Id:          relationIds[index],
				ApiPluginId: groupId,
				TargetId:    value.ID,
				TargetType:  body.Type,
			})
		}
	}
	if err = s.db.RemoveApiPluginRelation(result); err != nil {
		return fmt.Errorf("cannot write database api relation: %+v", err)
	}

	klog.Infof("unbind apis success :%+v", result)
	return nil
}

//add transformer by kong
func (s *Service) AddTransformerByKong(apiId string, existed *ApiPlugin) (kongPluginId string,err error) {
	switch existed.Type {
	case "response-transformer":
		kongPluginId, err := s.AddResponseTransformerByKong(apiId, existed)
		if err != nil {
			klog.Errorf("add response transformer by kong err : %+v", err)
		}
		return kongPluginId, nil
	case "request-transformer":
		kongPluginId, err := s.AddRequestTransformerByKong(apiId, existed)
		if err != nil {
			klog.Errorf("add request transformer by kong err : %+v", err)
		}
		return kongPluginId, nil
	default:
		return "", nil
	}
}
//add response transformer by kong
func (s *Service) AddResponseTransformerByKong(apiId string, existed *ApiPlugin) (kongPluginId string, err error) {
	//id := db.Spec.KongApi.KongID
	klog.Infof("begin create response transformer,the route id is %s", apiId)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s%s%s", schema, "kong-kong-admin", 8001, "/routes/", apiId, "/plugins"))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &ResTransformerRequestBody{}
	requestBody.Name = "response-transformer"
	if len(existed.ConsumerId) != 0 {
		requestBody.ConsumerId = existed.ConsumerId
	}
	econfig, err := json.Marshal(existed.Config)
	if err != nil {
		return "", fmt.Errorf("json.Marshal error,: %v", err)
	}
	var config ResTransformerConfig
	if err = json.Unmarshal(econfig, &config); err != nil {
		return "", fmt.Errorf("json.Unmarshal error,: %v", err)
	}

	//remove
	if len(config.Remove.Json) != 0 {
		for _, j := range config.Remove.Json {
			requestBody.Config.Remove.Json = append(requestBody.Config.Remove.Json, j)
		}
	}
	if len(config.Remove.Headers) != 0 {
		for _, h := range config.Remove.Headers {
			requestBody.Config.Remove.Headers = append(requestBody.Config.Remove.Headers, h)
		}
	}
	//rename
	if len(config.Rename.Headers) != 0 {
		for _, h := range config.Rename.Headers {
			requestBody.Config.Rename.Headers = append(requestBody.Config.Rename.Headers, h)
		}
	}
	//replace
	if len(config.Replace.Json) != 0 {
		for _, j := range config.Replace.Json {
			requestBody.Config.Replace.Json = append(requestBody.Config.Replace.Json, j)
		}
	}
	if len(config.Replace.Json_types) != 0 {
		for _, jt := range config.Replace.Json_types {
			requestBody.Config.Replace.Json_types = append(requestBody.Config.Replace.Json_types, jt)
		}
	}
	if len(config.Replace.Headers) != 0 {
		for _, h := range config.Replace.Headers {
			requestBody.Config.Replace.Headers = append(requestBody.Config.Replace.Headers, h)
		}
	}
	//add
	if len(config.Add.Json) != 0 {
		for _, j := range config.Add.Json {
			requestBody.Config.Add.Json = append(requestBody.Config.Add.Json, j)
		}
	}
	if len(config.Add.Json_types) != 0 {
		for _, jt := range config.Add.Json_types {
			requestBody.Config.Add.Json_types = append(requestBody.Config.Add.Json_types, jt)
		}
	}
	if len(config.Add.Headers) != 0 {
		for _, h := range config.Add.Headers {
			requestBody.Config.Add.Headers = append(requestBody.Config.Add.Headers, h)
		}
	}
	//append
	if len(config.Append.Json) != 0 {
		for _, j := range config.Append.Json {
			requestBody.Config.Append.Json = append(requestBody.Config.Append.Json, j)
		}
	}
	if len(config.Append.Json_types) != 0 {
		for _, jt := range config.Append.Json_types {
			requestBody.Config.Append.Json_types = append(requestBody.Config.Append.Json_types, jt)
		}
	}
	if len(config.Append.Headers) != 0 {
		for _, h := range config.Append.Headers {
			requestBody.Config.Append.Headers = append(requestBody.Config.Append.Headers, h)
		}
	}
	responseBody := &ResTransformerResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return "", fmt.Errorf("request for add response transformer error: %+v", errs)
	}
	klog.V(5).Infof("creation response transformer code: %d, body: %s ", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.V(5).Infof("create response transformer failed msg: %s\n", responseBody.Message)
		return "", fmt.Errorf("request for create response transformer error: receive wrong status code: %s", string(body))
	}
	//(*db).Spec.ResponseTransformer.Id = responseBody.ID
	kongPluginId = responseBody.ID
	if err != nil {
		return "", fmt.Errorf("create response transformer error %s", responseBody.Message)
	}
	return kongPluginId, nil
}
//add request transformer by kong
func (s *Service) AddRequestTransformerByKong(apiId string, existed *ApiPlugin) (kongPluginId string,err error) {
	//id := db.Spec.KongApi.KongID
	klog.Infof("begin create reqest transformer,the route id is %s", apiId)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Post(fmt.Sprintf("%s://%s:%d%s%s%s", schema, "kong-kong-admin", 8001, "/routes/", apiId, "/plugins"))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &ReqTransformerRequestBody{}
	requestBody.Name = "request-transformer"
	if len(existed.ConsumerId) != 0 {
		requestBody.ConsumerId = existed.ConsumerId
	}
	econfig, err := json.Marshal(existed.Config)
	if err != nil {
		return "", fmt.Errorf("json.Marshal error,: %v", err)
	}
	var config ReqTransformerConfig
	if err = json.Unmarshal(econfig, &config); err != nil {
		return "", fmt.Errorf("json.Unmarshal error,: %v", err)
	}
	//remove
	if len(config.Remove.Body) != 0 {
		for _, j := range config.Remove.Body {
			requestBody.Config.Remove.Body = append(requestBody.Config.Remove.Body, j)
		}
	}
	if len(config.Remove.Headers) != 0 {
		for _, j := range config.Remove.Headers {
			requestBody.Config.Remove.Headers = append(requestBody.Config.Remove.Headers, j)
		}
	}
	if len(config.Remove.Querystring) != 0 {
		for _, j := range config.Remove.Querystring {
			requestBody.Config.Remove.Querystring = append(requestBody.Config.Remove.Querystring, j)
		}
	}

	//rename
	if len(config.Rename.Body) != 0 {
		for _, h := range config.Rename.Body {
			requestBody.Config.Rename.Body = append(requestBody.Config.Rename.Body, h)
		}
	}
	if len(config.Rename.Headers) != 0 {
		for _, h := range config.Rename.Headers {
			requestBody.Config.Rename.Headers = append(requestBody.Config.Rename.Headers, h)
		}
	}
	if len(config.Rename.Querystring) != 0 {
		for _, h := range config.Rename.Querystring {
			requestBody.Config.Rename.Querystring = append(requestBody.Config.Rename.Querystring, h)
		}
	}
	//replace
	if len(config.Replace.Body) != 0 {
		for _, h := range config.Replace.Body {
			requestBody.Config.Replace.Body = append(requestBody.Config.Replace.Body, h)
		}
	}
	if len(config.Replace.Headers) != 0 {
		for _, h := range config.Replace.Headers {
			requestBody.Config.Replace.Headers = append(requestBody.Config.Replace.Headers, h)
		}
	}
	if len(config.Replace.Querystring) != 0 {
		for _, h := range config.Replace.Querystring {
			requestBody.Config.Replace.Querystring = append(requestBody.Config.Replace.Querystring, h)
		}
	}
	requestBody.Config.Replace.Uri = config.Replace.Uri

	//add
	if len(config.Add.Body) != 0 {
		for _, h := range config.Add.Body {
			requestBody.Config.Add.Body = append(requestBody.Config.Add.Body, h)
		}
	}
	if len(config.Add.Headers) != 0 {
		for _, h := range config.Add.Headers {
			requestBody.Config.Add.Headers = append(requestBody.Config.Add.Headers, h)
		}
	}
	if len(config.Add.Querystring) != 0 {
		for _, h := range config.Add.Querystring {
			requestBody.Config.Add.Querystring = append(requestBody.Config.Add.Querystring, h)
		}
	}
	//append
	if len(config.Append.Body) != 0 {
		for _, h := range config.Append.Body {
			requestBody.Config.Append.Body = append(requestBody.Config.Append.Body, h)
		}
	}
	if len(config.Append.Headers) != 0 {
		for _, h := range config.Append.Headers {
			requestBody.Config.Append.Headers = append(requestBody.Config.Append.Headers, h)
		}
	}
	if len(config.Append.Querystring) != 0 {
		for _, h := range config.Append.Querystring {
			requestBody.Config.Append.Querystring = append(requestBody.Config.Append.Querystring, h)
		}
	}
	responseBody := &ReqTransformerResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return "", fmt.Errorf("request for add response transformer error: %+v", errs)
	}
	klog.V(5).Infof("creation response transformer code: %d, body: %s ", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.V(5).Infof("create response transformer failed msg: %s\n", responseBody.Message)
		return "", fmt.Errorf("request for create response transformer error: receive wrong status code: %s", string(body))
	}
	//(*db).Spec.ResponseTransformer.Id = responseBody.ID
	kongPluginId = responseBody.ID
	if err != nil {
		return "", fmt.Errorf("create response transformer error %s", responseBody.Message)
	}
	return kongPluginId, nil
}

func (s *Service) DeleteTransformerByKong(apiId string, kongPluginId string) (err error) {
	klog.Infof("delete response transformer the id of api is %s,the kong_id of response transformer %s", apiId, kongPluginId)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	for k, v := range headers {
		request = request.Set(k, v)
	}
	klog.Infof("delete kongPluginId id %s %s", kongPluginId, fmt.Sprintf("%s://%s:%d%s/%s", schema, "kong-kong-admin", 8001, "/plugins", kongPluginId))
	response, body, errs := request.Delete(fmt.Sprintf("%s://%s:%d%s/%s", schema, "kong-kong-admin", 8001, "/plugins", kongPluginId)).End()
	request = request.Retry(3, 5*time.Second, retryStatus...)
	if len(errs) > 0 {
		return fmt.Errorf("request for delete response transformer error: %+v", errs)
	}
	klog.V(5).Infof("delete response transformer response code: %d%s", response.StatusCode, string(body))
	if response.StatusCode != 204 {
		return fmt.Errorf("request for delete response transformer error: receive wrong status code: %d", response.StatusCode)
	}
	return nil
}

//update response transformer
func (s *Service) UpdateResponseTransformerByKong(kongPluginId string, existed *ApiPlugin) (err error) {
	klog.Infof("update kongPluginId is:%s, Host:%s, Port:%d", kongPluginId, "kong-kong-admin", 8001)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	klog.Infof("update response transformer id %s %s", kongPluginId, fmt.Sprintf("%s://%s:%d%s/%s", schema, "kong-kong-admin", 8001, "/plugins", kongPluginId))
	request = request.Patch(fmt.Sprintf("%s://%s:%d%s/%s", schema, "kong-kong-admin", 8001, "/plugins", kongPluginId))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &ResTransformerRequestBody{}
	requestBody.Name = "response-transformer"
	if len(existed.ConsumerId) != 0 {
		requestBody.ConsumerId = existed.ConsumerId
	}
	econfig, err := json.Marshal(existed.Config)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err)
	}
	var config ResTransformerConfig
	if err = json.Unmarshal(econfig, &config); err != nil {
		return fmt.Errorf("json.Unmarshal error,: %v", err)
	}

	//remove
	if len(config.Remove.Json) != 0 {
		for _, j := range config.Remove.Json {
			requestBody.Config.Remove.Json = append(requestBody.Config.Remove.Json, j)
		}
	}
	if len(config.Remove.Headers) != 0 {
		for _, h := range config.Remove.Headers {
			requestBody.Config.Remove.Headers = append(requestBody.Config.Remove.Headers, h)
		}
	}
	//rename
	if len(config.Rename.Headers) != 0 {
		for _, h := range config.Rename.Headers {
			requestBody.Config.Rename.Headers = append(requestBody.Config.Rename.Headers, h)
		}
	}
	//replace
	if len(config.Replace.Json) != 0 {
		for _, j := range config.Replace.Json {
			requestBody.Config.Replace.Json = append(requestBody.Config.Replace.Json, j)
		}
	}
	if len(config.Replace.Json_types) != 0 {
		for _, jt := range config.Replace.Json_types {
			requestBody.Config.Replace.Json_types = append(requestBody.Config.Replace.Json_types, jt)
		}
	}
	if len(config.Replace.Headers) != 0 {
		for _, h := range config.Replace.Headers {
			requestBody.Config.Replace.Headers = append(requestBody.Config.Replace.Headers, h)
		}
	}
	//add
	if len(config.Add.Json) != 0 {
		for _, j := range config.Add.Json {
			requestBody.Config.Add.Json = append(requestBody.Config.Add.Json, j)
		}
	}
	if len(config.Add.Json_types) != 0 {
		for _, jt := range config.Add.Json_types {
			requestBody.Config.Add.Json_types = append(requestBody.Config.Add.Json_types, jt)
		}
	}
	if len(config.Add.Headers) != 0 {
		for _, h := range config.Add.Headers {
			requestBody.Config.Add.Headers = append(requestBody.Config.Add.Headers, h)
		}
	}
	//append
	if len(config.Append.Json) != 0 {
		for _, j := range config.Append.Json {
			requestBody.Config.Append.Json = append(requestBody.Config.Append.Json, j)
		}
	}
	if len(config.Append.Json_types) != 0 {
		for _, jt := range config.Append.Json_types {
			requestBody.Config.Append.Json_types = append(requestBody.Config.Append.Json_types, jt)
		}
	}
	if len(config.Append.Headers) != 0 {
		for _, h := range config.Append.Headers {
			requestBody.Config.Append.Headers = append(requestBody.Config.Append.Headers, h)
		}
	}
	responseBody := &ResTransformerResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for update response transformer error: %+v", errs)
	}
	klog.V(5).Infof("update response transformer code: %d, body: %s ", response.StatusCode, string(body))
	if response.StatusCode != 200 {
		klog.V(5).Infof("update response transformer failed msg: %s\n", responseBody.Message)
		return fmt.Errorf("request for update response transformer error: receive wrong status code: %s", string(body))
	}
	if err != nil {
		return fmt.Errorf("create response transformer error %s", responseBody.Message)
	}
	return nil
}

//update request transformer
func (s *Service) UpdateRequestTransformerByKong(kongPluginId string, existed *ApiPlugin) (err error) {
	klog.Infof("update kongPluginId is:%s, Host:%s, Port:%d", kongPluginId, "kong-kong-admin", 8001)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	klog.Infof("update request transformer id %s %s", kongPluginId, fmt.Sprintf("%s://%s:%d%s/%s", schema, "kong-kong-admin", 8001, "/plugins", kongPluginId))
	request = request.Patch(fmt.Sprintf("%s://%s:%d%s/%s", schema, "kong-kong-admin", 8001, "/plugins", kongPluginId))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &ReqTransformerRequestBody{}
	requestBody.Name = "request-transformer"
	if len(existed.ConsumerId) != 0 {
		requestBody.ConsumerId = existed.ConsumerId
	}
	econfig, err := json.Marshal(existed.Config)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err)
	}
	var config ReqTransformerConfig
	if err = json.Unmarshal(econfig, &config); err != nil {
		return fmt.Errorf("json.Unmarshal error,: %v", err)
	}

	//remove
	if len(config.Remove.Body) != 0 {
		for _, j := range config.Remove.Body {
			requestBody.Config.Remove.Body = append(requestBody.Config.Remove.Body, j)
		}
	}
	if len(config.Remove.Headers) != 0 {
		for _, j := range config.Remove.Headers {
			requestBody.Config.Remove.Headers = append(requestBody.Config.Remove.Headers, j)
		}
	}
	if len(config.Remove.Querystring) != 0 {
		for _, j := range config.Remove.Querystring {
			requestBody.Config.Remove.Querystring = append(requestBody.Config.Remove.Querystring, j)
		}
	}

	//rename
	if len(config.Rename.Body) != 0 {
		for _, h := range config.Rename.Body {
			requestBody.Config.Rename.Body = append(requestBody.Config.Rename.Body, h)
		}
	}
	if len(config.Rename.Headers) != 0 {
		for _, h := range config.Rename.Headers {
			requestBody.Config.Rename.Headers = append(requestBody.Config.Rename.Headers, h)
		}
	}
	if len(config.Rename.Querystring) != 0 {
		for _, h := range config.Rename.Querystring {
			requestBody.Config.Rename.Querystring = append(requestBody.Config.Rename.Querystring, h)
		}
	}
	//replace
	if len(config.Replace.Body) != 0 {
		for _, h := range config.Replace.Body {
			requestBody.Config.Replace.Body = append(requestBody.Config.Replace.Body, h)
		}
	}
	if len(config.Replace.Headers) != 0 {
		for _, h := range config.Replace.Headers {
			requestBody.Config.Replace.Headers = append(requestBody.Config.Replace.Headers, h)
		}
	}
	if len(config.Replace.Querystring) != 0 {
		for _, h := range config.Replace.Querystring {
			requestBody.Config.Replace.Querystring = append(requestBody.Config.Replace.Querystring, h)
		}
	}
	requestBody.Config.Replace.Uri = config.Replace.Uri

	//add
	if len(config.Add.Body) != 0 {
		for _, h := range config.Add.Body {
			requestBody.Config.Add.Body = append(requestBody.Config.Add.Body, h)
		}
	}
	if len(config.Add.Headers) != 0 {
		for _, h := range config.Add.Headers {
			requestBody.Config.Add.Headers = append(requestBody.Config.Add.Headers, h)
		}
	}
	if len(config.Add.Querystring) != 0 {
		for _, h := range config.Add.Querystring {
			requestBody.Config.Add.Querystring = append(requestBody.Config.Add.Querystring, h)
		}
	}
	//append
	if len(config.Append.Body) != 0 {
		for _, h := range config.Append.Body {
			requestBody.Config.Append.Body = append(requestBody.Config.Append.Body, h)
		}
	}
	if len(config.Append.Headers) != 0 {
		for _, h := range config.Append.Headers {
			requestBody.Config.Append.Headers = append(requestBody.Config.Append.Headers, h)
		}
	}
	if len(config.Append.Querystring) != 0 {
		for _, h := range config.Append.Querystring {
			requestBody.Config.Append.Querystring = append(requestBody.Config.Append.Querystring, h)
		}
	}

	responseBody := &ReqTransformerResponseBody{}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for update request transformer error: %+v", errs)
	}
	klog.V(5).Infof("update request transformer code: %d, body: %s ", response.StatusCode, string(body))
	if response.StatusCode != 201 {
		klog.V(5).Infof("update request transformer failed msg: %s\n", responseBody.Message)
		return fmt.Errorf("request for update request transformer error: receive wrong status code: %s", string(body))
	}
	if err != nil {
		return fmt.Errorf("create request transformer error %s", responseBody.Message)
	}
	return nil
}

