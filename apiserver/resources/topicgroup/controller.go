package topicgroup

import (
	"fmt"
	"github.com/chinamobile/nlpt/pkg/util"
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

const (
	success = iota
	fail
)

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
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type PingResponse = DeleteResponse

type TopicgroupList []*service.Topicgroup

func (tgs TopicgroupList) Len() int {
	return len(tgs)
}

func (tgs TopicgroupList) GetItem(i int) (interface{}, error) {
	if i >= len(tgs) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return tgs[i], nil
}

func (c *controller) CreateTopicgroup(req *restful.Request) (int, *CreateResponse) {
	body := &service.Topicgroup{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if tg, err := c.service.CreateTopicgroup(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: success,
			Data: tg,
		}
	}
}

func (c *controller) GetTopicgroup(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if tp, err := c.service.GetTopicgroup(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    fail,
			Message: fmt.Errorf("get database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: success,
			Data: tp,
		}
	}
}

//批量删除topicgroups
func (c *controller) DeleteTopicgroups(req *restful.Request) (int, *ListResponse) {
	ids := req.QueryParameters("ids")
	for _, id := range ids {
		if _, err := c.service.DeleteTopicgroup(id); err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    1,
				Message: fmt.Errorf("delete topicgroup error: %+v", err).Error(),
			}
		}
	}
	return http.StatusOK, &ListResponse{
		Code:    0,
		Message: "delete success",
	}
}
func (c *controller) DeleteTopicgroup(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if tp, err := c.service.DeleteTopicgroup(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    fail,
			Message: fmt.Errorf("delete topicgroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: success,
			Data: tp,
		}
	}
}

func (c *controller) ListTopicgroup(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")

	if tg, err := c.service.ListTopicgroup(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    fail,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var tps TopicgroupList = tg

		data, err := util.PageWrap(tps, page, size)
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
