package topicgroup

import (
	"fmt"
	tgerror "github.com/chinamobile/nlpt/apiserver/resources/topicgroup/error"
	"github.com/chinamobile/nlpt/apiserver/resources/topicgroup/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	"github.com/chinamobile/nlpt/pkg/util"
	"net/http"
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
	Code      int                 `json:"code"`
	ErrorCode string              `json:"errorCode"`
	Message   string              `json:"message"`
	Data      *service.Topicgroup `json:"data,omitempty"`
	Detail    string              `json:"detail"`
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
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	Detail    string      `json:"detail"`
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

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: tgerror.Error_Auth_Error,
			Message:   fmt.Sprintf("auth model error: %+v", err)}
	}

	body.Users = user.InitWithOwner(authuser.Name)
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

func (c *controller) GetTopics(req *restful.Request) (int, *ListResponse) {
	id := req.PathParameter("id")
	if tps, err := c.service.GetTopics(id); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    fail,
			Message: fmt.Errorf("get database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &ListResponse{
			Code: success,
			Data: tps,
		}
	}
}

//修改topicgroup的策略
func (c *controller) ModifyTopicgroup(req *restful.Request) (int, *CreateResponse) {
	id := req.PathParameter("id")
	if len(id) == 0 {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    fail,
			Message: fmt.Sprintf("parameter id is required"),
		}
	}

	policies := service.NewPolicies()
	if err := req.ReadEntity(policies); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Sprintf("cannot read entity: %+v", err),
		}
	}

	data, err := c.service.ModifyTopicgroup(id, policies)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Sprintf("modify topic group error: %+v", err),
		}
	}

	return http.StatusOK, &CreateResponse{
		Code:    0,
		Message: "success",
		Data:    data,
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
