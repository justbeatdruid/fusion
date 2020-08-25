package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/database"
	"github.com/chinamobile/nlpt/apiserver/resources/datasource/rdb/driver"
	appconfig "github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/crds/api/api/v1"
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	dsv1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	dw "github.com/chinamobile/nlpt/pkg/datawarehouse"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var defaultNamespace = "default"

type Service struct {
	kubeClient        *clientset.Clientset
	client            dynamic.NamespaceableResourceInterface
	serviceunitClient dynamic.NamespaceableResourceInterface
	applicationClient dynamic.NamespaceableResourceInterface
	datasourceClient  dynamic.NamespaceableResourceInterface

	tenantEnabled bool

	db *database.DatabaseConnection

	dataService dw.Connector
	localConfig appconfig.ErrorConfig
}

func NewService(client dynamic.Interface, dsConnector dw.Connector, kubeClient *clientset.Clientset, tenantEnabled bool, localConfig appconfig.ErrorConfig, db *database.DatabaseConnection) *Service {
	return &Service{
		kubeClient:        kubeClient,
		client:            client.Resource(v1.GetOOFSGVR()),
		serviceunitClient: client.Resource(suv1.GetOOFSGVR()),
		applicationClient: client.Resource(appv1.GetOOFSGVR()),
		datasourceClient:  client.Resource(dsv1.GetOOFSGVR()),

		dataService:   dsConnector,
		db:            db,
		tenantEnabled: tenantEnabled,
		localConfig:   localConfig,
	}
}

func (s *Service) GetClient() dynamic.NamespaceableResourceInterface {
	return s.client
}

func (s *Service) CreateApi(model *Api) (*Api, error, string) {
	if err := s.Validate(model); err != nil {
		return nil, err, "001000016"
	}
	var crdNamespace = model.Namespace
	// check serviceunit
	//get serviceunit kongID
	var su *suv1.Serviceunit
	su, err := s.getServiceunit(model.Serviceunit.ID, crdNamespace)
	if err != nil {
		return nil, fmt.Errorf("get serviceunit error: %+v", err), "001000017"
	}
	if !s.tenantEnabled {
		if !user.WritePermitted(model.Users.Owner.ID, su.ObjectMeta.Labels) {
			return nil, fmt.Errorf("user %s has no permission for writting serviceunit %s",
				model.Users.Owner.ID, su.ObjectMeta.Name), "001000018"
		}
	}

	// refill datawarehouse query
	if su.Spec.Type == "data" && su.Spec.DatasourceID != nil {
		ds, err := s.getDatasource(crdNamespace, su.Spec.DatasourceID.ID)
		if err != nil {
			return nil, fmt.Errorf("get datasource error: %+v", err), "001000019"
		}
		if model.DataWarehouseQuery != nil && ds.Spec.Type == dsv1.DataWarehouseType {
			if err := model.DataWarehouseQuery.RefillQuery(ds.Spec.DataWarehouse); err != nil {
				return nil, fmt.Errorf("cannot refill datawarehouse query: %+v", err), "001000020"
			}
		}
	}

	model.Serviceunit.KongID = su.Spec.KongService.ID
	model.Serviceunit.Port = su.Spec.KongService.Port
	model.Serviceunit.Host = su.Spec.KongService.Host
	model.Serviceunit.Protocol = su.Spec.KongService.Protocol
	model.Serviceunit.Type = string(su.Spec.Type)
	model.Serviceunit.FissionFnName = su.Spec.FissionRefInfo.FnName

	//init publish count
	model.PublishInfo.PublishCount = 0
	// create api
	api, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err), "001000021"
	}
	//update service unit
	//if _, err = s.updateServiceApi(api.Spec.Serviceunit.ID, api.ObjectMeta.Name, api.Spec.Name); err != nil {
	//	if e := s.ForceDelete(api.ObjectMeta.Name); e != nil {
	//		klog.Errorf("cannot delete error api: %+v", e)
	//	}
	//	return nil, fmt.Errorf("cannot update related service unit: %+v", err)
	//}
	return ToModel(api), nil, "0"
}
func CanExeAction(api *v1.Api, action v1.Action) error {
	if api.Status.Status == v1.Running {
		klog.V(5).Infof("api status is running: %+v", api)
		return fmt.Errorf("api status is running and retry latter")
	}
	switch action {
	//API发布后不允删除，只能先下线再删除，解绑是单独操作 发布后可执行
	//API发布后允许更新，再发布，更新后要提示发布才能生效
	case v1.Delete:
		if api.Status.PublishStatus == v1.Released {
			klog.Infof("api status is not ok %s", api.Status.PublishStatus)
			return fmt.Errorf("api has published and cannot exec")
		}
	//只有API发布后才允许执行API下线 发布后才可以绑定
	//添加、更新、删除插件时api必选是发布状态
	case v1.Offline, v1.Bind, v1.AddPlugins, v1.UpdatePlugins, v1.DeletePlugins:
		if api.Status.PublishStatus != v1.Released {
			klog.Infof("api status is not ok %s", api.Status.PublishStatus)
			return fmt.Errorf("api has not published and cannot exec")
		}
	default:
		klog.V(5).Infof("default api status is ok: %+v", api)
	}

	klog.Infof("api status is ok %+v", api)
	return nil
}

