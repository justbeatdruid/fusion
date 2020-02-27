package topicgroup

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/topicgroup/service"
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
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    *service.Topicgroup `json:"data,omitempty"`
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
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	Data    []*service.Topicgroup `json:"data"`
}
type PingResponse = DeleteResponse

func (c *controller) CreateTopicGroup(req *restful.Request) (int, *CreateResponse) {
	body := &service.Topicgroup{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if tp, err := c.service.CreateTopicGroup(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: tp,
		}
	}
}

func (c *controller) GetTopicgroup(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if tp, err := c.service.GetTopicgroup(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: tp,
		}
	}
}

//删除所有topics
func (c *controller) DeleteAllTopicgroups(req *restful.Request) (int, *ListResponse) {
	if tps, err := c.service.DeleteAllTopicgroups(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("delete topicgroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: tps,
		}
	}
}
func (c *controller) DeleteTopicgroup(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if tp, err := c.service.DeleteTopicgroup(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete topicgroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: 0,
			Data: tp,
		}
	}
}

func (c *controller) ListTopicgroup(req *restful.Request) (int, *ListResponse) {
	if tp, err := c.service.ListTopicgroup(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: tp,
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
