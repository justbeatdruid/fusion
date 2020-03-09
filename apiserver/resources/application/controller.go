package application

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/application/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/emicklei/go-restful"
)

type controller struct {
	service *service.Service
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient()),
	}
}

type Wrapped struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    *service.Application `json:"data,omitempty"`
}

type CreateRequest = Wrapped
type CreateResponse = Wrapped
type DeleteResponse = Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
type PingResponse = DeleteResponse

func (c *controller) CreateApplication(req *restful.Request) (int, *CreateResponse) {
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
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	if app, err := c.service.CreateApplication(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create application error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: app,
		}
	}
}

func (c *controller) GetApplication(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if app, err := c.service.GetApplication(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get application error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: app,
		}
	}
}

func (c *controller) PatchApplication(req *restful.Request) (int, *DeleteResponse) {
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, ok := reqBody["data"]
	if !ok {
		data, ok = reqBody["data,omitempty"]
	}
	if !ok {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if app, err := c.service.PatchApplication(req.PathParameter("id"), data); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("patch application error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: 0,
			Data: app,
		}
	}
}

func (c *controller) DeleteApplication(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if app, err := c.service.DeleteApplication(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete application error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: 0,
			Data: app,
		}
	}
}

func (c *controller) ListApplication(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	group := req.QueryParameter("group")
	name := req.QueryParameter("name")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: "auth model error",
		}
	}
	if app, err := c.service.ListApplication(util.WithGroup(group), util.WithNameLike(name), util.WithUser(authuser.Name)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list application error: %+v", err).Error(),
		}
	} else {
		var apps ApplicationList = app
		data, err := util.PageWrap(apps, page, size)
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

type ApplicationList []*service.Application

func (apps ApplicationList) Len() int {
	return len(apps)
}

func (apps ApplicationList) GetItem(i int) (interface{}, error) {
	if i >= len(apps) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return apps[i], nil
}

func (c *controller) AddUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if len(body.Data.ID) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "read entity error: id in data is null",
		}
	}
	if len(body.Data.Role) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "read entity error: role in data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "auth model error",
		}
	}
	if err := c.service.AddUser(id, authuser.Name, body.Data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    2,
			Message: fmt.Errorf("add user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code: 0,
		}
	}
}

func (c *controller) RemoveUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	userid := req.PathParameter("userid")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "id in path parameter is null",
		}
	}
	if len(userid) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "user id in path parameter is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "auth model error",
		}
	}
	if err := c.service.RemoveUser(id, authuser.Name, userid); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    2,
			Message: fmt.Errorf("remove user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code: 0,
		}
	}
}

func (c *controller) ChangeOwner(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "id in path parameter is null",
		}
	}
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if len(body.Data.ID) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "read entity error: id in data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "auth model error",
		}
	}
	if err := c.service.ChangeOwner(id, authuser.Name, body.Data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    2,
			Message: fmt.Errorf("change owner error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code: 0,
		}
	}
}

func (c *controller) ChangeUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	userid := req.PathParameter("userid")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "id in path parameter is null",
		}
	}
	if len(userid) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "user id in path parameter is null",
		}
	}
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if len(body.Data.Role) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "read entity error: role in data is null",
		}
	}
	body.Data.ID = userid
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    1,
			Message: "auth model error",
		}
	}
	if err := c.service.ChangeUser(id, authuser.Name, body.Data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    2,
			Message: fmt.Errorf("change user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code: 0,
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
