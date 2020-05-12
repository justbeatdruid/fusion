package application

import (
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	"net/http"
	"time"

	"github.com/chinamobile/nlpt/apiserver/mutex"
	"github.com/chinamobile/nlpt/apiserver/resources/application/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	"github.com/chinamobile/nlpt/pkg/util"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
	lock    mutex.Mutex
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.GetKubeClient(), cfg.TenantEnabled),
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

type CreateRequest = Wrapped
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
			ErrorCode: "002000001",
			Message:   c.errMsg.Application["002000001"],
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			Detail:    "read entity error: data is null",
			ErrorCode: "002000002",
			Message:   c.errMsg.Application["002000002"],
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: "002000003",
			Message:   c.errMsg.Application["002000003"],
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	body.Data.Namespace = authuser.Namespace
	if app, err, code := c.service.CreateApplication(body.Data); err != nil {
		if errors.IsNameDuplicated(err) {
			code = "002000021"
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
			ErrorCode: "0",
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
			ErrorCode: "002000003",
			Message:   c.errMsg.Application["002000003"],
		}
	}
	if app, err := c.service.GetApplication(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      1,
			Detail:    fmt.Errorf("get application error: %+v", err).Error(),
			ErrorCode: "002000005",
			Message:   c.errMsg.Application["002000005"],
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code:      0,
			ErrorCode: "0",
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
			ErrorCode: "002000001",
			Message:   c.errMsg.Application["002000001"],
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
			ErrorCode: "002000002",
			Message:   c.errMsg.Application["002000002"],
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: "002000003",
			Message:   c.errMsg.Application["002000003"],
		}
	}
	if app, err := c.service.PatchApplication(req.PathParameter("id"), data, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		code := "002000007"
		if errors.IsNameDuplicated(err) {
			code = "002000021"
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
			ErrorCode: "0",
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
			ErrorCode: "002000003",
			Message:   c.errMsg.Application["002000003"],
		}
	}
	if app, err := c.service.DeleteApplication(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      1,
			Detail:    fmt.Errorf("delete application error: %+v", err).Error(),
			ErrorCode: "002000006",
			Message:   c.errMsg.Application["002000006"],
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      app,
		}
	}
}

func (c *controller) ListApplication(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	group := req.QueryParameter("group")
	name := req.QueryParameter("name")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: "002000003",
			Message:   c.errMsg.Application["002000003"],
		}
	}
	if app, err := c.service.ListApplication(util.WithGroup(group), util.WithNameLike(name),
		util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			Detail:    fmt.Errorf("list application error: %+v", err).Error(),
			ErrorCode: "002000008",
			Message:   c.errMsg.Application["002000008"],
		}
	} else {
		var apps ApplicationList = app
		data, err := util.PageWrap(apps, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      1,
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
				ErrorCode: "002000009",
				Message:   c.errMsg.Application["002000009"],
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: "0",
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
			ErrorCode: "002000001",
			Message:   c.errMsg.Application["002000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, err := user.ToData(body.Data)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000002",
			Message:   c.errMsg.Application["002000002"],
			Detail:    "read entity error: " + err.Error(),
		}
	}
	if len(data.ID) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000011",
			Message:   c.errMsg.Application["002000011"],
			Detail:    "read entity error: id in data is null",
		}
	}
	if len(data.Role) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000012",
			Message:   c.errMsg.Application["002000012"],
			Detail:    "read entity error: role in data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000003",
			Message:   c.errMsg.Application["002000003"],
			Detail:    "auth model error",
		}
	}
	if err := c.service.AddUser(id, authuser.Name, data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: "002000013",
			Message:   c.errMsg.Application["002000013"],
			Detail:    fmt.Errorf("add user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code:      0,
			ErrorCode: "0",
		}
	}
}

func (c *controller) RemoveUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	userid := req.PathParameter("userid")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000014",
			Message:   c.errMsg.Application["002000014"],
			Detail:    "id in path parameter is null",
		}
	}
	if len(userid) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000015",
			Message:   c.errMsg.Application["002000015"],
			Detail:    "user id in path parameter is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000003",
			Message:   c.errMsg.Application["002000003"],
			Detail:    "auth model error",
		}
	}
	if err := c.service.RemoveUser(id, authuser.Name, userid); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: "002000016",
			Message:   c.errMsg.Application["002000016"],
			Detail:    fmt.Errorf("remove user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code:      0,
			ErrorCode: "0",
		}
	}
}

func (c *controller) ChangeOwner(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000014",
			Message:   c.errMsg.Application["002000014"],
			Detail:    "id in path parameter is null",
		}
	}
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000001",
			Message:   c.errMsg.Application["002000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, err := user.ToData(body.Data)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000002",
			Message:   c.errMsg.Application["002000002"],
			Detail:    "read entity error: " + err.Error(),
		}
	}
	if len(data.ID) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000011",
			Message:   c.errMsg.Application["002000011"],
			Detail:    "read entity error: id in data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000003",
			Message:   c.errMsg.Application["002000003"],
			Detail:    "auth model error",
		}
	}
	if err := c.service.ChangeOwner(id, authuser.Name, data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: "002000018",
			Message:   c.errMsg.Application["002000018"],
			Detail:    fmt.Errorf("change owner error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code:      0,
			ErrorCode: "0",
		}
	}
}

func (c *controller) ChangeUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	userid := req.PathParameter("userid")
	if len(id) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000014",
			Message:   c.errMsg.Application["002000014"],
			Detail:    "id in path parameter is null",
		}
	}
	if len(userid) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000015",
			Message:   c.errMsg.Application["002000015"],
			Detail:    "user id in path parameter is null",
		}
	}
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000001",
			Message:   c.errMsg.Application["002000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, err := user.ToData(body.Data)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000002",
			Message:   c.errMsg.Application["002000002"],
			Detail:    "read entity error: " + err.Error(),
		}
	}
	if len(data.Role) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000012",
			Message:   c.errMsg.Application["002000012"],
			Detail:    "read entity error: role in data is null",
		}
	}
	data.ID = userid
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "002000003",
			Message:   c.errMsg.Application["002000003"],
			Detail:    "auth model error",
		}
	}
	if err := c.service.ChangeUser(id, authuser.Name, data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: "002000017",
			Message:   c.errMsg.Application["002000017"],
			Detail:    fmt.Errorf("change user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code:      0,
			ErrorCode: "0",
		}
	}
}

func (c *controller) DoStatisticsOncApps(req *restful.Request) (int, *StatisticsResponse) {
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &StatisticsResponse{
			Code:      1,
			ErrorCode: "002000003",
			Message:   c.errMsg.Application["002000003"],
			Detail:    "auth model error",
		}
	}
	appList, err := c.service.List(util.WithNamespace(authuser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &StatisticsResponse{
			Code:      1,
			ErrorCode: "002000008",
			Message:   c.errMsg.Application["002000008"],
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