func (s *Service) PatchApi(id string, data interface{}, opts ...util.OpOption) (*Api, error) {
	api, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	err = CanExeAction(api, v1.Update)
	if err != nil {
		return nil, fmt.Errorf("cannot exec: %+v", err)
	}
	if ok, err := s.writable(api, opts...); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("user has no write permission for api")
	}
	if err = s.assignment(api, data); err != nil {
		return nil, err
	}
	//更新API
	api.Status.Status = v1.Init
	api.Status.Action = v1.Update

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(api.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd: %+v", err)
	}
	return ToModel(api), err
}

func (s *Service) AddApiPlugins(id string, name string, config interface{}, opts ...util.OpOption) (*Api, error) {
	api, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	err = CanExeAction(api, v1.AddPlugins)
	if err != nil {
		return nil, fmt.Errorf("cannot exec: %+v", err)
	}
	if ok, err := s.writable(api, opts...); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("user has no write permission for api")
	}
	if err = s.assignmentConfig(api, name, config); err != nil {
		return nil, err
	}
	//更新API
	api.Status.Status = v1.Init
	api.Status.Action = v1.AddPlugins

	klog.V(5).Infof("api plugins details %+v", api)

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(api.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd: %+v", err)
	}
	return ToModel(api), err
}

func (s *Service) DeleteApiPlugins(api_id string, plugin_id string, opts ...util.OpOption) (*Api, error) {
	api, err := s.Get(api_id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	err = CanExeAction(api, v1.DeletePlugins)
	if err != nil {
		return nil, fmt.Errorf("cannot exec: %+v", err)
	}
	if ok, err := s.writable(api, opts...); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("user has no write permission for api")
	}
	/*
		if err = s.assignmentConfig(api, name, config); err != nil {
			return nil, err
		}
	*/
	//更新API
	api.Status.Status = v1.Init
	api.Status.Action = v1.DeletePlugins

	klog.V(5).Infof("api plugins details %+v", api)

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(api.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd: %+v", err)
	}
	return ToModel(api), err
}

func (s *Service) PatchApiPlugins(api_id string, plugin_id string, config interface{}, opts ...util.OpOption) (*Api, error) {
	api, err := s.Get(api_id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	err = CanExeAction(api, v1.UpdatePlugins)
	if err != nil {
		return nil, fmt.Errorf("cannot exec: %+v", err)
	}
	if ok, err := s.writable(api, opts...); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("user has no write permission for api")
	}
	if err = s.assignmentConfig(api, api.Spec.ResponseTransformer.Name, config); err != nil {
		return nil, err
	}
	//更新API
	api.Status.Status = v1.Init
	api.Status.Action = v1.UpdatePlugins

	klog.V(5).Infof("api plugins details %+v", api)

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(api.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd: %+v", err)
	}
	return ToModel(api), err
}

