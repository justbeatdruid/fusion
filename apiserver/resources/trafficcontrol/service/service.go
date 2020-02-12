package service

import (
	"fmt"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	"github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

var crdNamespace = "default"

type Service struct {
	client    dynamic.NamespaceableResourceInterface
	apiClient dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{
		client:    client.Resource(v1.GetOOFSGVR()),
		apiClient: client.Resource(apiv1.GetOOFSGVR()),
	}
}

func (s *Service) CreateTrafficcontrol(model *Trafficcontrol) (*Trafficcontrol, error) {
	if err := s.Validate(model); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	su, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(su), nil
}

func (s *Service) ListTrafficcontrol() ([]*Trafficcontrol, error) {
	sus, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(sus), nil
}

func (s *Service) GetTrafficcontrol(id string) (*Trafficcontrol, error) {
	su, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(su), nil
}

func (s *Service) DeleteTrafficcontrol(id string) error {
	err := s.Delete(id)
	if err != nil {
		return fmt.Errorf("cannot delete traffic control: %+v", err)
	}
	return nil
}

// + update_sunyu
func (s *Service) UpdateTrafficcontrol(model *Trafficcontrol, id string) (*Trafficcontrol, error) {
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

func (s *Service) Create(su *v1.Trafficcontrol) (*v1.Trafficcontrol, error) {
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
	klog.V(5).Infof("get v1.trafficcontrol of creating: %+v", su)
	return su, nil
}

func (s *Service) List() (*v1.TrafficcontrolList, error) {
	var options metav1.ListOptions
	crd, err := s.client.Namespace(crdNamespace).List(options)
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	sus := &v1.TrafficcontrolList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), sus); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunitList: %+v", sus)
	return sus, nil
}

func (s *Service) Get(id string) (*v1.Trafficcontrol, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	su := &v1.Trafficcontrol{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.trafficcontrol: %+v", su)
	return su, nil
}

func (s *Service) ForceDelete(id string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) Delete(id string) error {
	su, err := s.Get(id)
	if err != nil {
		return fmt.Errorf("get crd by id error: %+v", err)
	}
	klog.V(5).Infof("get v1.trafficcontrol: %+v", su)
	err = s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	//TODO need check status !!!
	//su.Status.Status = v1.Delete
	return nil
}

// + update_sunyu
func (s *Service) Update(su *v1.Trafficcontrol) (*v1.Trafficcontrol, error) {
	return s.UpdateStatus(su)
}

func (s *Service) UpdateSpec(su *v1.Trafficcontrol) (*v1.Trafficcontrol, error) {
	return s.UpdateStatus(su)
}

func (s *Service) UpdateStatus(su *v1.Trafficcontrol) (*v1.Trafficcontrol, error) {
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
	su = &v1.Trafficcontrol{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.trafficcontrol: %+v", su)

	return su, nil
}

func (s *Service) getAPi(id string) (*apiv1.Api, error) {
	crd, err := s.apiClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	su := &apiv1.Api{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.api info: %+v", su)
	return su, nil
}

func (s *Service) updateApi(apiid string, traffic *v1.Trafficcontrol, status v1.Status) (*apiv1.Api, error) {
	api, err := s.getAPi(apiid)
	if err != nil {
		return nil, fmt.Errorf("cannot update api: %+v", err)
	}
	if v1.Bind == status {
		api.Spec.Traffic = apiv1.Traffic{
			ID:   traffic.ObjectMeta.Name,
			Name: traffic.Spec.Name,
		}
	} else {
		api.Spec.Traffic = apiv1.Traffic{
			ID:   "",
			Name: "",
		}
	}
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(api)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd api: %+v", crd)
	crd, err = s.apiClient.Namespace(api.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd api status: %+v", err)
	}
	api = &apiv1.Api{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), api); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	return api, nil
}
func (s *Service) BindApi(id string, apis []v1.Api) (*Trafficcontrol, error) {
	traffic, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get traffic error: %+v", err)
	}

	for _, api := range apis {
		apiSource, err := s.getAPi(api.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot get api: %+v", err)
		}
		if len(apiSource.Spec.Traffic.ID) > 0 {
			return nil, fmt.Errorf("api alrady bound to traffic")
		}
		//绑定api
		traffic.ObjectMeta.Labels[api.ID] = "true"
		traffic.Spec.Apis = append(traffic.Spec.Apis, v1.Api{
			ID:     api.ID,
			Name:   apiSource.Spec.Name,
			KongID: apiSource.Spec.KongApi.KongID,
			Result: v1.INIT,
		})

		if _, err = s.updateApi(api.ID, traffic, v1.Bind); err != nil {
			return nil, fmt.Errorf("cannot update api traffic")
		}
		//update status bind
		traffic.Status.Status = v1.Bind
		traffic, err = s.UpdateSpec(traffic)
		traffic, err = s.UpdateStatus(traffic)
	}
	return ToModel(traffic), err
}

func (s *Service) UnBindApi(id string, apis []v1.Api) (*Trafficcontrol, error) {
	traffic, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get traffic error: %+v", err)
	}

	for _, api := range apis {
		apiSource, err := s.getAPi(api.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot get api: %+v", err)
		}
		if len(apiSource.Spec.Traffic.ID) == 0 {
			return nil, fmt.Errorf("api has no bound to traffic")
		}
		//解除绑定
		traffic.ObjectMeta.Labels[api.ID] = "false"
		if _, err = s.updateApi(api.ID, traffic, v1.UnBind); err != nil {
			return nil, fmt.Errorf("cannot update api traffic")
		}
		//update status unbind
		traffic.Status.Status = v1.UnBind
		traffic, err = s.UpdateSpec(traffic)
		traffic, err = s.UpdateStatus(traffic)
	}
	return ToModel(traffic), err
}
