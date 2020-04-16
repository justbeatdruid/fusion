package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/kubernetes"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/service"
	tgerror "github.com/chinamobile/nlpt/apiserver/resources/topicgroup/error"
	topicv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/chinamobile/nlpt/pkg/util"
	"strconv"
	"strings"

	"github.com/chinamobile/nlpt/crds/topicgroup/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
	clientset "k8s.io/client-go/kubernetes"
)

const (
	LabelTenant = "nlpt.cmcc.com/pulsarTenant"
)

var crdNamespace = "default"

var oofsGVR = schema.GroupVersionResource{
	Group:    v1.GroupVersion.Group,
	Version:  v1.GroupVersion.Version,
	Resource: "topicgroups",
}

type Service struct {
	kubeClient  *clientset.Clientset
	client      dynamic.NamespaceableResourceInterface
	topicClient dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface, kubeClient *clientset.Clientset) *Service {
	return &Service{client: client.Resource(oofsGVR),
		topicClient: client.Resource(topicv1.GetOOFSGVR()),
	    kubeClient:  kubeClient}
}

func (s *Service) CreateTopicgroup(model *Topicgroup) (*Topicgroup, tgerror.TopicgroupError) {
	if err := model.Validate(); err != nil {
		return nil, tgerror.TopicgroupError{
			Err:       fmt.Errorf("bad request: %+v", err),
			ErrorCode: tgerror.ErrorBadRequest,
		}
	}

	tg, tgErr := s.Create(ToAPI(model))
	if tgErr.Err != nil {
		return nil, tgErr
	}

	return ToModel(tg), tgerror.TopicgroupError{}
}

func (s *Service) ListTopicgroup(opts ...util.OpOption) ([]*Topicgroup, error) {
	tgs, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(tgs), nil
}

func (s *Service) SearchTopicgroup(tgList []*Topicgroup, opts ...util.OpOption) ([]*Topicgroup, error) {
	nameLike := util.OpList(opts...).NameLike()
	topic := util.OpList(opts...).Topic()
	available := util.OpList(opts...).Available()

	var names []string
	if len(nameLike) > 0 {
		names = append(names, nameLike)
	}

	if len(topic) > 0 {
		tps, err := s.listTopics(topic)
		if err != nil {
			return nil, err
		}

		if len(tps) == 0 {
			return nil, nil
		}
		for _, tp := range tps {
			names = append(names, tp.TopicGroup)
		}
	}

	var finalSearchResult []*Topicgroup
	var finalIdList []string
	if len(names) > 0 {
		for _, name := range names {
			for _, tg := range tgList {
				if strings.Contains(tg.Name, name) {
					if !isTopicgroupIdExist(finalIdList, tg.Name) {
						finalIdList = append(finalIdList, tg.Name)
						finalSearchResult = append(finalSearchResult, tg)
					}
				}
			}
		}
	} else {
		//未输入name搜索条件
		finalSearchResult = tgList
	}

	if len(available) > 0 {
		var tgss []*Topicgroup
		a, _ := strconv.ParseBool(available)

		for _, tg := range finalSearchResult {
			if tg.Available == a {
				tgss = append(tgss, tg)
			}
		}
		finalSearchResult = tgss
	}

	return finalSearchResult, nil

}

func isTopicgroupIdExist(idList []string, id string) bool {
	for _, i := range idList {
		if i == id {
			return true
		}
	}
	return false
}

func (s *Service) GetTopicgroup(id string, opt ...util.OpOption) (*Topicgroup, error) {
	tg, err := s.Get(id, opt...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(tg), nil
}

//查询topicgroup下的所有topic
func (s *Service) GetTopics(id string, opts ...util.OpOption) ([]*service.Topic, error) {
	tps, err := s.getTopicsCrd(id, opts...)
	if err != nil {
		return nil, err
	}
	return service.ToListModel(tps), nil
}

func (s *Service) listTopics(topicName string) ([]*service.Topic, error) {
	crd, err := s.topicClient.List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	tps := &topicv1.TopicList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}

	var tpsResult []*service.Topic
	for _, tp := range tps.Items {
		var name = strings.ToLower(tp.Spec.Name)
		if strings.Contains(name, topicName) {
			tpsResult = append(tpsResult, service.ToModel(&tp))
		}
	}
	return tpsResult, nil
}
func (s *Service) getTopicsCrd(id string, opts ...util.OpOption) (*topicv1.TopicList, error) {
	op := util.OpList(opts...)
	tg, err := s.GetTopicgroup(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	if tg == nil {
		return nil, fmt.Errorf("topicgroup not exist: %+v", id)
	}
	options := metav1.ListOptions{}
	options.LabelSelector = fmt.Sprintf("topicgroup=%s", tg.Name)
	//查询所有的topic
	crd, err := s.topicClient.Namespace(op.Namespace()).List(options)
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	tps := &topicv1.TopicList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}

	return tps, nil
}
func (s *Service) DeleteTopicgroup(id string, opts ...util.OpOption) (*Topicgroup, error) {
	tg, err := s.Delete(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}
	return ToModel(tg), nil
}

func (s *Service) ModifyTopicgroup(id string, policies *Policies, opts ...util.OpOption) (*Topicgroup, error) {
	crd, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}

	if crd == nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}

	if policies == nil {
		return nil, fmt.Errorf("bad request:policies is required")
	}

	if err = policies.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	crd.Spec.Policies = s.MergePolicies(policies, crd.Spec.Policies)
	crd.Status.Status = v1.Update
	crd.Status.Message = "accepted update topic group policies"
	crd, err = s.UpdateStatus(crd)
	if err != nil {
		return nil, fmt.Errorf("modify topicgroup failed:%+v", err)
	}

	return ToModel(crd), nil

}