func (s *Service) ListApi(suid, appid, status string, opts ...util.OpOption) ([]*Api, error) {
	if len(suid) > 0 && len(appid) > 0 {
		return nil, fmt.Errorf("i am not sure if it is suitable to get api with application id and service unit id, but it make no sense")
	}
	// check opts.user is permitted for application and serviceunit
	if !s.tenantEnabled {
		op := util.OpList(opts...)
		userid := op.User()
		ns := defaultNamespace
		if len(appid) > 0 {
			app, err := s.getApplication(appid, ns)
			if err != nil {
				return nil, fmt.Errorf("cannot get application: %+v", err)
			}
			if !user.ReadPermitted(userid, app.ObjectMeta.Labels) {
				return nil, fmt.Errorf("permitted denied of application %s", appid)
			}
		}
		if len(suid) > 0 {
			su, err := s.getServiceunit(suid, ns)
			if err != nil {
				return nil, fmt.Errorf("cannot get serviceunit: %+v", err)
			}
			if !user.ReadPermitted(userid, su.ObjectMeta.Labels) {
				return nil, fmt.Errorf("permitted denied of service unit %s", suid)
			}
		}
	}
	apis, err := s.List(suid, appid, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}

	//publishedOnly false 表示查询所有，true表示查询已发布的api
	publishedOnly := false
	//大数据默认查询发布的api，汇聚平台查询所有api
	//需要确认一下，大数据场景若设置status查询则需要设置publishedOnly为false
	if len(suid) == 0 && len(appid) == 0 && len(status) == 0 && !s.tenantEnabled {
		publishedOnly = true
	}
	result := ToListModel(apis, publishedOnly, status, opts...)
	if len(appid) > 0 {
		for i := range result {
			if apis.Items[i].Status.Applications != nil {
				appstatus, ok := apis.Items[i].Status.Applications[appid]
				if ok {
					result[i].ApplicationBindStatus = &appstatus
				}
			}
		}
	}
	return result, nil
}

func (s *Service) GetApi(id string, opts ...util.OpOption) (*Api, error) {
	api, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(api), nil
}

func (s *Service) DeleteApi(id string, opts ...util.OpOption) (*Api, error) {
	api, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("get crd by id error: %+v", err)
	}
	err = CanExeAction(api, v1.Delete)
	if err != nil {
		return nil, fmt.Errorf("cannot exec: %+v", err)
	}
	if ok, err := s.writable(api, opts...); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("user has no write permission for api")
	}
	api, err = s.Delete(api)
	if err != nil {
		return nil, err
	}
	util.WaitDelete(s, api.ObjectMeta)
	return ToModel(api), nil
}

func (s *Service) BatchDeleteApi(apis []v1.ApiBind, opts ...util.OpOption) error {
	//先校验是否所有API满足删除条件，有一个不满足直接返回错误
	for _, value := range apis {
		api, err := s.Get(value.ID, opts...)
		if err != nil {
			return fmt.Errorf("cannot get api: %+v", err)
		}
		err = CanExeAction(api, v1.Delete)
		if err != nil {
			return fmt.Errorf("cannot exec: %+v", err)
		}
		if ok, err := s.writable(api, opts...); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("user has no write permission for api")
		}
	}
	for _, value := range apis {
		api, err := s.Get(value.ID, opts...)
		if err != nil {
			return fmt.Errorf("cannot get api: %+v", err)
		}

		api, err = s.Delete(api)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) PublishApi(id string, opts ...util.OpOption) (*Api, error) {
	api, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	err = CanExeAction(api, v1.Publish)
	if err != nil {
		return nil, fmt.Errorf("cannot exec: %+v", err)
	}
	if ok, err := s.writable(api, opts...); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("user has no write permission for api")
	}
	//TODO version随机生成
	api.Spec.PublishInfo.Version = names.NewID()
	//发布API
	api.Status.Status = v1.Init
	api.Status.Action = v1.Publish
	api.Spec.PublishInfo.PublishCount = api.Spec.PublishInfo.PublishCount + 1
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(api.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd: %+v", err)
	}

	return ToModel(api), err
}

func (s *Service) OfflineApi(id string, opts ...util.OpOption) (*Api, error) {
	api, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	err = CanExeAction(api, v1.Offline)
	if err != nil {
		return nil, fmt.Errorf("cannot exec: %+v", err)
	}
	if ok, err := s.writable(api, opts...); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("user has no write permission for api")
	}
	//下线API
	api.Status.Status = v1.Init
	api.Status.Action = v1.Offline
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(api.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error offline crd: %+v", err)
	}

	return ToModel(api), err
}

