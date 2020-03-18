package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/chinamobile/nlpt/apiserver/resources/api/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.DataserviceConnector, cfg.GetKubeClient(), cfg.TenantEnabled, cfg.LocalConfig),
		cfg.LocalConfig,
	}
}

const (
	serviceunit = "serviceunit"
	application = "application"
)

type Wrapped struct {
	Code    string       `json:"code"`
	Msg     string       `json:"msg"`
	Message string       `json:"message"`
	Data    *service.Api `json:"data,omitempty"`
}

type BindRequest struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Operation string `json:"operation"`
	} `json:"data,omitempty"`
}
type BindResponse = Wrapped
type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse = Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code    string      `json:"code"`
	Msg     string      `json:"msg"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type PingResponse = DeleteResponse

type TestApiResponse = struct {
	Code       string      `json:"code"`
	Msg        string      `json:"msg"`
	Message    string      `json:"message"`
	TestResult interface{} `json:"data,omitempty"`
}

func (c *controller) CreateApi(req *restful.Request) (int, interface{}) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000001",
			Msg:     c.errMsg.Api["001000001"],
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000002",
			Msg:     c.errMsg.Api["001000002"],
			Message: "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000003",
			Msg:     c.errMsg.Api["001000003"],
			Message: "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	body.Data.Namespace = authuser.Namespace
	if api, err, code := c.service.CreateApi(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    code,
			Msg:     c.errMsg.Api[code],
			Message: fmt.Errorf("create api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: "0",
			Data: api,
		}
	}
}

func (c *controller) PatchApi(req *restful.Request) (int, interface{}) {
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000001",
			Msg:     c.errMsg.Api["001000001"],
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, ok := reqBody["data"]
	if !ok {
		data, ok = reqBody["data,omitempty"]
	}
	if !ok {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000002",
			Msg:     c.errMsg.Api["001000002"],
			Message: "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000003",
			Msg:     c.errMsg.Api["001000003"],
			Message: "auth model error",
		}
	}
	if api, err := c.service.PatchApi(req.PathParameter("id"), data, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000005",
			Msg:     c.errMsg.Api["001000005"],
			Message: fmt.Errorf("patch api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: "0",
			Data: api,
		}
	}
}

func (c *controller) GetApi(req *restful.Request) (int, interface{}) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000003",
			Msg:     c.errMsg.Api["001000003"],
			Message: "auth model error",
		}
	}
	if api, err := c.service.GetApi(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    "001000006",
			Msg:     c.errMsg.Api["001000006"],
			Message: fmt.Errorf("get api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: "0",
			Data: api,
		}
	}
}

func (c *controller) DeleteApi(req *restful.Request) (int, interface{}) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000003",
			Msg:     c.errMsg.Api["001000003"],
			Message: "auth model error",
		}
	}
	if data, err := c.service.DeleteApi(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    "001000007",
			Msg:     c.errMsg.Api["001000007"],
			Message: fmt.Errorf("delete api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: "0",
			Data: data,
		}
	}
}

func (c *controller) PublishApi(req *restful.Request) (int, interface{}) {
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000003",
			Msg:     c.errMsg.Api["001000003"],
			Message: "auth model error",
		}
	}
	if su, err := c.service.PublishApi(req.PathParameter("id"), util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000008",
			Msg:     c.errMsg.Api["001000008"],
			Message: fmt.Errorf("publish api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: "0",
			Data: su,
		}
	}
}

func (c *controller) OfflineApi(req *restful.Request) (int, interface{}) {
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000003",
			Msg:     c.errMsg.Api["001000003"],
			Message: "auth model error",
		}
	}
	if su, err := c.service.OfflineApi(req.PathParameter("id"), util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000009",
			Msg:     c.errMsg.Api["001000009"],
			Message: fmt.Errorf("offline api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: "0",
			Data: su,
		}
	}
}

func (c *controller) ListApi(req *restful.Request) (int, interface{}) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	name := req.QueryParameter("name")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    "001000003",
			Msg:     c.errMsg.Api["001000003"],
			Message: "auth model error",
		}
	}
	if api, err := c.service.ListApi(req.QueryParameter(serviceunit), req.QueryParameter(application), util.WithNameLike(name), util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    "001000010",
			Msg:     c.errMsg.Api["001000010"],
			Message: fmt.Errorf("list api error: %+v", err).Error(),
		}
	} else {
		var apis ApiList = api
		data, err := util.PageWrap(apis, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    "001000011",
				Msg:     c.errMsg.Api["001000011"],
				Message: fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code: "0",
			Data: data,
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

func (c *controller) BindApi(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:    "001000001",
			Msg:     c.errMsg.Api["001000001"],
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000003",
			Msg:     c.errMsg.Api["001000003"],
			Message: "auth model error",
		}
	}
	apiID := req.PathParameter("id")
	appID := req.PathParameter("appid")
	if api, err := c.service.BindOrRelease(apiID, appID, body.Data.Operation, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:    "001000013",
			Msg:     c.errMsg.Api["001000013"],
			Message: fmt.Errorf("bind api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &BindResponse{
			Code: "0",
			Data: api,
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
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "001000003",
			Msg:     c.errMsg.Api["001000003"],
			Message: "auth model error",
		}
	}
	limit := req.QueryParameter("limit")
	if data, err := c.service.Query(apiid, req.Request.Form, limit, util.WithNamespace(authuser.Namespace)); err == nil {
		return http.StatusOK, struct {
			Code    string       `json:"code"`
			Msg     string       `json:"msg"`
			Message string       `json:"message"`
			Data    service.Data `json:"data"`
		}{
			Code: "0",
			Data: data,
		}
	} else {
		return http.StatusInternalServerError, struct {
			Code    string `json:"code"`
			Msg     string `json:"msg"`
			Message string `json:"message"`
		}{
			Code:    "001000014",
			Msg:     c.errMsg.Api["001000014"],
			Message: fmt.Sprintf("query data error:%+v", err),
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
			Code     string       `json:"code"`
			Message  string       `json:"message"`
			Data     service.Data `json:"data"`
			TimeUsed int          `json:"timeUserInMilliSeconds"`
		}{
			Code:     "0",
			Data:     data,
			TimeUsed: int(time.Since(now) / time.Millisecond),
		}
	} else {
		return http.StatusInternalServerError, struct {
			Code    string `json:"code"`
			Msg     string `json:"msg"`
			Message string `json:"message"`
		}{
			Code:    "001000014",
			Msg:     c.errMsg.Api["001000014"],
			Message: fmt.Sprintf("query data error:%+v", err),
		}
	}
}

func (c *controller) TestApi(req *restful.Request) (int, interface{}) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &TestApiResponse{
			Code:    "001000001",
			Msg:     c.errMsg.Api["001000001"],
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &TestApiResponse{
			Code:    "001000002",
			Msg:     c.errMsg.Api["001000002"],
			Message: "read entity error: data is null",
		}
	}
	if resp, err := c.service.TestApi(body.Data); err != nil {
		return http.StatusInternalServerError, &TestApiResponse{
			Code:    "001000015",
			Msg:     c.errMsg.Api["001000015"],
			Message: fmt.Errorf("Test api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &TestApiResponse{
			Code:       "0",
			TestResult: resp,
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
