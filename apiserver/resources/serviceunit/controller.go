package serviceunit

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/serviceunit/service"
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
		service.NewService(cfg.GetDynamicClient(), cfg.GetKubeClient(), cfg.TenantEnabled, cfg.LocalConfig),
		cfg.LocalConfig,
	}
}

type Wrapped struct {
	Code    string               `json:"code"`
	Msg     string               `json:"msg"`
	Message string               `json:"message"`
	Data    *service.Serviceunit `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse = Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code    string      `json:"code"`
	Msg     string      `json:"msg"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
type PingResponse = DeleteResponse

// + update_sunyu
type UpdateRequest = Wrapped
type UpdateResponse = Wrapped

func (c *controller) CreateServiceunit(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "008000001",
			Msg:     c.errMsg.Serviceunit["008000001"],
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "008000002",
			Msg:     c.errMsg.Serviceunit["008000002"],
			Message: "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "008000003",
			Msg:     c.errMsg.Serviceunit["008000003"],
			Message: "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	body.Data.Namespace = authuser.Namespace
	if su, err, code := c.service.CreateServiceunit(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    code,
			Msg:     c.errMsg.Serviceunit["code"],
			Message: fmt.Errorf("create serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: "0",
			Data: su,
		}
	}
}

func (c *controller) GetServiceunit(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "008000003",
			Msg:     c.errMsg.Serviceunit["008000003"],
			Message: "auth model error",
		}
	}
	if su, err := c.service.GetServiceunit(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    "008000005",
			Msg:     c.errMsg.Serviceunit["008000005"],
			Message: fmt.Errorf("get serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: "0",
			Data: su,
		}
	}
}

func (c *controller) PatchServiceunit(req *restful.Request) (int, *DeleteResponse) {
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "008000001",
			Msg:     c.errMsg.Serviceunit["008000001"],
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, ok := reqBody["data,omitempty"]
	if !ok {
		data, ok = reqBody["data"]
	}
	if !ok {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "008000002",
			Msg:     c.errMsg.Serviceunit["008000002"],
			Message: "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "008000003",
			Msg:     c.errMsg.Serviceunit["008000003"],
			Message: "auth model error",
		}
	}
	if su, err := c.service.PatchServiceunit(req.PathParameter("id"), data, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    "008000007",
			Msg:     c.errMsg.Serviceunit["008000007"],
			Message: fmt.Errorf("patch serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: "0",
			Data: su,
		}
	}
}

func (c *controller) DeleteServiceunit(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "008000003",
			Msg:     c.errMsg.Serviceunit["008000003"],
			Message: "auth model error",
		}
	}
	if data, err := c.service.DeleteServiceunit(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    "008000006",
			Msg:     c.errMsg.Serviceunit["008000006"],
			Message: fmt.Errorf("delete serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: "0",
			Data: data,
		}
	}
}

func (c *controller) ListServiceunit(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	group := req.QueryParameter("group")
	name := req.QueryParameter("name")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    "008000003",
			Msg:     c.errMsg.Serviceunit["008000003"],
			Message: "auth model error",
		}
	}
	if su, err := c.service.ListServiceunit(util.WithGroup(group), util.WithNameLike(name), util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    "008000008",
			Msg:     c.errMsg.Serviceunit["008000008"],
			Message: fmt.Errorf("list serviceunit error: %+v", err).Error(),
		}
	} else {
		var sus ServiceunitList = su
		data, err := util.PageWrap(sus, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    "008000009",
				Msg:     c.errMsg.Serviceunit["008000009"],
				Message: fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code: "0",
			Data: data,
		}
	}
}

type ServiceunitList []*service.Serviceunit

func (sus ServiceunitList) Len() int {
	return len(sus)
}

func (sus ServiceunitList) GetItem(i int) (interface{}, error) {
	if i >= len(sus) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return sus[i], nil
}

func (c *controller) PublishServiceunit(req *restful.Request) (int, *CreateResponse) {
	id := req.PathParameter("id")
	body := &struct {
		Published bool `json:"published"`
	}{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "008000001",
			Msg:     c.errMsg.Serviceunit["008000001"],
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "008000003",
			Msg:     c.errMsg.Serviceunit["008000003"],
			Message: "auth model error",
		}
	}
	if su, err := c.service.PublishServiceunit(id, body.Published, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    "008000010",
			Msg:     c.errMsg.Serviceunit["008000010"],
			Message: fmt.Errorf("publish serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: "0",
			Data: su,
		}
	}
}

