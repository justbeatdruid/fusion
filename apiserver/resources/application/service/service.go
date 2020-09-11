package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/resources/application/mqservice"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/parnurzeal/gorequest"
	"strings"

	"github.com/chinamobile/nlpt/apiserver/database"
	k8s "github.com/chinamobile/nlpt/apiserver/kubernetes"
	"github.com/chinamobile/nlpt/crds/application/api/v1"
	groupv1 "github.com/chinamobile/nlpt/crds/applicationgroup/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var defaultNamespace = "default"

type Service struct {
	kubeClient  *clientset.Clientset
	client      dynamic.NamespaceableResourceInterface
	groupClient dynamic.NamespaceableResourceInterface

	tenantEnabled bool

	db    *database.DatabaseConnection
	token *mqservice.Token
}

func NewService(client dynamic.Interface, kubeClient *clientset.Clientset, tenantEnabled bool, db *database.DatabaseConnection, topicConfig *config.TopicConfig) *Service {
	return &Service{
		kubeClient:    kubeClient,
		client:        client.Resource(v1.GetOOFSGVR()),
		groupClient:   client.Resource(groupv1.GetOOFSGVR()),
		tenantEnabled: tenantEnabled,
		db:            db,
		token:         mqservice.NewToken(topicConfig.TokenSecret),
	}
}

func (s *Service) GetClient() dynamic.NamespaceableResourceInterface {
	return s.client
}

func (s *Service) CreateApplication(model *Application) (*Application, error, string) {
	if err := s.Validate(model); err != nil {
		return nil, err, "002000019"
	}
	app, err := s.Create(ToAPI(model))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err), "002000020"
	}
	return ToModel(app), nil, "0"
}

func (s *Service) ListApplication(opts ...util.OpOption) ([]*Application, error) {
	apps, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	groupMap, err := s.GetGroupMap(util.OpList(opts...).Namespace())
	if err != nil {
		return nil, fmt.Errorf("get groups error: %+v", err)
	}
	return ToListModel(apps, groupMap, opts...), nil
}

func (s *Service) ListApplicationByRelation(resourceType, resourceId string, opts ...util.OpOption) ([]*Application, error) {
	if !s.db.Enabled() {
		return nil, fmt.Errorf("not support if database disabled")
	}
	return s.ListByRelationFromDatabase(resourceType, resourceId, opts...)
}

func (s *Service) GetApplication(id string, opts ...util.OpOption) (*Application, error) {
	app, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(app, opts...), nil
}

func (s *Service) DeleteApplication(id string, opts ...util.OpOption) (*Application, error) {
	app, err := s.Delete(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot delete object: %+v", err)
	}
	util.WaitDelete(s, app.ObjectMeta)
	return ToModel(app), err
}

func (s *Service) Create(app *v1.Application) (*v1.Application, error) {
	(*app).Spec.Result = v1.CREATING
	var crdNamespace = defaultNamespace
	if s.tenantEnabled {
		crdNamespace = app.ObjectMeta.Namespace
		if len(crdNamespace) == 0 {
			return nil, fmt.Errorf("namespace not set")
		}
	} else {
		app.ObjectMeta.Namespace = defaultNamespace
	}
	token, errs := s.token.Create(app.ObjectMeta.Name)
	if errs != nil {
		return nil, fmt.Errorf("cannot create topic token: %+v", errs)
	}

	if len(token) == 0 {
		return nil, fmt.Errorf("get topic token len is null")
	}

	(*app).Spec.TopicAuth.Token = token
	if err := k8s.EnsureNamespace(s.kubeClient, crdNamespace); err != nil {
		return nil, fmt.Errorf("cannot ensure k8s namespace: %+v", err)
	}
	if group, ok := app.ObjectMeta.Labels[v1.GroupLabel]; !ok {
		//return nil, fmt.Errorf("group not found")
	} else {
		if _, err := s.GetGroup(group, crdNamespace); err != nil {
			return nil, fmt.Errorf("get group error: %+v", err)
		}
	}

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(app)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.application of creating: %+v", app)

	return app, nil
}

