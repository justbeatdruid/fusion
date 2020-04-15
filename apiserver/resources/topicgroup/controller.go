package topicgroup

import (
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/resources/topic"
	topicservice "github.com/chinamobile/nlpt/apiserver/resources/topic/service"
	tgerror "github.com/chinamobile/nlpt/apiserver/resources/topicgroup/error"
	"github.com/chinamobile/nlpt/apiserver/resources/topicgroup/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	"github.com/chinamobile/nlpt/pkg/util"
	"net/http"
	"sort"
	"strings"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient()),
		cfg.LocalConfig,
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
type TopicgroupSlice TopicgroupList
type TopicList []*topicservice.Topic

func (tps TopicList) Len() int {
	return len(tps)
}

func (tps TopicList) GetItem(i int) (interface{}, error) {
	if i >= len(tps) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return tps[i], nil
}

//重写Interface的len方法
func (t TopicgroupSlice) Len() int {
	return len(t)
}

//重写Interface的Swap方法
func (t TopicgroupSlice) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

//重写Interface的Less方法
func (t TopicgroupSlice) Less(i, j int) bool {
	return t[i].CreatedAt > t[j].CreatedAt
}

func (tgs TopicgroupList) Len() int {
	return len(tgs)
}

func (tgs TopicgroupList) GetItem(i int) (interface{}, error) {
	if i >= len(tgs) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return tgs[i], nil
}

func (c *controller) newCreateResponse(code int, errorCode string, detail string, msg string) *CreateResponse {
	resp := &CreateResponse{
		Code:      code,
		Detail:    detail,
		ErrorCode: errorCode,
		Message:   msg,
	}

	if len(msg) == 0 {
		resp.Message = c.errMsg.TopicGroup[resp.ErrorCode]
	}
	return resp
}

func (c *controller) CreateTopicgroup(req *restful.Request) (int, *CreateResponse) {
	body := &service.Topicgroup{}
	body.Policies = *service.NewPolicies(true)
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      fail,
			ErrorCode: tgerror.ErrorReadEntity,
			Message:   c.errMsg.TopicGroup[tgerror.ErrorReadEntity],
			Detail:    fmt.Sprintf("cannot read entity: %+v", err),
		}
	}

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      fail,
			ErrorCode: tgerror.ErrorAuthError,
			Message:   c.errMsg.TopicGroup[tgerror.ErrorAuthError],
			Detail:    fmt.Sprintf("auth model error: %+v", err)}
	}

	body.Users = user.InitWithOwner(authuser.Name)
	body.Namespace = authuser.Namespace
	if tg, tgErr := c.service.CreateTopicgroup(body); tgErr.Err != nil {
		return http.StatusInternalServerError, c.newCreateResponse(fail, tgErr.ErrorCode, fmt.Sprintf("create database error: %+v", tgErr.Err), tgErr.Message)
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      success,
			ErrorCode: tgerror.Success,
			Data:      tg,
		}
	}
}

func (c *controller) GetTopicgroup(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if tp, err := c.service.GetTopicgroup(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      fail,
			ErrorCode: tgerror.ErrorQueryTopicgroupInfo,
			Message:   c.errMsg.TopicGroup[tgerror.ErrorQueryTopicgroupInfo],
			Detail:    fmt.Sprintf("get database error: %+v", err),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code:      success,
			ErrorCode: tgerror.Success,
			Data:      tp,
		}
	}
}

func (c *controller) GetTopics(req *restful.Request) (int, *ListResponse) {
	id := req.PathParameter("id")
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	authUser, err := auth.GetAuthUser(req)
	if err!=nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      fail,
			ErrorCode: tgerror.ErrorAuthError,
			Message:   c.errMsg.Topic[tgerror.ErrorAuthError],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	if tps, err := c.service.GetTopics(id, util.WithNamespace(authUser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    fail,
			Message: fmt.Errorf("get database error: %+v", err).Error(),
		}
	} else {
		var tl TopicList = tps
		sort.Sort(topic.TopicSlice(tl))
		data, err := util.PageWrap(tl, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      fail,
				ErrorCode: tgerror.ErrorPageParamInvalid,
				Message:   c.errMsg.Topic[tgerror.ErrorPageParamInvalid],
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      success,
			ErrorCode: tgerror.Success,
			Data:      data,
		}
	}
}

