package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/crds/api/api/v1"
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	dsv1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	dw "github.com/chinamobile/nlpt/pkg/datawarehouse"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

var crdNamespace = "default"

type Service struct {
	client            dynamic.NamespaceableResourceInterface
	serviceunitClient dynamic.NamespaceableResourceInterface
	applicationClient dynamic.NamespaceableResourceInterface
	datasourceClient  dynamic.NamespaceableResourceInterface

	dataService dw.Connector
}

func NewService(client dynamic.Interface, dsConfig *config.DataserviceConfig) *Service {
	return &Service{
		client:            client.Resource(v1.GetOOFSGVR()),
		serviceunitClient: client.Resource(suv1.GetOOFSGVR()),
		applicationClient: client.Resource(appv1.GetOOFSGVR()),
		datasourceClient:  client.Resource(dsv1.GetOOFSGVR()),

		dataService: dw.NewConnector(dsConfig.Host, dsConfig.Port),
	}
}

func (s *Service) CreateApi(model *Api) (*Api, error) {
	if err := s.Validate(model); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	// check serviceunit
	//get serviceunit kongID
	var su *suv1.Serviceunit
	su, err := s.getServiceunit(model.Serviceunit.ID)
	if err != nil {
		return nil, fmt.Errorf("get serviceunit error: %+v", err)
	}
	model.Serviceunit.KongID = su.Spec.KongService.ID
	model.Serviceunit.Port = su.Spec.KongService.Port
	model.Serviceunit.Host = su.Spec.KongService.Host
	model.Serviceunit.Type = string(su.Spec.Type)

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

func (s *Service) PatchApi(id string, data interface{}) (*Api, error) {
	api, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	if err = s.assignment(api, data); err != nil {
		return nil, err
	}
	//更新API时将API的状态修改为 Updating
	api.Status.Status = v1.Updating
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(crdNamespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd: %+v", err)
	}
	return ToModel(api), err
}

func (s *Service) ListApi(suid, appid string, opts ...util.OpOption) ([]*Api, error) {
	apis, err := s.List(suid, appid)
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

func (s *Service) GetApi(id string) (*Api, error) {
	api, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(api), nil
}

func (s *Service) DeleteApi(id string) (*Api, error) {
	api, err := s.Delete(id)
	return ToModel(api), err
}

func (s *Service) PublishApi(id string) (*Api, error) {
	api, err := s.Get(id)
	//发布API时将API的状态修改为
	api.Status.Status = v1.Creating
	//TODO version随机生成
	api.Spec.PublishInfo.Version = "11111"
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(crdNamespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd: %+v", err)
	}
	return ToModel(api), err
}

func (s *Service) OfflineApi(id string) (*Api, error) {
	api, err := s.Get(id)
	//下线API时将API的状态修改为
	api.Status.Status = v1.Offing
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(crdNamespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error offline crd: %+v", err)
	}
	return ToModel(api), err
}

func (s *Service) Create(api *v1.Api) (*v1.Api, error) {
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

func (s *Service) List(suid, appid string) (*v1.ApiList, error) {
	conditions := []string{}
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

func (s *Service) Get(id string) (*v1.Api, error) {
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

func (s *Service) ForceDelete(id string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) Delete(id string) (*v1.Api, error) {
	api, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get crd by id error: %+v", err)
	}
	//TODO need check status !!!
	api.Status.Status = v1.Delete
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
	klog.V(5).Infof("get v1.serviceunit: %+v", api)

	return api, nil
}

func (s *Service) getServiceunit(id string) (*suv1.Serviceunit, error) {
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

func (s *Service) getApplication(id string) (*appv1.Application, error) {
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

func (s *Service) updateServiceApi(svcid, apiid, apiname string) (*suv1.Serviceunit, error) {
	su, err := s.getServiceunit(svcid)
	if err != nil {
		return nil, fmt.Errorf("cannot get service unit: %+v", err)
	}
	su.Spec.APIs = append(su.Spec.APIs, suv1.Api{
		ID:   apiid,
		Name: apiname,
	})

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(su)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.serviceunitClient.Namespace(su.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	su = &suv1.Serviceunit{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	return su, nil
}

func (s *Service) updateApplicationApi(appid, apiid, apiname string) (*appv1.Application, error) {
	app, err := s.getApplication(appid)
	if err != nil {
		return nil, fmt.Errorf("cannot get application: %+v", err)
	}
	app.Spec.APIs = append(app.Spec.APIs, appv1.Api{
		ID:   apiid,
		Name: apiname,
	})

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(app)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.applicationClient.Namespace(app.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	app = &appv1.Application{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	return app, nil
}

func (s *Service) BindOrRelease(apiid, appid, operation string) (*Api, error) {
	switch operation {
	case "bind":
		return s.BindApi(apiid, appid)
	case "release":
		return s.ReleaseApi(apiid, appid)
	default:
		return nil, fmt.Errorf("unknown operation %s, expect bind or release", operation)
	}
}

func (s *Service) BindApi(apiid, appid string) (*Api, error) {
	api, err := s.Get(apiid)
	if err != nil {
		return nil, fmt.Errorf("get api error: %+v", err)
	}
	if _, err = s.getApplication(appid); err != nil {
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
	api, err = s.UpdateSpec(api)
	//if _, err = s.updateApplicationApi(appid, api.ObjectMeta.Name, api.Spec.Name); err != nil {
	//	return fmt.Errorf("cannot update")
	//}
	return ToModel(api), err
}

func (s *Service) ReleaseApi(apiid, appid string) (*Api, error) {
	api, err := s.Get(apiid)
	if err != nil {
		return nil, fmt.Errorf("get api error: %+v", err)
	}
	if _, err = s.getApplication(appid); err != nil {
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

	//解绑的时候先设置false TODO 之后controller 里面删除
	api.ObjectMeta.Labels[v1.ApplicationLabel(appid)] = "false"
	api, err = s.UpdateSpec(api)
	//if _, err = s.updateApplicationApi(appid, api.ObjectMeta.Name, api.Spec.Name); err != nil {
	//	return fmt.Errorf("cannot update")
	//}
	return ToModel(api), err
}

func (s *Service) Query(apiid string, params map[string][]string) (Data, error) {
	d := Data{
		Headers: make([]string, 0),
		Columns: make(map[string]string, 0),
		Data:    make([]map[string]string, 0),
	}
	api, err := s.Get(apiid)
	if err != nil {
		return d, fmt.Errorf("get api error: %+v", err)
	}
	// data service API
	if api.Spec.Query != nil {
		q := api.Spec.Query.ToApiQuery(params)
		q.UserID = "admin"
		su, err := s.getServiceunit(api.Spec.Serviceunit.ID)
		if err != nil {
			return d, fmt.Errorf("get serviceunit error: %+v", err)
		}
		ds, err := s.getDatasource(su.Spec.DatasourceID.ID)
		if err != nil {
			return d, fmt.Errorf("get datasource error: %+v", err)
		}
		if ds.Spec.DataWarehouse == nil {
			return d, fmt.Errorf("datasource datawarehouse is null")
		}
		q.DatabaseName = ds.Spec.DataWarehouse.Name

		data, err := s.dataService.QueryData(q)
		if err != nil {
			return d, fmt.Errorf("query data error: %+v", err)
		}
		d.Headers = data.Headers
		d.Columns = data.ColumnDic
		d.Data = data.Data
		return d, nil
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
	for i := range model.WebParams {
		body[model.WebParams[i].Name] = model.WebParams[i].Example
	}
	bytesData, _ := json.Marshal(body)

	reqest, err := http.NewRequest(string(model.Method), remoteUrl, bytes.NewReader(bytesData))
	for i := range model.ApiRequestParameters {
		reqest.Header.Add(model.ApiRequestParameters[i].Name, model.ApiRequestParameters[i].Example)

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
