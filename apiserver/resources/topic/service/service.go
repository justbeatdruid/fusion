package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/parnurzeal/gorequest"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/kubernetes"
	tperror "github.com/chinamobile/nlpt/apiserver/resources/topic/error"
	"github.com/chinamobile/nlpt/apiserver/resources/topic/pulsargo"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	clientauthv1 "github.com/chinamobile/nlpt/crds/clientauth/api/v1"
	"github.com/chinamobile/nlpt/crds/topic/api/v1"
	topicgroupv1 "github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
	"strconv"
)

var crdNamespace = "default"

const Label = "nlpt.cmcc.com/topicgroup"

var oofsGVR = schema.GroupVersionResource{
	Group:    v1.GroupVersion.Group,
	Version:  v1.GroupVersion.Version,
	Resource: "topics",
}

type Service struct {
	kubeClient        *clientset.Clientset
	client            dynamic.NamespaceableResourceInterface
	clientAuthClient  dynamic.NamespaceableResourceInterface
	topicGroupClient  dynamic.NamespaceableResourceInterface
	applicationClient dynamic.NamespaceableResourceInterface
	errMsg            config.ErrorConfig
	ip                string
	port              int
	authEnable        bool
	adminToken        string
}

const (
	ResetPositionByIdUrl   = "/admin/v2/persistent/%s/%s/%s/subscription/%s/resetcursor"
	ResetPositionByTimeUrl = "/admin/v2/persistent/%s/%s/%s/subscription/%s/resetcursor/%d"
)

type requestLogger struct {
	prefix string
}

var logger = &requestLogger{}

func (r *requestLogger) SetPrefix(prefix string) {
	r.prefix = prefix
}

func (r *requestLogger) Printf(format string, v ...interface{}) {
	klog.V(4).Infof(format, v...)
}

func (r *requestLogger) Println(v ...interface{}) {
	klog.V(4).Infof("%+v", v)
}

func (s *Service) GetClient() dynamic.NamespaceableResourceInterface {
	return s.client
}

func NewService(client dynamic.Interface, kubeClient *clientset.Clientset, topConfig *config.TopicConfig, errMsg config.ErrorConfig) *Service {
	return &Service{client: client.Resource(oofsGVR),
		clientAuthClient:  client.Resource(clientauthv1.GetOOFSGVR()),
		topicGroupClient:  client.Resource(topicgroupv1.GetOOFSGVR()),
		applicationClient: client.Resource(appv1.GetOOFSGVR()),
		errMsg:            errMsg,
		ip:                topConfig.Host,
		port:              topConfig.Port,
		authEnable:        topConfig.AuthEnable,
		adminToken:        topConfig.AdminToken,
		kubeClient:        kubeClient,
	}
}

func (s *Service) CreateTopic(model *Topic) (*Topic, tperror.TopicError) {
	if tpErr := model.Validate(); tpErr.Err != nil {
		return nil, tpErr
	}
	tp, tpErr := s.Create(ToAPI(model))
	if tpErr.Err != nil {
		return nil, tpErr
	}
	return ToModel(tp), tperror.TopicError{}
}

func (s *Service) ListTopic(opts ...util.OpOption) ([]*Topic, error) {
	tps, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(tps), nil
}

func (s *Service) GetTopic(id string, opts ...util.OpOption) (*Topic, error) {
	tp, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}

	return ToModel(tp), nil
}

func (s *Service) DeleteTopic(id string, opts ...util.OpOption) (*Topic, error) {
	tp, err := s.Delete(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}

	util.WaitDelete(s, tp.ObjectMeta)
	return ToModel(tp), nil
}

func (s *Service) DeletePermissions(id string, authUserId string, opts ...util.OpOption) (*Topic, error) {
	tp, err := s.DeletePer(id, authUserId, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete permission: %+v", err)
	}
	return ToModel(tp), nil
}