func (s *Service) Create(api *v1.Api) (*v1.Api, error) {
	var crdNamespace = defaultNamespace
	if s.tenantEnabled {
		crdNamespace = api.ObjectMeta.Namespace
		if len(crdNamespace) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
	} else {
		api.ObjectMeta.Namespace = defaultNamespace
	}
	// no need to ensure namespace in k8s. for serviceunit already in this namespace

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), api); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.api of creating: %+v", api)
	return api, nil
}

func (s *Service) List(suid, appid string, opts ...util.OpOption) (*v1.ApiList, error) {
	op := util.OpList(opts...)
	group := op.Group()
	ns := op.Namespace()
	var crdNamespace = defaultNamespace
	if s.tenantEnabled {
		if len(ns) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
		crdNamespace = ns
	}

	conditions := []string{}
	if len(group) > 0 {
		conditions = append(conditions, fmt.Sprintf("%s=%s", appv1.GroupLabel, group))
	}
	if len(suid) > 0 {
		conditions = append(conditions, fmt.Sprintf("%s=%s", v1.ServiceunitLabel, suid))
	}
	if len(appid) > 0 {
		conditions = append(conditions, fmt.Sprintf("%s=%s", v1.ApplicationLabel(appid), "true"))
	}
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{
		LabelSelector: strings.Join(conditions, ","),
	})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apis := &v1.ApiList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apis); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.apiList: %+v", apis)
	return apis, nil
}

func (s *Service) ListApis(crdNamespace string) (*v1.ApiList, error) {

	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apis := &v1.ApiList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apis); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("====test get v1.ApiList: %+v", apis)
	return apis, nil
}

func (s *Service) Get(id string, opts ...util.OpOption) (*v1.Api, error) {
	var crdNamespace = defaultNamespace
	if s.tenantEnabled {
		crdNamespace = util.OpList(opts...).Namespace()
		if len(crdNamespace) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
	}
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	api := &v1.Api{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), api); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.api: %+v", api)
	return api, nil
}

func (s *Service) ForceDelete(id, crdNamespace string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) Delete(api *v1.Api) (*v1.Api, error) {
	//TODO need check status !!!
	//删除API
	api.Status.Status = v1.Init
	api.Status.Action = v1.Delete
	return s.UpdateStatus(api)
}

func (s *Service) UpdateSpec(api *v1.Api) (*v1.Api, error) {
	return s.UpdateStatus(api)
}

func (s *Service) UpdateStatus(api *v1.Api) (*v1.Api, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	//TODO method client.Namespace().UpdateStatus() always returns error
	//     however method Update() can also update status
	crd, err = s.client.Namespace(api.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	api = &v1.Api{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), api); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.api: %+v", api)

	return api, nil
}

