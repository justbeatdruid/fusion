package serviceunit

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/serviceunit/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

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
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    []*service.Serviceunit `json:"data"`
}
type PingResponse = DeleteResponse

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
	if db, err := c.service.CreateServiceunit(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) GetServiceunit(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if db, err := c.service.GetServiceunit(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: db,
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
	if db, err := c.service.ListServiceunit(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) PublishServiceunit(req *restful.Request) (int, *CreateResponse) {
	id := req.PathParameter("id")
	if su, err := c.service.PublishServiceunit(id); err != nil {
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

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
