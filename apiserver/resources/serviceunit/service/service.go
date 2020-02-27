package service

import (
	"fmt"

	datav1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	groupv1 "github.com/chinamobile/nlpt/crds/serviceunitgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

var crdNamespace = "default"

type Service struct {
	client           dynamic.NamespaceableResourceInterface
	datasourceClient dynamic.NamespaceableResourceInterface
	groupClient      dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{
		client:           client.Resource(v1.GetOOFSGVR()),
		datasourceClient: client.Resource(datav1.GetOOFSGVR()),
		groupClient:      client.Resource(groupv1.GetOOFSGVR()),
	}
}

func (s *Service) CreateServiceunit(model *Serviceunit) (*Serviceunit, error) {
	if err := s.Validate(model); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	su, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(su), nil
}

func (s *Service) ListServiceunit(group string, opts ...util.OpOption) ([]*Serviceunit, error) {
	sus, err := s.List(group)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	groupMap, err := s.GetGroupMap()
	if err != nil {
		return nil, fmt.Errorf("get groups error: %+v", err)
	}
	dataMap, err := s.GetDatasourceMap()
	if err != nil {
		return nil, fmt.Errorf("get datasource error: %+v", err)
	}
	return ToListModel(sus, groupMap, dataMap, opts...), nil
}

func (s *Service) GetServiceunit(id string) (*Serviceunit, error) {
	su, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(su), nil
}

func (s *Service) DeleteServiceunit(id string) (*Serviceunit, error) {
	su, err := s.Delete(id)
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}
	return ToModel(su), nil
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

func (s *Service) PublishServiceunit(id string, published bool) (*Serviceunit, error) {
	su, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get crd by id error: %+v", err)
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
	if group, ok := su.ObjectMeta.Labels[v1.GroupLabel]; !ok {
		//return nil, fmt.Errorf("group not found")
	} else {
		if _, err := s.GetGroup(group); err != nil {
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
	return su, nil
}

func (s *Service) PatchServiceunit(id string, data interface{}) (*Serviceunit, error) {
	su, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	if err = s.assignment(su, data); err != nil {
		return nil, err
	}
	su.Status.Status = v1.Update
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

func (s *Service) List(group string) (*v1.ServiceunitList, error) {
	var options metav1.ListOptions
	if len(group) > 0 {
		options.LabelSelector = fmt.Sprintf("%s=%s", v1.GroupLabel, group)
	}
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

func (s *Service) Get(id string) (*v1.Serviceunit, error) {
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
			group, err := s.GetGroup(gid)
			if err != nil {
				return nil, fmt.Errorf("get group error: %+v", err)
			}
			su.Spec.Group.ID = group.ObjectMeta.Name
			su.Spec.Group.Name = group.Spec.Name
		}
	}
	return su, nil
}

func (s *Service) ForceDelete(id string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) Delete(id string) (*v1.Serviceunit, error) {
	su, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get crd by id error: %+v", err)
	}
	//TODO need check status !!!
	su.Status.Status = v1.Delete
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

func (s *Service) getDatasource(id string) (*datav1.DatasourceSpec, error) {
	crd, err := s.datasourceClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	ds := &datav1.Datasource{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ds); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasource: %+v", ds)
	return &ds.Spec, nil
}

func (s *Service) GetGroup(id string) (*groupv1.ServiceunitGroup, error) {
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

func (s *Service) GetGroupMap() (map[string]string, error) {
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

func (s *Service) GetDatasourceMap() (map[string]v1.Datasource, error) {
	crd, err := s.datasourceClient.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	datas := &datav1.DatasourceList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), datas); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasourcelist: %+v", datas)
	m := make(map[string]v1.Datasource)
	for _, data := range datas.Items {
		ds := v1.Datasource{
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
