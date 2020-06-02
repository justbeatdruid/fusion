package service

import (
	"fmt"
	"strings"

	k8s "github.com/chinamobile/nlpt/apiserver/kubernetes"
	datav1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	groupv1 "github.com/chinamobile/nlpt/crds/serviceunitgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"

	appconfig "github.com/chinamobile/nlpt/cmd/apiserver/app/config"
)

var defaultNamespace = "default"

type Service struct {
	kubeClient       *clientset.Clientset
	client           dynamic.NamespaceableResourceInterface
	datasourceClient dynamic.NamespaceableResourceInterface
	groupClient      dynamic.NamespaceableResourceInterface

	tenantEnabled bool
}

func NewService(client dynamic.Interface, kubeClient *clientset.Clientset, tenantEnabled bool, localConfig appconfig.ErrorConfig) *Service {
	return &Service{
		kubeClient:       kubeClient,
		client:           client.Resource(v1.GetOOFSGVR()),
		datasourceClient: client.Resource(datav1.GetOOFSGVR()),
		groupClient:      client.Resource(groupv1.GetOOFSGVR()),

		tenantEnabled: tenantEnabled,
	}
}

func (s *Service) GetClient() dynamic.NamespaceableResourceInterface {
	return s.client
}

func (s *Service) CreateServiceunit(model *Serviceunit) (*Serviceunit, error, string) {
	if err := s.Validate(model); err != nil {
		return nil, err, "008000019"
	}
	su, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err), "008000020"
	}
	return ToModel(su), nil, "0"
}

func (s *Service) ListServiceunit(opts ...util.OpOption) ([]*Serviceunit, error) {
	sus, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	groupMap, err := s.GetGroupMap(util.OpList(opts...).Namespace())
	if err != nil {
		return nil, fmt.Errorf("get groups error: %+v", err)
	}
	dataMap, err := s.GetDatasourceMap()
	if err != nil {
		return nil, fmt.Errorf("get datasource error: %+v", err)
	}
	return ToListModel(sus, groupMap, dataMap, opts...), nil
}

func (s *Service) GetServiceunit(id string, opts ...util.OpOption) (*Serviceunit, error) {
	su, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	if su.Spec.Type == v1.DataService && len(su.Spec.DatasourceID.ID) > 0 {
		data, err := s.getDatasource(su.ObjectMeta.Namespace, su.Spec.DatasourceID.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot get datasource: %+v", err)
		}
		ds := &v1.Datasource{
			ID:   data.ObjectMeta.Name,
			Name: data.Spec.Name,
		}
		if data.Spec.Type == datav1.DataWarehouseType && data.Spec.DataWarehouse != nil {
			ds.DataWarehouse = &v1.DataWarehouse{
				DatabaseName:        data.Spec.DataWarehouse.Name,
				DatabaseDisplayName: data.Spec.DataWarehouse.DisplayName,
				SubjectName:         data.Spec.DataWarehouse.SubjectName,
				SubjectDisplayName:  data.Spec.DataWarehouse.SubjectDisplayName,
			}
		}
		su.Spec.DatasourceID = ds
	}
	return ToModel(su, opts...), nil
}

func (s *Service) DeleteServiceunit(id string, opts ...util.OpOption) (*Serviceunit, error) {
	su, err := s.Delete(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}
	util.WaitDelete(s, su.ObjectMeta)
	return ToModel(su), err
}

// + update_sunyu
func (s *Service) UpdateServiceunit(model *Serviceunit, id string) (*Serviceunit, error) {
	//db, err := s.GetServiceunit(id)
	crd, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	su, err := s.Update(ToAPIUpdate(model, crd)) //model是传入的，db是原始的
	if err != nil {
		return nil, fmt.Errorf("cannot update status to update: %+v", err)
	}
	return ToModel(su), nil
}

func (s *Service) PublishServiceunit(id string, published bool, opts ...util.OpOption) (*Serviceunit, error) {
	su, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("get crd by id error: %+v", err)
	}
	if !s.tenantEnabled {
		u := util.OpList(opts...).User()
		if !user.WritePermitted(u, su.ObjectMeta.Labels) {
			return nil, fmt.Errorf("write permission denied")
		}
	}
	var status string
	if published {
		status = "published"
	} else {
		status = "unpublished"
	}
	if su.Status.Published == published {
		return nil, fmt.Errorf("serviceunit already %s", status)
	}
	su.Status.Published = published
	su, err = s.UpdateStatus(su)
	return ToModel(su), err
}

func (s *Service) Create(su *v1.Serviceunit) (*v1.Serviceunit, error) {
	var crdNamespace = defaultNamespace
	if s.tenantEnabled {
		crdNamespace = su.ObjectMeta.Namespace
		if len(crdNamespace) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
	} else {
		su.ObjectMeta.Namespace = defaultNamespace
	}
	if err := k8s.EnsureNamespace(s.kubeClient, crdNamespace); err != nil {
		return nil, fmt.Errorf("cannot ensure k8s namespace: %+v", err)
	}

	if group, ok := su.ObjectMeta.Labels[v1.GroupLabel]; !ok {
		//return nil, fmt.Errorf("group not found")
	} else {
		if _, err := s.GetGroup(group, crdNamespace); err != nil {
			return nil, fmt.Errorf("get group error: %+v", err)
		}
	}

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(su)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})

	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunit of creating: %+v", su)
	(*su).Spec.Result = v1.CREATING
	return su, nil
}

