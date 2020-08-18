package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/database/model"
	"github.com/chinamobile/nlpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/chinamobile/nlpt/apiserver/database"
	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
	// "k8s.io/klog"
)

var crdNamespace = "default"

type Service struct {
	tenantEnabled bool
	db            *database.DatabaseConnection
	apiClient     dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface, tenantEnabled bool, db *database.DatabaseConnection) *Service {
	return &Service{
		tenantEnabled: tenantEnabled,
		db:            db,
		apiClient:     client.Resource(apiv1.GetOOFSGVR()),
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

func (s *Service) BatchBindOrRelease(groupId, operation string, apis []ApiBind, opts ...util.OpOption) error {
	switch operation {
	case "bind":
		return s.BatchBindApi(groupId, apis, opts...)
	case "unbind":
		return s.BatchUnbindApi(groupId, apis, opts...)
	default:
		return fmt.Errorf("unknown operation %s, expect bind or unbind", operation)
	}
}

func (s *Service) BatchBindApi(groupId string, apis []ApiBind, opts ...util.OpOption) error {
	if len(apis) == 0 {
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
	for _, api := range apis {
		apiSource, err := s.getAPi(api.ID, crdNamespace)
		if err != nil {
			return fmt.Errorf("cannot get api: %+v", err)
		}
		if apiSource.Status.PublishStatus != apiv1.Released {
			return fmt.Errorf("api not released: %s", apiSource.Spec.Name)
		}
		isBind := false
		for _, relation := range existed.ApiRelation {
			if relation.ApiId == api.ID {
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
	for index, value := range apis {
		api, err := s.getAPi(value.ID, crdNamespace)
		if err != nil {
			return fmt.Errorf("cannot get api: %+v", err)
		}
		//检测是否已经绑定，已经绑定的api跳过
		if !status[index] {
			klog.Infof("apiplugins no bound to api and need bind %+v", api)
			result = append(result, model.ApiPluginRelation{
				ApiPluginId: groupId,
				ApiId:       value.ID,
			})
		}
	}
	if err = s.db.AddApiPluginRelation(result); err != nil {
		return fmt.Errorf("cannot write database api relation: %+v", err)
	}

	klog.Infof("bind apis success :%+v", result)
	return nil
}

func (s *Service) BatchUnbindApi(groupId string, apis []ApiBind, opts ...util.OpOption) error {
	if len(apis) == 0 {
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
	for _, api := range apis {
		_, err := s.getAPi(api.ID, crdNamespace)
		if err != nil {
			return fmt.Errorf("cannot get api: %+v", err)
		}
		isBind := false
		for _, relation := range existed.ApiRelation {
			if relation.ApiId == api.ID {
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
	for index, value := range apis {
		api, err := s.getAPi(value.ID, crdNamespace)
		if err != nil {
			return fmt.Errorf("cannot get api: %+v", err)
		}
		//已经检测是否绑定 只有都绑定才需要解绑
		if status[index] {
			klog.Infof("apiplugins has bound to api and need unbind %+v", api)
			result = append(result, model.ApiPluginRelation{
				Id:          relationIds[index],
				ApiPluginId: groupId,
				ApiId:       value.ID,
			})
		}

	}
	if err = s.db.RemoveApiPluginRelation(result); err != nil {
		return fmt.Errorf("cannot write database api relation: %+v", err)
	}

	klog.Infof("unbind apis success :%+v", result)
	return nil
}
