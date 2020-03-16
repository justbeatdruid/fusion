package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	datasource "github.com/chinamobile/nlpt/apiserver/resources/datasource/service"
	"github.com/chinamobile/nlpt/crds/api/api/v1"
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	dsv1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	dw "github.com/chinamobile/nlpt/pkg/datawarehouse"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"

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

	dataService dw.Connector
}

func NewService(client dynamic.Interface, dsConnector dw.Connector, kubeClient *clientset.Clientset, tenantEnabled bool) *Service {
	return &Service{
		kubeClient:        kubeClient,
		client:            client.Resource(v1.GetOOFSGVR()),
		serviceunitClient: client.Resource(suv1.GetOOFSGVR()),
		applicationClient: client.Resource(appv1.GetOOFSGVR()),
		datasourceClient:  client.Resource(dsv1.GetOOFSGVR()),

		dataService: dsConnector,

		tenantEnabled: tenantEnabled,
	}
}

func (s *Service) CreateApi(model *Api) (*Api, error) {
	if err := s.Validate(model); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	var crdNamespace = model.Namespace
	// check serviceunit
	//get serviceunit kongID
	var su *suv1.Serviceunit
	su, err := s.getServiceunit(model.Serviceunit.ID, crdNamespace)
	if err != nil {
		return nil, fmt.Errorf("get serviceunit error: %+v", err)
	}
	if !s.tenantEnabled {
		if !user.WritePermitted(model.Users.Owner.ID, su.ObjectMeta.Labels) {
			return nil, fmt.Errorf("user %s has no permission for writting serviceunit %s", model.Users.Owner.ID, su.ObjectMeta.Name)
		}
	}

	// refill datawarehouse query
	if su.Spec.DatasourceID != nil {
		ds, err := s.getDatasource(su.Spec.DatasourceID.ID)
		if err != nil {
			return nil, fmt.Errorf("get datasource error: %+v", err)
		}
		if model.DataWarehouseQuery != nil && ds.Spec.Type == dsv1.DataWarehouseType {
			if err := model.DataWarehouseQuery.RefillQuery(ds.Spec.DataWarehouse); err != nil {
				return nil, fmt.Errorf("cannot refill datawarehouse query: %+v", err)
			}
		}
	}

	model.Serviceunit.KongID = su.Spec.KongService.ID
	model.Serviceunit.Port = su.Spec.KongService.Port
	model.Serviceunit.Host = su.Spec.KongService.Host
	model.Serviceunit.Protocol = su.Spec.KongService.Protocol
	model.Serviceunit.Type = string(su.Spec.Type)
	//api协议依赖服务单元
	if model.Serviceunit.Protocol == "https" {
		model.Protocol = v1.HTTPS               //data type
		model.ApiDefineInfo.Protocol = v1.HTTPS //web type
	} else {
		model.Protocol = v1.HTTP
		model.ApiDefineInfo.Protocol = v1.HTTP
	}

	//init publish count
	model.PublishInfo.PublishCount = 0
	// create api
	api, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	//update service unit
	//if _, err = s.updateServiceApi(api.Spec.Serviceunit.ID, api.ObjectMeta.Name, api.Spec.Name); err != nil {
	//	if e := s.ForceDelete(api.ObjectMeta.Name); e != nil {
	//		klog.Errorf("cannot delete error api: %+v", e)
	//	}
	//	return nil, fmt.Errorf("cannot update related service unit: %+v", err)
	//}
	return ToModel(api), nil
}
func CanExeAction(api *v1.Api, action v1.Action) error {
	if api.Status.Status == v1.Running {
		klog.V(5).Infof("api status is running: %+v", api)
		return fmt.Errorf("api status is running and retry latter")
	}
	switch action {
	//API发布后不允许更新 发布 删除 解绑 只能先下线再操作最后重新发布
	case v1.Update, v1.Publish, v1.Delete, v1.UnBind:
		if api.Status.PublishStatus == v1.Released {
			klog.Infof("api status is not ok %s", api.Status.PublishStatus)
			return fmt.Errorf("api has published and cannot exec")
		}
	//只有API发布后才允许执行API下线
	case v1.Offline:
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
		return nil, fmt.Errorf("user has no write permission for serviceunit")
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

func (s *Service) ListApi(suid, appid string, opts ...util.OpOption) ([]*Api, error) {
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
	result := ToListModel(apis, opts...)
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
		return nil, fmt.Errorf("user has no write permission for serviceunit")
	}
	api, err = s.Delete(api)
	if err != nil {
		return nil, err
	}
	return ToModel(api), nil
}

func (s *Service) PublishApi(id string, opts ...util.OpOption) (*Api, error) {
	api, err := s.Get(id)
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
		return nil, fmt.Errorf("user has no write permission for serviceunit")
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
		return nil, fmt.Errorf("user has no write permission for serviceunit")
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

func (s *Service) getDatasource(id string) (*dsv1.Datasource, error) {
	crdNamespace := defaultNamespace
	crd, err := s.datasourceClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
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
		return s.BindApi(apiid, appid)
	case "release":
		return s.ReleaseApi(apiid, appid)
	default:
		return nil, fmt.Errorf("unknown operation %s, expect bind or release", operation)
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
	if ok, err := s.writable(api, opts...); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("user has no write permission for serviceunit")
	}
	if _, err = s.getApplication(appid, api.ObjectMeta.Namespace); err != nil {
		return nil, fmt.Errorf("get application error: %+v", err)
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

func (s *Service) ReleaseApi(apiid, appid string, opts ...util.OpOption) (*Api, error) {
	api, err := s.Get(apiid, opts...)
	if err != nil {
		return nil, fmt.Errorf("get api error: %+v", err)
	}
	if ok, err := s.writable(api, opts...); err != nil {
		return nil, err
	} else if !ok {
		return nil, fmt.Errorf("user has no write permission for serviceunit")
	}
	if _, err = s.getApplication(appid, api.ObjectMeta.Namespace); err != nil {
		return nil, fmt.Errorf("get application error: %+v", err)
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
	api.Status.ApplicationCount = api.Status.ApplicationCount - 1
	api, err = s.UpdateSpec(api)
	//if _, err = s.updateApplicationApi(appid, api.ObjectMeta.Name, api.Spec.Name); err != nil {
	//	return fmt.Errorf("cannot update")
	//}

	return ToModel(api), err
}

func (s *Service) Query(apiid string, params map[string][]string, opts ...util.OpOption) (Data, error) {
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
		typesMap := make(map[string]string)
		for _, w := range api.Spec.ApiQueryInfo.WebParams {
			typesMap[w.Name] = string(w.Type)
		}
		api.Spec.DataWarehouseQuery.RefillWhereFields(typesMap, params)
		q := api.Spec.DataWarehouseQuery.Query

		data, err := s.dataService.QueryData(*q)
		if err != nil {
			return d, fmt.Errorf("query data error: %+v", err)
		}
		d.Headers = data.Headers
		d.Columns = data.ColumnDic
		d.Data = data.Data
		return d, nil
	} else if api.Spec.RDBQuery != nil {
		su, err := s.getServiceunit(api.Spec.Serviceunit.ID, api.ObjectMeta.Namespace)
		if err != nil {
			return d, fmt.Errorf("get serviceunit error: %+v", err)
		}
		ds, err := s.getDatasource(su.Spec.DatasourceID.ID)
		if err != nil {
			return d, fmt.Errorf("get datasource error: %+v", err)
		}
		if ds.Spec.Type == "mysql" { //mysql查询
			//获取数据源连接信息
			queryFields := api.Spec.RDBQuery.QueryFields //查询字段
			//whereFields := api.Spec.RDBQuery.WhereFields //查询条件
			sql := strings.Builder{}
			sql.WriteString("select ")
			for _, v := range queryFields {
				sql.WriteString(v.Field + ", ")
			}
			newSql := sql.String()
			newSql = newSql[0 : len(newSql)-2] //窃取多余的","
			fmt.Println("newSql:" + newSql)
			sqlEnd := strings.Builder{}
			sqlEnd.WriteString(newSql)
			sqlEnd.WriteString(" from " + api.Spec.RDBQuery.Table)
			/*
				if len(whereFields) > 0 {
					sqlEnd.WriteString(" where ")
					for _, v := range whereFields {
						if v.ParameterEnabled {
							sqlEnd.WriteString(v.Field + "=" + "'" + v.Values[0] + "'" + " and ")
						} else {
							sqlEnd.WriteString(v.Field + "=" + "'" + req.QueryParameter(v.Field) + "'" + " and ")
						}
					}
				}
			*/
			sqlFanal := sqlEnd.String()
			sqlFanal = sqlFanal[0 : len(sqlFanal)-4]
			MysqlData, err := datasource.ConnectMysql(ds, sqlFanal)
			if err != nil {
				return d, fmt.Errorf("get mysqldata error: %+v", err)
			}
			d.Data = MysqlData
			return d, nil
		}

	} else {
		klog.V(4).Infof("api %s is not dataservice api", apiid)
	}
	return d, nil
}

//TestApi ...
func (s *Service) TestApi(model *Api) (interface{}, error) {
	client := &http.Client{}
	remoteUrl := strings.ToLower(string(model.Protocol)) + "://" + model.Serviceunit.Host + ":" + strconv.Itoa(model.Serviceunit.Port)
	for i := range model.KongApi.Paths {
		remoteUrl += model.KongApi.Paths[i]
	}

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
