package service

import (
	"fmt"

	k8s "github.com/chinamobile/nlpt/apiserver/kubernetes"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"github.com/chinamobile/nlpt/crds/serviceunitgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var defaultNamespace = "default"

var oofsGVR = schema.GroupVersionResource{
	Group:    v1.GroupVersion.Group,
	Version:  v1.GroupVersion.Version,
	Resource: "serviceunitgroups",
}

type Service struct {
	kubeClient *clientset.Clientset
	client     dynamic.NamespaceableResourceInterface
	suClient   dynamic.NamespaceableResourceInterface

	tenantEnabled bool
}

func NewService(client dynamic.Interface, kubeClient *clientset.Clientset, tenantEnabled bool) *Service {
	return &Service{
		kubeClient: kubeClient,
		client:     client.Resource(v1.GetOOFSGVR()),
		suClient:   client.Resource(suv1.GetOOFSGVR()),

		tenantEnabled: tenantEnabled,
	}
}

func (s *Service) CreateServiceunitGroup(model *ServiceunitGroup) (*ServiceunitGroup, error) {
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	// check if unique
	list, err := s.List(util.WithNamespace(model.Namespace))
	if err != nil {
		return nil, fmt.Errorf("cannot get application group list: %+v", err)
	}
	for _, item := range list.Items {
		if item.Spec.Name == model.Name {
			return nil, errors.NameDuplicatedError("name dumplicated: %s", model.Name)
		}
	}
	sug, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(sug), nil
}

func (s *Service) ListServiceunitGroup(opts ...util.OpOption) ([]*ServiceunitGroup, error) {
	sugs, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(sugs), nil
}

func (s *Service) GetServiceunitGroup(id string, opts ...util.OpOption) (*ServiceunitGroup, error) {
	sug, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(sug), nil
}

func (s *Service) DeleteServiceunitGroup(id string, opts ...util.OpOption) error {
	sus, err := s.getServiceunitList(id, opts...)
	klog.V(5).Infof("get %d serviceunits in group", len(sus.Items))
	if err != nil {
		return fmt.Errorf("cannot get serviceunit list: %+v", err)
	}
	if len(sus.Items) > 0 {
		return errors.ContentNotVoidError("content not void: %d serviceunits(s) still in group %s", len(sus.Items), id)
	}
	return s.Delete(id, opts...)
}

func (s *Service) UpdateServiceunitGroup(id string, model *ServiceunitGroup, opts ...util.OpOption) (*ServiceunitGroup, error) {
	sug, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("get serviceunitgroup error: %+v", err)
	}
	if len(model.Name) > 0 {
		// check if unique
		list, err := s.List(opts...)
		if err != nil {
			return nil, fmt.Errorf("cannot get application group list: %+v", err)
		}
		for _, item := range list.Items {
			if item.Spec.Name == model.Name {
				return nil, errors.NameDuplicatedError("name dumplicated: %s", model.Name)
			}
		}
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
	var crdNamespace = defaultNamespace
	if s.tenantEnabled {
		crdNamespace = sug.ObjectMeta.Namespace
		if len(crdNamespace) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
	} else {
		sug.ObjectMeta.Namespace = defaultNamespace
	}
	if err := k8s.EnsureNamespace(s.kubeClient, crdNamespace); err != nil {
		return nil, fmt.Errorf("cannot ensure k8s namespace: %+v", err)
	}

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

func (s *Service) List(opts ...util.OpOption) (*v1.ServiceunitGroupList, error) {
	op := util.OpList(opts...)
	ns := op.Namespace()
	var crdNamespace = defaultNamespace
	if s.tenantEnabled {
		if len(ns) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
		crdNamespace = ns
	}

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

func (s *Service) Get(id string, opts ...util.OpOption) (*v1.ServiceunitGroup, error) {
	crdNamespace := defaultNamespace
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
	sug := &v1.ServiceunitGroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), sug); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunitgroup: %+v", sug)
	return sug, nil
}

func (s *Service) Delete(id string, opts ...util.OpOption) error {
	crdNamespace := defaultNamespace
	if s.tenantEnabled {
		crdNamespace = util.OpList(opts...).Namespace()
		if len(crdNamespace) == 0 {
			return fmt.Errorf("namespace not set")
		}
	}

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

func (s *Service) getServiceunitList(groupid string, opts ...util.OpOption) (*suv1.ServiceunitList, error) {
	crdNamespace := defaultNamespace
	if s.tenantEnabled {
		crdNamespace = util.OpList(opts...).Namespace()
		if len(crdNamespace) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
	}

	labelSelector := fmt.Sprintf("%s=%s", suv1.GroupLabel, groupid)
	klog.V(5).Infof("list with label selector: %s", labelSelector)
	crd, err := s.suClient.Namespace(crdNamespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	sus := &suv1.ServiceunitList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), sus); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunitList: %+v", sus)
	return sus, nil
}
