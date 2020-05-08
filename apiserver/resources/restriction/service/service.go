package service

import (
	"fmt"
	k8s "github.com/chinamobile/nlpt/apiserver/kubernetes"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/util"
	clientset "k8s.io/client-go/kubernetes"
	"strings"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	"github.com/chinamobile/nlpt/crds/restriction/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"

	appconfig "github.com/chinamobile/nlpt/cmd/apiserver/app/config"
)

var defaultNamespace = "default"

type Service struct {
	kubeClient    *clientset.Clientset
	client        dynamic.NamespaceableResourceInterface
	apiClient     dynamic.NamespaceableResourceInterface
	tenantEnabled bool
	localConfig   appconfig.ErrorConfig
}

func NewService(client dynamic.Interface, kubeClient *clientset.Clientset, tenantEnabled bool, localConfig appconfig.ErrorConfig) *Service {
	return &Service{
		kubeClient:    kubeClient,
		client:        client.Resource(v1.GetOOFSGVR()),
		apiClient:     client.Resource(apiv1.GetOOFSGVR()),
		tenantEnabled: tenantEnabled,
		localConfig:   localConfig,
	}
}

func (s *Service) CreateRestriction(model *Restriction) (*Restriction, error, string) {
	if err := s.Validate(model); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err), "007000012"
	}
	su, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err), "007000013"
	}
	return ToModel(su), nil, "0"
}

func (s *Service) ListRestriction(opts ...util.OpOption) ([]*Restriction, error) {
	apps, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(apps, opts...), nil
}

func (s *Service) ListApis(crdNamespace string) (*apiv1.ApiList, error) {
	crd, err := s.apiClient.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apis := &apiv1.ApiList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apis); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("====test get v1.ApiList: %+v", apis)
	return apis, nil
}

func (s *Service) GetRestriction(id string, opts ...util.OpOption) (*Restriction, error) {
	su, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	var crdNamespace = defaultNamespace
	if s.tenantEnabled {
		crdNamespace = util.OpList(opts...).Namespace()
		if len(crdNamespace) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
	}
	apiList, err := s.ListApis(crdNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot list api object: %+v", err)
	}
	nameIds := make(map[string]string)
	for _, value := range apiList.Items {
		nameIds[value.ObjectMeta.Name] = value.Spec.Name
	}
	for index, value := range su.Spec.Apis {
		(*su).Spec.Apis[index].Name = nameIds[value.ID]
		klog.V(5).Infof("get api id and name : %s, %s", value.ID, su.Spec.Apis[index].Name)
	}
	return ToModel(su), nil
}

func (s *Service) DeleteRestriction(id string, opts ...util.OpOption) error {
	err := s.Delete(id, opts...)
	if err != nil {
		return fmt.Errorf("cannot delete traffic control: %+v", err)
	}
	return nil
}

// + update
func (s *Service) UpdateRestriction(id string, reqData interface{}, opts ...util.OpOption) (*Restriction, error) {
	crd, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	if err = s.assignment(crd, reqData); err != nil {
		return nil, err
	}
	//更新crd的状态为开始更新
	crd.Status.Status = v1.Update
	for index := range crd.Spec.Apis {
		crd.Spec.Apis[index].Result = v1.UPDATING
	}
	su, err := s.Update(crd)
	if err != nil {
		return nil, fmt.Errorf("cannot update status to update: %+v", err)
	}

	return ToModel(su), nil
}

func (s *Service) Create(su *v1.Restriction) (*v1.Restriction, error) {
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
	klog.V(5).Infof("get v1.restriction of creating: %+v", su)
	return su, nil
}

func (s *Service) List(opts ...util.OpOption) (*v1.RestrictionList, error) {
	var options metav1.ListOptions
	op := util.OpList(opts...)
	u := op.User()
	ns := op.Namespace()
	var labels []string
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
	apps := &v1.RestrictionList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.RestrictionList: %+v", apps)
	return apps, nil
}

