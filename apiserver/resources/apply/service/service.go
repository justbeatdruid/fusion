package service

import (
	"fmt"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	"github.com/chinamobile/nlpt/crds/apply/api/v1"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

var crdNamespace = "default"

type Service struct {
	client            dynamic.NamespaceableResourceInterface
	apiClient         dynamic.NamespaceableResourceInterface
	serviceunitClient dynamic.NamespaceableResourceInterface
	applicationClient dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{
		client:            client.Resource(v1.GetOOFSGVR()),
		apiClient:         client.Resource(apiv1.GetOOFSGVR()),
		serviceunitClient: client.Resource(suv1.GetOOFSGVR()),
		applicationClient: client.Resource(appv1.GetOOFSGVR()),
	}
}

func (s *Service) CreateApply(model *Apply) (*Apply, error) {
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}

	check := func(r *Resource) error {
		switch r.Type {
		case v1.Serviceunit:
			su, err := s.getServiceunit(r.ID)
			if err != nil {
				return fmt.Errorf("get serviceunit error: %+v", err)
			}
			r.Name = su.Spec.Name
		case v1.Api:
			api, err := s.getApi(r.ID)
			if err != nil {
				return fmt.Errorf("get api error: %+v", err)
			}
			r.Name = api.Spec.Name
		case v1.Application:
			app, err := s.getApplication(r.ID)
			if err != nil {
				return fmt.Errorf("get application error: %+v", err)
			}
			r.Name = app.Spec.Name
		default:
			return fmt.Errorf("unknown target type: %s", r.Type)
		}
		return nil
	}
	for n, r := range map[string]*Resource{
		"target": &model.Target,
		"source": &model.Source,
	} {
		if err := check(r); err != nil {
			return nil, fmt.Errorf("%s resource error: %+v", n, err)
		}
	}
	//TODO check if target api already bound to source application

	apl, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return s.ToModel(apl)
}

func (s *Service) ListApply() ([]*Apply, error) {
	apls, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return s.ToListModel(apls)
}

func (s *Service) GetApply(id string) (*Apply, error) {
	apl, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return s.ToModel(apl)
}

func (s *Service) DeleteApply(id string) error {
	return s.Delete(id)
}

func (s *Service) Create(apl *v1.Apply) (*v1.Apply, error) {
	apl.Status.OperationDone = false
	apl.Status.Retry = v1.OperationRetry
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(apl)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apl); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.apply of creating: %+v", apl)
	return apl, nil
}

func (s *Service) List() (*v1.ApplyList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apls := &v1.ApplyList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apls); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applyList: %+v", apls)
	return apls, nil
}

func (s *Service) Get(id string) (*v1.Apply, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	apl := &v1.Apply{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apl); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.apply: %+v", apl)
	return apl, nil
}

func (s *Service) Delete(id string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
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

func (s *Service) getApi(id string) (*apiv1.Api, error) {
	crd, err := s.apiClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	api := &apiv1.Api{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), api); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.api: %+v", api)
	return api, nil
}

func (s *Service) ApproveApply(id string, admitted bool, reason string) (*v1.Apply, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	apl := &v1.Apply{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apl); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.apply: %+v", apl)
	if apl.Status.Status != v1.Waiting {
		return nil, fmt.Errorf("wrong apply status, expect %s, have %s", v1.Waiting, apl.Status.Status)
	}

	if admitted {
		apl.Status.Status = v1.Admited
	} else {
		apl.Status.Status = v1.Denied
	}
	apl.Status.Reason = reason
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(apl)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd = &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apl); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.apply of creating: %+v", apl)
	return apl, nil
}
