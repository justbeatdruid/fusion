package service

import (
	"fmt"
	"strings"

	"github.com/chinamobile/nlpt/crds/application/api/v1"
	groupv1 "github.com/chinamobile/nlpt/crds/applicationgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

var crdNamespace = "default"

type Service struct {
	client      dynamic.NamespaceableResourceInterface
	groupClient dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{
		client:      client.Resource(v1.GetOOFSGVR()),
		groupClient: client.Resource(groupv1.GetOOFSGVR()),
	}
}

func (s *Service) CreateApplication(model *Application) (*Application, error) {
	if err := s.Validate(model); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	app, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(app), nil
}

func (s *Service) ListApplication(opts ...util.OpOption) ([]*Application, error) {
	apps, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	groupMap, err := s.GetGroupMap()
	if err != nil {
		return nil, fmt.Errorf("get groups error: %+v", err)
	}
	return ToListModel(apps, groupMap, opts...), nil
}

func (s *Service) GetApplication(id string) (*Application, error) {
	app, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(app), nil
}

func (s *Service) DeleteApplication(id string) (*Application, error) {
	app, err := s.Delete(id)
	if err != nil {
		return nil, fmt.Errorf("cannot delete object: %+v", err)
	}
	return ToModel(app), err
}

func (s *Service) Create(app *v1.Application) (*v1.Application, error) {
	if group, ok := app.ObjectMeta.Labels[v1.GroupLabel]; !ok {
		//return nil, fmt.Errorf("group not found")
	} else {
		if _, err := s.GetGroup(group); err != nil {
			return nil, fmt.Errorf("get group error: %+v", err)
		}
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
	klog.V(5).Infof("get v1.application of creating: %+v", app)
	return app, nil
}

func (s *Service) PatchApplication(id string, data interface{}) (*Application, error) {
	app, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	if err = s.assignment(app, data); err != nil {
		return nil, err
	}
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(app)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(crdNamespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd: %+v", err)
	}
	return ToModel(app), err
}

func (s *Service) List(opts ...util.OpOption) (*v1.ApplicationList, error) {
	var options metav1.ListOptions
	op := util.OpList(opts...)
	group := op.Group()
	u := op.User()
	var labels []string
	if len(group) > 0 {
		labels = append(labels, fmt.Sprintf("%s=%s", v1.GroupLabel, group))
	}
	if len(u) > 0 {
		labels = append(labels, user.GetLabelSelector(u))
	}
	options.LabelSelector = strings.Join(labels, ",")
	klog.V(5).Infof("list with label selector: %s", options.LabelSelector)
	crd, err := s.client.Namespace(crdNamespace).List(options)
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apps := &v1.ApplicationList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationList: %+v", apps)
	return apps, nil
}

func (s *Service) Get(id string) (*v1.Application, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	app := &v1.Application{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.application: %+v", app)
	if app.ObjectMeta.Labels != nil {
		if gid, ok := app.ObjectMeta.Labels[v1.GroupLabel]; ok {
			group, err := s.GetGroup(gid)
			if err != nil {
				return nil, fmt.Errorf("get group error: %+v", err)
			}
			app.Spec.Group.ID = group.ObjectMeta.Name
			app.Spec.Group.Name = group.Spec.Name
		}
	}
	return app, nil
}

func (s *Service) ForceDelete(id string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) Delete(id string) (*v1.Application, error) {
	app, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get crd by id error: %+v", err)
	}
	//TODO need check status !!!
	app.Status.Status = v1.Delete
	return s.UpdateStatus(app)
}

func (s *Service) UpdateSpec(app *v1.Application) (*v1.Application, error) {
	return s.UpdateStatus(app)
}

func (s *Service) UpdateStatus(app *v1.Application) (*v1.Application, error) {
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
	app = &v1.Application{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.application: %+v", app)

	return app, nil
}

func (s *Service) GetGroup(id string) (*groupv1.ApplicationGroup, error) {
	crd, err := s.groupClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	app := &groupv1.ApplicationGroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationgroup: %+v", app)
	return app, nil
}

func (s *Service) GetGroupMap() (map[string]string, error) {
	crd, err := s.groupClient.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apps := &groupv1.ApplicationGroupList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationgrouplist: %+v", apps)
	m := make(map[string]string)
	for _, app := range apps.Items {
		m[app.ObjectMeta.Name] = app.Spec.Name
	}
	return m, nil
}

func (s *Service) AddUser(id, operator string, data *user.Data) error {
	crd, err := s.Get(id)
	if err != nil {
		return fmt.Errorf("get crd error: %+v", err)
	}
	labels := crd.ObjectMeta.Labels
	if !user.IsOwner(operator, labels) && !user.IsManager(operator, labels) {
		return fmt.Errorf("only owner or manager can add user")
	}
	labels, err = user.AddUserLabels(data, labels)
	if err != nil {
		return fmt.Errorf("add user labels error: %+v", err)
	}
	crd.ObjectMeta.Labels = labels
	_, err = s.UpdateSpec(crd)
	if err != nil {
		return fmt.Errorf("update crd error: %+v", err)
	}
	return nil
}

func (s *Service) RemoveUser(id, operator, target string) error {
	crd, err := s.Get(id)
	if err != nil {
		return fmt.Errorf("get crd error: %+v", err)
	}
	labels := crd.ObjectMeta.Labels
	if !user.IsOwner(operator, labels) && !user.IsManager(operator, labels) {
		return fmt.Errorf("only owner or manager can remove user")
	}
	labels, err = user.RemoveUserLabels(target, labels)
	if err != nil {
		return fmt.Errorf("remove user labels error: %+v", err)
	}
	crd.ObjectMeta.Labels = labels
	_, err = s.UpdateSpec(crd)
	if err != nil {
		return fmt.Errorf("update crd error: %+v", err)
	}
	return nil
}

func (s *Service) ChangeOwner(id, operator string, data *user.Data) error {
	klog.V(5).Infof("change owner appid=%s, operator=%s, data=%+v", id, operator, *data)
	crd, err := s.Get(id)
	if err != nil {
		return fmt.Errorf("get crd error: %+v", err)
	}
	labels := crd.ObjectMeta.Labels
	if !user.IsOwner(operator, labels) {
		return fmt.Errorf("only owner or can change owner")
	}
	labels, err = user.ChangeOwner(data.ID, labels)
	if err != nil {
		return fmt.Errorf("change owner labels error: %+v", err)
	}
	crd.ObjectMeta.Labels = labels
	_, err = s.UpdateSpec(crd)
	if err != nil {
		return fmt.Errorf("update crd error: %+v", err)
	}
	return nil
}

func (s *Service) ChangeUser(id, operator string, data *user.Data) error {
	crd, err := s.Get(id)
	if err != nil {
		return fmt.Errorf("get crd error: %+v", err)
	}
	labels := crd.ObjectMeta.Labels
	if !user.IsOwner(operator, labels) && !user.IsManager(operator, labels) {
		return fmt.Errorf("only owner or manager can add user")
	}
	labels, err = user.ChangeUser(data, labels)
	if err != nil {
		return fmt.Errorf("change user labels error: %+v", err)
	}
	crd.ObjectMeta.Labels = labels
	_, err = s.UpdateSpec(crd)
	if err != nil {
		return fmt.Errorf("update crd error: %+v", err)
	}
	return nil
}
