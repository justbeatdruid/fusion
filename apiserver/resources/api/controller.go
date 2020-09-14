package api

import (
	"encoding/json"
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/resources/api/parser"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	v1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	"k8s.io/klog"

	"github.com/chinamobile/nlpt/apiserver/concurrency"
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
	lock    concurrency.Mutex
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.DataserviceConnector, cfg.GetKubeClient(), cfg.GetListers().ApiLister(), cfg.TenantEnabled, cfg.LocalConfig, cfg.Database),
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

type QueryRequest struct {
	Data struct {
		Page string   `json:"page,omitempty"`
		Size string   `json:"size,omitempty"`
		Apis []string `json:"apis"`
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

type ExportRequest = struct {
	Code      int            `json:"code"`
	ErrorCode string         `json:"errorCode"`
	Message   string         `json:"message"`
	Data      service.Export `json:"data"`
	Detail    string         `json:"detail"`
}

type AddPluginsRequest struct {
	Data struct {
		Name   string      `json:"name"`
		Config interface{} `json:"config"`
	} `json:"data,omitempty"`
}

type PatchPluginsRequest struct {
	Data struct {
		Config interface{} `json:"config"`
	} `json:"data,omitempty"`
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
	if api, err, code := c.service.CreateApi(body.Data); err != nil {
		if errors.IsNameDuplicated(err) {
			code = "001000022"
		} else if errors.IsUnpublished(err) {
			code = "001000026"
		}
		if strings.Contains(err.Error(), "path duplicated") {
			code = "001000023"
		}
		if strings.Contains(err.Error(), "path is illegal") {
			code = "001000024"
		}
		if strings.Contains(err.Error(), "KongApi.Paths is illegal") {
			code = "001000025"
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
		if strings.Contains(err.Error(), "path is illegal") {
			code = "001000024"
		}
		if strings.Contains(err.Error(), "KongApi.Paths is illegal") {
			code = "001000025"
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
		req.QueryParameter(publishstatus), true, util.WithNameLike(name), util.WithUser(authuser.Name),
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

func (c *controller) AddApiPlugins(req *restful.Request) (int, interface{}) {
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

	reqBody := &AddPluginsRequest{}
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
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
	klog.V(5).Infof("AddApiPlugins name config %s, %+v", reqBody.Data.Name, reqBody.Data.Config)
	if api, err := c.service.AddApiPlugins(req.PathParameter("id"), reqBody.Data.Name, reqBody.Data.Config, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		code := "001000005"
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

func (c *controller) DeleteApiPlugins(req *restful.Request) (int, interface{}) {
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
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	api_id := req.PathParameter("api_id")
	if api, err := c.service.DeleteApiPlugins(api_id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		code := "001000005"
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

func (c *controller) PatchApiPlugins(req *restful.Request) (int, interface{}) {
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

	reqBody := &PatchPluginsRequest{}
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
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
	api_id := req.PathParameter("api_id")
	plugin_id := req.PathParameter("plugin_id")
	klog.V(5).Infof("PatchApiPlugins api_id plugin_id config %s, %s, %+v", api_id, plugin_id, reqBody.Data.Config)
	if api, err := c.service.PatchApiPlugins(api_id, plugin_id, reqBody.Data.Config, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		code := "001000005"
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

type ApiList []*service.Api
type ApiResList []*service.ApiRes

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

func (apis ApiResList) Len() int {
	return len(apis)
}

func (apis ApiResList) GetItem(i int) (interface{}, error) {
	if i >= len(apis) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return apis[i], nil
}

func (apis ApiResList) Swap(i, j int) {
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

//ExportApis export apis
func (c *controller) ExportApis(req *restful.Request) (int, *ExportRequest) {
	body := &ExportRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &ExportRequest{
			Code:      1,
			ErrorCode: "001000001",
			Message:   c.errMsg.Api["001000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ExportRequest{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}

	file := excelize.NewFile()
	index := file.NewSheet("apis")
	//  数组建立
	s := []string{"id", "namespace", "name", "description", "serviceunit.id", "serviceunit.name", "users.owner.id",
		"users.owner.name", "apiType", "authType", "tags", "apiBackendType", "method", "protocol", "returnType",
		"apiDefineInfo.path", "apiDefineInfo.matchMode", "apiDefineInfo.method", "apiDefineInfo.protocol",
		"apiDefineInfo.cors", "apiQueryInfo.webParams", "publishStatus", "updatedAt", "releasedAt", "创建成功返回"}

	for i := range s {
		file.SetCellValue("apis", string(i+65)+"1", s[i])
	}

	row := 1
	for i := range body.Data.IDs {
		api, err := c.service.GetApi(body.Data.IDs[i], util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace))
		if err != nil {
			klog.Errorf("get Api failed, err: %+v", err)
			continue
		}
		webParams, _ := json.Marshal(api.ApiQueryInfo.WebParams)

		row++
		file.SetCellValue("apis", string(65)+strconv.Itoa(row), api.ID)
		file.SetCellValue("apis", string(66)+strconv.Itoa(row), api.Namespace)
		file.SetCellValue("apis", string(67)+strconv.Itoa(row), api.Name)
		file.SetCellValue("apis", string(68)+strconv.Itoa(row), api.Description)
		file.SetCellValue("apis", string(69)+strconv.Itoa(row), api.Serviceunit.ID)
		file.SetCellValue("apis", string(70)+strconv.Itoa(row), api.Serviceunit.Name)
		file.SetCellValue("apis", string(71)+strconv.Itoa(row), api.Users.Owner.ID)
		file.SetCellValue("apis", string(72)+strconv.Itoa(row), api.Users.Owner.Name)
		file.SetCellValue("apis", string(73)+strconv.Itoa(row), api.ApiType)
		file.SetCellValue("apis", string(74)+strconv.Itoa(row), api.AuthType)
		file.SetCellValue("apis", string(75)+strconv.Itoa(row), api.Tags)
		file.SetCellValue("apis", string(76)+strconv.Itoa(row), api.ApiBackendType)
		file.SetCellValue("apis", string(77)+strconv.Itoa(row), api.Method)
		file.SetCellValue("apis", string(78)+strconv.Itoa(row), api.Protocol)
		file.SetCellValue("apis", string(79)+strconv.Itoa(row), api.ReturnType)
		file.SetCellValue("apis", string(80)+strconv.Itoa(row), api.ApiDefineInfo.Path)
		file.SetCellValue("apis", string(81)+strconv.Itoa(row), api.ApiDefineInfo.MatchMode)
		file.SetCellValue("apis", string(82)+strconv.Itoa(row), api.ApiDefineInfo.Method)
		file.SetCellValue("apis", string(83)+strconv.Itoa(row), api.ApiDefineInfo.Protocol)
		file.SetCellValue("apis", string(84)+strconv.Itoa(row), api.ApiDefineInfo.Cors)
		file.SetCellValue("apis", string(85)+strconv.Itoa(row), webParams)
		file.SetCellValue("apis", string(86)+strconv.Itoa(row), api.PublishStatus)
		file.SetCellValue("apis", string(87)+strconv.Itoa(row), api.UpdatedAt)
		file.SetCellValue("apis", string(88)+strconv.Itoa(row), api.ReleasedAt)
		file.SetCellValue("apis", string(88)+strconv.Itoa(row), api.ApiReturnInfo.NormalExample)
	}

	file.SetActiveSheet(index)
	err = file.SaveAs("./tmp/api.xlsx")
	if err != nil {
		return http.StatusInternalServerError, &ExportRequest{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    fmt.Sprintf("save file error: %+v", err),
		}
	}
	return http.StatusOK, &ExportRequest{
		Code:      0,
		ErrorCode: "",
		Message:   "",
	}
}

//ImportApis import apis
func (c *controller) ImportApis(req *restful.Request, response *restful.Response) (int, *ImportResponse) {
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ImportResponse{
			Code:      1,
			ErrorCode: "0001000003",
			Message:   "0001000003",
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	authuserName := user.InitWithOwner(authuser.Name)
	nameSpace := authuser.Namespace

	spec := &parser.ApiExcelSpec{
		SheetName:        "apis",
		MultiPartFileKey: "uploadfile",
		TitleRowSpecList: []string{"api名称", "api类型", "认证方式", "标签", "描述",
			"请求path", "请求协议", "Method", "是否跨域", "后端请求path", "成功响应示例", "失败响应示例", "webParams"},
	}
	apis, err := parser.ParseApisFromExcel(req, response, spec)
	if err != nil {
		return http.StatusInternalServerError, &ImportResponse{
			Code:      1,
			ErrorCode: "001000027",
			Message:   "001000027",
			Detail:    fmt.Sprintf("import apis error: %+v", err),
		}
	}
	successdata := []service.Api{}
	faildata := []service.Api{}
	suid := req.PathParameter("suid")
	for _, ap := range apis.ParseData {
		if len(ap) < 12 {
			return http.StatusOK, &ImportResponse{
				Code:      0,
				ErrorCode: "0",
				Data:      successdata,
			}
		}
		webParams := &[]v1.WebParams{}
		err := json.Unmarshal([]byte(ap[12]), &webParams)
		if err != nil {
			return http.StatusInternalServerError, &ImportResponse{
				Code:      1,
				ErrorCode: "000100000sy",
				Message:   "webParams unmarshal failed",
				Detail:    fmt.Sprintf("webParams unmarshal failed",err),
			}
		}
		api := &service.Api{
			Namespace:   nameSpace,
			Name:        ap[0],
			Description: ap[4],
			Users:    authuserName,
			ApiType:  v1.ApiType(ap[1]),
			AuthType: v1.AuthType(ap[2]),
			Tags:     ap[3],

			Serviceunit: v1.Serviceunit{
				ID:    suid,
			},
			ApiDefineInfo: v1.ApiDefineInfo{
				Path:     ap[5],
				Protocol: v1.Protocol(ap[6]),
				Method:   v1.Method(ap[7]),
				Cors:     ap[8],
			},
			KongApi: v1.KongApiInfo{
				Paths: []string{ap[9]},
			},
			ApiReturnInfo: v1.ApiReturnInfo{
				NormalExample:  ap[10],
				FailureExample: ap[11],
			},
			ApiQueryInfo: v1.ApiQueryInfo{WebParams:*webParams},
		}

		if apic, err, code := c.service.CreateApi(api); err != nil {
			if errors.IsNameDuplicated(err) {
				code = "001000022"
			} else if errors.IsUnpublished(err) {
				code = "001000026"
			}
			if strings.Contains(err.Error(), "path duplicated") {
				code = "001000023"
			}
			if strings.Contains(err.Error(), "path is illegal") {
				code = "001000024"
			}
			if strings.Contains(err.Error(), "KongApi.Paths is illegal") {
				code = "001000025"
			}
			faildata = append(faildata, *api)
			return http.StatusInternalServerError, &ImportResponse{
				Code:      2,
				ErrorCode: code,
				Message:   c.errMsg.Api[code],
				Detail:    fmt.Errorf("import api error: %+v", err).Error(),
				Data:      faildata,
			}
		} else {
			successdata = append(successdata, *apic)
		}
	}
	return http.StatusOK, &ImportResponse{
		Code:      0,
		ErrorCode: "0",
		Data:      successdata,
	}
}
func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}

func (c *controller) ListAllApplicationApis(req *restful.Request) (int, interface{}) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	name := req.QueryParameter("name")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	if api, err := c.service.ListAllApplicationApis(
		util.WithNameLike(name), util.WithUser(authuser.Name),
		util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			ErrorCode: "001000010",
			Message:   c.errMsg.Api["001000010"],
			Detail:    fmt.Errorf("list api error: %+v", err).Error(),
		}
	} else {
		var apis ApplicationScopedApiList = api
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

func (c *controller) ListAllServiceunitApis(req *restful.Request) (int, interface{}) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	name := req.QueryParameter("name")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: "001000003",
			Message:   c.errMsg.Api["001000003"],
			Detail:    "auth model error",
		}
	}
	if api, err := c.service.ListAllServiceunitApis(
		util.WithNameLike(name), util.WithUser(authuser.Name),
		util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			ErrorCode: "001000010",
			Message:   c.errMsg.Api["001000010"],
			Detail:    fmt.Errorf("list api error: %+v", err).Error(),
		}
	} else {
		var apis ServiceunitScopedApiList = api
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

func (c *controller) ListApisByApiGroup(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: "",
			Message:   "",
		}
	}
	if api, err := c.service.ListApisByApiGroup(id, util.WithUser(authuser.Name),
		util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			Detail:    fmt.Errorf("list api error: %+v", err).Error(),
			ErrorCode: "",
			Message:   "",
		}
	} else {
		var apis ApiList = api
		data, err := util.PageWrap(apis, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      1,
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
				ErrorCode: "",
				Message:   "",
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      data,
		}
	}
}

func (c *controller) ListApisByApiPlugin(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: "",
			Message:   "",
		}
	}
	if api, err := c.service.ListApisByApiPlugin(id, util.WithUser(authuser.Name),
		util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			Detail:    fmt.Errorf("list api error: %+v", err).Error(),
			ErrorCode: "",
			Message:   "",
		}
	} else {
		var apis ApiResList = api
		data, err := util.PageWrap(apis, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      1,
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
				ErrorCode: "",
				Message:   "",
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      data,
		}
	}
}

func (c *controller) ListApisForCapability(req *restful.Request) (int, *ListResponse) {
	body := &QueryRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: "001000001",
			Message:   c.errMsg.Api["001000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	page := body.Data.Page
	size := body.Data.Size
	if len(size) == 0 {
		size = "-1"
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: "",
			Message:   "",
		}
	}
	if api, err := c.service.ListApisForCapability(body.Data.Apis, util.WithUser(authuser.Name),
		util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			Detail:    fmt.Errorf("list api error: %+v", err).Error(),
			ErrorCode: "",
			Message:   "",
		}
	} else {
		var apis ApiList = api
		data, err := util.PageWrap(apis, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      1,
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
				ErrorCode: "",
				Message:   "",
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      data,
		}
	}
}

type ApplicationScopedApiList []*service.ApplicationScopedApi

func (apis ApplicationScopedApiList) Len() int {
	return len(apis)
}

func (apis ApplicationScopedApiList) GetItem(i int) (interface{}, error) {
	if i >= len(apis) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return apis[i], nil
}

func (apis ApplicationScopedApiList) Less(i, j int) bool {
	return apis[i].ReleasedAt.Time.After(apis[j].ReleasedAt.Time)
}

func (apis ApplicationScopedApiList) Swap(i, j int) {
	apis[i], apis[j] = apis[j], apis[i]
}

type ServiceunitScopedApiList []*service.ServiceunitScopedApi

func (apis ServiceunitScopedApiList) Len() int {
	return len(apis)
}

func (apis ServiceunitScopedApiList) GetItem(i int) (interface{}, error) {
	if i >= len(apis) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return apis[i], nil
}

func (apis ServiceunitScopedApiList) Less(i, j int) bool {
	return apis[i].ReleasedAt.Time.After(apis[j].ReleasedAt.Time)
}

func (apis ServiceunitScopedApiList) Swap(i, j int) {
	apis[i], apis[j] = apis[j], apis[i]
}