func (s *Service) MergePolicies(req *Policies, db v1.Policies) v1.Policies {
	if req.NumBundles != NotSet {
		db.NumBundles = req.NumBundles
	}

	if req.RetentionPolicies.RetentionTimeInMinutes != NotSet {
		db.RetentionPolicies.RetentionTimeInMinutes = req.RetentionPolicies.RetentionTimeInMinutes
	}

	if req.RetentionPolicies.RetentionSizeInMB != NotSet {
		db.RetentionPolicies.RetentionSizeInMB = req.RetentionPolicies.RetentionSizeInMB
	}

	if req.MessageTtlInSeconds != NotSet {
		db.MessageTtlInSeconds = req.MessageTtlInSeconds
	}

	if req.BacklogQuota.Limit != NotSet {
		db.BacklogQuota.Limit = req.BacklogQuota.Limit
	}

	if req.BacklogQuota.Policy != NotSetString {
		db.BacklogQuota.Policy = req.BacklogQuota.Policy
	}
	return db
}

func (s *Service) IsTopicgroupExist(tp *v1.Topicgroup) (bool, error) {
	//判断是否已存在
	var options metav1.ListOptions
	options.LabelSelector = fmt.Sprintf("%s=%s", LabelTenant, tp.ObjectMeta.Namespace)
	tpList, err := s.ListWithOptions(options, util.WithNamespace(tp.Namespace))
	if err != nil {
		return false, err
	}

	if len(tpList.Items) > 0 {
		for _, t := range tpList.Items {
			if t.Spec.Name == tp.Spec.Name {
				return true, nil
			}
		}
	}
	return false, nil

}
func (s *Service) ListWithOptions(options metav1.ListOptions, opts ...util.OpOption) (*v1.TopicgroupList, error) {
	op := util.OpList(opts...)
	crd, err := s.client.Namespace(op.Namespace()).List(options)
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	tps := &v1.TopicgroupList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	//klog.V(5).Infof("get v1.topicgroupList: %+v", tps)
	return tps, nil
}
func (s *Service) Create(tp *v1.Topicgroup) (*v1.Topicgroup, tgerror.TopicgroupError) {
	//判断是否已存在
	isExist, err := s.IsTopicgroupExist(tp)
	if err != nil {
		return nil, tgerror.TopicgroupError{
			Err:       fmt.Errorf("cannot create object: %+v", err),
			ErrorCode: tgerror.ErrorCreateTopicgroup,
		}
	}
	if isExist {
		return nil, tgerror.TopicgroupError{
			Err:       fmt.Errorf("topicgroup already exists in tenant: %v", tp.ObjectMeta.Namespace),
			ErrorCode: tgerror.ErrorDuplicatedTopicgroup,
		}
	}
	//给Topicgroup打上租户的标签
	if tp.ObjectMeta.Labels == nil {
		tp.ObjectMeta.Labels = make(map[string]string)
	}
	tp.ObjectMeta.Labels[LabelTenant] = tp.ObjectMeta.Namespace

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tp)
	if err != nil {
		return nil, tgerror.TopicgroupError{
			Err:       fmt.Errorf("convert crd to unstructured error: %+v", err),
			ErrorCode: tgerror.ErrorCreateTopicgroup,
		}
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	err = kubernetes.EnsureNamespace(s.kubeClient,tp.Namespace)
	if err!=nil{
		return nil, tgerror.TopicgroupError{
			Err:       fmt.Errorf("cannot ensure k8s namespace: %+v", err),
			ErrorCode: tgerror.ErrorEnsureNamespace,
		}
	}
	crd, err = s.client.Namespace(tp.Namespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, tgerror.TopicgroupError{
			Err:       fmt.Errorf("error creating crd: %+v", err),
			ErrorCode: tgerror.ErrorCreateTopicgroup,
		}
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tp); err != nil {
		return nil, tgerror.TopicgroupError{
			Err:       fmt.Errorf("convert unstructured to crd error: %+v", err),
			ErrorCode: tgerror.ErrorCreateTopicgroup,
		}
	}
	klog.V(5).Infof("get v1.topicgroup of creating: %+v", tp)
	return tp, tgerror.TopicgroupError{}
}

func (s *Service) List(opts ...util.OpOption) (*v1.TopicgroupList, error) {
	op := util.OpList(opts...)
	crd, err := s.client.Namespace(op.Namespace()).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	tps := &v1.TopicgroupList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	//	klog.V(5).Infof("get v1.topicgroupList: %+v", tps)
	return tps, nil
}

func (s *Service) Get(id string, ops ...util.OpOption) (*v1.Topicgroup, error) {
	op := util.OpList(ops...)
	crd, err := s.client.Namespace(op.Namespace()).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	tp := &v1.Topicgroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tp); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	//klog.V(5).Infof("get v1.Topicgroup: %+v", tp)
	return tp, nil
}

func (s *Service) Delete(id string, opts ...util.OpOption) (*v1.Topicgroup, error) {
	tg, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("error delete crd: %+v", err)
	}
	tg.Status.Status = v1.Delete
	tg.Status.Message = "accepted delete request"
	return s.UpdateStatus(tg)
}

//更新状态
func (s *Service) UpdateStatus(tg *v1.Topicgroup) (*v1.Topicgroup, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tg)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.client.Namespace(tg.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	tg = &v1.Topicgroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tg); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.topicgroup: %+v", tg)

	return tg, nil
}

func (s *Service) UpdateTopicStatus(tp *topicv1.Topic) (*topicv1.Topic, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tp)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.topicClient.Namespace(tp.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	tp = &topicv1.Topic{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tp); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.topic: %+v", tp)

	return tp, nil
}