func (s *Service) getServiceunit(id, crdNamespace string) (*suv1.Serviceunit, error) {
	crd, err := s.serviceunitClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	su := &suv1.Serviceunit{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunit: %+v", su)
	return su, nil
}

func (s *Service) getApplication(id, crdNamespace string) (*appv1.Application, error) {
	crd, err := s.applicationClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	app := &appv1.Application{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.application: %+v", app)
	return app, nil
}

func (s *Service) getDatasource(namespace string, id string) (*dsv1.Datasource, error) {
	crd, err := s.datasourceClient.Namespace(namespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	ds := &dsv1.Datasource{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ds); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasource: %+v", ds)
	return ds, nil
}

func (s *Service) BindOrRelease(apiid, appid, operation string, opts ...util.OpOption) (*Api, error) {
	switch operation {
	case "bind":
		return s.BindApi(apiid, appid, opts...)
	case "release":
		return s.ReleaseApi(apiid, appid, opts...)
	default:
		return nil, fmt.Errorf("unknown operation %s, expect bind or release", operation)
	}
}

func (s *Service) BatchBindOrRelease(appid, operation string, apis []v1.ApiBind, opts ...util.OpOption) error {
	switch operation {
	case "bind":
		return s.BatchBindApi(appid, apis, opts...)
	case "release":
		return s.BatchReleaseApi(appid, apis, opts...)
	default:
		return fmt.Errorf("unknown operation %s, expect bind or release", operation)
	}
}

func (s *Service) BindApi(apiid, appid string, opts ...util.OpOption) (*Api, error) {
	if true {
		return nil, fmt.Errorf("do not bind api directly. make an application.")
	}
	api, err := s.Get(apiid, opts...)
	if err != nil {
		return nil, fmt.Errorf("get api error: %+v", err)
	}
	app, err := s.getApplication(appid, api.ObjectMeta.Namespace)
	if err != nil {
		return nil, fmt.Errorf("get application error: %+v", err)
	}
	if ok, err := s.bindPermitted(app, opts...); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("user has no write permission for application")
	}
	for _, existedapp := range api.Spec.Applications {
		if existedapp.ID == appid {
			return nil, fmt.Errorf("application alrady bound to api")
		}
	}
	api.ObjectMeta.Labels[v1.ApplicationLabel(appid)] = "true"
	api.Spec.Applications = append(api.Spec.Applications, v1.Application{
		ID: appid,
	})
	api.Status.ApplicationCount = api.Status.ApplicationCount + 1
	api, err = s.UpdateSpec(api)
	//if _, err = s.updateApplicationApi(appid, api.ObjectMeta.Name, api.Spec.Name); err != nil {
	//	return fmt.Errorf("cannot update")
	//}

	return ToModel(api), err
}

func (s *Service) BatchBindApi(appid string, apis []v1.ApiBind, opts ...util.OpOption) error {
	if len(apis) == 0 {
		return fmt.Errorf("at least one api must select to bind")
	}
	//先校验是否所有API满足绑定条件，有一个不满足直接返回错误
	for _, value := range apis {
		api, err := s.Get(value.ID, opts...)
		if err != nil {
			return fmt.Errorf("cannot get api: %+v", err)
		}
		//for _, existedapp := range api.Spec.Applications {
		//	if existedapp.ID == appid {
		//		return fmt.Errorf("application alrady bound to api")
		//	}
		//}
		app, err := s.getApplication(appid, api.ObjectMeta.Namespace)
		if err != nil {
			return fmt.Errorf("get application error: %+v", err)
		}
		if ok, err := s.bindPermitted(app, opts...); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("user has no write permission for application")
		}
		err = CanExeAction(api, v1.Bind)
		if err != nil {
			return fmt.Errorf("cannot exec: %+v", err)
		}
	}
	for _, value := range apis {
		api, err := s.Get(value.ID, opts...)
		if err != nil {
			return fmt.Errorf("cannot get api: %+v", err)
		}
		//检测是否已经绑定，已经绑定的api跳过
		isBind := false
		for _, existedapp := range api.Spec.Applications {
			if existedapp.ID == appid {
				isBind = true
				klog.Infof("application alrady bound to api and no need update status %+v", api)
				break
			}
		}
		if !isBind {
			//绑定API
			api.Status.Status = v1.Init
			api.Status.Action = v1.Bind
			api.ObjectMeta.Labels[v1.ApplicationLabel(appid)] = "true"
			api.Spec.Applications = append(api.Spec.Applications, v1.Application{
				ID: appid,
			})
			klog.Infof("application no bound to api and need update status %+v", api)
		}

		api, err = s.UpdateSpec(api)
		if err != nil {
			return fmt.Errorf("cannot update api bind: %+v", err)
		}
	}
	return nil
}

func (s *Service) ReleaseApi(apiid, appid string, opts ...util.OpOption) (*Api, error) {
	api, err := s.Get(apiid, opts...)
	if err != nil {
		return nil, fmt.Errorf("get api error: %+v", err)
	}
	app, err := s.getApplication(appid, api.ObjectMeta.Namespace)
	if err != nil {
		return nil, fmt.Errorf("get application error: %+v", err)
	}
	if ok, err := s.bindPermitted(app, opts...); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("user has no write permission for application")
	}
	exist := false
	for _, existedapp := range api.Spec.Applications {
		if existedapp.ID == appid {
			exist = true
		}
	}
	if !exist {
		return nil, fmt.Errorf("application not bound to api")
	}

	err = CanExeAction(api, v1.UnBind)
	if err != nil {
		return nil, fmt.Errorf("cannot exec: %+v", err)
	}
	//解绑API
	api.Status.Status = v1.Init
	api.Status.Action = v1.UnBind
	//解绑的时候先设置false TODO 之后controller 里面删除
	api.ObjectMeta.Labels[v1.ApplicationLabel(appid)] = "false"
	api, err = s.UpdateSpec(api)
	//if _, err = s.updateApplicationApi(appid, api.ObjectMeta.Name, api.Spec.Name); err != nil {
	//	return fmt.Errorf("cannot update")
	//}

	return ToModel(api), err
}
func (s *Service) BatchReleaseApi(appid string, apis []v1.ApiBind, opts ...util.OpOption) error {
	if len(apis) == 0 {
		return fmt.Errorf("at least one api must select to unbind")
	}
	for _, value := range apis {
		api, err := s.Get(value.ID, opts...)
		if err != nil {
			return fmt.Errorf("cannot get api: %+v", err)
		}
		app, err := s.getApplication(appid, api.ObjectMeta.Namespace)
		if err != nil {
			return fmt.Errorf("get application error: %+v", err)
		}
		if ok, err := s.bindPermitted(app, opts...); err != nil {
			return err
		} else if !ok {
			return fmt.Errorf("user has no write permission for application")
		}
		exist := false
		for _, existedapp := range api.Spec.Applications {
			if existedapp.ID == appid {
				exist = true
			}
		}
		if !exist {
			return fmt.Errorf("application not bound to api")
		}
		err = CanExeAction(api, v1.UnBind)
		if err != nil {
			return fmt.Errorf("cannot exec: %+v", err)
		}
	}
	for _, value := range apis {
		api, err := s.Get(value.ID, opts...)
		if err != nil {
			return fmt.Errorf("cannot get api: %+v", err)
		}
		//解绑API
		api.Status.Status = v1.Init
		api.Status.Action = v1.UnBind
		//解绑的时候先设置false TODO 之后controller 里面删除
		api.ObjectMeta.Labels[v1.ApplicationLabel(appid)] = "false"
		api.Status.ApplicationCount = api.Status.ApplicationCount - 1
		api, err = s.UpdateSpec(api)
		if err != nil {
			return fmt.Errorf("cannot update api unbind: %+v", err)
		}
	}
	return nil
}

