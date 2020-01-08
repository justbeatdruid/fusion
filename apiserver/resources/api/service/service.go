package service

import (
	"fmt"

	susvc "github.com/chinamobile/nlpt/apiserver/resources/application/service"
	appsvc "github.com/chinamobile/nlpt/apiserver/resources/serviceunit/service"
	"github.com/chinamobile/nlpt/crds/api/api/v1"
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

var crdNamespace = "default"

var oofsGVR = schema.GroupVersionResource{
	Group:    v1.GroupVersion.Group,
	Version:  v1.GroupVersion.Version,
	Resource: "apis",
}

type Service struct {
	client            dynamic.NamespaceableResourceInterface
	serviceunitClient dynamic.NamespaceableResourceInterface
	applicationClient dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{
		client:            client.Resource(oofsGVR),
		serviceunitClient: client.Resource(susvc.GetOOFSGVR()),
		applicationClient: client.Resource(appsvc.GetOOFSGVR()),
	}
}

func (s *Service) CreateApi(model *Api) (*Api, error) {
	if err := s.Validate(model); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
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

func (s *Service) ListApi() ([]*Api, error) {
	apis, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(apis), nil
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

func (s *Service) List() (*v1.ApiList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
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
	klog.V(5).Infof("get v1.serviceunit: %+v", app)
	return app, nil
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

func (s *Service) BindApi(apiid, appid string) (*Api, error) {
	api, err := s.Get(apiid)
	if err != nil {
		return nil, fmt.Errorf("get api error: %+v", err)
	}
	api.Spec.Applications = append(api.Spec.Applications, v1.Application{
		ID: appid,
	})
	api, err = s.UpdateSpec(api)
	//if _, err = s.updateApplicationApi(appid, api.ObjectMeta.Name, api.Spec.Name); err != nil {
	//	return fmt.Errorf("cannot update")
	//}
	return ToModel(api), err
}
