package application

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chinamobile/nlpt/apiserver/concurrency"
	aperror "github.com/chinamobile/nlpt/apiserver/resources/application/error"
	"github.com/chinamobile/nlpt/apiserver/resources/application/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	v1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	"github.com/chinamobile/nlpt/pkg/util"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
	lock    concurrency.Mutex
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.GetKubeClient(), cfg.TenantEnabled, cfg.Database, cfg.TopicConfig),
		cfg.LocalConfig,
		cfg.Mutex,
	}
}

type Wrapped struct {
	Code      int                  `json:"code"`
	ErrorCode string               `json:"errorCode"`
	Message   string               `json:"message"`
	Detail    string               `json:"detail"`
	Data      *service.Application `json:"data,omitempty"`
}

type RequestWrapped struct {
	Data *service.Application `json:"data,omitempty"`
}

func (RequestWrapped) SwaggerDoc() map[string]string {
	return map[string]string{
		"":     "Application Doc",
		"data": "Request data",
	}
}

func (Wrapped) SwaggerDoc() map[string]string {
	return map[string]string{
		"":        "Application Doc",
		"data":    "Response data",
		"code":    "code",
		"message": "message",
		"detail":  "detail",
	}
}

type CreateRequest = RequestWrapped
type CreateResponse = Wrapped
type DeleteResponse = Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Message   string      `json:"message"`
	Detail    string      `json:"detail"`
	Data      interface{} `json:"data,omitempty"`
}
type PingResponse = DeleteResponse

type StatisticsResponse = struct {
	Code      int                `json:"code"`
	ErrorCode string             `json:"errorCode"`
	Message   string             `json:"message"`
	Data      service.Statistics `json:"data"`
	Detail    string             `json:"detail"`
}

func (c *controller) CreateApplication(req *restful.Request) (int, *CreateResponse) {
	unlock, err := c.lock.Lock("application")
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   2,
			Detail: fmt.Sprintf("lock error: %+v", err),
		}
	}
	defer func() {
		_ = unlock()
	}()
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
			ErrorCode: aperror.FailedToReadMessageContent,
			Message:   c.errMsg.Application[aperror.FailedToReadMessageContent],
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			Detail:    "read entity error: data is null",
			ErrorCode: aperror.MessageBodyIsEmpty,
			Message:   c.errMsg.Application[aperror.MessageBodyIsEmpty],
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	body.Data.Namespace = authuser.Namespace
	if app, err, code := c.service.CreateApplication(body.Data); err != nil {
		if errors.IsNameDuplicated(err) {
			code = aperror.DuplicateApplicationName
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			Detail:    fmt.Errorf("create application error: %+v", err).Error(),
			ErrorCode: code,
			Message:   c.errMsg.Application[code],
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: aperror.Success,
			Data:      app,
		}
	}
}

func (c *controller) GetApplication(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
		}
	}
	if app, err := c.service.GetApplication(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      1,
			Detail:    fmt.Errorf("get application error: %+v", err).Error(),
			ErrorCode: aperror.QueryingASingleApplicationFailedById,
			Message:   c.errMsg.Application[aperror.QueryingASingleApplicationFailedById],
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code:      0,
			ErrorCode: aperror.Success,
			Data:      app,
		}
	}
}

func (c *controller) PatchApplication(req *restful.Request) (int, *DeleteResponse) {
	unlock, err := c.lock.Lock("application")
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:   2,
			Detail: fmt.Sprintf("lock error: %+v", err),
		}
	}
	defer func() {
		_ = unlock()
	}()
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
			ErrorCode: aperror.FailedToReadMessageContent,
			Message:   c.errMsg.Application[aperror.FailedToReadMessageContent],
		}
	}
	data, ok := reqBody["data"]
	if !ok {
		data, ok = reqBody["data,omitempty"]
	}
	if !ok {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			Detail:    "read entity error: data is null",
			ErrorCode: aperror.MessageBodyIsEmpty,
			Message:   c.errMsg.Application[aperror.MessageBodyIsEmpty],
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
		}
	}
	if app, err := c.service.PatchApplication(req.PathParameter("id"), data, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		code := aperror.UpdateApplicationFailed
		if errors.IsNameDuplicated(err) {
			code = aperror.DuplicateApplicationName
		}
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      1,
			Detail:    fmt.Errorf("patch application error: %+v", err).Error(),
			ErrorCode: code,
			Message:   c.errMsg.Application[code],
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      0,
			ErrorCode: aperror.Success,
			Data:      app,
		}
	}
}

func (c *controller) DeleteApplication(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
		}
	}
	if app, err := c.service.DeleteApplication(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      1,
			Detail:    fmt.Errorf("delete application error: %+v", err).Error(),
			ErrorCode: aperror.FailedToDeleteApplication,
			Message:   c.errMsg.Application[aperror.FailedToDeleteApplication],
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      0,
			ErrorCode: aperror.Success,
			Data:      app,
		}
	}
}

