package api

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/api/service"
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
		AppID string `json:"appID"`
		ApiID string `json:"apiID"`
	} `json:"data,omitempty"`
}
type BindResponse = Wrapped
type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse = Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    []*service.Api `json:"data"`
}
type PingResponse = DeleteResponse

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
	if su, err := c.service.PublishApi(body.Data.ID); err != nil {
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
	if su, err := c.service.OfflineApi(body.Data.ID); err != nil {
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
	if api, err := c.service.ListApi(req.QueryParameter(serviceunit), req.QueryParameter(application)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: api,
		}
	}
}

func (c *controller) BindApi(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if api, err := c.service.BindApi(body.Data.ApiID, body.Data.AppID); err != nil {
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

func (c *controller) ReleaseApi(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if api, err := c.service.ReleaseApi(body.Data.ApiID, body.Data.AppID); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:    2,
			Message: fmt.Errorf("release api error: %+v", err).Error(),
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
	form := req.Request.Form
	header := req.Request.Header
	return http.StatusOK, struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{
		Code:    0,
		Message: fmt.Sprintf("get your request with api %s, form: %+v, Header: %+v", apiid, form, header),
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
