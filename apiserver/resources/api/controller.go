package api

import (
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	"k8s.io/klog"
	"net/http"
	"time"

	"github.com/chinamobile/nlpt/apiserver/mutex"
	"github.com/chinamobile/nlpt/apiserver/resources/api/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
	lock    mutex.Mutex
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.DataserviceConnector, cfg.GetKubeClient(), cfg.TenantEnabled, cfg.LocalConfig),
		cfg.LocalConfig,
		cfg.Mutex,
	}
}

const (
	serviceunit   = "serviceunit"
	application   = "application"
	publishstatus = "status"
)

type Wrapped struct {
	Code      int          `json:"code"`
	ErrorCode string       `json:"errorCode"`
	Detail    string       `json:"detail"`
	Message   string       `json:"message"`
	Data      *service.Api `json:"data,omitempty"`
}

type RequestWrapped struct {
	Data *service.Api `json:"data,omitempty"`
}

type BindRequest struct {
	//Code    int    `json:"code"`
	//Message string `json:"message"`
	Data struct {
		Operation string       `json:"operation"`
		Apis      []v1.ApiBind `json:"apis"`
	} `json:"data,omitempty"`
}

type BindResponse = Wrapped
type CreateResponse = Wrapped
type CreateRequest = RequestWrapped
type DeleteResponse = Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Detail    string      `json:"detail"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}
type PingResponse = DeleteResponse

type TestApiResponse = struct {
	Code       int         `json:"code"`
	ErrorCode  string      `json:"errorCode"`
	Detail     string      `json:"detail"`
	Message    string      `json:"message"`
	TestResult interface{} `json:"data,omitempty"`
}

type StatisticsResponse = struct {
	Code      int                `json:"code"`
	ErrorCode string             `json:"errorCode"`
	Message   string             `json:"message"`
	Data      service.Statistics `json:"data"`
	Detail    string             `json:"detail"`
}

func (c *controller) CreateApi(req *restful.Request) (int, interface{}) {
	unlock, err := c.lock.Lock("api")
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   2,
			Detail: fmt.Sprintf("lock error: %+v", err),
		}
	}
	defer func() {
		_ = unlock()
	}()
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000001",
			Message:   c.errMsg.Api["001000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000002",
			Message:   c.errMsg.Api["001000002"],
			Detail:    "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	body.Data.Namespace = authuser.Namespace
	klog.V(5).Infof("body namespace is : %s", body.Data.Namespace)
	if api, err, code := c.service.CreateApi(body.Data); err != nil {
		if errors.IsNameDuplicated(err) {
			code = "001000022"
		} else if errors.IsUnpublished(err) {
			return http.StatusInternalServerError, &CreateResponse{
				Code:      2,
				ErrorCode: code,
				Message:   "服务单元未发布",
				Detail:    fmt.Errorf("create api error: %+v", err).Error(),
			}
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: code,
			Message:   c.errMsg.Api[code],
			Detail:    fmt.Errorf("create api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      api,
		}
	}
}

func (c *controller) PatchApi(req *restful.Request) (int, interface{}) {
	unlock, err := c.lock.Lock("api")
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   2,
			Detail: fmt.Sprintf("lock error: %+v", err),
		}
	}
	defer func() {
		_ = unlock()
	}()
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000001",
			Message:   c.errMsg.Api["001000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, ok := reqBody["data"]
	if !ok {
		data, ok = reqBody["data,omitempty"]
	}
	if !ok {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000002",
			Message:   c.errMsg.Api["001000002"],
			Detail:    "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	if api, err := c.service.PatchApi(req.PathParameter("id"), data, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		code := "001000005"
		if errors.IsNameDuplicated(err) {
			code = "001000022"
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: code,
			Message:   c.errMsg.Api[code],
			Detail:    fmt.Errorf("patch api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      api,
		}
	}
}

func (c *controller) GetApi(req *restful.Request) (int, interface{}) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	if api, err := c.service.GetApi(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      2,
			ErrorCode: "001000006",
			Message:   c.errMsg.Api["001000006"],
			Detail:    fmt.Errorf("get api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      api,
		}
	}
}

func (c *controller) DeleteApi(req *restful.Request) (int, interface{}) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	if data, err := c.service.DeleteApi(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      2,
			ErrorCode: "001000007",
			Message:   c.errMsg.Api["001000007"],
			Detail:    fmt.Errorf("delete api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      data,
		}
	}
}

func (c *controller) BatchDeleteApi(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      1,
			ErrorCode: "001000001",
			Message:   c.errMsg.Api["001000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data.Operation != "delete" {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000016",
			Message:   c.errMsg.Api["001000016"],
			Detail:    "operation params error",
		}
	}

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	if err := c.service.BatchDeleteApi(body.Data.Apis, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      2,
			ErrorCode: "001000007",
			Message:   c.errMsg.Api["001000007"],
			Detail:    fmt.Errorf("delete api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      0,
			ErrorCode: "0",
		}
	}
}

func (c *controller) PublishApi(req *restful.Request) (int, interface{}) {
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	if su, err := c.service.PublishApi(req.PathParameter("id"), util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: "001000008",
			Message:   c.errMsg.Api["001000008"],
			Detail:    fmt.Errorf("publish api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      su,
		}
	}
}

func (c *controller) OfflineApi(req *restful.Request) (int, interface{}) {
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	if su, err := c.service.OfflineApi(req.PathParameter("id"), util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: "001000009",
			Message:   c.errMsg.Api["001000009"],
			Detail:    fmt.Errorf("offline api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      su,
		}
	}
}

func (c *controller) ListApi(req *restful.Request) (int, interface{}) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	name := req.QueryParameter("name")
	res := req.QueryParameter("restriction")
	traff := req.QueryParameter("trafficcontrol")
	authType := req.QueryParameter("authType")
	apiBackendType := req.QueryParameter("apiBackendType")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	if api, err := c.service.ListApi(req.QueryParameter(serviceunit), req.QueryParameter(application),
		req.QueryParameter(publishstatus), util.WithNameLike(name), util.WithUser(authuser.Name),
		util.WithNamespace(authuser.Namespace), util.WithRestriction(res), util.WithTrafficcontrol(traff),
		util.WithAuthType(authType), util.WithApiBackendType(apiBackendType)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			ErrorCode: "001000010",
			Message:   c.errMsg.Api["001000010"],
			Detail:    fmt.Errorf("list api error: %+v", err).Error(),
		}
	} else {
		var apis ApiList = api
		data, err := util.PageWrap(apis, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      3,
				ErrorCode: "001000011",
				Message:   c.errMsg.Api["001000011"],
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      data,
		}
	}
}

type ApiList []*service.Api

func (apis ApiList) Len() int {
	return len(apis)
}

func (apis ApiList) GetItem(i int) (interface{}, error) {
	if i >= len(apis) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return apis[i], nil
}

func (apis ApiList) Less(i, j int) bool {
	return apis[i].ReleasedAt.Time.After(apis[j].ReleasedAt.Time)
}

func (apis ApiList) Swap(i, j int) {
	apis[i], apis[j] = apis[j], apis[i]
}

func (c *controller) BindApi(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      1,
			ErrorCode: "001000001",
			Message:   c.errMsg.Api["001000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	apiID := req.PathParameter("id")
	appID := req.PathParameter("appid")
	if api, err := c.service.BindOrRelease(apiID, appID, body.Data.Operation, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      2,
			ErrorCode: "001000013",
			Message:   c.errMsg.Api["001000013"],
			Detail:    fmt.Errorf("bind api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &BindResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      api,
		}
	}
}

func (c *controller) BatchBindApi(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      1,
			ErrorCode: "001000001",
			Message:   c.errMsg.Api["001000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	appID := req.PathParameter("appid")
	if err := c.service.BatchBindOrRelease(appID, body.Data.Operation, body.Data.Apis, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      2,
			ErrorCode: "001000013",
			Message:   c.errMsg.Api["001000013"],
			Detail:    fmt.Errorf("bind api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &BindResponse{
			Code:      0,
			ErrorCode: "0",
			//Data:      api,
		}
	}
}

func (c *controller) Query(req *restful.Request) (int, interface{}) {
	apiid := req.PathParameter(apiidPath)
	req.Request.ParseForm()
	// pass an array to query parameter example:
	// http://localhost:8080?links[]=http://www.baidu.com&links[]=http://www.google.cn
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &Wrapped{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	limit := req.QueryParameter("limit")
	// Always return 200
	if data, err := c.service.Query(apiid, req.Request.Form, limit, util.WithNamespace(authuser.Namespace)); err == nil {
		return http.StatusOK, struct {
			Code      int          `json:"code"`
			ErrorCode string       `json:"errorCode"`
			Message   string       `json:"message"`
			Data      service.Data `json:"data"`
		}{
			Code:      0,
			ErrorCode: "0",
			Data:      data,
		}
	} else {
		//return http.StatusInternalServerError, struct {
		return http.StatusOK, struct {
			Code      int          `json:"code"`
			ErrorCode string       `json:"errorCode"`
			Data      service.Data `json:"data"`
		}{
			Code:      0,
			ErrorCode: "0",
			Data: service.Data{
				Data: []map[string]string{{"调试结果": "查询异常"}},
			},
		}
	}
}

func (c *controller) KongQuery(req *restful.Request) (int, interface{}) {
	now := time.Now()
	tenantid := req.PathParameter(tenantidPath)
	apiid := req.PathParameter(apiidPath)
	req.Request.ParseForm()
	// pass an array to query parameter example:
	// http://localhost:8080?links[]=http://www.baidu.com&links[]=http://www.google.cn
	limit := req.QueryParameter("limit")
	if data, err := c.service.Query(apiid, req.Request.Form, limit, util.WithNamespace(tenantid)); err == nil {
		return http.StatusOK, struct {
			Code      int          `json:"code"`
			ErrorCode string       `json:"errorCode"`
			Message   string       `json:"message"`
			Data      service.Data `json:"data"`
			TimeUsed  int          `json:"timeUsedInMilliSeconds"`
		}{
			Code:      0,
			ErrorCode: "0",
			Data:      data,
			TimeUsed:  int(time.Since(now) / time.Millisecond),
		}
	} else {
		return http.StatusInternalServerError, struct {
			Code      int    `json:"code"`
			ErrorCode string `json:"errorCode"`
			Message   string `json:"message"`
			Detail    string `json:"detail"`
		}{
			Code:      1,
			ErrorCode: "001000014",
			Message:   c.errMsg.Api["001000014"],
			Detail:    fmt.Sprintf("query data error:%+v", err),
		}
	}
}

func (c *controller) TestApi(req *restful.Request) (int, interface{}) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &TestApiResponse{
			Code:      1,
			ErrorCode: "001000001",
			Message:   c.errMsg.Api["001000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &TestApiResponse{
			Code:      1,
			ErrorCode: "001000002",
			Message:   c.errMsg.Api["001000002"],
			Detail:    "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	body.Data.Namespace = authuser.Namespace
	if resp, err := c.service.TestApi(body.Data); err != nil {
		return http.StatusInternalServerError, &TestApiResponse{
			Code:      2,
			ErrorCode: "001000015",
			Message:   c.errMsg.Api["001000015"],
			Detail:    fmt.Errorf("Test api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &TestApiResponse{
			Code:       0,
			ErrorCode:  "0",
			TestResult: resp,
		}
	}
}

func (c *controller) DoStatisticsOncApis(req *restful.Request) (int, *StatisticsResponse) {
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &StatisticsResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	apiList, err := c.service.ListApis(authuser.Namespace)
	if err != nil {
		return http.StatusInternalServerError, &StatisticsResponse{
			Code:      1,
			ErrorCode: "001000010",
			Message:   c.errMsg.Api["001000010"],
			Detail:    fmt.Sprintf("do statistics on apis error, %+v", err),
		}
	}

	data := service.Statistics{}
	data.Total = len(apiList.Items)
	data.Increment, data.TotalCalled = c.CountApisIncrement(apiList.Items)
	return http.StatusOK, &StatisticsResponse{
		Code:      0,
		ErrorCode: "",
		Message:   "",
		Data:      data,
		Detail:    "do statistics on apis successfully",
	}
}

func (c *controller) CountApisIncrement(apis []v1.Api) (int, int) {
	var increment int
	var totalCalled int
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	for _, t := range apis {
		if t.Status.ReleasedAt.Unix() < end.Unix() && t.Status.ReleasedAt.Unix() >= start.Unix() {
			increment++
		}
		totalCalled = totalCalled + t.Status.CalledCount

	}

	return increment, totalCalled
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