func (c *controller) ListApplication(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	group := req.QueryParameter("group")
	name := req.QueryParameter("name")
	id := req.QueryParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
		}
	}
	if app, err := c.service.ListApplication(util.WithGroup(group), util.WithNameLike(name),
		util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace), util.WithId(id)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			Detail:    fmt.Errorf("list application error: %+v", err).Error(),
			ErrorCode: aperror.QueryApplicationListFailed,
			Message:   c.errMsg.Application[aperror.QueryApplicationListFailed],
		}
	} else {
		var apps ApplicationList = app
		data, err := util.PageWrap(apps, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      1,
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
				ErrorCode: aperror.QueryApplicationPagingParameterError,
				Message:   c.errMsg.Application[aperror.QueryApplicationPagingParameterError],
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: aperror.Success,
			Data:      data,
		}
	}
}

func (c *controller) ListApplicationByRelation(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	group := req.QueryParameter("group")
	resourceType := req.QueryParameter("resourceType")
	resourceId := req.QueryParameter("resourceId")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
		}
	}
	if app, err := c.service.ListApplicationByRelation(resourceType, resourceId, util.WithGroup(group),
		util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			Detail:    fmt.Errorf("list application error: %+v", err).Error(),
			ErrorCode: aperror.QueryApplicationListFailed,
			Message:   c.errMsg.Application[aperror.QueryApplicationListFailed],
		}
	} else {
		var apps ApplicationList = app
		data, err := util.PageWrap(apps, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      1,
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
				ErrorCode: aperror.QueryApplicationPagingParameterError,
				Message:   c.errMsg.Application[aperror.QueryApplicationPagingParameterError],
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: aperror.Success,
			Data:      data,
		}
	}
}

type ApplicationList []*service.Application

func (apps ApplicationList) Len() int {
	return len(apps)
}

func (apps ApplicationList) GetItem(i int) (interface{}, error) {
	if i >= len(apps) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return apps[i], nil
}

func (apps ApplicationList) Less(i, j int) bool {
	return apps[i].CreatedAt.Time.After(apps[j].CreatedAt.Time)
}

func (apps ApplicationList) Swap(i, j int) {
	apps[i], apps[j] = apps[j], apps[i]
}

func (c *controller) AddUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.FailedToReadMessageContent,
			Message:   c.errMsg.Application[aperror.FailedToReadMessageContent],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, err := user.ToData(body.Data)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.MessageBodyIsEmpty,
			Message:   c.errMsg.Application[aperror.MessageBodyIsEmpty],
			Detail:    "read entity error: " + err.Error(),
		}
	}
	if len(data.ID) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.TheIdInTheMessageBodyIsEmpty,
			Message:   c.errMsg.Application[aperror.TheIdInTheMessageBodyIsEmpty],
			Detail:    "read entity error: id in data is null",
		}
	}
	if len(data.Role) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.TheRoleInTheMessageBodyIsEmpty,
			Message:   c.errMsg.Application[aperror.TheRoleInTheMessageBodyIsEmpty],
			Detail:    "read entity error: role in data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if err := c.service.AddUser(id, authuser.Name, data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: aperror.FailedToAddUser,
			Message:   c.errMsg.Application[aperror.FailedToAddUser],
			Detail:    fmt.Errorf("add user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code:      0,
			ErrorCode: aperror.Success,
		}
	}
}

func (c *controller) RemoveUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	userid := req.PathParameter("userid")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.TheIdInTheUrlParameterIsEmpty,
			Message:   c.errMsg.Application[aperror.TheIdInTheUrlParameterIsEmpty],
			Detail:    "id in path parameter is null",
		}
	}
	if len(userid) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.UserIdInUrlParameterIsEmpty,
			Message:   c.errMsg.Application[aperror.UserIdInUrlParameterIsEmpty],
			Detail:    "user id in path parameter is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if err := c.service.RemoveUser(id, authuser.Name, userid); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: aperror.FailedToRemoveUser,
			Message:   c.errMsg.Application[aperror.FailedToRemoveUser],
			Detail:    fmt.Errorf("remove user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code:      0,
			ErrorCode: aperror.Success,
		}
	}
}

func (c *controller) ChangeOwner(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.TheIdInTheUrlParameterIsEmpty,
			Message:   c.errMsg.Application[aperror.TheIdInTheUrlParameterIsEmpty],
			Detail:    "id in path parameter is null",
		}
	}
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.FailedToReadMessageContent,
			Message:   c.errMsg.Application[aperror.FailedToReadMessageContent],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, err := user.ToData(body.Data)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.MessageBodyIsEmpty,
			Message:   c.errMsg.Application[aperror.MessageBodyIsEmpty],
			Detail:    "read entity error: " + err.Error(),
		}
	}
	if len(data.ID) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.TheIdInTheMessageBodyIsEmpty,
			Message:   c.errMsg.Application[aperror.TheIdInTheMessageBodyIsEmpty],
			Detail:    "read entity error: id in data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if err := c.service.ChangeOwner(id, authuser.Name, data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: aperror.FailedToChangeOwner,
			Message:   c.errMsg.Application[aperror.FailedToChangeOwner],
			Detail:    fmt.Errorf("change owner error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code:      0,
			ErrorCode: aperror.Success,
		}
	}
}