func (s *Service) PatchApplication(id string, data interface{}, opts ...util.OpOption) (*Application, error) {
	app, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	if err = s.assignment(app, data); err != nil {
		return nil, err
	}
	crdNamespace := app.ObjectMeta.Namespace

	if !s.tenantEnabled {
		u := util.OpList(opts...).User()
		if !user.WritePermitted(u, app.ObjectMeta.Labels) {
			return nil, fmt.Errorf("write permission denied")
		}
	}

	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(app)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(crdNamespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		(*app).Spec.Result = v1.UPDATEFAILED
		return nil, fmt.Errorf("error update crd: %+v", err)
	}
	(*app).Spec.Result = v1.UPDATESUCCESS
	return ToModel(app), err
}

func (s *Service) List(opts ...util.OpOption) (*v1.ApplicationList, error) {
	if s.db.Enabled() {
		return s.ListFromDatabase(opts...)
	}
	var options metav1.ListOptions
	op := util.OpList(opts...)
	group := op.Group()
	ns := op.Namespace()
	u := op.User()
	var labels []string
	if len(group) > 0 {
		labels = append(labels, fmt.Sprintf("%s=%s", v1.GroupLabel, group))
	}
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
	apps := &v1.ApplicationList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationList: %+v", apps)
	return apps, nil
}

func (s *Service) Get(id string, opts ...util.OpOption) (*v1.Application, error) {
	crdNamespace := defaultNamespace
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
	app := &v1.Application{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.application: %+v", app)
	if app.ObjectMeta.Labels != nil {
		if gid, ok := app.ObjectMeta.Labels[v1.GroupLabel]; ok {
			group, err := s.GetGroup(gid, crdNamespace)
			if err != nil {
				return nil, fmt.Errorf("get group error: %+v", err)
			}
			app.Spec.Group.ID = group.ObjectMeta.Name
			app.Spec.Group.Name = group.Spec.Name
		}
	}
	return app, nil
}

func (s *Service) ForceDelete(id, crdNamespace string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}


/**
王晓婧更改部分**/
const (
	tenantId = "Tenantid"
)

//curl -i http://10.160.32.24:30800/api/v1/topics?application=cef854c589376a7d -H 'tenantId:148bf93169f7422bbf2e1eee2afa981f'
type TopicData struct {
	Page      int         `json:"page"`
	Size      int         `json:"size"`
	Content   interface{} `json:"content"`
	TotalSize int         `json:"totalSize"`
	TotalPage int         `json:"totalPage"`
}
type TopicResponseBody struct {
	Code      int       `json:"code"`
	Data      TopicData `json:"data"`
	ErrorCode string    `json:"errorCode"`
	Message   string    `json:"message"`
	Detail    string    `json:"detail"`
}

func GetTopicByApplication(ID string, opts ...util.OpOption) (rsp int, err error) {
	klog.Infof("begin get topic from application %s", ID)
	request := gorequest.New().SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)

	klog.Infof("% +v ", request.Header)
	schema := "http"
	TopicResponseBody := &TopicResponseBody{}
	request.Header.Add("tenantId", "148bf93169f7422bbf2e1eee2afa981f")
	klog.Infof("begin into request.Get ")
	response, body, errs := request.Get(fmt.Sprintf("%s://%s:%d%s%s", schema, "10.160.32.24", 30800, "/api/v1/topics?application=", ID)).EndStruct(TopicResponseBody)
	klog.Infof("end request.Get ")

	if len(errs) > 0 {
		return 0, fmt.Errorf("request for get route info error: %+v", errs)

	}
	klog.Infof("begin get topic from application %+v", response)
	if response.StatusCode != 200 {
		return 0, fmt.Errorf("request for get route error: receive wrong status code: %s", string(body))
	}

	klog.Infof("application: %+v", response)
	klog.Infof("TopicResponseBody: %+v", TopicResponseBody)
	klog.Infof("TotalSize: %d", TopicResponseBody.Data.TotalSize)
	return TopicResponseBody.Data.TotalSize, nil

}


