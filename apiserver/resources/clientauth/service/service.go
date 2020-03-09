package service

import (
	"fmt"

	"github.com/chinamobile/nlpt/crds/clientauth/api/v1"

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
	Resource: "clientauths",
}

type Service struct {
	client dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{client: client.Resource(oofsGVR)}
}

func (s *Service) CreateClientauth(model *Clientauth) (*Clientauth, error) {
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	ca, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(ca), nil
}

func (s *Service) ListClientauth() ([]*Clientauth, error) {
	cas, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(cas), nil
}

func (s *Service) GetClientauth(id string) (*Clientauth, error) {
	ca, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(ca), nil
}

func (s *Service) DeleteClientauth(id string) (*Clientauth, error) {
	ca, err := s.Delete(id)
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}
	return ToModel(ca), nil
}


func (s *Service) DeleteAllClientauths() ([]*Clientauth, error) {
	cas, err := s.DeleteClientauths()
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}
	return ToListModel(cas), nil
}

func (s *Service) Create(ca *v1.Clientauth) (*v1.Clientauth, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ca)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ca); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.clientauth of creating: %+v", ca)
	return ca, nil
}

func (s *Service) List() (*v1.ClientauthList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	cas := &v1.ClientauthList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), cas); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.clientauthList: %+v", cas)
	return cas, nil
}

func (s *Service) Get(id string) (*v1.Clientauth, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	ca := &v1.Clientauth{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ca); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.Clientauth: %+v", ca)
	return ca, nil
}

func (s *Service) Delete(id string) (*v1.Clientauth, error) {
	ca, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("error delete crd: %+v", err)
	}
	ca.Status.Status = v1.Delete
	return s.UpdateStatus(ca)
}

//将所有Clientauth的status置为delete
func (s *Service) DeleteClientauths() (*v1.ClientauthList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	cas := &v1.ClientauthList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), cas); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	for i := range cas.Items {
		cas.Items[i].Status.Status = v1.Delete
	}
	for j := range cas.Items {
		content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&(cas.Items[j]))
		if err != nil {
			return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
		}
		crd := &unstructured.Unstructured{}
		crd.SetUnstructuredContent(content)
		klog.V(5).Infof("try to update status for crd: %+v", crd)
		crd, err = s.client.Namespace(cas.Items[j].ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("error update crd status: %+v", err)
		}
	}
	cas = &v1.ClientauthList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), cas); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}

	return cas, nil
}

//更新状态
func (s *Service) UpdateStatus(ca *v1.Clientauth) (*v1.Clientauth, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ca)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.client.Namespace(ca.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	ca = &v1.Clientauth{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ca); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.clientauth: %+v", ca)

	return ca, nil
}
