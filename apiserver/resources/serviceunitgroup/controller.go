package serviceunitgroup

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/serviceunitgroup/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.GetKubeClient(), cfg.TenantEnabled),
		cfg.LocalConfig,
	}
}

type Wrapped struct {
	Code      int                       `json:"code"`
	Message   string                    `json:"message"`
	ErrorCode string                    `json:"errorCode"`
	Detail    string                    `json:"detail"`
	Data      *service.ServiceunitGroup `json:"data,omitempty"`
}

type RequestWrapped struct {
	Data *service.ServiceunitGroup `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = RequestWrapped

type DeleteResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	ErrorCode string `json:"errorCode"`
	Detail    string `json:"detail"`
}
type GetResponse = Wrapped
type ListResponse = struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	ErrorCode string      `json:"errorCode"`
	Detail    string      `json:"detail"`
	Data      interface{} `json:"data"`
}
type PingResponse = DeleteResponse

func (c *controller) CreateServiceunitGroup(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "auth model error",
		}
	}
	body.Data.Namespace = authuser.Namespace

	if db, err := c.service.CreateServiceunitGroup(body.Data); err != nil {
		message := "创建操作执行错误"
		if errors.IsNameDuplicated(err) {
			message = "名字重复"
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: message,
			Detail:  fmt.Errorf("create serviceunitgroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) GetServiceunitGroup(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "auth model error",
		}
	}

	if db, err := c.service.GetServiceunitGroup(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:   1,
			Detail: fmt.Errorf("get serviceunitgroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) DeleteServiceunitGroup(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:   1,
			Detail: "auth model error",
		}
	}

	if err := c.service.DeleteServiceunitGroup(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		message := "删除操作执行错误"
		if errors.IsContentNotVoid(err) {
			message = "分组下包含正在使用的项目"
		}
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: message,
			Detail:  fmt.Errorf("delete serviceunitgroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:   0,
			Detail: "",
		}
	}
}

func (c *controller) ListServiceunitGroup(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:   1,
			Detail: "auth model error",
		}
	}
	db, err := c.service.ListServiceunitGroup(util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:   1,
			Detail: fmt.Errorf("list serviceunitgroup error: %+v", err).Error(),
		}
	}
	if len(page) == 0 && len(size) == 0 {
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: db,
		}
	}
	var sus ServiceunitList = db
	data, err := util.PageWrap(sus, page, size)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:   1,
			Detail: fmt.Errorf("page error error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &ListResponse{
		Code: 0,
		Data: data,
	}
}

type ServiceunitList []*service.ServiceunitGroup

func (sus ServiceunitList) Len() int {
	return len(sus)
}

func (sus ServiceunitList) GetItem(i int) (interface{}, error) {
	if i >= len(sus) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return sus[i], nil
}

func (c *controller) UpdateServiceunitGroup(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "auth model error",
		}
	}
	if db, err := c.service.UpdateServiceunitGroup(id, body.Data, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		message := "更新操作执行错误"
		if errors.IsNameDuplicated(err) {
			message = "名字重复"
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: message,
			Detail:  fmt.Errorf("update serviceunitgroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: db,
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
