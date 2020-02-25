package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/crds/serviceunitgroup/api/v1"

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
	Resource: "serviceunitgroups",
}

type Service struct {
	client dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{client: client.Resource(oofsGVR)}
}

func (s *Service) CreateServiceunitGroup(model *ServiceunitGroup) (*ServiceunitGroup, error) {
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	sug, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(sug), nil
}

func (s *Service) ListServiceunitGroup() ([]*ServiceunitGroup, error) {
	sugs, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(sugs), nil
}

func (s *Service) GetServiceunitGroup(id string) (*ServiceunitGroup, error) {
	sug, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(sug), nil
}

func (s *Service) DeleteServiceunitGroup(id string) error {
	return s.Delete(id)
}

func (s *Service) UpdateServiceunitGroup(id string, model *ServiceunitGroup) (*ServiceunitGroup, error) {
	sug, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get serviceunitgroup error: %+v", err)
	}
	if len(model.Name) > 0 {
		sug.Spec.Name = model.Name
	}
	if len(model.Description) > 0 {
		sug.Spec.Description = model.Description
	}
	sug, err = s.UpdateSpec(sug)
	if err != nil {
		return nil, fmt.Errorf("cannot update object: %+v", err)
	}
	return ToModel(sug), nil
}

func (s *Service) Create(sug *v1.ServiceunitGroup) (*v1.ServiceunitGroup, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(sug)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), sug); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunitgroup of creating: %+v", sug)
	return sug, nil
}

func (s *Service) List() (*v1.ServiceunitGroupList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	sugs := &v1.ServiceunitGroupList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), sugs); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunitgroupList: %+v", sugs)
	return sugs, nil
}

func (s *Service) Get(id string) (*v1.ServiceunitGroup, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	sug := &v1.ServiceunitGroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), sug); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunitgroup: %+v", sug)
	return sug, nil
}

func (s *Service) Delete(id string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) UpdateSpec(sug *v1.ServiceunitGroup) (*v1.ServiceunitGroup, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(sug)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.client.Namespace(sug.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	sug = &v1.ServiceunitGroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), sug); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunitgroup: %+v", sug)

	return sug, nil
}
