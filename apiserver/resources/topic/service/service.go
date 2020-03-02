package service

import (
	"context"
	"fmt"
	"github.com/apache/pulsar/pulsar-client-go/pulsar"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/crds/topic/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
	"strconv"
)

var crdNamespace = "default"

var oofsGVR = schema.GroupVersionResource{
	Group:    v1.GroupVersion.Group,
	Version:  v1.GroupVersion.Version,
	Resource: "topics",
}

type Service struct {
	client dynamic.NamespaceableResourceInterface
	ip     string
	host   int
}

func NewService(client dynamic.Interface, topConfig *config.TopicConfig) *Service {
	return &Service{client: client.Resource(oofsGVR),
		ip:   topConfig.Host,
		host: topConfig.Port,
	}
}
func (s *Service) CreateTopic(model *Topic) (*Topic, error) {
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	tp, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(tp), nil
}

func (s *Service) ListTopic() ([]*Topic, error) {
	tps, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(tps), nil
}

func (s *Service) GetTopic(id string) (*Topic, error) {
	tp, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(tp), nil
}

func (s *Service) DeleteTopic(id string) (*Topic, error) {
	tp, err := s.Delete(id)
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}
	return ToModel(tp), nil
}

//删除全部topic
func (s *Service) DeleteAllTopics() ([]*Topic, error) {
	tps, err := s.DeleteTopics()
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}
	return ToListModel(tps), nil
}

//带时间查询
func (s *Service) ListMessagesTime(topicUrls []string, start int64, end int64) ([]Message, error) {
	messages, err := s.ListTopicMessagesTime(topicUrls, start, end)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return messages, nil
}

//不带时间查询
func (s *Service) ListMessages(topicUrls []string) ([]Message, error) {
	messages, err := s.ListTopicMessages(topicUrls)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return messages, nil
}

func (s *Service) Create(tp *v1.Topic) (*v1.Topic, error) {
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
	klog.V(5).Infof("get v1.topic of creating: %+v", tp)
	return tp, nil
}

func (s *Service) List() (*v1.TopicList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	tps := &v1.TopicList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	//klog.V(5).Infof("get v1.topicList: %+v", tps)
	return tps, nil
}

func (s *Service) Get(id string) (*v1.Topic, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	tp := &v1.Topic{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tp); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.topic: %+v", tp)
	return tp, nil
}

func (s *Service) Delete(id string) (*v1.Topic, error) {
	tp, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("error delete crd: %+v", err)
	}
	tp.Status.Status = v1.Delete
	return s.UpdateStatus(tp)
}

//将所有topic的status置为delete
func (s *Service) DeleteTopics() (*v1.TopicList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	tps := &v1.TopicList{}
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
	tps = &v1.TopicList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}

	return tps, nil
}

//更新状态
func (s *Service) UpdateStatus(tp *v1.Topic) (*v1.Topic, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tp)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.client.Namespace(tp.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	tp = &v1.Topic{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tp); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.topic: %+v", tp)

	return tp, nil
}

//带时间查询topic中的所有消息
func (s *Service) ListTopicMessagesTime(topicUrls []string, start int64, end int64) ([]Message, error) {
	// Instantiate a Pulsar client
	ip := s.ip
	host := s.host
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://" + ip + ":" + strconv.Itoa(host),
	})
	if err != nil {
		fmt.Printf("Could not create client: %v", err)
	}
	defer client.Close()
	var messageStructs []Message
	var messageStruct Message
	var timeStamp int64
	for _, topicUrl := range topicUrls {
		reader, err := client.CreateReader(pulsar.ReaderOptions{
			Topic:          topicUrl,
			StartMessageID: pulsar.EarliestMessage,
		})
		if err != nil {
			fmt.Println("create reader error")
			continue
		}
		ctx := context.Background()
		for {
			if flag, _ := reader.HasNext(); flag == false {
				break
			}
			msg, err := reader.Next(ctx)
			if err != nil {
				fmt.Printf("Error reading from topic: %v", err)
			}
			// Process the message
			messageStruct.TopicName = msg.Topic()
			timeStamp = messageStruct.Time.Unix()
			if timeStamp >= start && timeStamp <= end {
				messageStruct.ID = msg.ID()
				messageStruct.Messages = string(msg.Payload()[:])
				messageStructs = append(messageStructs, messageStruct)
			}
		}
		reader.Close()
	}
	return messageStructs, err
}

//不带时间查询多个topic中的所有消息
func (s *Service) ListTopicMessages(topicUrls []string) ([]Message, error) {
	// Instantiate a Pulsar client
	ip := s.ip
	host := s.host
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://" + ip + ":" + strconv.Itoa(host),
	})
	if err != nil {
		fmt.Printf("Could not create client: %v", err)
	}
	defer client.Close()
	var messageStructs []Message
	var messageStruct Message
	for _, topicUrl := range topicUrls {
		fmt.Println(topicUrl)
		reader, err := client.CreateReader(pulsar.ReaderOptions{
			Topic:          topicUrl,
			StartMessageID: pulsar.EarliestMessage,
		})
		if err != nil {
			fmt.Println("create reader error")
			continue
		}
		ctx := context.Background()
		for {
			if flag, _ := reader.HasNext(); flag == false {
				break
			}
			msg, err := reader.Next(ctx)
			if err != nil {
				fmt.Printf("Error reading from topic: %v", err)
			}
			// Process the message
			messageStruct.TopicName = msg.Topic()
			messageStruct.Time = msg.PublishTime()
			messageStruct.ID = msg.ID()
			messageStruct.Messages = string(msg.Payload()[:])
			messageStructs = append(messageStructs, messageStruct)
		}
		reader.Close()
	}

	return messageStructs, err
}