//带时间查询
func (s *Service) ListMessagesTime(topicUrls []string, start int64, end int64) ([]Message, error) {
	messages, err := s.ListTopicMessagesTime(topicUrls, start, end)
	if err != nil {
		return nil, fmt.Errorf("cannot get TopicMessages: %+v", err)
	}
	return messages, nil
}

//不带时间查询
func (s *Service) ListMessages(topicUrls []string) ([]Message, error) {
	topicUrls = make([]string, 0)
	messages, err := s.ListTopicMessages(topicUrls)
	if err != nil {
		return nil, fmt.Errorf("cannot get TopicMessages: %+v", err)
	}
	return messages, nil
}

func (s *Service) IsTopicExist(tp *v1.Topic) bool {
	return s.IsTopicUrlExist(tp.GetUrl(), util.WithNamespace(tp.Namespace))
}

func (s *Service) IsTopicUrlExist(url string, opts ...util.OpOption) bool {
	tps, err := s.List(opts...)
	if err != nil {
		return false
	}

	//判重复
	if len(tps.Items) > 0 {
		for _, t := range tps.Items {
			if t.GetUrl() == url {
				return true
			}
		}
	}
	return false
}
func (s *Service) Create(tp *v1.Topic) (*v1.Topic, tperror.TopicError) {
	//get topicgroup name from id
	tg, err := s.GetTopicgroup(tp.Spec.TopicGroup, tp.Namespace)
	if err != nil {
		return nil, tperror.TopicError{
			Err:       fmt.Errorf("topic exists: %+v", tp.GetUrl()),
			ErrorCode: tperror.ErrorCannotFindTopicgroup,
			Message:   fmt.Sprintf(s.errMsg.Topic[tperror.ErrorCannotFindTopicgroup], tp.GetUrl()),
		}
	}
	tp.Spec.TopicGroup = tg.Spec.Name

	if s.IsTopicExist(tp) {
		return nil, tperror.TopicError{
			Err:       fmt.Errorf("topic exists: %+v", tp.GetUrl()),
			ErrorCode: tperror.ErrorTopicExists,
			Message:   fmt.Sprintf(s.errMsg.Topic[tperror.ErrorTopicExists], tp.GetUrl()),
		}
	}

	//添加标签
	if tp.ObjectMeta.Labels == nil {
		tp.ObjectMeta.Labels = make(map[string]string)
	}
	tp.ObjectMeta.Labels[Label] = tp.Spec.TopicGroup

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tp)
	if err != nil {
		return nil, tperror.TopicError{
			Err:       fmt.Errorf("convert crd to unstructured error: %+v", err),
			ErrorCode: tperror.ErrorCreateTopic,
		}
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	err = kubernetes.EnsureNamespace(s.kubeClient, tp.Namespace)
	if err != nil {
		return nil, tperror.TopicError{
			Err:       fmt.Errorf("cannot ensure k8s namespace: %+v", err),
			ErrorCode: tperror.ErrorEnsureNamespace,
			Message:   "",
		}
	}
	crd, err = s.client.Namespace(tp.Namespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, tperror.TopicError{
			Err:       fmt.Errorf("error creating crd: %+v", err),
			ErrorCode: tperror.ErrorCreateTopic,
		}
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tp); err != nil {
		return nil, tperror.TopicError{
			Err:       fmt.Errorf("convert unstructured to crd error: %+v", err),
			ErrorCode: tperror.ErrorCreateTopic,
		}
	}
	klog.V(5).Infof("get v1.topic of creating: %+v", tp)
	return tp, tperror.TopicError{}
}