func (s *Service) Query(apiid string, params map[string][]string, limitstr string, opts ...util.OpOption) (Data, error) {
	d := Data{
		Headers: make([]string, 0),
		Columns: make(map[string]string, 0),
		Data:    make([]map[string]string, 0),
	}
	api, err := s.Get(apiid, opts...)
	if err != nil {
		return d, fmt.Errorf("get api error: %+v", err)
	}
	// data service API
	if api.Spec.DataWarehouseQuery != nil {
		if api.Spec.DataWarehouseQuery.Type == "hql" {
			hql, databaseName := api.Spec.DataWarehouseQuery.HQL, api.Spec.DataWarehouseQuery.Database
			data, err := s.dataService.QueryHqlData(hql, databaseName, api.Spec.Name)
			if err != nil {
				return d, fmt.Errorf("query data error: %+v", err)
			}
			d.Headers = data.Headers
			d.Columns = data.ColumnDic
			d.Data = data.Data
		} else {
			typesMap := make(map[string]string)
			for _, w := range api.Spec.ApiQueryInfo.WebParams {
				typesMap[w.Name] = string(w.Type)
			}
			api.Spec.DataWarehouseQuery.RefillWhereFields(typesMap, params)
			q := api.Spec.DataWarehouseQuery.Query
			if len(limitstr) > 0 {
				limit, err := strconv.Atoi(limitstr)
				if err != nil {
					return d, fmt.Errorf("cannot parse limit parameter %s to int: %+v", limitstr, err)
				}
				q.Limit = limit
			}

			data, err := s.dataService.QueryData(*q, api.Spec.Name)
			if err != nil {
				return d, fmt.Errorf("query data error: %+v", err)
			}
			d.Headers = data.Headers
			d.Columns = data.ColumnDic
			d.Data = data.Data
		}
		return d, nil
	} else if api.Spec.RDBQuery != nil {
		su, err := s.getServiceunit(api.Spec.Serviceunit.ID, api.ObjectMeta.Namespace)
		if err != nil {
			return d, fmt.Errorf("get serviceunit error: %+v", err)
		}
		ds, err := s.getDatasource(api.ObjectMeta.Namespace, su.Spec.DatasourceID.ID)
		if err != nil {
			return d, fmt.Errorf("get datasource error: %+v", err)
		}
		if ds.Spec.RDB == nil {
			return d, fmt.Errorf("cannot find rdb info in datasource")
		}
		//获取数据源连接信息
		queryFields := api.Spec.RDBQuery.QueryFields //查询字段
		//whereFields := api.Spec.RDBQuery.WhereFields //查询条件
		sqlBuilder := strings.Builder{}
		sqlBuilder.WriteString("select ")
		fields := make([]string, len(queryFields))
		for i := range queryFields {
			fields[i] = queryFields[i].Field
		}
		sqlBuilder.WriteString(strings.Join(fields, ","))
		switch ds.Spec.RDB.Type {
		case "mysql":
			sqlBuilder.WriteString(" from " + api.Spec.RDBQuery.Table)
		case "postgres", "potgresql":
			sqlBuilder.WriteString(" from " + ds.Spec.RDB.Schema + "." + api.Spec.RDBQuery.Table)
		default:
			return d, fmt.Errorf("unsupported rdb type %s", ds.Spec.RDB.Type)
		}
		whereFields := make([]string, 0)
		for _, w := range api.Spec.ApiQueryInfo.WebParams {
			if vs, ok := params[w.Name]; ok {
				if len(vs) > 0 {
					whereFields = append(whereFields, fmt.Sprintf("%s='%s'", w.Name, vs[0]))
				}
			}
		}
		if len(whereFields) > 0 {
			sqlBuilder.WriteString(" where ")
			sqlBuilder.WriteString(strings.Join(whereFields, " and "))
		}
		sql := sqlBuilder.String()
		klog.V(4).Infof("query rdb data with sql %s", sql)
		MysqlData, err := driver.GetRDBData(ds, sql)
		if err != nil {
			return d, fmt.Errorf("get mysqldata error: %+v", err)
		}
		d.Data = MysqlData
		return d, nil

	} else {
		klog.V(4).Infof("api %s is not dataservice api", apiid)
	}
	return d, nil
}

