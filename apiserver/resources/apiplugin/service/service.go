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
	klog.Infof("query api group by condition :%+v", p)
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
		if err := s.UpdateResponseTransformerByKong(kongPluginId, existed); err != nil {
			klog.Errorf("patch response transformer by kong err : %+v", err)
		}
	case "request-transformer":
		if err := s.UpdateRequestTransformerByKong(kongPluginId, existed); err != nil {
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
		bindStatus := BindInit
		detail := BindInit
		enable := false
		//检测是否已经绑定，已经绑定的api跳过
		if !status[index] {
			klog.Infof("apiplugins no bound to target and need bind %+v", value)
			kongPluginId, err := s.AddTransformerByKong(value.ID, existed)
			if err != nil {
				klog.Errorf("add transformer by kong err : %+v", err)
				bindStatus = BindFailed
				detail = fmt.Sprintf("add transformer by kong err : %+v", err)
			} else {
				bindStatus = BindSuccess
				detail = "bind success"
				enable = true
			}
			result = append(result, model.ApiPluginRelation{
				ApiPluginId:  groupId,
				TargetId:     value.ID,
				TargetType:   body.Type,
				KongPluginId: kongPluginId,
				Status:       bindStatus,
				Detail:       detail,
				Enable:       enable,
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
	updateResult := make([]model.ApiPluginRelation, 0)
	for index, value := range body.Apis {
		//已经检测是否绑定 只有都绑定才需要解绑
		bindStatus := UnbindInit
		detail := UnbindInit
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
				bindStatus = UnbindFailed
				detail = fmt.Sprintf("delete rsp transformer by kong err : %+v", err)
				updateResult = append(result, model.ApiPluginRelation{
					Id:          relationIds[index],
					ApiPluginId: groupId,
					TargetId:    value.ID,
					TargetType:  body.Type,
					Status:      bindStatus,
					Detail:      detail,
					Enable:      false,
				})
			}
			result = append(result, model.ApiPluginRelation{
				Id:          relationIds[index],
				ApiPluginId: groupId,
				TargetId:    value.ID,
				TargetType:  body.Type,
			})
		}
	}
	//解绑时分2种情况 解绑成功的需要删除 解绑失败的更新状态为失败不删除
	if err = s.db.RemoveApiPluginRelation(result); err != nil {
		return fmt.Errorf("cannot delete database api plugin relation: %+v", err)
	}
	if err = s.db.UpdateApiPluginRelation(updateResult); err != nil {
		return fmt.Errorf("cannot update database api plugin relation: %+v", err)
	}

	klog.Infof("unbind apis success :%+v", result)
	return nil
}

//add transformer by kong
func (s *Service) AddTransformerByKong(apiId string, existed *ApiPlugin) (kongPluginId string, err error) {
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
		requestBody.Consumer = &Consumer{}
		requestBody.Consumer.Id = existed.ConsumerId
	}
	econfig, err := json.Marshal(existed.Config)
	if err != nil {
		return "", fmt.Errorf("json.Marshal error,: %v", err)
	}
	klog.Infof("xxxxxxx econfig is %s", string(econfig))
	var responseTransformer OutResponseTransformer
	if err = json.Unmarshal(econfig, &responseTransformer); err != nil {
		return "", fmt.Errorf("json.Unmarshal error,: %v", err)
	}
	klog.Infof("xxxxxx config is  %+v", responseTransformer)

	//remove
	if len(responseTransformer.Config.Remove.Json) != 0 {
		for _, j := range responseTransformer.Config.Remove.Json {
			requestBody.Config.Remove.Json = append(requestBody.Config.Remove.Json, j)
		}
	}
	if len(responseTransformer.Config.Remove.Headers) != 0 {
		for _, h := range responseTransformer.Config.Remove.Headers {
			requestBody.Config.Remove.Headers = append(requestBody.Config.Remove.Headers, h)
		}
	}
	//rename
	for i, _ := range responseTransformer.Config.Rename.Headers {
		requestBody.Config.Rename.Headers = append(requestBody.Config.Rename.Headers,
			responseTransformer.Config.Rename.Headers[i].Key+":"+responseTransformer.Config.Rename.Headers[i].Value)
	}
	//replace
	for i, _ := range responseTransformer.Config.Replace.Headers {
		requestBody.Config.Replace.Headers = append(requestBody.Config.Replace.Headers,
			responseTransformer.Config.Replace.Headers[i].Key+":"+responseTransformer.Config.Replace.Headers[i].Value)
	}
	for i, _ := range responseTransformer.Config.Replace.Json {
		requestBody.Config.Replace.Json = append(requestBody.Config.Replace.Json,
			responseTransformer.Config.Replace.Json[i].Key+":"+responseTransformer.Config.Replace.Json[i].Value)
		requestBody.Config.Replace.Json_types = append(
			requestBody.Config.Replace.Json_types, responseTransformer.Config.Replace.Json[i].Type)
	}

	//add
	for i, _ := range responseTransformer.Config.Add.Headers {
		requestBody.Config.Add.Headers = append(requestBody.Config.Add.Headers,
			responseTransformer.Config.Add.Headers[i].Key+":"+responseTransformer.Config.Add.Headers[i].Value)
	}
	for i, _ := range responseTransformer.Config.Add.Json {
		requestBody.Config.Add.Json = append(requestBody.Config.Add.Json,
			responseTransformer.Config.Add.Json[i].Key+":"+responseTransformer.Config.Add.Json[i].Value)
		requestBody.Config.Add.Json_types = append(requestBody.Config.Add.Json_types,
			responseTransformer.Config.Add.Json[i].Type)
	}
	//append
	/*
		for i, _ := range config.Append.Headers {
			requestBody.Config.Append.Headers = append(requestBody.Config.Append.Headers,
				config.Add.Headers[i].Key+":"+config.Add.Headers[i].Value)
		}
		for i, _ := range  config.Append.Json {
			requestBody.Config.Add.Json = append(requestBody.Config.Add.Json,
				config.Append.Json[i].Key+":"+config.Append.Json[i].Value)
			requestBody.Config.Add.Json_types = append(requestBody.Config.Add.Json_types,
				config.Append.Json[i].Type)
		}
	*/
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
func (s *Service) AddRequestTransformerByKong(apiId string, existed *ApiPlugin) (kongPluginId string, err error) {
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
		requestBody.Consumer = &Consumer{}
		requestBody.Consumer.Id = existed.ConsumerId
	}
	econfig, err := json.Marshal(existed.Config)
	if err != nil {
		return "", fmt.Errorf("json.Marshal error,: %v", err)
	}
	var requestTransformer OutRequestTransformer
	if err = json.Unmarshal(econfig, &requestTransformer); err != nil {
		return "", fmt.Errorf("json.Unmarshal error,: %v", err)
	}
	//remove
	if len(requestTransformer.Config.HttpMethod) != 0 {
		requestBody.Config.HttpMethod = requestTransformer.Config.HttpMethod
	}
	for i, _ := range requestTransformer.Config.Remove.Body {
		requestBody.Config.Remove.Body = append(requestBody.Config.Remove.Body, requestTransformer.Config.Remove.Body[i])
	}
	for i, _ := range requestTransformer.Config.Remove.Headers {
		requestBody.Config.Remove.Headers = append(requestBody.Config.Remove.Headers, requestTransformer.Config.Remove.Headers[i])
	}
	for i, _ := range requestTransformer.Config.Remove.Querystring {
		requestBody.Config.Remove.Querystring = append(requestBody.Config.Remove.Querystring, requestTransformer.Config.Remove.Querystring[i])
	}
	//rename
	for i, _ := range requestTransformer.Config.Rename.Body {
		requestBody.Config.Rename.Body = append(requestBody.Config.Rename.Body,
			requestTransformer.Config.Rename.Body[i].Key+":"+requestTransformer.Config.Rename.Body[i].Value)
	}
	for i, _ := range requestTransformer.Config.Rename.Headers {
		requestBody.Config.Rename.Headers = append(requestBody.Config.Rename.Headers,
			requestTransformer.Config.Rename.Headers[i].Key+":"+requestTransformer.Config.Rename.Headers[i].Value)
	}
	for i, _ := range requestTransformer.Config.Rename.Querystring {
		requestBody.Config.Rename.Querystring = append(requestBody.Config.Rename.Querystring,
			requestTransformer.Config.Rename.Querystring[i].Key+":"+requestTransformer.Config.Rename.Querystring[i].Value)
	}
	//replace
	for i, _ := range requestTransformer.Config.Replace.Body {
		requestBody.Config.Replace.Body = append(requestBody.Config.Replace.Body,
			requestTransformer.Config.Replace.Body[i].Key+":"+requestTransformer.Config.Replace.Body[i].Value)
	}
	for i, _ := range requestTransformer.Config.Replace.Headers {
		requestBody.Config.Replace.Headers = append(requestBody.Config.Replace.Headers,
			requestTransformer.Config.Replace.Headers[i].Key+":"+requestTransformer.Config.Replace.Headers[i].Value)
	}
	for i, _ := range requestTransformer.Config.Replace.Querystring {
		requestBody.Config.Replace.Querystring = append(requestBody.Config.Replace.Querystring,
			requestTransformer.Config.Replace.Querystring[i].Key+":"+requestTransformer.Config.Replace.Querystring[i].Value)
	}
	requestBody.Config.Replace.Uri = existed.ReplaceUri

	//add
	for i, _ := range requestTransformer.Config.Add.Body {
		requestBody.Config.Add.Body = append(requestBody.Config.Add.Body,
			requestTransformer.Config.Add.Body[i].Key+":"+requestTransformer.Config.Add.Body[i].Value)
	}
	for i, _ := range requestTransformer.Config.Add.Headers {
		requestBody.Config.Add.Headers = append(requestBody.Config.Add.Headers,
			requestTransformer.Config.Add.Headers[i].Key+":"+requestTransformer.Config.Add.Headers[i].Value)
	}
	for i, _ := range requestTransformer.Config.Add.Querystring {
		requestBody.Config.Add.Querystring = append(requestBody.Config.Add.Querystring,
			requestTransformer.Config.Add.Querystring[i].Key+":"+requestTransformer.Config.Add.Querystring[i].Value)
	}
	//append
	for i, _ := range requestTransformer.Config.Append.Body {
		requestBody.Config.Append.Body = append(requestBody.Config.Append.Body,
			requestTransformer.Config.Append.Body[i].Key+":"+requestTransformer.Config.Append.Body[i].Value)
	}
	for i, _ := range requestTransformer.Config.Append.Headers {
		requestBody.Config.Append.Headers = append(requestBody.Config.Append.Headers,
			requestTransformer.Config.Append.Headers[i].Key+":"+requestTransformer.Config.Append.Headers[i].Value)
	}
	for i, _ := range requestTransformer.Config.Append.Querystring {
		requestBody.Config.Append.Querystring = append(requestBody.Config.Append.Querystring,
			requestTransformer.Config.Append.Querystring[i].Key+":"+requestTransformer.Config.Append.Querystring[i].Value)
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
		requestBody.Consumer = &Consumer{}
		requestBody.Consumer.Id = existed.ConsumerId
	}
	econfig, err := json.Marshal(existed.Config)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err)
	}
	var responseTransformer OutResponseTransformer
	if err = json.Unmarshal(econfig, &responseTransformer); err != nil {
		return fmt.Errorf("json.Unmarshal error,: %v", err)
	}

	//remove
	if len(responseTransformer.Config.Remove.Json) != 0 {
		for _, j := range responseTransformer.Config.Remove.Json {
			requestBody.Config.Remove.Json = append(requestBody.Config.Remove.Json, j)
		}
	}
	if len(responseTransformer.Config.Remove.Headers) != 0 {
		for _, h := range responseTransformer.Config.Remove.Headers {
			requestBody.Config.Remove.Headers = append(requestBody.Config.Remove.Headers, h)
		}
	}
	//rename
	for i, _ := range responseTransformer.Config.Rename.Headers {
		requestBody.Config.Rename.Headers = append(requestBody.Config.Rename.Headers,
			responseTransformer.Config.Rename.Headers[i].Key+":"+responseTransformer.Config.Rename.Headers[i].Value)
	}
	//replace
	for i, _ := range responseTransformer.Config.Replace.Headers {
		requestBody.Config.Replace.Headers = append(requestBody.Config.Replace.Headers,
			responseTransformer.Config.Replace.Headers[i].Key+":"+responseTransformer.Config.Replace.Headers[i].Value)
	}
	for i, _ := range responseTransformer.Config.Replace.Json {
		requestBody.Config.Replace.Json = append(requestBody.Config.Replace.Json,
			responseTransformer.Config.Replace.Json[i].Key+":"+responseTransformer.Config.Replace.Json[i].Value)
		requestBody.Config.Replace.Json_types = append(
			requestBody.Config.Replace.Json_types, responseTransformer.Config.Replace.Json[i].Type)
	}

	//add
	for i, _ := range responseTransformer.Config.Add.Headers {
		requestBody.Config.Add.Headers = append(requestBody.Config.Add.Headers,
			responseTransformer.Config.Add.Headers[i].Key+":"+responseTransformer.Config.Add.Headers[i].Value)
	}
	for i, _ := range responseTransformer.Config.Add.Json {
		requestBody.Config.Add.Json = append(requestBody.Config.Add.Json,
			responseTransformer.Config.Add.Json[i].Key+":"+responseTransformer.Config.Add.Json[i].Value)
		requestBody.Config.Add.Json_types = append(requestBody.Config.Add.Json_types,
			responseTransformer.Config.Add.Json[i].Type)
	}
	//append
	/*
		for i, _ := range config.Append.Headers {
			requestBody.Config.Append.Headers = append(requestBody.Config.Append.Headers,
				config.Add.Headers[i].Key+":"+config.Add.Headers[i].Value)
		}
		for i, _ := range  config.Append.Json {
			requestBody.Config.Add.Json = append(requestBody.Config.Add.Json,
				config.Append.Json[i].Key+":"+config.Append.Json[i].Value)
			requestBody.Config.Add.Json_types = append(requestBody.Config.Add.Json_types,
				config.Append.Json[i].Type)
		}
	*/
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
		requestBody.Consumer = &Consumer{}
		requestBody.Consumer.Id = existed.ConsumerId
	}
	econfig, err := json.Marshal(existed.Config)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err)
	}
	var config ReqTransformerConfig
	if err = json.Unmarshal(econfig, &config); err != nil {
		return fmt.Errorf("json.Unmarshal error,: %v", err)
	}
	if len(config.HttpMethod) != 0 {
		requestBody.Config.HttpMethod = config.HttpMethod
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

func (s *Service) ListPluginRelationFromDatabase(id, types string, opts ...util.OpOption) ([]*ApiRes, error) {
	if !s.tenantEnabled {
		return nil, fmt.Errorf("unspported for apiserver with tenant disabled")
	}
	sql := ""
	switch types {
	case ApiType:
		sqlTpl := `SELECT api.id, api.name, api_plugin_relation.status, api_plugin_relation.detail, api_plugin_relation.enable, api_plugin_relation.kong_plugin_id FROM api_plugin_relation LEFT JOIN api on api.id = api_plugin_relation.target_id 
              where api_plugin_relation.api_plugin_id="%s" and api.namespace="%s" and api_plugin_relation.target_type="%s"`
		sql = fmt.Sprintf(sqlTpl, id, util.OpList(opts...).Namespace(), types)
		klog.Infof("query api sql: %s", sql)
	case ServiceunitType:
		sqlTpl := `SELECT serviceunit.id, serviceunit.name, api_plugin_relation.status, api_plugin_relation.detail, api_plugin_relation.enable, api_plugin_relation.kong_plugin_id FROM api_plugin_relation LEFT JOIN serviceunit on serviceunit.id = api_plugin_relation.target_id 
              where api_plugin_relation.api_plugin_id="%s" and serviceunit.namespace="%s" and api_plugin_relation.target_type="%s"`
		sql = fmt.Sprintf(sqlTpl, id, util.OpList(opts...).Namespace(), types)
		klog.Infof("query serviceunit sql: %s", sql)
	}
	mResult := make([]*ApiRes, 0)
	_, err := s.db.Raw(sql).QueryRows(&mResult)
	if err != nil {
		return nil, fmt.Errorf("query from database error: %+v", err)
	}
	return mResult, nil
}

func (s *Service) ListRelationsByApiPlugin(id, types string, opts ...util.OpOption) ([]*ApiRes, error) {
	if !s.db.Enabled() {
		return nil, fmt.Errorf("not support if database disabled")
	}
	return s.ListPluginRelationFromDatabase(id, types, opts...)
}

func (s *Service) UpdatePluginEnableByKongId(pluginId string, data EnableReq, opts ...util.OpOption) error {
	existed, err := s.GetApiPlugin(pluginId)
	if err != nil {
		return fmt.Errorf("cannot find apiplugin with id %s: %+v", pluginId, err)
	}
	crdNamespace := util.OpList(opts...).Namespace()
	if existed.Namespace != crdNamespace {
		return fmt.Errorf("permission denied: wrong tenant")
	}

	updateResult := make([]model.ApiPluginRelation, 0)
	//先校验是否所有API满足绑定条件，有一个不满足直接返回错误
	for _, kong := range data.Ids {
		for _, relation := range existed.ApiRelation {
			if relation.KongPluginId == kong {
				klog.Infof("kong %s has bind apiplugin %s ", kong, pluginId)
				//TODO 调用Kong接口生效失效 成功后更新状态，否则不需要更新
				err := s.UpdateEnableByKong(relation.KongPluginId, existed.Type, data.Enable)
				if err != nil {
					return fmt.Errorf("update enable transformer failed %+v", err)
				}
				updateResult = append(updateResult, model.ApiPluginRelation{
					Id:           relation.Id,
					ApiPluginId:  pluginId,
					TargetId:     relation.TargetId,
					TargetType:   relation.TargetType,
					Status:       relation.Status,
					Detail:       relation.Detail,
					Enable:       data.Enable,
					KongPluginId: relation.KongPluginId,
				})
				break
			}
		}
	}
	if err = s.db.UpdateApiPluginRelation(updateResult); err != nil {
		return fmt.Errorf("cannot write database api relation: %+v", err)
	}
	return nil
}

func (s *Service) UpdateEnableByKong(pluginId string, transType string, enable bool) (err error) {
	//id := db.Spec.KongApi.KongID
	klog.Infof("update enable plugin id is %s", pluginId)
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Patch(fmt.Sprintf("%s://%s:%d%s/%s", schema, "10.160.32.24", 30081, "/plugins", pluginId))
	//request = request.Patch(fmt.Sprintf("%s://%s:%d%s/%s", schema, "kong-kong-admin", 8001, "/plugins", pluginId))
	for k, v := range headers {
		request = request.Set(k, v)
	}
	request = request.Retry(3, 5*time.Second, retryStatus...)
	requestBody := &EnableRequestBody{}
	requestBody.Enabled = enable
	var responseBody interface{}
	switch transType {
	case "request-transformer":
		responseBody = &ReqTransformerResponseBody{}
	case "response-transformer":
		responseBody = &ResTransformerResponseBody{}
	}
	response, body, errs := request.Send(requestBody).EndStruct(responseBody)
	if len(errs) > 0 {
		return fmt.Errorf("request for enable transformer error: %+v", errs)
	}
	klog.V(5).Infof("enable transformer code: %d, body: %s ", response.StatusCode, string(body))
	if response.StatusCode != 200 {
		klog.V(5).Infof("enable transformer failed")
		return fmt.Errorf("request for enable transformer error: receive wrong status code: %s", string(body))
	}
	//(*db).Spec.ResponseTransformer.Id = responseBody.ID
	if err != nil {
		return fmt.Errorf("enable transformer error")
	}
	return nil
}