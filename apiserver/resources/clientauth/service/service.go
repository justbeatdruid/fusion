package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/pkg/util"
	"time"

	"github.com/chinamobile/nlpt/crds/clientauth/api/v1"

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
	Resource: "clientauths",
}

type Service struct {
	client dynamic.NamespaceableResourceInterface
}

func NewService(client dynamic.Interface) *Service {
	return &Service{client: client.Resource(oofsGVR)}
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
			return nil, fmt.Errorf("用户名已存在")
		}
	}
	nowTime := util.Now().Unix()
	//model.CreateTime = nowTime
	model.IssuedAt = nowTime
	//创建token
	t := &Token{
		Sub: model.Name,
		Iat: model.IssuedAt,
		Exp: model.ExpireAt,
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

func (s *Service) ListClientauth() ([]*Clientauth, error) {
	cas, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(cas), nil
}

func (s *Service) GetClientauth(id string) (*Clientauth, error) {
	ca, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(ca), nil
}

func (s *Service) DeleteClientauth(id string) (*Clientauth, error) {
	ca, err := s.Delete(id)
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

	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ca); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.clientauth of creating: %+v", ca)
	return ca, nil
}

func (s *Service) List() (*v1.ClientauthList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
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

func (s *Service) Get(id string) (*v1.Clientauth, error) {
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
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

func (s *Service) Delete(id string) (*v1.Clientauth, error) {
	ca, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("error delete crd: %+v", err)
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
	cad, err := s.Get(ca.ID)
	if err != nil {
		return nil, fmt.Errorf("clientauth id is not exist, id : %+v, error : %+v", ca.ID, err)
	}
	//校验时间，token的过期时间必须大于当前时间
	if ca.ExpireAt <= time.Now().Unix() {
		return nil, fmt.Errorf("token expire time:%d must be greater than now", a.ExpireAt)
	}
	cad.Spec.ExipreAt = ca.ExpireAt
	now := time.Now().Unix()
	t := &Token{
		Sub: cad.Spec.Name,
		Iat: now,
		Exp: cad.Spec.ExipreAt,
	}

	token, err := t.Create()
	if err != nil {
		return nil, fmt.Errorf("generate token error, %+v", err)
	}
	cad.Spec.IssuedAt = now
	cad.Spec.Token = token
	cad.Status.Status = v1.Updated
	cad.Status.Message = "success"
	cad, err = s.UpdateStatus(cad)
	if err != nil {
		return nil, fmt.Errorf("update token error, %+v", err)
	}
	return ToModel(cad), nil

}
