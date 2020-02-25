package api

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/api/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/emicklei/go-restful"
)

type controller struct {
	service *service.Service
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.DataserviceConfig),
	}
}

const (
	serviceunit = "serviceunit"
	application = "application"
)

type Wrapped struct {
	Code    int          `json:"code"`
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
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type PingResponse = DeleteResponse

type TestApiResponse = struct {
	Code       int         `json:"code"`
	Message    string      `json:"message"`
	TestResult interface{} `json:"data,omitempty"`
}

func (c *controller) CreateApi(req *restful.Request) (int, interface{}) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}

	if api, err := c.service.CreateApi(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: api,
		}
	}
}

func (c *controller) PatchApi(req *restful.Request) (int, interface{}) {
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, ok := reqBody["data"]
	if !ok {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if api, err := c.service.PatchApi(req.PathParameter("id"), data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: api,
		}
	}
}

func (c *controller) GetApi(req *restful.Request) (int, interface{}) {
	id := req.PathParameter("id")
	if api, err := c.service.GetApi(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: api,
		}
	}
}

func (c *controller) DeleteApi(req *restful.Request) (int, interface{}) {
	id := req.PathParameter("id")
	if data, err := c.service.DeleteApi(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: 0,
			Data: data,
		}
	}
}

func (c *controller) PublishApi(req *restful.Request) (int, interface{}) {
	if su, err := c.service.PublishApi(req.PathParameter("id")); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("publish api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: su,
		}
	}
}

func (c *controller) OfflineApi(req *restful.Request) (int, interface{}) {
	if su, err := c.service.OfflineApi(req.PathParameter("id")); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("publish api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: su,
		}
	}
}

func (c *controller) ListApi(req *restful.Request) (int, interface{}) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	if api, err := c.service.ListApi(req.QueryParameter(serviceunit), req.QueryParameter(application)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list api error: %+v", err).Error(),
		}
	} else {
		var apis ApiList = api
		data, err := util.PageWrap(apis, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    1,
				Message: fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: data,
		}
	}
}

type ApiList []*service.Api

func (apis ApiList) Length() int {
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
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	apiID := req.PathParameter("id")
	appID := req.PathParameter("appid")
	if api, err := c.service.BindOrRelease(apiID, appID, body.Data.Operation); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:    2,
			Message: fmt.Errorf("bind api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &BindResponse{
			Code: 0,
			Data: api,
		}
	}
}

func (c *controller) Query(req *restful.Request) (int, interface{}) {
	apiid := req.PathParameter(apiidPath)
	req.Request.ParseForm()
	// pass an array to query parameter example:
	// http://localhost:8080?links[]=http://www.baidu.com&links[]=http://www.google.cn
	if data, err := c.service.Query(apiid, req.Request.Form); err == nil {
		return http.StatusOK, struct {
			Code    int          `json:"code"`
			Message string       `json:"message"`
			Data    service.Data `json:"data"`
		}{
			Code: 0,
			Data: data,
		}
	} else {
		return http.StatusInternalServerError, struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}{
			Code:    1,
			Message: fmt.Sprintf("query data error:%+v", err),
		}
	}
}

func (c *controller) TestApi(req *restful.Request) (int, interface{}) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &TestApiResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &TestApiResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if resp, err := c.service.TestApi(body.Data); err != nil {
		return http.StatusInternalServerError, &TestApiResponse{
			Code:    2,
			Message: fmt.Errorf("Test api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &TestApiResponse{
			Code:       0,
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
