package applicationgroup

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/applicationgroup/service"
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
	Data      *service.ApplicationGroup `json:"data,omitempty"`
}

type RequestWrapped struct {
	Data *service.ApplicationGroup `json:"data,omitempty"`
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

func (c *controller) CreateApplicationGroup(req *restful.Request) (int, *CreateResponse) {
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
	if db, err := c.service.CreateApplicationGroup(body.Data); err != nil {
		message := "创建操作执行错误"
		if errors.IsNameDuplicated(err) {
			message = "名字重复"
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: message,
			Detail:  fmt.Errorf("create applicationgroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) GetApplicationGroup(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "auth model error",
		}
	}
	if db, err := c.service.GetApplicationGroup(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:   1,
			Detail: fmt.Errorf("get applicationgroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) DeleteApplicationGroup(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:   1,
			Detail: "auth model error",
		}
	}
	if err := c.service.DeleteApplicationGroup(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		message := "删除操作执行错误"
		if errors.IsContentNotVoid(err) {
			message = "分组下包含正在使用的项目"
		}
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: message,
			Detail:  fmt.Errorf("delete applicationgroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:   0,
			Detail: "",
		}
	}
}

func (c *controller) ListApplicationGroup(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:   1,
			Detail: "auth model error",
		}
	}
	db, err := c.service.ListApplicationGroup(util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:   1,
			Detail: fmt.Errorf("list applicationgroup error: %+v", err).Error(),
		}
	}
	if len(page) == 0 && len(size) == 0 {
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: db,
		}
	}
	var apps ApplicationList = db
	data, err := util.PageWrap(apps, page, size)
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

type ApplicationList []*service.ApplicationGroup

func (apps ApplicationList) Len() int {
	return len(apps)
}

func (apps ApplicationList) GetItem(i int) (interface{}, error) {
	if i >= len(apps) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return apps[i], nil
}

func (c *controller) UpdateApplicationGroup(req *restful.Request) (int, *CreateResponse) {
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
	if db, err := c.service.UpdateApplicationGroup(id, body.Data, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		message := "更新操作执行错误"
		if errors.IsNameDuplicated(err) {
			message = "名字重复"
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: message,
			Detail:  fmt.Errorf("update applicationgroup error: %+v", err).Error(),
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