//修改topicgroup的策略
func (c *controller) ModifyTopicgroup(req *restful.Request) (int, *CreateResponse) {
	id := req.PathParameter("id")
	if len(id) == 0 {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      fail,
			ErrorCode: tgerror.ErrorModifyIDInvalid,
			Message:   c.errMsg.TopicGroup[tgerror.ErrorModifyIDInvalid],
			Detail:    fmt.Sprintf("parameter id is required"),
		}
	}

	policies := service.NewPolicies(false)
	if err := req.ReadEntity(policies); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      fail,
			ErrorCode: tgerror.ErrorReadEntity,
			Message:   c.errMsg.TopicGroup[tgerror.ErrorReadEntity],
			Detail:    fmt.Sprintf("cannot read entity: %+v", err),
		}
	}
    authUser, err := auth.GetAuthUser(req)
    if err!=nil{
		return http.StatusInternalServerError, &CreateResponse{
			Code:      fail,
			ErrorCode: tgerror.ErrorAuthError,
			Message:   c.errMsg.Topic[tgerror.ErrorAuthError],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	data, err := c.service.ModifyTopicgroup(id, policies,util.WithNamespace(authUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: tgerror.ErrorModify,
			Message:   c.errMsg.TopicGroup[tgerror.ErrorModify],
			Detail:    fmt.Sprintf("modify topic group error: %+v", err),
		}
	}

	return http.StatusOK, &CreateResponse{
		Code:      success,
		Message:   "accepted modify policies request",
		ErrorCode: tgerror.Success,
		Data:      data,
	}

}

//批量删除topicgroups
/*func (c *controller) DeleteTopicgroups(req *restful.Request) (int, *ListResponse) {
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
		Code:      0,
		ErrorCode: tgerror.Success,
		Message:   "delete success",
	}
}*/
func (c *controller) DeleteTopicgroup(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authUser, err := auth.GetAuthUser(req)
	if err!=nil{
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      fail,
			ErrorCode: tgerror.ErrorAuthError,
			Message:   c.errMsg.Topic[tgerror.ErrorAuthError],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	if tp, err := c.service.DeleteTopicgroup(id, util.WithNamespace(authUser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      fail,
			ErrorCode: tgerror.ErrorDelete,
			Message:   c.errMsg.TopicGroup[tgerror.ErrorDelete],
			Detail:    fmt.Sprintf("delete topicgroup error: %+v", err),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      success,
			ErrorCode: tgerror.Success,
			Data:      tp,
		}
	}
}

func (c *controller) ListTopicgroup(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	name := req.QueryParameter("name")
	topicName := req.QueryParameter("topicName")
	available := req.QueryParameter("available")

	if len(topicName) > 0 {
		topicName = strings.ToLower(strings.Trim(topicName, " "))
	}

	if len(name) > 0 {
		name = strings.ToLower(strings.Trim(name, " "))
	}

	if len(available) > 0 {
		switch available {
		case "true":
		case "false":
			break
		default:
			return http.StatusInternalServerError, &ListResponse{
				Code:      fail,
				ErrorCode: tgerror.ErrorGetTopicgroupList,
				Message:   c.errMsg.TopicGroup[tgerror.ErrorGetTopicgroupList],
				Detail:    fmt.Sprintf("list database error: available param error"),
			}
		}

	}
    authUser, err := auth.GetAuthUser(req)
    if err!=nil{
		return http.StatusInternalServerError, &ListResponse{
			Code:      fail,
			ErrorCode: tgerror.ErrorAuthError,
			Message:   c.errMsg.Topic[tgerror.ErrorAuthError],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	if tg, err := c.service.ListTopicgroup(util.WithNamespace(authUser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      fail,
			ErrorCode: tgerror.ErrorGetTopicgroupList,
			Message:   c.errMsg.TopicGroup[tgerror.ErrorGetTopicgroupList],
			Detail:    fmt.Sprintf("list database error: %+v", err),
		}
	} else {
		tg, err = c.service.SearchTopicgroup(tg, util.WithNameLike(name), util.WithTopic(topicName), util.WithAvailable(available))
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      fail,
				ErrorCode: tgerror.ErrorGetTopicgroupList,
				Message:   c.errMsg.TopicGroup[tgerror.ErrorGetTopicgroupList],
				Detail:    fmt.Sprintf("list database error: %+v", err),
			}
		}
		var tps TopicgroupList = tg
		sort.Sort(TopicgroupSlice(tps))
		data, err := util.PageWrap(tps, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      fail,
				ErrorCode: tgerror.ErrorPageParamInvalid,
				Message:   c.errMsg.TopicGroup[tgerror.ErrorPageParamInvalid],
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		} else {
			return http.StatusOK, &ListResponse{
				Code:      success,
				ErrorCode: tgerror.Success,
				Data:      data,
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
