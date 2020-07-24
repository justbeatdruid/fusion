package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/kubernetes"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/util"
	"time"

	"github.com/chinamobile/nlpt/crds/clientauth/api/v1"

	topicv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var crdNamespace = "default"

var oofsGVR = schema.GroupVersionResource{
	Group:    v1.GroupVersion.Group,
	Version:  v1.GroupVersion.Version,
	Resource: "clientauths",
}

type Service struct {
	kubeClient  *clientset.Clientset
	client      dynamic.NamespaceableResourceInterface
	topicClient dynamic.NamespaceableResourceInterface
	tokenSecret string
}

func NewService(client dynamic.Interface, topicConfig *config.TopicConfig, kubeClient *clientset.Clientset) *Service {
	return &Service{
		client:      client.Resource(oofsGVR),
		topicClient: client.Resource(topicv1.GetOOFSGVR()),
		kubeClient:  kubeClient,
		tokenSecret: topicConfig.TokenSecret,
	}
}

func (s *Service) CreateClientauth(model *Clientauth) (*Clientauth, error) {
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	//判断用户名是否已经存在
	cas, err := s.ListClientauth()
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	for _, ca := range cas {
		if ca.Name == model.Name {
			return nil, fmt.Errorf("The username already exists. ")
		}
	}
	nowTime := util.Now().Unix()
	//model.CreateTime = nowTime
	model.IssuedAt = nowTime
	//创建token
	t := &Token{
		Sub:    model.Name,
		Iat:    model.IssuedAt,
		Exp:    model.ExpireAt,
		Secret: s.tokenSecret,
	}
	model.Token, err = t.Create()
	if err != nil {
		return nil, fmt.Errorf("cannot create token: %+v", err)
	}
	ca, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(ca), nil
}

func (s *Service) ListClientauth(opts ...util.OpOption) ([]*Clientauth, error) {
	cas, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(cas), nil
}

func (s *Service) GetClientauth(id string, opts ...util.OpOption) (*Clientauth, error) {
	ca, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(ca), nil
}

func (s *Service) DeleteClientauth(id string, opts ...util.OpOption) (*Clientauth, error) {

	tps, err := s.ListTopicsWithAuthId(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot listTopicsWithAuthId: %+v", err)
	}

	if len(tps.Items) > 0 {
		return nil, fmt.Errorf("cannot delete authorized client auth user")
	}
	ca, err := s.Delete(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot update status to delete: %+v", err)
	}
	return ToModel(ca), nil
}

func (s *Service) Create(ca *v1.Clientauth) (*v1.Clientauth, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ca)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	err = kubernetes.EnsureNamespace(s.kubeClient, ca.Namespace)
	if err != nil {
		if err != nil {
			return nil, fmt.Errorf("cannot ensure k8s namespace: %+v", err)
		}
	}
	crd, err = s.client.Namespace(ca.Namespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ca); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.clientauth of creating: %+v", ca)
	return ca, nil
}

func (s *Service) List(opts ...util.OpOption) (*v1.ClientauthList, error) {
	op := util.OpList(opts...)
	crd, err := s.client.Namespace(op.Namespace()).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	cas := &v1.ClientauthList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), cas); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.clientauthList: %+v", cas)
	return cas, nil
}

func (s *Service) GetTopic(id string, opts ...util.OpOption) (*topicv1.Topic, error) {
	op := util.OpList(opts...)
	crd, err := s.topicClient.Namespace(op.Namespace()).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	tp := &topicv1.Topic{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tp); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	//klog.V(5).Infof("get v1.topic: %+v", tp)
	return tp, nil
}
func (s *Service) ListTopicsWithAuthId(id string, ops ...util.OpOption) (*topicv1.TopicList, error) {
	op := util.OpList(ops...)
	var opts metav1.ListOptions
	opts.LabelSelector = id
	crd, err := s.topicClient.Namespace(op.Namespace()).List(opts)
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}

	tps := &topicv1.TopicList{}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), tps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.topicList: %+v", tps)
	return tps, nil

}
func (s *Service) Get(id string, opts ...util.OpOption) (*v1.Clientauth, error) {
	op := util.OpList(opts...)
	crd, err := s.client.Namespace(op.Namespace()).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	ca := &v1.Clientauth{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ca); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.Clientauth: %+v", ca)
	return ca, nil
}

func (s *Service) Delete(id string, opts ...util.OpOption) (*v1.Clientauth, error) {
	ca, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("error delete crd: %+v", err)
	}
	//判断是否有授权
	if ca.Spec.AuthorizedMap != nil {
		return nil, fmt.Errorf("Existence authorization ")
	}

	ca.Status.Status = v1.Delete
	return s.UpdateStatus(ca)
}

//更新状态
func (s *Service) UpdateStatus(ca *v1.Clientauth) (*v1.Clientauth, error) {

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ca)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.client.Namespace(ca.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	ca = &v1.Clientauth{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ca); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.clientauth: %+v", ca)

	return ca, nil
}

func (s *Service) RegenerateToken(ca *Clientauth) (*Clientauth, error) {
	cad, err := s.Get(ca.ID, util.WithNamespace(ca.Namespace))
	if err != nil {
		return nil, fmt.Errorf("clientauth id is not exist, id : %+v, error : %+v", ca.ID, err)
	}
	if cad.Spec.ExipreAt > time.Now().Unix() {
		// 说明当前token仍然有效，不能重新生成token
		return nil, fmt.Errorf("token is valid, cannot regenerate")
	}

	//校验时间，token的过期时间必须大于当前时间
	if ca.ExpireAt <= time.Now().Unix() && ca.IsPermanent == false {
		return nil, fmt.Errorf("token expire time:%d must be greater than now", ca.ExpireAt)
	}
	cad.Spec.ExipreAt = ca.ExpireAt
	now := time.Now().Unix()
	t := &Token{
		Sub:    cad.Spec.Name,
		Iat:    now,
		Exp:    cad.Spec.ExipreAt,
		Secret: s.tokenSecret,
	}

	token, err := t.Create()
	if err != nil {
		return nil, fmt.Errorf("generate token error, %+v", err)
	}
	cad.Spec.IssuedAt = now
	cad.Spec.Token = token
	cad.Spec.IsPermanent = ca.IsPermanent
	cad.Status.Status = v1.Updated
	cad.Status.Message = "success"
	cad, err = s.UpdateStatus(cad)
	if err != nil {
		return nil, fmt.Errorf("update token error, %+v", err)
	}
	return ToModel(cad), nil

}