//TestApi ...
func (s *Service) TestApi(model *Api) (interface{}, error) {
	client := &http.Client{}
	var crdNamespace = model.Namespace
	var su *suv1.Serviceunit
	su, err := s.getServiceunit(model.Serviceunit.ID, crdNamespace)
	if err != nil {
		return nil, fmt.Errorf("get serviceunit error: %+v", err)
	}
	if !s.tenantEnabled {
		if !user.WritePermitted(model.Users.Owner.ID, su.ObjectMeta.Labels) {
			return nil, fmt.Errorf("user %s has no permission for writting serviceunit %s",
				model.Users.Owner.ID, su.ObjectMeta.Name)
		}
	}
	remoteUrl := strings.ToLower(string(su.Spec.KongService.Protocol)) + "://" + su.Spec.KongService.Host + ":" + strconv.Itoa(su.Spec.KongService.Port)
	for i := range model.KongApi.Paths {
		remoteUrl += model.KongApi.Paths[i]
	}
	klog.V(5).Infof("host is %s, remote url is %s", su.Spec.KongService.Host, remoteUrl)

	body := map[string]interface{}{}
	headers := map[string]string{}
	querys := map[string]string{}
	for _, v := range model.ApiQueryInfo.WebParams {
		switch v.Location {
		case v1.Body:
			body[v.Name] = v.Example
		case v1.Path:
			strings.Replace(remoteUrl, "{"+v.Name+"}", v.Example, -1)

		case v1.Header:
			headers[v.Name] = v.Example

		case v1.Query:
			querys[v.Name] = v.Example
		}

	}
	if len(querys) != 0 {
		count := 0
		remoteUrl += "?"
		for k, v := range querys {
			remoteUrl = remoteUrl + k + "=" + v
			if count < len(querys)-1 {
				remoteUrl += "&"
				count += 1
			}
		}

	}

	bytesData, _ := json.Marshal(body)

	reqest, err := http.NewRequest(string(model.Method), remoteUrl, bytes.NewReader(bytesData))
	for k, v := range headers {
		reqest.Header.Add(k, v)

	}
	response, err := client.Do(reqest)
	if err != nil {
		return nil, fmt.Errorf("request failed: %+v", err)
	}
	defer response.Body.Close()

	respbody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("ReadAll failed: %+v", err)
	}
	respReturn := map[string]interface{}{}
	if err = json.Unmarshal(respbody, &respReturn); err != nil {
		return nil, fmt.Errorf("Unmarshal failed: %+v", err)
	}
	return respReturn, nil
}