////大数据
type TopicPage struct{
	TotalCount   int             `json:"totalCounte"`
	PageSize     int             `json:"pageSize"`
	TotalPage    int             `json:"totalPage"`
	CurrPage     int             `json:"currPage"`
	List         interface{}     `json:"list"`
}
type TopicResponseBodyBigData struct {
	Code      int         `json:"code"`
	Msg       string      `json:"msg"`
	Data      interface{} `json:"data"`
	Page      TopicPage    `json:"page"`

}



func GetTopic(ID string) (rsp int, err error) {
	klog.Infof("begin get topic from application %s", ID)
	request := gorequest.New().SetDebug(true).SetCurlCommand(true).SetDoNotClearSuperAgent(true)

	klog.Infof("% +v ", request.Header)
	schema := "http"
	TopicResponseBodyBigData := &TopicResponseBodyBigData {}
	//request.Header.Add("tenantId", "148bf93169f7422bbf2e1eee2afa981f")
	klog.Infof("begin into request.Get ")
	response, body, errs := request.Get(fmt.Sprintf("%s://%s:%d%s%s%s%s", schema, "10.160.32.24", 30892, "/cmcc/api/top/app/messagetopicapp/queryApp?appId=", ID, "&page=1", "&limit=10")).EndStruct(TopicResponseBodyBigData )
	klog.Infof("end request.Get ")

	if len(errs) > 0 {
		return 0, fmt.Errorf("request for get route info error: %+v", errs)

	}
	klog.Infof("begin get topic from application %+v", response)
	if response.StatusCode != 200 {
		return 0, fmt.Errorf("request for get route error: receive wrong status code: %s", string(body))
	}

	klog.Infof("application: %+v", response)
	klog.Infof("TopicResponseBody: %+v", TopicResponseBodyBigData)
	klog.Infof("TotalSize: %d", TopicResponseBodyBigData.Page.TotalCount)
	return TopicResponseBodyBigData.Page.TotalCount, nil

}

///////////////




func (s *Service) Delete(id string, opts ...util.OpOption) (*v1.Application, error) {
	//大数据
	if !s.tenantEnabled{
		klog.Infof("GET IN TO BIGDATA ")
		totalCount, err := GetTopic(id)
		if err != nil {
			return nil, fmt.Errorf("get crd by id error TOPIC: %+v", err)
		}

		app, err := s.Get(id, opts...)
		if err != nil {
			return nil, fmt.Errorf("get crd by id error: %+v", err)
		}

		if totalCount > 0 {
			return nil, fmt.Errorf("Binding topic,can not delete")
		}
		if len(app.Spec.APIs) > 0 {
			return nil, fmt.Errorf("Binding api,can not delete")
		}
		//TODO need check status !!!
		app.Status.Status = v1.Delete
		(*app).Spec.Result = v1.DELETING
		return s.UpdateStatus(app)
	} else
	//业务汇聚
	{
		klog.Infof("GET IN TO 业务汇聚 ")
		totalSize, err := GetTopicByApplication(id, opts...)
		if err != nil {
			return nil, fmt.Errorf("get crd by id error TOPIC: %+v", err)
		}

		app, err := s.Get(id, opts...)

		if err != nil {
			return nil, fmt.Errorf("get crd by id error: %+v", err)
		}

		if !s.tenantEnabled {
			u := util.OpList(opts...).User()
			if !user.WritePermitted(u, app.ObjectMeta.Labels) {
				return nil, fmt.Errorf("write permission denied")
			}
		}

		if totalSize > 0 {
			return nil, fmt.Errorf("Binding topic,can not delete")
		}
		if len(app.Spec.APIs) > 0 {
			return nil, fmt.Errorf("Binding api,can not delete")
		}
		//TODO need check status !!!
		app.Status.Status = v1.Delete
		(*app).Spec.Result = v1.DELETING
		return s.UpdateStatus(app)
	}
}

func (s *Service) UpdateSpec(app *v1.Application) (*v1.Application, error) {
	return s.UpdateStatus(app)
}