func (s *Service) List(opts ...util.OpOption) (*v1.TopicList, error) {
	op := util.OpList(opts...)
	crd, err := s.client.Namespace(op.Namespace()).List(metav1.ListOptions{})
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

func (s *Service) ListByLabel(key string, value string) (*v1.TopicList, error) {
	var options metav1.ListOptions
	options.LabelSelector = fmt.Sprintf("%s=%s", key, value)
	return s.ListByOptions(options)
}

func (s *Service) ListByOptions(options metav1.ListOptions) (*v1.TopicList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(options)
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	tps := &v1.TopicList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.topicList: %+v", tps)
	return tps, nil
}

func (s *Service) Get(id string, opts ...util.OpOption) (*v1.Topic, error) {
	op := util.OpList(opts...)
	crd, err := s.client.Namespace(op.Namespace()).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	tp := &v1.Topic{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tp); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	//klog.V(5).Infof("get v1.topic: %+v", tp)
	return tp, nil
}

func (s *Service) Delete(id string, opts ...util.OpOption) (*v1.Topic, error) {
	tp, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("error delete crd: %+v", err)
	}
	tp.Status.Status = v1.Deleting
	return s.UpdateStatus(tp)
}

func (s *Service) DeletePer(id string, authUserId string, opts ...util.OpOption) (*v1.Topic, error) {
	tp, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("error delete crd: %+v", err)
	}
	//查询授权用户id的标签
	v, ok := tp.ObjectMeta.Labels[authUserId]
	if ok {
		if v == "true" {
			tp.ObjectMeta.Labels[authUserId] = "false"
		} else {
			//TODO 如果value非true，则认为该权限已经在回收中，禁止重复操作
			return nil, fmt.Errorf("revoke permission error, permission has already revoking")
		}
		delete(tp.ObjectMeta.Labels, authUserId)
	}

	for index := range tp.Spec.Permissions {
		if tp.Spec.Permissions[index].AuthUserID == authUserId {
			tp.Spec.Permissions[index].Status.Status = v1.DeletingAuthorization
			break
		}
	}
	tp.Status.AuthorizationStatus = v1.DeletingAuthorization
	return s.UpdateStatus(tp)
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
	err = kubernetes.EnsureNamespace(s.kubeClient, tp.Namespace)
	if err != nil {
		return nil, fmt.Errorf("cannot ensure k8s namespace: %+v", err)
	}
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
	client, err := s.GetPulsarClient()
	if err != nil {
		return nil, err
	}
	defer client.Close()
	var messageStructs []Message
	var messageStruct Message
	var timeStamp int64
	for _, topicUrl := range topicUrls {
		reader, err := client.CreateReader(pulsar.ReaderOptions{
			Topic:          topicUrl,
			StartMessageID: pulsar.EarliestMessageID(),
		})

		if err != nil {
			return nil, fmt.Errorf("create reader error: %+v", err)
		}
		ctx := context.Background()
		for reader.HasNext() {
			msg, err := reader.Next(ctx)
			if err != nil {
				return nil, fmt.Errorf("Error reading from topic: %+v ", err)
			}
			// Process the message
			messageStruct.TopicName = msg.Topic()
			timeStamp = messageStruct.Time.Unix()
			if timeStamp >= start && timeStamp <= end {
				messageStruct.ID = msg.ID()
				messageStruct.Messages = string(msg.Payload()[:])
				messageStruct.Size = len(msg.Payload())
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
	client, err := s.GetPulsarClient()
	if err != nil {
		fmt.Printf("Could not create client: %v", err)
		return nil, err
	}
	defer client.Close()
	var messageStructs []Message
	var messageStruct Message
	for _, topicUrl := range topicUrls {
		fmt.Println(topicUrl)
		reader, err := client.CreateReader(pulsar.ReaderOptions{
			Topic:          topicUrl,
			StartMessageID: pulsar.EarliestMessageID(),
		})
		if err != nil {
			fmt.Printf("create reader error: %+v", err)
			continue
		}
		for reader.HasNext() {
			ctx := context.Background()
			msg, err := reader.Next(ctx)
			if err != nil {
				fmt.Printf("Error reading from topic: %v", err)
			}

			// Process the message
			messageStruct.TopicName = msg.Topic()
			messageStruct.Time = util.NewTime(msg.PublishTime())
			messageStruct.ID, _ = pulsar.DeserializeMessageID(msg.ID().Serialize())
			messageStruct.Messages = string(msg.Payload()[:])
			messageStruct.Size = len(msg.Payload())
			messageStructs = append(messageStructs, messageStruct)

		}
		reader.Close()
	}

	return messageStructs, err
}

func (s *Service) GetPulsarClient() (pulsar.Client, error) {
	if s.authEnable {
		return pulsar.NewClient(pulsar.ClientOptions{
			URL:            "pulsar://" + s.ip + ":" + strconv.Itoa(s.port),
			Authentication: pulsar.NewAuthenticationToken(s.adminToken),
		})
	}
	return pulsar.NewClient(pulsar.ClientOptions{
		URL: "pulsar://" + s.ip + ":" + strconv.Itoa(s.port),
	})
}

//func (s *Service) IsTopicExist(url string) bool {
//	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{
//		FieldSelector: url,
//	})
//	if err != nil {
//		return false
//	}
//	tps := &v1.TopicList{}
//	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
//		return false
//	}
//
//	if len(tps.Items) == 0 {
//		return false
//	} else {
//		return true
//	}
//
//}

func (s *Service) GrantPermissions(topicId string, authUserId string, actions v1.Actions, opts ...util.OpOption) (*Topic, error) {
	//1.根据id查询topic
	tp, err := s.Get(topicId, opts...)
	if err != nil {
		klog.Errorf("error get crd: %+v", err)
		return nil, fmt.Errorf("error get crd: %+v", err)
	}

	//2.处理actions
	as := v1.Actions{}
	for _, ac := range actions {
		as = append(as, ac)
	}
	//3.给topic加auth id的标签，key：auth id
	if tp.ObjectMeta.Labels == nil {
		tp.ObjectMeta.Labels = make(map[string]string)
	}

	if _, ok := tp.ObjectMeta.Labels[authUserId]; ok {
		return nil, fmt.Errorf("already grant this user permissions:%+v", authUserId)
	}
	tp.ObjectMeta.Labels[authUserId] = "true"

	//4.根据auth id查询name
	authUserName, err := s.QueryAuthUserNameById(authUserId, opts...)
	if err != nil {
		return nil, fmt.Errorf("grant permission error:%+v", err)
	}

	perm := v1.Permission{
		AuthUserID:   authUserId,
		AuthUserName: authUserName,
		Actions:      as,
		Status: v1.PermissionStatus{
			Status:  v1.Authorizing,
			Message: "",
		},
	}
	tp.Spec.Permissions = append(tp.Spec.Permissions, perm)
	tp.Status.AuthorizationStatus = v1.Authorizing
	v1Tp, err := s.UpdateStatus(tp)
	if err != nil {
		return nil, fmt.Errorf("cannot update object: %+v", err)
	}

	return ToModel(v1Tp), nil
}

func (s *Service) ModifyPermissions(topicId string, authUserId string, actions v1.Actions, opts ...util.OpOption) (*Topic, error) {
	//1.根据id查询topic
	tp, err := s.Get(topicId, opts...)
	if err != nil {
		klog.Errorf("error get crd: %+v", err)
		return nil, fmt.Errorf("error get crd: %+v", err)
	}

	//2.处理actions
	as := v1.Actions{}
	for _, ac := range actions {
		as = append(as, ac)
	}
	//3.给topic加auth id的标签，key：auth id
	if tp.ObjectMeta.Labels == nil {
		return nil, fmt.Errorf("not authorization:%+v", authUserId)
	}

	if _, ok := tp.ObjectMeta.Labels[authUserId]; !ok {
		return nil, fmt.Errorf("not authorization:%+v", authUserId)
	}

	//4.根据auth id查询name
	_, err = s.QueryAuthUserNameById(authUserId, opts...)
	if err != nil {
		return nil, fmt.Errorf("grant permission error:%+v", err)
	}

	for i, p := range tp.Spec.Permissions {
		if p.AuthUserID == authUserId {
			p.Status.Status = v1.UpdatingAuthorization
			p.Actions = actions
			tp.Spec.Permissions[i] = p
			break
		}

	}

	tp.Status.AuthorizationStatus = v1.UpdatingAuthorization

	v1Tp, err := s.UpdateStatus(tp)
	if err != nil {
		return nil, fmt.Errorf("cannot update object: %+v", err)
	}

	return ToModel(v1Tp), nil
}

func (s *Service) QueryAuthUserNameById(id string, opts ...util.OpOption) (string, error) {
	ca, err := s.QueryAuthUserById(id, opts...)
	if err != nil {
		return "", err
	}
	return ca.Spec.Name, nil

}

func (s *Service) QueryAuthUserById(id string, opts ...util.OpOption) (*clientauthv1.Clientauth, error) {
	op := util.OpList(opts...)
	crd, err := s.clientAuthClient.Namespace(op.Namespace()).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	ca := &clientauthv1.Clientauth{}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ca); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}

	//klog.V(5).Infof("get auth user name: %+v", ca)
	return ca, nil

}
func (s *Service) GetTopicgroup(id string, namespace string) (*topicgroupv1.Topicgroup, error) {
	crd, err := s.topicGroupClient.Namespace(namespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	tg := &topicgroupv1.Topicgroup{}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tg); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}

	return tg, nil

}

