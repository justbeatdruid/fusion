package service

import (
	"fmt"

	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	"github.com/chinamobile/nlpt/crds/applicationgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/util"

	k8s "github.com/chinamobile/nlpt/apiserver/kubernetes"
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
	Resource: "applicationgroups",
}

type Service struct {
	kubeClient *clientset.Clientset
	client     dynamic.NamespaceableResourceInterface
	appClient  dynamic.NamespaceableResourceInterface

	tenantEnabled bool
}

func NewService(client dynamic.Interface, kubeClient *clientset.Clientset, tenantEnabled bool) *Service {
	return &Service{
		kubeClient: kubeClient,
		client:     client.Resource(v1.GetOOFSGVR()),
		appClient:  client.Resource(appv1.GetOOFSGVR()),

		tenantEnabled: tenantEnabled,
	}
}

func (s *Service) CreateApplicationGroup(model *ApplicationGroup) (*ApplicationGroup, error) {
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

	app, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(app), nil
}

func (s *Service) ListApplicationGroup(opts ...util.OpOption) ([]*ApplicationGroup, error) {
	apps, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(apps), nil
}

func (s *Service) GetApplicationGroup(id string, opts ...util.OpOption) (*ApplicationGroup, error) {
	app, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(app), nil
}

func (s *Service) DeleteApplicationGroup(id string, opts ...util.OpOption) error {
	apps, err := s.getApplicationList(id, opts...)
	klog.V(5).Infof("get %d applications in group", len(apps.Items))
	if err != nil {
		return fmt.Errorf("cannot get application list: %+v", err)
	}
	if len(apps.Items) > 0 {
		return errors.ContentNotVoidError("content not void: %d application(s) still in group %s", len(apps.Items), id)
	}
	return s.Delete(id, opts...)
}

func (s *Service) UpdateApplicationGroup(id string, model *ApplicationGroup, opts ...util.OpOption) (*ApplicationGroup, error) {
	app, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("get applicationgroup error: %+v", err)
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
	var crdNamespace = defaultNamespace
	if s.tenantEnabled {
		crdNamespace = app.ObjectMeta.Namespace
		if len(crdNamespace) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
	} else {
		app.ObjectMeta.Namespace = defaultNamespace
	}
	if err := k8s.EnsureNamespace(s.kubeClient, crdNamespace); err != nil {
		return nil, fmt.Errorf("cannot ensure k8s namespace: %+v", err)
	}
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

func (s *Service) List(opts ...util.OpOption) (*v1.ApplicationGroupList, error) {
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
	apps := &v1.ApplicationGroupList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationgrouplist: %+v", apps)
	return apps, nil
}

func (s *Service) Get(id string, opts ...util.OpOption) (*v1.ApplicationGroup, error) {
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
	app := &v1.ApplicationGroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationgroup: %+v", app)
	return app, nil
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

func (s *Service) getApplicationList(groupid string, opts ...util.OpOption) (*appv1.ApplicationList, error) {
	crdNamespace := defaultNamespace
	if s.tenantEnabled {
		crdNamespace = util.OpList(opts...).Namespace()
		if len(crdNamespace) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
	}
	labelSelector := fmt.Sprintf("%s=%s", appv1.GroupLabel, groupid)
	klog.V(5).Infof("list with label selector: %s", labelSelector)
	crd, err := s.appClient.Namespace(crdNamespace).List(metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apps := &appv1.ApplicationList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationList: %+v", apps)
	return apps, nil
}