func (s *Service) writable(api *v1.Api, opts ...util.OpOption) (bool, error) {
	if !s.tenantEnabled {
		uid := util.OpList(opts...).User()
		var su *suv1.Serviceunit
		su, err := s.getServiceunit(api.Spec.Serviceunit.ID, api.ObjectMeta.Namespace)
		if err != nil {
			return false, fmt.Errorf("get serviceunit error: %+v", err)
		}
		if !user.WritePermitted(uid, su.ObjectMeta.Labels) {
			return false, nil
		}
	}
	return true, nil
}

func (s *Service) bindPermitted(app *appv1.Application, opts ...util.OpOption) (bool, error) {
	if !s.tenantEnabled {
		uid := util.OpList(opts...).User()
		if !user.WritePermitted(uid, app.ObjectMeta.Labels) {
			return false, nil
		}
	}
	return true, nil
}

func (s *Service) listApplications(opts ...util.OpOption) ([]appv1.Application, error) {
	op := util.OpList(opts...)
	ns := op.Namespace()
	u := op.User()
	var options metav1.ListOptions
	var labels []string
	labels = append(labels, user.GetLabelSelector(u))
	options.LabelSelector = strings.Join(labels, ",")
	klog.V(5).Infof("list with label selector: %s", options.LabelSelector)
	crd, err := s.applicationClient.Namespace(ns).List(options)
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apps := &appv1.ApplicationList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationList: %+v", apps)
	return apps.Items, nil
}

func (s *Service) ListAllApplicationApis(opts ...util.OpOption) ([]*ApplicationScopedApi, error) {
	apps, err := s.listApplications(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list application: %+v", err)
	}
	klog.V(5).Infof("get %d applications", len(apps))
	result := make([]*ApplicationScopedApi, 0)
	for _, app := range apps {
		apis, err := s.ListApi("", app.ObjectMeta.Name, "", opts...)
		if err != nil {
			return nil, fmt.Errorf("list api error: %+v", err)
		}
		for _, api := range apis {
			result = append(result, &ApplicationScopedApi{
				BoundApplicationId:   app.ObjectMeta.Name,
				BoundApplicationName: app.Spec.Name,
				Api:                  *api,
			})
		}
	}
	return result, nil
}

func (s *Service) listServiceunits(opts ...util.OpOption) ([]suv1.Serviceunit, error) {
	op := util.OpList(opts...)
	ns := op.Namespace()
	u := op.User()
	var options metav1.ListOptions
	var labels []string
	labels = append(labels, user.GetLabelSelector(u))
	options.LabelSelector = strings.Join(labels, ",")
	klog.V(5).Infof("list with label selector: %s", options.LabelSelector)
	crd, err := s.serviceunitClient.Namespace(ns).List(options)
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apps := &suv1.ServiceunitList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunitList: %+v", apps)
	return apps.Items, nil
}

func (s *Service) ListAllServiceunitApis(opts ...util.OpOption) ([]*ServiceunitScopedApi, error) {
	sus, err := s.listServiceunits(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list serviceunit: %+v", err)
	}
	klog.V(5).Infof("get %d serviceunits", len(sus))
	result := make([]*ServiceunitScopedApi, 0)
	for _, su := range sus {
		apis, err := s.ListApi(su.ObjectMeta.Name, "", "", opts...)
		if err != nil {
			return nil, fmt.Errorf("list api error: %+v", err)
		}
		for _, api := range apis {
			result = append(result, &ServiceunitScopedApi{
				BoundServiceunitId:   su.ObjectMeta.Name,
				BoundServiceunitName: su.Spec.Name,
				Api:                  *api,
			})
		}
	}
	return result, nil
}

func (s *Service) ListApisByApiGroup(id string, opts ...util.OpOption) ([]*Api, error) {
	if !s.db.Enabled() {
		return nil, fmt.Errorf("not support if database disabled")
	}
	return s.ListByApiRelationFromDatabase(id, opts...)
}