func (c *controller) ChangeUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	userid := req.PathParameter("userid")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.TheIdInTheUrlParameterIsEmpty,
			Message:   c.errMsg.Application[aperror.TheIdInTheUrlParameterIsEmpty],
			Detail:    "id in path parameter is null",
		}
	}
	if len(userid) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.UserIdInUrlParameterIsEmpty,
			Message:   c.errMsg.Application[aperror.UserIdInUrlParameterIsEmpty],
			Detail:    "user id in path parameter is null",
		}
	}
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.FailedToReadMessageContent,
			Message:   c.errMsg.Application[aperror.FailedToReadMessageContent],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, err := user.ToData(body.Data)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.MessageBodyIsEmpty,
			Message:   c.errMsg.Application[aperror.MessageBodyIsEmpty],
			Detail:    "read entity error: " + err.Error(),
		}
	}
	if len(data.Role) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.TheRoleInTheMessageBodyIsEmpty,
			Message:   c.errMsg.Application[aperror.TheRoleInTheMessageBodyIsEmpty],
			Detail:    "read entity error: role in data is null",
		}
	}
	data.ID = userid
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if err := c.service.ChangeUser(id, authuser.Name, data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: aperror.FailedToChangeUser,
			Message:   c.errMsg.Application[aperror.FailedToChangeUser],
			Detail:    fmt.Errorf("change user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code:      0,
			ErrorCode: aperror.Success,
		}
	}
}

func (c *controller) GetUsers(req *restful.Request) (int, map[string]string, *user.ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	id := req.PathParameter("id")
	if len(id) == 0 {
		return http.StatusInternalServerError, nil, &user.ListResponse{
			Code:      1,
			ErrorCode: aperror.TheIdInTheUrlParameterIsEmpty,
			Message:   c.errMsg.Application[aperror.TheIdInTheUrlParameterIsEmpty],
			Detail:    "id in path parameter is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, nil, &user.ListResponse{
			Code:      1,
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if ul, isAdmin, err := c.service.GetUsers(id, authuser.Name); err != nil {
		return http.StatusInternalServerError, nil, &user.ListResponse{
			Code:      2,
			ErrorCode: "000000001",
			Message:   c.errMsg.Application["000000001"],
			Detail:    fmt.Errorf("get application user error: %+v", err).Error(),
		}
	} else {
		data, err := util.PageWrap(ul, page, size)
		if err != nil {
			return http.StatusInternalServerError, nil, &user.ListResponse{
				Code:      1,
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
				ErrorCode: aperror.QueryApplicationPagingParameterError,
				Message:   c.errMsg.Application[aperror.QueryApplicationPagingParameterError],
			}
		}
		return http.StatusOK, map[string]string{
				"isadmin": strconv.FormatBool(isAdmin),
			}, &user.ListResponse{
				Code:      0,
				ErrorCode: aperror.Success,
				Data:      data,
			}
	}
}

func (c *controller) DoStatisticsOncApps(req *restful.Request) (int, *StatisticsResponse) {
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &StatisticsResponse{
			Code:      1,
			ErrorCode: aperror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Application[aperror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	appList, err := c.service.List(util.WithNamespace(authuser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &StatisticsResponse{
			Code:      1,
			ErrorCode: aperror.QueryApplicationListFailed,
			Message:   c.errMsg.Application[aperror.QueryApplicationListFailed],
			Detail:    fmt.Sprintf("do statistics on apps error, %+v", err),
		}
	}

	data := service.Statistics{}
	data.Total = len(appList.Items)
	data.Increment, data.Percentage = c.CountAppsIncrement(appList.Items)
	return http.StatusOK, &StatisticsResponse{
		Code:      0,
		ErrorCode: "",
		Message:   "",
		Data:      data,
		Detail:    "do statistics on apps successfully",
	}
}

func (c *controller) CountAppsIncrement(apps []v1.Application) (int, string) {
	var increment int
	var percentage string
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	for _, t := range apps {
		createTime := util.NewTime(t.ObjectMeta.CreationTimestamp.Time)
		if createTime.Unix() < end.Unix() && createTime.Unix() >= start.Unix() {
			increment++
		}
	}
	total := len(apps)
	pre := float64(increment) / float64(total) * 100
	percentage = fmt.Sprintf("%.0f%s", pre, "%")

	return increment, percentage
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
