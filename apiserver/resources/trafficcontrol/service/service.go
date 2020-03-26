package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/util"
	"strings"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	"github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"

	appconfig "github.com/chinamobile/nlpt/cmd/apiserver/app/config"
)

var crdNamespace = "default"

type Service struct {
	client            dynamic.NamespaceableResourceInterface
	apiClient         dynamic.NamespaceableResourceInterface
	applicationClient dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface, localConfig appconfig.ErrorConfig) *Service {
	return &Service{
		client:            client.Resource(v1.GetOOFSGVR()),
		apiClient:         client.Resource(apiv1.GetOOFSGVR()),
		applicationClient: client.Resource(appv1.GetOOFSGVR()),
	}
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

func (s *Service) ListTrafficcontrol(opts ...util.OpOption) ([]*Trafficcontrol, error) {
	apps, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(apps, opts...), nil
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
func (s *Service) UpdateTrafficcontrol(id string, reqData interface{}) (*Trafficcontrol, error) {
	crd, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	if err = s.assignment(crd, reqData); err != nil {
		return nil, err
	}
	//更新crd的状态为开始更新
	crd.Status.Status = v1.Update
	for index := range crd.Spec.Apis {
		crd.Spec.Apis[index].Result = v1.INIT
	}
	su, err := s.Update(crd)
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

func (s *Service) List(opts ...util.OpOption) (*v1.TrafficcontrolList, error) {
	var options metav1.ListOptions
	op := util.OpList(opts...)
	u := op.User()
	var labels []string
	if len(u) > 0 {
		labels = append(labels, user.GetLabelSelector(u))
	}
	options.LabelSelector = strings.Join(labels, ",")
	klog.V(5).Infof("list with label selector: %s", options.LabelSelector)
	crd, err := s.client.Namespace(crdNamespace).List(options)
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apps := &v1.TrafficcontrolList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.trafficcontrolList: %+v", apps)
	return apps, nil
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
	if len(su.Spec.Apis) != 0 {
		return fmt.Errorf("please unbind apis")
	}
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

func (s *Service) updateApi(apiid string, traffic *v1.Trafficcontrol) (*apiv1.Api, error) {
	api, err := s.getAPi(apiid)
	if err != nil {
		return nil, fmt.Errorf("cannot update api: %+v", err)
	}
	switch traffic.Spec.Type {
	case v1.APIC, v1.APPC, v1.IPC, v1.USERC:
		if v1.Bind == traffic.Status.Status {
			api.Spec.Traffic.ID = traffic.ObjectMeta.Name
			api.Spec.Traffic.Name = traffic.Spec.Name
		} else {
			api.Spec.Traffic.ID = ""
			api.Spec.Traffic.Name = ""
		}
	case v1.SPECAPPC:
		if v1.Bind == traffic.Status.Status {
			api.Spec.Traffic.SpecialID = append(api.Spec.Traffic.SpecialID, traffic.ObjectMeta.Name)
		} else {
			for index := 0; index < len(api.Spec.Traffic.SpecialID); index++ {
				if api.Spec.Traffic.SpecialID[index] == traffic.ObjectMeta.Name {
					api.Spec.Traffic.SpecialID = append(api.Spec.Traffic.SpecialID[:index], api.Spec.Traffic.SpecialID[index+1:]...)
					break
				}
			}
		}

	default:
		return nil, fmt.Errorf("wrong type: %s.", traffic.Spec.Type)
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

func (s *Service) BindOrUnbindApis(operation, id string, apis []v1.Api) (*Trafficcontrol, error) {
	if operation == "unbind" {
		return s.UnBindApi(id, apis)
	} else if operation == "bind" {
		return s.BindApi(id, apis)
	} else {
		return nil, fmt.Errorf("error operation type")
	}

}

func (s *Service) BindApi(id string, apis []v1.Api) (*Trafficcontrol, error) {
	traffic, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get traffic error: %+v", err)
	}

	//先校验是否所有API满足绑定条件，有一个不满足直接返回错误
	for _, api := range apis {
		apiSource, err := s.getAPi(api.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot get api: %+v", err)
		}
		//SPECAPPC 1个API可以绑定多个
		if len(apiSource.Spec.Traffic.ID) > 0 && traffic.Spec.Type != v1.SPECAPPC {
			return nil, fmt.Errorf("api alrady bound to traffic")
		}
	}
	//update status bind
	traffic.Status.Status = v1.Bind
	for _, api := range apis {
		apiSource, err := s.getAPi(api.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot get api: %+v", err)
		}
		//绑定api
		traffic.ObjectMeta.Labels[api.ID] = "true"
		isFisrtBind := true
		for index, v := range traffic.Spec.Apis {
			if v.ID == apiSource.ObjectMeta.Name {
				traffic.Spec.Apis[index].Result = v1.INIT
				isFisrtBind = false
				break
			}
		}
		if isFisrtBind == true {
			traffic.Spec.Apis = append(traffic.Spec.Apis, v1.Api{
				ID:     api.ID,
				Name:   apiSource.Spec.Name,
				KongID: apiSource.Spec.KongApi.KongID,
				Result: v1.INIT,
			})
		}
		//update api 操作时会判断是绑定还是解绑所以先将状态设置成bind
		if _, err = s.updateApi(api.ID, traffic); err != nil {
			return nil, fmt.Errorf("cannot update api traffic")
		}
		//update traffic
	}
	//update traffic 所有api绑定完成后更新数据库的状态
	traffic, err = s.UpdateSpec(traffic)
	return ToModel(traffic), err
}

func (s *Service) UnBindApi(id string, apis []v1.Api) (*Trafficcontrol, error) {
	traffic, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("get traffic error: %+v", err)
	}

	//先校验是否所有API满足绑定条件，有一个不满足直接返回错误
	for _, api := range apis {
		apiSource, err := s.getAPi(api.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot get api: %+v", err)
		}
		if len(apiSource.Spec.Traffic.ID) == 0 && traffic.Spec.Type != v1.SPECAPPC {
			return nil, fmt.Errorf("api has no bound to traffic")
		}
		if len(apiSource.Spec.Traffic.SpecialID) == 0 && traffic.Spec.Type == v1.SPECAPPC {
			return nil, fmt.Errorf("api has no bound to special traffic")
		}
	}
	//update status unbind
	traffic.Status.Status = v1.UnBind
	for _, api := range apis {
		apiSource, err := s.getAPi(api.ID)
		if err != nil {
			return nil, fmt.Errorf("cannot get api: %+v", err)
		}
		//解除绑定
		for index, v := range traffic.Spec.Apis {
			if v.ID == apiSource.ObjectMeta.Name {
				traffic.Spec.Apis[index].Result = v1.INIT
			}
		}
		//update api 操作时会判断是绑定还是解绑所以先将状态设置成unbind
		traffic.ObjectMeta.Labels[api.ID] = "false"
		if _, err = s.updateApi(api.ID, traffic); err != nil {
			return nil, fmt.Errorf("cannot update api traffic")
		}
	}
	//update traffic
	traffic, err = s.UpdateSpec(traffic)
	return ToModel(traffic), err
}
