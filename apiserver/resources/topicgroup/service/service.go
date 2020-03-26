package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/service"
	topicv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"

	"github.com/chinamobile/nlpt/crds/topicgroup/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

var crdNamespace = "default"

var oofsGVR = schema.GroupVersionResource{
	Group:    v1.GroupVersion.Group,
	Version:  v1.GroupVersion.Version,
	Resource: "topicgroups",
}

type Service struct {
	client      dynamic.NamespaceableResourceInterface
	topicClient dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{client: client.Resource(oofsGVR),
		topicClient: client.Resource(topicv1.GetOOFSGVR())}
}

func (s *Service) CreateTopicgroup(model *Topicgroup) (*Topicgroup, error) {
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	tg, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(tg), nil
}

func (s *Service) ListTopicgroup() ([]*Topicgroup, error) {
	tgs, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(tgs), nil
}

func (s *Service) GetTopicgroup(id string) (*Topicgroup, error) {
	tg, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(tg), nil
}

//查询topicgroup下的所有topic
func (s *Service) GetTopics(id string) ([]*service.Topic, error) {
	//查询topicgroup，得到名字
	tg, err := s.GetTopicgroup(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	tgName := tg.Name

	//查询所有的topic
	crd, err := s.topicClient.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	topicList := &topicv1.TopicList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), topicList); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	//遍历所有的topic的topicGroup
	tps := &topicv1.TopicList{}
	for _, tp := range topicList.Items {
		if tp.Spec.TopicGroup == tgName {
			tps.Items = append(tps.Items, tp)
		}
	}
	return service.ToListModel(tps), nil
}

func (s *Service) DeleteTopicgroup(id string) (*Topicgroup, error) {
	tg, err := s.Delete(id)
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}
	return ToModel(tg), nil
}

func (s *Service) ModifyTopicgroup(id string, policies *Policies) (*Topicgroup, error) {
	crd, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}

	if crd == nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}

	if policies == nil {
		return nil, fmt.Errorf("bad request:policies is required")
	}
	crd.Spec.Policies = ToPolicesApi(policies)
	crd.Status.Status = v1.Update
	crd, err = s.UpdateStatus(crd)
	if err != nil {
		return nil, fmt.Errorf("modify topicgroup failed:%+v", err)
	}

	return ToModel(crd), nil

}

func (s *Service) Create(tp *v1.Topicgroup) (*v1.Topicgroup, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tp)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tp); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.topicgroup of creating: %+v", tp)
	return tp, nil
}

func (s *Service) List() (*v1.TopicgroupList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	tps := &v1.TopicgroupList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.topicgroupList: %+v", tps)
	return tps, nil
}

func (s *Service) Get(id string) (*v1.Topicgroup, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	tp := &v1.Topicgroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tp); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.Topicgroup: %+v", tp)
	return tp, nil
}

func (s *Service) Delete(id string) (*v1.Topicgroup, error) {
	tg, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("error delete crd: %+v", err)
	}
	tg.Status.Status = v1.Delete
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
