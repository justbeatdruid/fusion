package service

import (
	"fmt"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	"github.com/chinamobile/nlpt/crds/apply/api/v1"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

var crdNamespace = "default"

type Service struct {
	client            dynamic.NamespaceableResourceInterface
	apiClient         dynamic.NamespaceableResourceInterface
	serviceunitClient dynamic.NamespaceableResourceInterface
	applicationClient dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{
		client:            client.Resource(v1.GetOOFSGVR()),
		apiClient:         client.Resource(apiv1.GetOOFSGVR()),
		serviceunitClient: client.Resource(suv1.GetOOFSGVR()),
		applicationClient: client.Resource(appv1.GetOOFSGVR()),
	}
}

func (s *Service) CreateApply(model *Apply) (*Apply, error) {
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}

	var (
		targetName string
		sourceName string
	)
	var targetApi *apiv1.Api
	var applicationId string
	check := func(r *Resource) error {
		switch r.Type {
		case v1.Serviceunit:
			su, err := s.getServiceunit(r.ID)
			if err != nil {
				return fmt.Errorf("get serviceunit error: %+v", err)
			}
			r.Name = su.Spec.Name
			r.Owner = user.GetOwner(su.ObjectMeta.Labels)
			r.Labels = su.ObjectMeta.Labels
		case v1.Api:
			api, err := s.getApi(r.ID)
			if err != nil {
				return fmt.Errorf("get api error: %+v", err)
			}
			r.Name = api.Spec.Name
			r.Owner = user.GetOwner(api.ObjectMeta.Labels)
			r.Labels = api.ObjectMeta.Labels
			targetName = api.Spec.Name
			targetApi = api
		case v1.Application:
			app, err := s.getApplication(r.ID)
			if err != nil {
				return fmt.Errorf("get application error: %+v", err)
			}
			r.Name = app.Spec.Name
			r.Owner = user.GetOwner(app.ObjectMeta.Labels)
			r.Labels = app.ObjectMeta.Labels
			sourceName = app.Spec.Name
			applicationId = app.ObjectMeta.Name
		default:
			return fmt.Errorf("unknown target type: %s", r.Type)
		}
		return nil
	}
	for n, r := range map[string]*Resource{
		"target": &model.Target,
		"source": &model.Source,
	} {
		if err := check(r); err != nil {
			return nil, fmt.Errorf("%s resource error: %+v", n, err)
		}
	}
	// check if target api already bound to source application
	if targetApi == nil {
		return nil, fmt.Errorf("cannot find api")
	}
	if len(applicationId) == 0 {
		return nil, fmt.Errorf("cannot find application")
	}
	if _, ok := targetApi.ObjectMeta.Labels[apiv1.ApplicationLabel(applicationId)]; ok {
		return nil, errors.AlreadyBoundError("api already bound to application")
	}

	klog.V(5).Infof("applicant: %s", model.Users.AppliedBy.ID)
	if !user.WritePermitted(model.Users.AppliedBy.ID, model.Source.Labels) {
		return nil, errors.PermissionDeniedError("user has no permission for source %s", model.Source.Name)
	}
	model.Users.ApprovedBy.ID = model.Target.Owner
	if len(model.Users.ApprovedBy.ID) == 0 {
		return nil, fmt.Errorf("cannot get approved user")
	}
	if len(model.Users.AppliedBy.ID) == 0 {
		return nil, fmt.Errorf("cannot get applied user")
	}
	apl, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	apl.Spec.TargetName = targetName
	apl.Spec.SourceName = sourceName
	return s.ToModel(apl)
}

