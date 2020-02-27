package topic

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/chinamobile/nlpt/apiserver/resources/topic/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
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
	Code    int            `json:"code"`
	Message string         `json:"message"`
	Data    *service.Topic `json:"data,omitempty"`
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
type MessageResponse = struct {
	Code     int         `json:"code"`
	Message  string      `json:"message"`
	Messages interface{} `json:"messages"`
}
type PingResponse = DeleteResponse

func (c *controller) CreateTopic(req *restful.Request) (int, *CreateResponse) {
	body := &service.Topic{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if tp, err := c.service.CreateTopic(body); err != nil {
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

func (c *controller) GetTopic(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if tp, err := c.service.GetTopic(id); err != nil {
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

//批量删除topics
func (c *controller) DeleteAllTopics(req *restful.Request) (int, *ListResponse) {
	if tps, err := c.service.DeleteAllTopics(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("delete topic error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: tps,
		}
	}
}
func (c *controller) DeleteTopic(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if _, err := c.service.DeleteTopic(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete topic error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:    0,
			Message: "deleting",
		}
	}
}

func (c *controller) ListTopic(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	if tp, err := c.service.ListTopic(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var tps TopicList = tp
		data, err := util.PageWrap(tps, page, size)
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

type TopicList []*service.Topic

func (ts TopicList) Length() int {
	return len(ts)
}

func (ts TopicList) GetItem(i int) (interface{}, error) {
	if i >= len(ts) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return ts[i], nil
}

//查询topic的消息
func (c *controller) ListMessages(req *restful.Request) (int, *MessageResponse) {
	topicUrl := req.QueryParameter("topicUrl")
	startTime := req.QueryParameter("startTime")
	endTime := req.QueryParameter("endTime")
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	//startTime、endTime参数都存在
	if len(startTime)>0 &&len(endTime)>0 {
		start, _ := strconv.ParseInt(startTime, 10, 64)
		end, _ := strconv.ParseInt(endTime, 10, 64)
		if messages, err := c.service.ListMessagesTime(topicUrl, start, end); err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:    1,
				Message: fmt.Errorf("list database error: %+v", err).Error(),
			}
		} else {
			var ms MessageList = messages
			data, err := util.PageWrap(ms, page, size)
			if err != nil {
				return http.StatusInternalServerError, &MessageResponse{
					Code:    1,
					Message: fmt.Sprintf("page parameter error: %+v", err),
				}
			}
			return http.StatusOK, &MessageResponse{
				Code:     0,
				Messages: data,
			}
		}
	} else {
		if messages, err := c.service.ListMessages(topicUrl); err != nil {
			return http.StatusInternalServerError, &MessageResponse{
				Code:    1,
				Message: fmt.Errorf("list database error: %+v", err).Error(),
			}
		} else {
			var ms MessageList = messages
			data, err := util.PageWrap(ms, page, size)
			if err != nil {
				return http.StatusInternalServerError, &MessageResponse{
					Code:    1,
					Message: fmt.Sprintf("page parameter error: %+v", err),
				}
			}
			return http.StatusOK, &MessageResponse{
				Code:     0,
				Messages: data,
			}
		}
	}

}

type MessageList []service.Message

func (ms MessageList) Length() int {
	return len(ms)
}

func (ms MessageList) GetItem(i int) (interface{}, error) {
	if i >= len(ms) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return ms[i], nil
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