func (s *Service) Get(id string, opts ...util.OpOption) (*v1.Restriction, error) {
	var crdNamespace = defaultNamespace
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
	su := &v1.Restriction{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.restriction: %+v", su)
	return su, nil
}

func (s *Service) ForceDelete(id string, crdNamespace string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) Delete(id string, opts ...util.OpOption) error {
	var crdNamespace = defaultNamespace
	if s.tenantEnabled {
		crdNamespace = util.OpList(opts...).Namespace()
		if len(crdNamespace) == 0 {
			return fmt.Errorf("namespace not set")
		}
	}
	su, err := s.Get(id, opts...)
	if err != nil {
		return fmt.Errorf("get crd by id error: %+v", err)
	}
	klog.V(5).Infof("get v1.restriction: %+v", su)
	if len(su.Spec.Apis) != 0 {
		return fmt.Errorf("please unbind apis")
	}
	if !s.tenantEnabled {
		u := util.OpList(opts...).User()
		if !user.WritePermitted(u, su.ObjectMeta.Labels) {
			return fmt.Errorf("write permission denied")
		}
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
func (s *Service) Update(su *v1.Restriction) (*v1.Restriction, error) {
	return s.UpdateStatus(su)
}

func (s *Service) UpdateSpec(su *v1.Restriction) (*v1.Restriction, error) {
	return s.UpdateStatus(su)
}

func (s *Service) UpdateStatus(su *v1.Restriction) (*v1.Restriction, error) {
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
	su = &v1.Restriction{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.restriction: %+v", su)

	return su, nil
}

func (s *Service) getAPi(id string, crdNamespace string) (*apiv1.Api, error) {
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

func (s *Service) updateApi(apiid string, restriction *v1.Restriction, crdNamespace string) (*apiv1.Api, error) {
	api, err := s.getAPi(apiid, crdNamespace)
	if err != nil {
		return nil, fmt.Errorf("cannot update api: %+v", err)
	}

	if v1.Bind == restriction.Status.Status {
		api.Spec.Restriction.ID = restriction.ObjectMeta.Name
		api.Spec.Restriction.Name = restriction.Spec.Name
	} else {
		api.Spec.Restriction.ID = ""
		api.Spec.Restriction.Name = ""
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

func (s *Service) BindOrUnbindApis(operation, id string, apis []v1.Api, opts ...util.OpOption) (*Restriction, error) {
	if operation == "unbind" {
		return s.UnBindApi(id, apis, opts...)
	} else if operation == "bind" {
		return s.BindApi(id, apis, opts...)
	} else {
		return nil, fmt.Errorf("error operation type")
	}

}

func (s *Service) BindApi(id string, apis []v1.Api, opts ...util.OpOption) (*Restriction, error) {
	restriction, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("get restriction error: %+v", err)
	}
	crdNamespace := util.OpList(opts...).Namespace()
	if !s.tenantEnabled {
		u := util.OpList(opts...).User()
		if !user.WritePermitted(u, restriction.ObjectMeta.Labels) {
			return nil, fmt.Errorf("write permission denied")
		}
	}
	//先校验完所有API再执行操作
	for _, api := range apis {
		apiSource, err := s.getAPi(api.ID, crdNamespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get api: %+v", err)
		}
		// 1个API只能绑定一个
		if len(apiSource.Spec.Restriction.ID) > 0 {
			return nil, fmt.Errorf("api alrady bound to restriction")
		}
	}
	//update status bind
	restriction.Status.Status = v1.Bind
	for _, api := range apis {
		apiSource, err := s.getAPi(api.ID, crdNamespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get api: %+v", err)
		}
		//绑定api
		restriction.ObjectMeta.Labels[api.ID] = "true"
		isFisrtBind := true
		for index, v := range restriction.Spec.Apis {
			if v.ID == apiSource.ObjectMeta.Name {
				restriction.Spec.Apis[index].Result = v1.BINDING
				isFisrtBind = false
				break
			}
		}
		if isFisrtBind == true {
			restriction.Spec.Apis = append(restriction.Spec.Apis, v1.Api{
				ID:     api.ID,
				Name:   apiSource.Spec.Name,
				KongID: apiSource.Spec.KongApi.KongID,
				Result: v1.BINDING,
			})
		}

		fmt.Printf("kongID:%s\n", apiSource.Spec.KongApi.KongID)
		fmt.Printf("crdNamespace: %s\n", crdNamespace)

		//update api 操作时会判断是绑定还是解绑所以先将状态设置成bind
		if _, err = s.updateApi(api.ID, restriction, crdNamespace); err != nil {
			return nil, fmt.Errorf("cannot update api restriction")
		}
	}

	//update traffic 所有api绑定完成后更新数据库的状态
	restriction, err = s.UpdateSpec(restriction)
	return ToModel(restriction), err
}

func (s *Service) UnBindApi(id string, apis []v1.Api, opts ...util.OpOption) (*Restriction, error) {
	restriction, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("get restriction error: %+v", err)
	}
	crdNamespace := util.OpList(opts...).Namespace()
	if !s.tenantEnabled {
		u := util.OpList(opts...).User()
		if !user.WritePermitted(u, restriction.ObjectMeta.Labels) {
			return nil, fmt.Errorf("write permission denied")
		}
	}
	//校验API
	for _, api := range apis {
		apiSource, err := s.getAPi(api.ID, crdNamespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get api: %+v", err)
		}
		//
		if len(apiSource.Spec.Restriction.ID) == 0 {
			return nil, fmt.Errorf("api has no bound to restriction")
		}
	}
	//解除绑定
	//update status unbind
	restriction.Status.Status = v1.UnBind
	for _, api := range apis {
		apiSource, err := s.getAPi(api.ID, crdNamespace)
		if err != nil {
			return nil, fmt.Errorf("cannot get api: %+v", err)
		}
		//解除绑定
		for index, v := range restriction.Spec.Apis {
			if v.ID == apiSource.ObjectMeta.Name {
				restriction.Spec.Apis[index].Result = v1.UNBINDING
			}
		}
		////update api 操作时会判断是绑定还是解绑所以先将状态设置成unbind
		restriction.ObjectMeta.Labels[api.ID] = "false"
		if _, err = s.updateApi(api.ID, restriction, crdNamespace); err != nil {
			return nil, fmt.Errorf("cannot update api restriction")
		}
	}
	//update traffic
	restriction, err = s.UpdateSpec(restriction)
	return ToModel(restriction), err
}
