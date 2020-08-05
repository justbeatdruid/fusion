package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/database"
	"github.com/chinamobile/nlpt/apiserver/kubernetes"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/service"
	tgerror "github.com/chinamobile/nlpt/apiserver/resources/topicgroup/error"
	topicv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
	"github.com/chinamobile/nlpt/pkg/util"
	"strconv"
	"strings"
	"time"

	"github.com/chinamobile/nlpt/crds/topicgroup/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

const (
	LabelTenant     = "nlpt.cmcc.com/pulsarTenant"
	LabelTopicgroup = "nlpt.cmcc.com/topicgroup"
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
	db          *database.DatabaseConnection
}

func (s *Service) GetClient() dynamic.NamespaceableResourceInterface {
	return s.client
}

func NewService(client dynamic.Interface, kubeClient *clientset.Clientset, db *database.DatabaseConnection) *Service {
	return &Service{client: client.Resource(oofsGVR),
		topicClient: client.Resource(topicv1.GetOOFSGVR()),
		kubeClient:  kubeClient,
		db:          db}
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
	klog.Info(">>>>>>>>ListTopicgroup in time:", time.Now())
	tgs, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	klog.Info(">>>>>>>>>ListTopicgroup out time:", time.Now())

	return s.ToListModel(tgs), nil
}
func (s *Service) ToListModel(items *v1.TopicgroupList) []*Topicgroup {
	var app = make([]*Topicgroup, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
		//app[i].TopicCount = s.GetTopicCountOfTopicgroup(app[i].Name, util.WithNamespace(app[i].Namespace))
	}
	return app
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

func (s *Service) GetTopicCountOfTopicgroup(name string, opts ...util.OpOption) int {
	options := metav1.ListOptions{}
	options.LabelSelector = fmt.Sprintf("%s=%s", LabelTopicgroup, name)
	//查询所有的topic
	op := util.OpList(opts...)
	crd, err := s.topicClient.Namespace(op.Namespace()).List(options)
	if err != nil {
		return 0
	}
	tps := &topicv1.TopicList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return 0
	}

	return len(tps.Items)
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
	options.LabelSelector = fmt.Sprintf("%s=%s", LabelTopicgroup, tg.Name)
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
func (s *Service) DeleteTopicgroup(id string, opts ...util.OpOption) (*Topicgroup, string, error) {
	tg, message, err := s.Delete(id, opts...)
	if err != nil {
		return nil, message, err
	}
	util.WaitDelete(s, tg.ObjectMeta)
	return ToModel(tg), message, nil
}

func (s *Service) ModifyTopicgroup(id string, topicgroup *Topicgroup, opts ...util.OpOption) (*Topicgroup, string, error) {
	crd, err := s.Get(id, opts...)
	if err != nil {
		return nil, "Topic分组不存在", fmt.Errorf("cannot get object: %+v", err)
	}

	if crd == nil {
		return nil, "Topic分组不存在", fmt.Errorf("cannot get object: %+v", err)
	}

	if topicgroup == nil {
		return nil, "参数policies为空", fmt.Errorf("bad request:policies is required")
	}

	if err = topicgroup.ValidateModifyBody(); err != nil {
		return nil, fmt.Sprintf("参数错误：%+v", err), err
	}
	crd.Spec.Description = topicgroup.Description
	crd.Spec.Policies = s.MergePolicies(topicgroup.Policies, crd.Spec.Policies)
	crd.Status.Status = v1.Updating
	crd.Status.Message = "updating topic group policies"
	crd, msg, err := s.UpdateStatus(crd)
	if err != nil {
		return nil, msg, fmt.Errorf("modify topicgroup failed:%+v", err)
	}

	return ToModel(crd), msg, nil

}

func (s *Service) MergePolicies(req *Policies, db *v1.Policies) *v1.Policies {
	p := ToPolicesApi(req)
	if db == nil {
		db = &v1.Policies{}
	}
	if p.SubscriptionAuthMode != nil {
		db.SubscriptionAuthMode = p.SubscriptionAuthMode
	}
	if p.EncryptionRequired != nil {
		db.EncryptionRequired = p.EncryptionRequired
	}
	if p.OffloadDeletionLagMs != nil {
		db.OffloadDeletionLagMs = p.OffloadDeletionLagMs
	}
	if p.SchemaValidationEnforced != nil {
		db.SchemaValidationEnforced = p.SchemaValidationEnforced
	}
	if p.SchemaCompatibilityStrategy != nil {
		db.SchemaCompatibilityStrategy = p.SchemaCompatibilityStrategy
	}
	if p.IsAllowAutoUpdateSchema != nil {
		db.IsAllowAutoUpdateSchema = p.IsAllowAutoUpdateSchema
	}
	if p.MaxProducersPerTopic != nil {
		db.MaxProducersPerTopic = p.MaxProducersPerTopic
	}

	if p.MaxConsumersPerTopic != nil {
		db.MaxConsumersPerTopic = p.MaxConsumersPerTopic
	}

	if p.MaxConsumersPerSubscription != nil {
		db.MaxConsumersPerSubscription = p.MaxConsumersPerSubscription
	}

	if p.DeduplicationEnabled != nil {
		db.DeduplicationEnabled = p.DeduplicationEnabled
	}
	if p.SubscriptionAuthMode != nil {
		db.SubscriptionAuthMode = p.SubscriptionAuthMode
	}
	if p.OffloadThreshold != nil {
		db.OffloadThreshold = p.OffloadThreshold
	}

	if p.CompactionThreshold != nil {
		db.CompactionThreshold = p.CompactionThreshold
	}

	if p.Persistence != nil {
		db.Persistence = p.Persistence
	}

	if p.RetentionPolicies != nil {
		db.RetentionPolicies = p.RetentionPolicies
	}
	if p.ClusterSubscribeRate != nil {
		db.ClusterSubscribeRate = p.ClusterSubscribeRate
	}
	if p.SubscriptionDispatchRate != nil {
		db.SubscriptionDispatchRate = p.SubscriptionDispatchRate
	}
	if p.TopicDispatchRate != nil {
		db.TopicDispatchRate = p.TopicDispatchRate
	}
	if p.BacklogQuota != nil {
		db.BacklogQuota = p.BacklogQuota
	}
	if p.Bundles != nil {
		db.Bundles = p.Bundles
	}
	if p.MessageTtlInSeconds != nil {
		db.MessageTtlInSeconds = p.MessageTtlInSeconds
	}

	return db
}

func (s *Service) IsTopicgroupExist(tp *v1.Topicgroup) (bool, error) {
	//判断是否已存在
	//var options metav1.ListOptions
	//options.LabelSelector = fmt.Sprintf("%s=%s", LabelTenant, tp.ObjectMeta.Namespace)

	tpList, err := s.List(util.WithNamespace(tp.Namespace), util.WithNameLike(tp.Spec.Name))
	if err != nil {
		return false, err
	}

	if len(tpList.Items) > 0 {
		return true, nil
	}
	return false, nil

}
//func (s *Service) ListWithOptions(options metav1.ListOptions, opts ...util.OpOption) (*v1.TopicgroupList, error) {
//	op := util.OpList(opts...)
//	crd, err := s.client.Namespace(op.Namespace()).List(options)
//	if err != nil {
//		return nil, fmt.Errorf("error list crd: %+v", err)
//	}
//	tps := &v1.TopicgroupList{}
//	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
//		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
//	}
//	//klog.V(5).Infof("get v1.topicgroupList: %+v", tps)
//	return tps, nil
//}
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
	err = kubernetes.EnsureNamespace(s.kubeClient, tp.Namespace)
	if err != nil {
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
	klog.V(5).Infof("get v1.topicgroup of creating: %+v, time: %+v", tp, time.Now().Unix())
	return tp, tgerror.TopicgroupError{}
}

func (s *Service) List(opts ...util.OpOption) (*v1.TopicgroupList, error) {
	if s.db.Enabled() {
		return s.ListFromDatabase(opts...)
	}
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

func (s *Service) Delete(id string, opts ...util.OpOption) (*v1.Topicgroup, string, error) {
	tg, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Sprintf("Topic分组不存在"), fmt.Errorf("资源查询失败: %+v", err)
	}

	if s.GetTopicCountOfTopicgroup(tg.Spec.Name, opts...) > 0 {
		return nil, "不能删除Topic分组，当前Topic分组不为空", fmt.Errorf("topic group is not empty")
	}
	tg.Status.Status = v1.Deleting
	tg.Status.Message = "deleting"
	return s.UpdateStatus(tg)
}

//更新状态
func (s *Service) UpdateStatus(tg *v1.Topicgroup) (*v1.Topicgroup, string, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tg)
	if err != nil {
		return nil, "数据库错误", fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.client.Namespace(tg.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, "数据库错误", fmt.Errorf("error update crd status: %+v", err)
	}
	tg = &v1.Topicgroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tg); err != nil {
		return nil, "数据库错误", fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.topicgroup: %+v", tg)

	return tg, "", nil
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