/*
// +update_sunyu
func (c *controller) UpdateServiceunit(req *restful.Request) (int, *UpdateResponse) {
	if true {
		return http.StatusNotImplemented, &UpdateResponse{
			Code:    "008000001",
			Msg: c.errMsg.Serviceunit["008000001"],
			Message: "interface not supported",
		}
	}
	body := &UpdateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    "008000001",
			Msg: c.errMsg.Serviceunit["008000001"],
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    "008000002",
			Msg: c.errMsg.Serviceunit["008000002"],
			Message: "read entity error: data is null",
		}
	}
	id := req.PathParameter("id")
	if su, err := c.service.UpdateServiceunit(body.Data, id); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    "008000007",
			Msg: c.errMsg.Serviceunit["008000007"],
			Message: fmt.Errorf("update serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &UpdateResponse{
			Code: "0",
			Data: su,
		}
	}
}

*/

func (c *controller) AddUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000001",
			Msg:     c.errMsg.Serviceunit["008000001"],
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000002",
			Msg:     c.errMsg.Serviceunit["008000002"],
			Message: "read entity error: data is null",
		}
	}
	if len(body.Data.ID) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000011",
			Msg:     c.errMsg.Serviceunit["008000011"],
			Message: "read entity error: id in data is null",
		}
	}
	if len(body.Data.Role) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000012",
			Msg:     c.errMsg.Serviceunit["008000012"],
			Message: "read entity error: role in data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000003",
			Msg:     c.errMsg.Serviceunit["008000003"],
			Message: "auth model error",
		}
	}
	if err := c.service.AddUser(id, authuser.Name, body.Data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000013",
			Msg:     c.errMsg.Serviceunit["008000013"],
			Message: fmt.Errorf("add user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code: "0",
		}
	}
}

func (c *controller) RemoveUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	userid := req.PathParameter("userid")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000014",
			Msg:     c.errMsg.Serviceunit["008000014"],
			Message: "id in path parameter is null",
		}
	}
	if len(userid) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000015",
			Msg:     c.errMsg.Serviceunit["008000015"],
			Message: "user id in path parameter is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000003",
			Msg:     c.errMsg.Serviceunit["008000003"],
			Message: "auth model error",
		}
	}
	if err := c.service.RemoveUser(id, authuser.Name, userid); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000016",
			Msg:     c.errMsg.Serviceunit["008000016"],
			Message: fmt.Errorf("remove user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code: "0",
		}
	}
}

func (c *controller) ChangeOwner(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000014",
			Msg:     c.errMsg.Serviceunit["008000014"],
			Message: "id in path parameter is null",
		}
	}
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000001",
			Msg:     c.errMsg.Serviceunit["008000001"],
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000002",
			Msg:     c.errMsg.Serviceunit["008000002"],
			Message: "read entity error: data is null",
		}
	}
	if len(body.Data.ID) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000011",
			Msg:     c.errMsg.Serviceunit["008000011"],
			Message: "read entity error: id in data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000003",
			Msg:     c.errMsg.Serviceunit["008000003"],
			Message: "auth model error",
		}
	}
	if err := c.service.ChangeOwner(id, authuser.Name, body.Data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000018",
			Msg:     c.errMsg.Serviceunit["008000018"],
			Message: fmt.Errorf("change owner error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code: "0",
		}
	}
}

func (c *controller) ChangeUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	userid := req.PathParameter("userid")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000014",
			Msg:     c.errMsg.Serviceunit["008000014"],
			Message: "id in path parameter is null",
		}
	}
	if len(userid) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000015",
			Msg:     c.errMsg.Serviceunit["008000015"],
			Message: "user id in path parameter is null",
		}
	}
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000001",
			Msg:     c.errMsg.Serviceunit["008000001"],
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000002",
			Msg:     c.errMsg.Serviceunit["008000002"],
			Message: "read entity error: data is null",
		}
	}
	if len(body.Data.Role) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000012",
			Msg:     c.errMsg.Serviceunit["008000012"],
			Message: "read entity error: role in data is null",
		}
	}
	body.Data.ID = userid
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000003",
			Msg:     c.errMsg.Serviceunit["008000003"],
			Message: "auth model error",
		}
	}
	if err := c.service.ChangeUser(id, authuser.Name, body.Data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:    "008000017",
			Msg:     c.errMsg.Serviceunit["008000017"],
			Message: fmt.Errorf("change user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code: "0",
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