func (s *Service) ListApply(role string, opts ...util.OpOption) ([]*Apply, error) {
	apls, err := s.List(role, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	// this part limit that source is application and target is api
	// do some merge things
	{
		appMaps, err := s.getApplications()
		if err != nil {
			return nil, fmt.Errorf("cannot get applications: %+v", err)
		}
		apiMaps, err := s.getApis()
		if err != nil {
			return nil, fmt.Errorf("cannot get apis: %+v", err)
		}
		for i, item := range apls.Items {
			var sobj Object
			var found bool
			if item.Spec.SourceType == "api" {
				sobj, found = apiMaps[item.Spec.SourceID]
			} else if item.Spec.SourceType == "application" {
				sobj, found = appMaps[item.Spec.SourceID]
			}
			if found {
				apls.Items[i].Spec.SourceName = sobj.Name
				apls.Items[i].Spec.SourceGroup = sobj.Group
			} else {
				apls.Items[i].Spec.SourceName = "已删除"
				apls.Items[i].Spec.SourceGroup = ""
			}
			var tobj Object
			if item.Spec.TargetType == "api" {
				tobj, found = apiMaps[item.Spec.TargetID]
			} else if item.Spec.TargetType == "application" {
				tobj, found = appMaps[item.Spec.TargetID]
			}
			if found {
				apls.Items[i].Spec.TargetName = tobj.Name
				apls.Items[i].Spec.TargetGroup = tobj.Group
			} else {
				apls.Items[i].Spec.TargetName = "已删除"
				apls.Items[i].Spec.TargetGroup = ""
			}
		}
	}
	return s.ToListModel(apls)
}

func (s *Service) GetApply(id string) (*Apply, error) {
	apl, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	app, err := s.getApplication(apl.Spec.SourceID)
	if err != nil {
		return nil, fmt.Errorf("cannot get application: %+v", err)
	}
	apl.Spec.SourceName = app.Spec.Name
	api, err := s.getApi(apl.Spec.TargetID)
	if err != nil {
		return nil, fmt.Errorf("cannot get api: %+v", err)
	}
	apl.Spec.TargetName = api.Spec.Name
	return s.ToModel(apl)
}

func (s *Service) DeleteApply(id string) error {
	return s.Delete(id)
}

func (s *Service) Create(apl *v1.Apply) (*v1.Apply, error) {
	apl.Status.OperationDone = false
	apl.Status.Retry = v1.OperationRetry
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(apl)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apl); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.apply of creating: %+v", apl)
	return apl, nil
}

func (s *Service) List(role string, opts ...util.OpOption) (*v1.ApplyList, error) {
	var labels string
	operator := util.OpList(opts...).User()
	switch role {
	case "approver":
		labels = user.GetApproverLabelSelector(operator)
	case "applicant":
		labels = user.GetApplicantLabelSelector(operator)
	default:
		return nil, fmt.Errorf("wrong role: %s", role)
	}
	klog.V(5).Infof("list api with label selector: %s", labels)
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{
		LabelSelector: labels,
	})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apls := &v1.ApplyList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apls); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applyList: %+v", apls)
	return apls, nil
}

func (s *Service) Get(id string) (*v1.Apply, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	apl := &v1.Apply{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apl); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.apply: %+v", apl)
	return apl, nil
}

func (s *Service) Delete(id string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) getServiceunit(id string) (*suv1.Serviceunit, error) {
	crd, err := s.serviceunitClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	su := &suv1.Serviceunit{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), su); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.serviceunit: %+v", su)
	return su, nil
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

func (s *Service) getApi(id string) (*apiv1.Api, error) {
	crd, err := s.apiClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	api := &apiv1.Api{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), api); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.api: %+v", api)
	return api, nil
}

func (s *Service) getApplications() (map[string]Object, error) {
	crd, err := s.applicationClient.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	apps := &appv1.ApplicationList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationList: %+v", apps)
	result := make(map[string]Object)
	for _, app := range apps.Items {
		result[app.ObjectMeta.Name] = Object{
			Group: app.ObjectMeta.Labels[appv1.GroupLabel],
			Name:  app.Spec.Name,
		}
	}
	return result, nil
}

func (s *Service) getApis() (map[string]Object, error) {
	crd, err := s.apiClient.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	apis := &apiv1.ApiList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apis); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.apiList: %+v", apis)
	result := make(map[string]Object)
	for _, api := range apis.Items {
		result[api.ObjectMeta.Name] = Object{
			Name: api.Spec.Name,
		}
	}
	return result, nil
}

func (s *Service) getApiOwners(id string) (map[string]string, error) {
	crd, err := s.apiClient.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	apis := &apiv1.ApiList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apis); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.apiList: %+v", apis)
	m := make(map[string]string)
	for _, api := range apis.Items {
		m[api.ObjectMeta.Name] = user.GetOwner(api.ObjectMeta.Labels)
	}
	return m, nil
}

func (s *Service) ApproveApply(id string, admitted bool, reason string, opts ...util.OpOption) (*v1.Apply, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	apl := &v1.Apply{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apl); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.apply: %+v", apl)

	apluser := user.GetApplyUserFromLabels(apl.ObjectMeta.Labels)
	operator := util.OpList(opts...).User()
	if apluser.ApprovedBy.ID != operator {
		return nil, errors.PermissionDeniedError("user %s has no permission to approve this apply", operator)
	}

	if apl.Status.Status != v1.Waiting {
		return nil, fmt.Errorf("wrong apply status, expect %s, have %s", v1.Waiting, apl.Status.Status)
	}

	if admitted {
		apl.Status.Status = v1.Admited
	} else {
		apl.Status.Status = v1.Denied
	}
	apl.Status.Reason = reason
	apl.Status.ApprovedAt = metav1.Now()
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(apl)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd = &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apl); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.apply of creating: %+v", apl)
	return apl, nil
}
