package service

import (
	"fmt"

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
	client dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{client: client.Resource(oofsGVR)}
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

func (s *Service) DeleteTopicgroup(id string) (*Topicgroup, error) {
	tg, err := s.Delete(id)
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}
	return ToModel(tg), nil
}

//删除全部topic
func (s *Service) DeleteAllTopicgroups() ([]*Topicgroup, error) {
	tps, err := s.DeleteTopics()
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}
	return ToListModel(tps), nil
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

//将所有topic的status置为delete
func (s *Service) DeleteTopics() (*v1.TopicgroupList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	tps := &v1.TopicgroupList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	for i := range tps.Items {
		tps.Items[i].Status.Status = v1.Delete
	}
	for j := range tps.Items {
		content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&(tps.Items[j]))
		if err != nil {
			return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
		}
		crd := &unstructured.Unstructured{}
		crd.SetUnstructuredContent(content)
		klog.V(5).Infof("try to update status for crd: %+v", crd)
		crd, err = s.client.Namespace(tps.Items[j].ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("error update crd status: %+v", err)
		}
	}
	tps = &v1.TopicgroupList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}

	return tps, nil
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
