package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/crds/applicationgroup/api/v1"

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
	Resource: "applicationgroups",
}

type Service struct {
	client dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{client: client.Resource(oofsGVR)}
}

func (s *Service) CreateApplicationGroup(model *ApplicationGroup) (*ApplicationGroup, error) {
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	app, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(app), nil
}

func (s *Service) ListApplicationGroup() ([]*ApplicationGroup, error) {
	apps, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(apps), nil
}

func (s *Service) GetApplicationGroup(id string) (*ApplicationGroup, error) {
	app, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(app), nil
}

func (s *Service) DeleteApplicationGroup(id string) error {
	return s.Delete(id)
}

func (s *Service) UpdateApplicationGroup(id string, model *ApplicationGroup) (*ApplicationGroup, error) {
	app, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get applicationgroup error: %+v", err)
	}
	if len(model.Name) > 0 {
		app.Spec.Name = model.Name
	}
	if len(model.Description) > 0 {
		app.Spec.Description = model.Description
	}
	app, err = s.UpdateSpec(app)
	if err != nil {
		return nil, fmt.Errorf("cannot update object: %+v", err)
	}
	return ToModel(app), nil
}

func (s *Service) Create(app *v1.ApplicationGroup) (*v1.ApplicationGroup, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(app)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationgroup of creating: %+v", app)
	return app, nil
}

func (s *Service) List() (*v1.ApplicationGroupList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apps := &v1.ApplicationGroupList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationgrouplist: %+v", apps)
	return apps, nil
}

func (s *Service) Get(id string) (*v1.ApplicationGroup, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	app := &v1.ApplicationGroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationgroup: %+v", app)
	return app, nil
}

func (s *Service) Delete(id string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) UpdateSpec(app *v1.ApplicationGroup) (*v1.ApplicationGroup, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(app)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.client.Namespace(app.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	app = &v1.ApplicationGroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationgroup: %+v", app)

	return app, nil
}
