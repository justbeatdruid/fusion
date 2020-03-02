package serviceunit

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/serviceunit/service"
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
	Data    *service.Serviceunit `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse = Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code    int         `json:"code"`
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
	if su, err := c.service.CreateServiceunit(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: su,
		}
	}
}

func (c *controller) GetServiceunit(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if su, err := c.service.GetServiceunit(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: su,
		}
	}
}

func (c *controller) PatchServiceunit(req *restful.Request) (int, *DeleteResponse) {
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, ok := reqBody["data,omitempty"]
	if !ok {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if su, err := c.service.PatchServiceunit(req.PathParameter("id"), data); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("patch serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: 0,
			Data: su,
		}
	}
}

func (c *controller) DeleteServiceunit(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if data, err := c.service.DeleteServiceunit(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: 0,
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
			Code:    1,
			Message: "auth model error",
		}
	}
	if su, err := c.service.ListServiceunit(util.WithGroup(group), util.WithNameLike(name), util.WithUser(authuser.Name)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list serviceunit error: %+v", err).Error(),
		}
	} else {
		var sus ServiceunitList = su
		data, err := util.PageWrap(sus, page, size)
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

type ServiceunitList []*service.Serviceunit

func (sus ServiceunitList) Length() int {
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
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if su, err := c.service.PublishServiceunit(id, body.Published); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: su,
		}
	}
}

// +update_sunyu
func (c *controller) UpdateServiceunit(req *restful.Request) (int, *UpdateResponse) {
	if true {
		return http.StatusNotImplemented, &UpdateResponse{
			Code:    1,
			Message: "interface not supported",
		}
	}
	body := &UpdateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	id := req.PathParameter("id")
	if su, err := c.service.UpdateServiceunit(body.Data, id); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    2,
			Message: fmt.Errorf("update serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &UpdateResponse{
			Code: 0,
			Data: su,
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