func (s *Service) PatchServiceunit(id string, data interface{}, opts ...util.OpOption) (*Serviceunit, error) {
	su, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	if err = s.assignment(su, data); err != nil {
		return nil, err
	}
	crdNamespace := su.ObjectMeta.Namespace

	if !s.tenantEnabled {
		u := util.OpList(opts...).User()
		if !user.WritePermitted(u, su.ObjectMeta.Labels) {
			return nil, fmt.Errorf("write permission denied")
		}
	}

	su.Status.Status = v1.Update
	(*su).Spec.Result = v1.UPDATING
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(su)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(crdNamespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd: %+v", err)
	}
	return ToModel(su), err
}

func (s *Service) List(opts ...util.OpOption) (*v1.ServiceunitList, error) {
	var options metav1.ListOptions
	op := util.OpList(opts...)
	group := op.Group()
	ns := op.Namespace()
	u := op.User()
	var labels []string
	if len(group) > 0 {
		labels = append(labels, fmt.Sprintf("%s=%s", v1.GroupLabel, group))
	}
	var crdNamespace = defaultNamespace
	if s.tenantEnabled {
		if len(ns) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
		crdNamespace = ns
	} else {
		if len(u) > 0 {
			labels = append(labels, user.GetLabelSelector(u))
		}
	}
	options.LabelSelector = strings.Join(labels, ",")
	klog.V(5).Infof("list with label selector: %s", options.LabelSelector)
	crd, err := s.client.Namespace(crdNamespace).List(options)
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	sus := &v1.ServiceunitList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), sus); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunitList: %+v", sus)
	return sus, nil
}

func (s *Service) Get(id string, opts ...util.OpOption) (*v1.Serviceunit, error) {
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
	su := &v1.Serviceunit{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunit: %+v", su)
	if su.ObjectMeta.Labels != nil {
		if gid, ok := su.ObjectMeta.Labels[v1.GroupLabel]; ok {
			group, err := s.GetGroup(gid, crdNamespace)
			if err != nil {
				return nil, fmt.Errorf("get group error: %+v", err)
			}
			su.Spec.Group.ID = group.ObjectMeta.Name
			su.Spec.Group.Name = group.Spec.Name
		}
	}
	return su, nil
}

func (s *Service) ForceDelete(id, crdNamespace string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) Delete(id string, opts ...util.OpOption) (*v1.Serviceunit, error) {
	su, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("get crd by id error: %+v", err)
	}
	if su.Status.APICount != 0 {
		return nil, fmt.Errorf("existing apis references this serviceunit")
	}
	if !s.tenantEnabled {
		u := util.OpList(opts...).User()
		if !user.WritePermitted(u, su.ObjectMeta.Labels) {
			return nil, fmt.Errorf("write permission denied")
		}
	}

	//TODO need check status !!!
	su.Status.Status = v1.Delete
	(*su).Spec.Result = v1.DELETING
	return s.UpdateStatus(su)
}

// + update_sunyu
func (s *Service) Update(su *v1.Serviceunit) (*v1.Serviceunit, error) {
	return s.UpdateStatus(su)
}

func (s *Service) UpdateSpec(su *v1.Serviceunit) (*v1.Serviceunit, error) {
	return s.UpdateStatus(su)
}

func (s *Service) UpdateStatus(su *v1.Serviceunit) (*v1.Serviceunit, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(su)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.client.Namespace(su.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	su = &v1.Serviceunit{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunit: %+v", su)

	return su, nil
}

func (s *Service) getDatasource(crdNamespace, id string) (*datav1.Datasource, error) {
	// TODO
	crd, err := s.datasourceClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	ds := &datav1.Datasource{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ds); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasource: %+v", ds)
	return ds, nil
}

func (s *Service) GetGroup(id, crdNamespace string) (*groupv1.ServiceunitGroup, error) {
	crd, err := s.groupClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	su := &groupv1.ServiceunitGroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunitgroup: %+v", su)
	return su, nil
}

func (s *Service) GetGroupMap(crdNamespace string) (map[string]string, error) {
	crd, err := s.groupClient.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	sus := &groupv1.ServiceunitGroupList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), sus); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunitgrouplist: %+v", sus)
	m := make(map[string]string)
	for _, su := range sus.Items {
		m[su.ObjectMeta.Name] = su.Spec.Name
	}
	return m, nil
}

func (s *Service) GetDatasourceMap() (map[string]*v1.Datasource, error) {
	// TODO
	crdNamespace := defaultNamespace
	crd, err := s.datasourceClient.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	datas := &datav1.DatasourceList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), datas); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasourcelist: %+v", datas)
	m := make(map[string]*v1.Datasource)
	for _, data := range datas.Items {
		ds := &v1.Datasource{
			ID:   data.ObjectMeta.Name,
			Name: data.Spec.Name,
		}
		if data.Spec.Type == datav1.DataWarehouseType && data.Spec.DataWarehouse != nil {
			ds.DataWarehouse = &v1.DataWarehouse{
				DatabaseName:        data.Spec.DataWarehouse.Name,
				DatabaseDisplayName: data.Spec.DataWarehouse.DisplayName,
				SubjectName:         data.Spec.DataWarehouse.SubjectName,
				SubjectDisplayName:  data.Spec.DataWarehouse.SubjectDisplayName,
			}
		}
		m[data.ObjectMeta.Name] = ds
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

func (s *Service) GetUsers(id, operator string) (user.UserList, bool, error) {
	crd, err := s.Get(id)
	if err != nil {
		return nil, false, fmt.Errorf("get crd error: %+v", err)
	}
	labels := crd.ObjectMeta.Labels
	return user.GetCasUsers(labels), user.IsOwner(operator, labels), nil
}