func (s *Service) AddPartitionsOfTopic(id string, partitionNum int, ops ...util.OpOption) (*Topic, error) {
	tp, err := s.Get(id, ops...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	if tp.Spec.PartitionNum > partitionNum {
		return nil, errors.New("The number of partitions must be larger than the original ")
	}
	if tp.Spec.Partitioned == false {
		return nil, errors.New("Topic is not partitioned topic ")
	}
	tp.Status.Status = v1.Updating
	tp.Spec.OldPartitionNum = tp.Spec.PartitionNum
	tp.Spec.PartitionNum = partitionNum
	v1tp, err := s.UpdateStatus(tp)
	if err != nil {
		return nil, fmt.Errorf("cannot update object: %+v", err)
	}
	return ToModel(v1tp), nil
}

func (s *Service) getApplication(id, crdNamespace string) (*appv1.Application, error) {
	crd, err := s.applicationClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	app := &appv1.Application{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	//klog.V(5).Infof("get v1.application: %+v", app)
	return app, nil
}

func (s *Service) BatchBindOrRelease(appid, operation string, topics []BindInfo, opts ...util.OpOption) error {
	switch operation {
	case "bind":
		return s.BatchBindApi(appid, topics, opts...)
	case "release":
		return s.BatchReleaseApi(appid, topics, opts...)
	default:
		return fmt.Errorf("unknown operation %s, only bind or release can be accepted", operation)
	}
}

func (s *Service) BatchBindApi(appid string, topics []BindInfo, opts ...util.OpOption) error {
	op := util.OpList(opts...)

	_, err := s.getApplication(appid, op.Namespace())
	if err != nil {
		return fmt.Errorf("get application error: %+v", err)
	}

	for _, t := range topics {
		tp, err := s.Get(t.ID, opts...)
		if err != nil {
			return fmt.Errorf("query topic by id error: %+v", err)
		}
		for _, boundAppID := range tp.Spec.Applications {
			if boundAppID.ID == appid {
				return fmt.Errorf("application has already bound to topic")
			}

		}
		application := v1.Application{
			ID:      appid,
			Status:  v1.Binding,
			Actions: t.Actions,
		}
		tp.Spec.Applications = append(tp.Spec.Applications, application)
		tp.Status.BindStatus = v1.BindingOrUnBinding
		if tp, err = s.UpdateStatus(tp); err != nil {
			return fmt.Errorf("application bind topic failed, error: %+v", err)
		}
	}
	return nil
}

func (s *Service) BatchReleaseApi(appid string, topics []BindInfo, opts ...util.OpOption) error {
	op := util.OpList(opts...)

	app, err := s.getApplication(appid, op.Namespace())
	if err != nil {
		return fmt.Errorf("get application(%+v) error: %+v", appid, err)
	}

	for _, t := range topics {
		tp, err := s.Get(t.ID, opts...)
		if err != nil {
			return fmt.Errorf("query topic by id(%+v) error: %+v", t.ID, err)
		}

		var topicIsBound = false
		for _, boundAppID := range tp.Spec.Applications {
			if boundAppID.ID == appid {
				topicIsBound = true
				break
			}
		}

		if !topicIsBound {
			return fmt.Errorf("topic(%+v) does not bind the application(%+v)", tp.Spec.Name, app.Spec.Name)
		}

		application := v1.Application{
			ID:     appid,
			Status: v1.Unbinding,
		}
		tp.Spec.Applications = append(tp.Spec.Applications, application)
		tp.Status.BindStatus = v1.BindingOrUnBinding
		if tp, err = s.UpdateStatus(tp); err != nil {
			return fmt.Errorf("application unbind topic failed, error: %+v", err)
		}
	}
	return nil
}

func (s *Service) GetSubscriptionsOfTopic(topic *Topic) *SubscriptionsInfo {
	return topic.ToSubscriptionsModel()
}
func (s *Service) SendMessages(topicUrl string, messagesBody string, key string) (pulsar.MessageID, error) {
	client, err := s.GetPulsarClient()
	if err != nil {
		return nil, fmt.Errorf("create pulsar client error: %s", err)
	}
	messageId, err := pulsargo.SendMessages(client, topicUrl, messagesBody, key)
	if err != nil {
		return nil, fmt.Errorf("SendMessages error %s", err)
	}
	return messageId, nil
}

func (s *Service) ResetPositionById(RP *ResetPosition, tp *Topic) error {
	body := make(map[string]interface{})
	body["ledgerId"] = RP.LedgerId
	body["entryId"] = RP.EntryId
	body["partitionIndex"] = RP.PartitionIndex
	topicGroup := tp.TopicGroup
	tenant := tp.Namespace
	topic := tp.Name
	subName := RP.SubName
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)
	RPUrl := fmt.Sprintf(ResetPositionByIdUrl, tenant, topicGroup, topic, subName)
	url := fmt.Sprintf("%s://%s:%d%s", "http", s.ip, s.port, RPUrl)
	request.Data = body
	if s.authEnable {
		request.Header.Set("Authorization", "Bearer "+s.adminToken)
	}
	response, b, err := request.Post(url).End()
	if err != nil {
		return fmt.Errorf("reset position error: %+v", err)
	}
	if response.StatusCode == http.StatusNoContent {
		return nil
	} else {
		errMsg := fmt.Errorf("reset position error, url: %s, Error code: %d, Error Message: %+v", url, response.StatusCode, b)
		return errMsg
	}
}

func (s *Service) ResetPositionByTime(RP *ResetPosition, tp *Topic) error {
	timestamp := RP.Timestamp
	topicGroup := tp.TopicGroup
	tenant := tp.Namespace
	topic := tp.Name
	subName := RP.SubName
	request := gorequest.New().SetLogger(logger).SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)
	RPUrl := fmt.Sprintf(ResetPositionByTimeUrl, tenant, topicGroup, topic, subName, timestamp)
	url := fmt.Sprintf("%s://%s:%d%s", "http", s.ip, s.port, RPUrl)
	if s.authEnable {
		request.Header.Set("Authorization", "Bearer "+s.adminToken)
	}
	response, b, err := request.Post(url).End()
	if err != nil {
		return fmt.Errorf("reset position error: %+v", err)
	}
	if response.StatusCode == http.StatusNoContent {
		return nil
	} else {
		errMsg := fmt.Errorf("reset position error, url: %s, Error code: %d, Error Message: %+v", url, response.StatusCode, b)
		return errMsg
	}
}
