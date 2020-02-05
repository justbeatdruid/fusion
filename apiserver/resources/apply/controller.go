package apply

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/apply/service"
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
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    *service.Apply `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type GetResponse = Wrapped
type ListResponse = struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Data    []*service.Apply `json:"data"`
}
type PingResponse = DeleteResponse
type ApproveRequest struct {
	Data *struct {
		Admitted bool   `json:"admitted"`
		Reason   string `json:"reason"`
	} `json:"data"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type ApproveResponse = ApproveRequest

func (c *controller) CreateApply(req *restful.Request) (int, *CreateResponse) {
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
	if apl, err := c.service.CreateApply(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: apl,
		}
	}
}

func (c *controller) GetApply(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if apl, err := c.service.GetApply(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: apl,
		}
	}
}

func (c *controller) DeleteApply(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if err := c.service.DeleteApply(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:    0,
			Message: "",
		}
	}
}

func (c *controller) ListApply(req *restful.Request) (int, *ListResponse) {
	if apl, err := c.service.ListApply(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: apl,
		}
	}
}

func (c *controller) ApproveApply(req *restful.Request) (int, *ApproveResponse) {
	id := req.PathParameter("id")
	body := &ApproveRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &ApproveResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &ApproveResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if _, err := c.service.ApproveApply(id, body.Data.Admitted, body.Data.Reason); err != nil {
		return http.StatusInternalServerError, &ApproveResponse{
			Code:    2,
			Message: fmt.Errorf("create database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &ApproveResponse{
			Code: 0,
			Data: nil,
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