func (s *Service) UpdateStatus(app *v1.Application) (*v1.Application, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(app)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	klog.V(5).Infof("try to update status for crd: %+v", crd)
	crd, err = s.client.Namespace(app.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error update crd status: %+v", err)
	}
	app = &v1.Application{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.application: %+v", app)

	return app, nil
}

func (s *Service) GetGroup(id, crdNamespace string) (*groupv1.ApplicationGroup, error) {
	crd, err := s.groupClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	app := &groupv1.ApplicationGroup{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), app); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationgroup: %+v", app)
	return app, nil
}

func (s *Service) GetGroupMap(crdNamespace string) (map[string]string, error) {
	crd, err := s.groupClient.Namespace(crdNamespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	apps := &groupv1.ApplicationGroupList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), apps); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.applicationgrouplist: %+v", apps)
	m := make(map[string]string)
	for _, app := range apps.Items {
		m[app.ObjectMeta.Name] = app.Spec.Name
	}
	return m, nil
}

func (s *Service) AddUser(id, operator string, data *user.Data) error {
	crd, err := s.Get(id)
	if err != nil {
		return fmt.Errorf("get crd error: %+v", err)
	}
	labels := crd.ObjectMeta.Labels
	if !user.IsOwner(operator, labels) && !user.IsManager(operator, labels) {
		return fmt.Errorf("only owner or manager can add user")
	}
	labels, err = user.AddUserLabels(data, labels)
	if err != nil {
		return fmt.Errorf("add user labels error: %+v", err)
	}
	crd.ObjectMeta.Labels = labels
	_, err = s.UpdateSpec(crd)
	if err != nil {
		return fmt.Errorf("update crd error: %+v", err)
	}
	return nil
}

func (s *Service) RemoveUser(id, operator, target string) error {
	crd, err := s.Get(id)
	if err != nil {
		return fmt.Errorf("get crd error: %+v", err)
	}
	labels := crd.ObjectMeta.Labels
	if !user.IsOwner(operator, labels) && !user.IsManager(operator, labels) {
		return fmt.Errorf("only owner or manager can remove user")
	}
	labels, err = user.RemoveUserLabels(target, labels)
	if err != nil {
		return fmt.Errorf("remove user labels error: %+v", err)
	}
	crd.ObjectMeta.Labels = labels
	_, err = s.UpdateSpec(crd)
	if err != nil {
		return fmt.Errorf("update crd error: %+v", err)
	}
	return nil
}

func (s *Service) ChangeOwner(id, operator string, data *user.Data) error {
	klog.V(5).Infof("change owner appid=%s, operator=%s, data=%+v", id, operator, *data)
	crd, err := s.Get(id)
	if err != nil {
		return fmt.Errorf("get crd error: %+v", err)
	}
	labels := crd.ObjectMeta.Labels
	if !user.IsOwner(operator, labels) {
		return fmt.Errorf("only owner or can change owner")
	}
	labels, err = user.ChangeOwner(data.ID, labels)
	if err != nil {
		return fmt.Errorf("change owner labels error: %+v", err)
	}
	crd.ObjectMeta.Labels = labels
	_, err = s.UpdateSpec(crd)
	if err != nil {
		return fmt.Errorf("update crd error: %+v", err)
	}
	return nil
}

func (s *Service) ChangeUser(id, operator string, data *user.Data) error {
	crd, err := s.Get(id)
	if err != nil {
		return fmt.Errorf("get crd error: %+v", err)
	}
	labels := crd.ObjectMeta.Labels
	if !user.IsOwner(operator, labels) && !user.IsManager(operator, labels) {
		return fmt.Errorf("only owner or manager can add user")
	}
	labels, err = user.ChangeUser(data, labels)
	if err != nil {
		return fmt.Errorf("change user labels error: %+v", err)
	}
	crd.ObjectMeta.Labels = labels
	_, err = s.UpdateSpec(crd)
	if err != nil {
		return fmt.Errorf("update crd error: %+v", err)
	}
	return nil
}

func (s *Service) GetUsers(id, operator string) (user.UserList, bool, error) {
	crd, err := s.Get(id)
	if err != nil {
		return nil, false, fmt.Errorf("get crd error: %+v", err)
	}
	labels := crd.ObjectMeta.Labels
	return user.GetCasUsers(labels), user.IsOwner(operator, labels), nil
}
