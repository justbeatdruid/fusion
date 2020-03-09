package clientauth

import (
	"fmt"
	"github.com/chinamobile/nlpt/pkg/util"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/clientauth/service"
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
	success = iota
	fail
)

type Wrapped struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    *service.Clientauth `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse = Wrapped

/*type DeleteResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}*/
type GetResponse = Wrapped
type ListResponse = struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type PingResponse = DeleteResponse

type ClientauthList []*service.Clientauth

func (cas ClientauthList) Len() int {
	return len(cas)
}

func (cas ClientauthList) GetItem(i int) (interface{}, error) {
	if i >= len(cas) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return cas[i], nil
}

func (c *controller) CreateClientauth(req *restful.Request) (int, *CreateResponse) {
	body := &service.Clientauth{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if ca, err := c.service.CreateClientauth(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: success,
			Data: ca,
		}
	}
}

func (c *controller) GetClientauth(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if ca, err := c.service.GetClientauth(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    fail,
			Message: fmt.Errorf("get database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: success,
			Data: ca,
		}
	}
}

//删除所有clientauths
func (c *controller) DeleteAllClientauths(req *restful.Request) (int, *ListResponse) {
	if cas, err := c.service.DeleteAllClientauths(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("delete clientauth error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &ListResponse{
			Code: success,
			Data: cas,
		}
	}
}
func (c *controller) DeleteClientauth(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if ca, err := c.service.DeleteClientauth(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    fail,
			Message: fmt.Errorf("delete clientauth error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: success,
			Data: ca,
		}
	}
}

func (c *controller) ListClientauth(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")

	if ca, err := c.service.ListClientauth(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    fail,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var cas ClientauthList = ca

		data, err := util.PageWrap(cas, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Sprintf("page parameter error: %+v", err),
			}
		} else {
			return http.StatusOK, &ListResponse{
				Code: success,
				Data: data,
			}
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
