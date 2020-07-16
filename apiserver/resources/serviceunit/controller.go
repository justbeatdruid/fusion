package serviceunit

import (
	"fmt"
	"io"
	"k8s.io/klog"
	"net/http"
	"os"
	"strconv"

	"github.com/chinamobile/nlpt/apiserver/concurrency"
	"github.com/chinamobile/nlpt/apiserver/resources/serviceunit/service"
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
	lock    concurrency.Mutex
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.GetKubeClient(), cfg.TenantEnabled, cfg.LocalConfig, cfg.Database),
		cfg.LocalConfig,
		cfg.Mutex,
	}
}

type Wrapped struct {
	Code      int                  `json:"code"`
	ErrorCode string               `json:"errorCode"`
	Detail    string               `json:"detail"`
	Message   string               `json:"message"`
	Data      *service.Serviceunit `json:"data,omitempty"`
}

type RequestWrapped struct {
	Data *service.Serviceunit `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = RequestWrapped
type DeleteResponse = Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Detail    string      `json:"detail"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
}
type PingResponse = DeleteResponse

// + update_sunyu
type UpdateRequest = Wrapped
type UpdateResponse = Wrapped
type TestFnResponse = ImportResponse
type ImportResponse struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
	Detail    string `json:"detail"`
	Data      string `json:"data,omitempty"`
}

type GetFnLogsResponse struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
	Detail    string `json:"detail"`
	Logs      string `json:"logs"`
}

const UploadPath string = "/data/upload/serviceunit/"

func (c *controller) CreateServiceunit(req *restful.Request) (int, *CreateResponse) {
	unlock, err := c.lock.Lock("serviceunit")
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
			ErrorCode: "008000001",
			Message:   c.errMsg.Serviceunit["008000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "008000002",
			Message:   c.errMsg.Serviceunit["008000002"],
			Detail:    "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	body.Data.Namespace = authuser.Namespace
	if su, err, code := c.service.CreateServiceunit(body.Data); err != nil {
		if errors.IsNameDuplicated(err) {
			code = "008000021"
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: code,
			Message:   c.errMsg.Serviceunit[code],
			Detail:    fmt.Errorf("create serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      su,
		}
	}
}

func (c *controller) GetServiceunit(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    "auth model error",
		}
	}
	if su, err := c.service.GetServiceunit(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      2,
			ErrorCode: "008000005",
			Message:   c.errMsg.Serviceunit["008000005"],
			Detail:    fmt.Errorf("get serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      su,
		}
	}
}

func (c *controller) PatchServiceunit(req *restful.Request) (int, *DeleteResponse) {
	unlock, err := c.lock.Lock("serviceunit")
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
			ErrorCode: "008000001",
			Message:   c.errMsg.Serviceunit["008000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, ok := reqBody["data,omitempty"]
	if !ok {
		data, ok = reqBody["data"]
	}
	if !ok {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "008000002",
			Message:   c.errMsg.Serviceunit["008000002"],
			Detail:    "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    "auth model error",
		}
	}
	if su, err := c.service.PatchServiceunit(req.PathParameter("id"), data, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		code := "008000007"
		if errors.IsNameDuplicated(err) {
			code = "008000021"
		}
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      2,
			ErrorCode: code,
			Message:   c.errMsg.Serviceunit[code],
			Detail:    fmt.Errorf("patch serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      su,
		}
	}
}

func (c *controller) DeleteServiceunit(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    "auth model error",
		}
	}
	if data, err := c.service.DeleteServiceunit(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      2,
			ErrorCode: "008000006",
			Message:   c.errMsg.Serviceunit["008000006"],
			Detail:    fmt.Errorf("delete serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      data,
		}
	}
}

func (c *controller) ListServiceunit(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	group := req.QueryParameter("group")
	name := req.QueryParameter("name")
	stype := req.QueryParameter("type")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    "auth model error",
		}
	}
	if su, err := c.service.ListServiceunit(util.WithGroup(group), util.WithNameLike(name), util.WithUser(authuser.Name),
		util.WithNamespace(authuser.Namespace), util.WithStype(stype)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			ErrorCode: "008000008",
			Message:   c.errMsg.Serviceunit["008000008"],
			Detail:    fmt.Errorf("list serviceunit error: %+v", err).Error(),
		}
	} else {
		var sus ServiceunitList = su
		data, err := util.PageWrap(sus, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      3,
				ErrorCode: "008000009",
				Message:   c.errMsg.Serviceunit["008000009"],
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      data,
		}
	}
}

type ServiceunitList []*service.Serviceunit

func (sus ServiceunitList) Len() int {
	return len(sus)
}

func (sus ServiceunitList) GetItem(i int) (interface{}, error) {
	if i >= len(sus) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return sus[i], nil
}

func (sus ServiceunitList) Less(i, j int) bool {
	return sus[i].CreatedAt.Time.After(sus[j].CreatedAt.Time)
}

func (sus ServiceunitList) Swap(i, j int) {
	sus[i], sus[j] = sus[j], sus[i]
}

func (c *controller) PublishServiceunit(req *restful.Request) (int, *CreateResponse) {
	id := req.PathParameter("id")
	body := &struct {
		Published bool `json:"published"`
	}{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "008000001",
			Message:   c.errMsg.Serviceunit["008000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    "auth model error",
		}
	}
	if su, err := c.service.PublishServiceunit(id, body.Published, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: "008000010",
			Message:   c.errMsg.Serviceunit["008000010"],
			Detail:    fmt.Errorf("publish serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      su,
		}
	}
}

/*
// +update_sunyu
func (c *controller) UpdateServiceunit(req *restful.Request) (int, *UpdateResponse) {
	if true {
		return http.StatusNotImplemented, &UpdateResponse{
			ErrorCode:    "008000001",
			Message: c.errMsg.Serviceunit["008000001"],
			Detail: "interface not supported",
		}
	}
	body := &UpdateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			ErrorCode:    "008000001",
			Message: c.errMsg.Serviceunit["008000001"],
			Detail: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &UpdateResponse{
			ErrorCode:    "008000002",
			Message: c.errMsg.Serviceunit["008000002"],
			Detail: "read entity error: data is null",
		}
	}
	id := req.PathParameter("id")
	if su, err := c.service.UpdateServiceunit(body.Data, id); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			ErrorCode:    "008000007",
			Message: c.errMsg.Serviceunit["008000007"],
			Detail: fmt.Errorf("update serviceunit error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &UpdateResponse{
			ErrorCode: "0",
			Data: su,
		}
	}
}

*/

func (c *controller) AddUser(req *restful.Request) (int, *user.UserResponse) {
	id := req.PathParameter("id")
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000001",
			Message:   c.errMsg.Serviceunit["008000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, err := user.ToData(body.Data)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000002",
			Message:   c.errMsg.Serviceunit["008000002"],
			Detail:    "read entity error: " + err.Error(),
		}
	}
	if len(data.ID) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000011",
			Message:   c.errMsg.Serviceunit["008000011"],
			Detail:    "read entity error: id in data is null",
		}
	}
	if len(data.Role) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000012",
			Message:   c.errMsg.Serviceunit["008000012"],
			Detail:    "read entity error: role in data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    "auth model error",
		}
	}
	if err := c.service.AddUser(id, authuser.Name, data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: "008000013",
			Message:   c.errMsg.Serviceunit["008000013"],
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
			ErrorCode: "008000014",
			Message:   c.errMsg.Serviceunit["008000014"],
			Detail:    "id in path parameter is null",
		}
	}
	if len(userid) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000015",
			Message:   c.errMsg.Serviceunit["008000015"],
			Detail:    "user id in path parameter is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    "auth model error",
		}
	}
	if err := c.service.RemoveUser(id, authuser.Name, userid); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: "008000016",
			Message:   c.errMsg.Serviceunit["008000016"],
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
			ErrorCode: "008000014",
			Message:   c.errMsg.Serviceunit["008000014"],
			Detail:    "id in path parameter is null",
		}
	}
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000001",
			Message:   c.errMsg.Serviceunit["008000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, err := user.ToData(body.Data)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000002",
			Message:   c.errMsg.Serviceunit["008000002"],
			Detail:    "read entity error: " + err.Error(),
		}
	}
	if len(data.ID) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000011",
			Message:   c.errMsg.Serviceunit["008000011"],
			Detail:    "read entity error: id in data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    "auth model error",
		}
	}
	if err := c.service.ChangeOwner(id, authuser.Name, data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: "008000018",
			Message:   c.errMsg.Serviceunit["008000018"],
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
			ErrorCode: "008000014",
			Message:   c.errMsg.Serviceunit["008000014"],
			Detail:    "id in path parameter is null",
		}
	}
	if len(userid) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000015",
			Message:   c.errMsg.Serviceunit["008000015"],
			Detail:    "user id in path parameter is null",
		}
	}
	body := &user.UserRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000001",
			Message:   c.errMsg.Serviceunit["008000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	data, err := user.ToData(body.Data)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000002",
			Message:   c.errMsg.Serviceunit["008000002"],
			Detail:    "read entity error: " + err.Error(),
		}
	}
	if len(data.Role) == 0 {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000012",
			Message:   c.errMsg.Serviceunit["008000012"],
			Detail:    "read entity error: role in data is null",
		}
	}
	data.ID = userid
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    "auth model error",
		}
	}
	if err := c.service.ChangeUser(id, authuser.Name, data); err != nil {
		return http.StatusInternalServerError, &user.UserResponse{
			Code:      2,
			ErrorCode: "008000017",
			Message:   c.errMsg.Serviceunit["008000017"],
			Detail:    fmt.Errorf("change user error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &user.UserResponse{
			Code:      0,
			ErrorCode: "0",
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
			ErrorCode: "008000014",
			Message:   c.errMsg.Application["008000014"],
			Detail:    "id in path parameter is null",
		}
	}
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, nil, &user.ListResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Application["008000003"],
			Detail:    "auth model error",
		}
	}
	if ul, isAdmin, err := c.service.GetUsers(id, authUser.Name); err != nil {
		return http.StatusInternalServerError, nil, &user.ListResponse{
			Code:      2,
			ErrorCode: "000000001",
			Message:   c.errMsg.Application["000000001"],
			Detail:    fmt.Errorf("get serviceunit user error: %+v", err).Error(),
		}
	} else {
		data, err := util.PageWrap(ul, page, size)
		if err != nil {
			return http.StatusInternalServerError, nil, &user.ListResponse{
				Code:      1,
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
				ErrorCode: "008000009",
				Message:   c.errMsg.Application["008000009"],
			}
		}
		return http.StatusOK, map[string]string{
				"isadmin": strconv.FormatBool(isAdmin),
			}, &user.ListResponse{
				Code:      0,
				ErrorCode: "0",
				Data:      data,
			}
	}
}

func (c *controller) ImportServiceunits(req *restful.Request, response *restful.Response) (int, *ImportResponse) {
	if err := req.Request.ParseMultipartForm(32 << 20); err != nil {
		if err != nil {
			return http.StatusInternalServerError, &ImportResponse{
				Code:      1,
				ErrorCode: "008000022",
				Message:   c.errMsg.Serviceunit["008000022"],
				Detail:    fmt.Sprintf("import file  error: %+v", err),
			}
		}
	}
	file, handler, err := req.Request.FormFile("uploadfile")
	klog.Infof("File name: %+v", handler.Filename)
	if err != nil {
		klog.Error("File error.")
		return http.StatusInternalServerError, &ImportResponse{
			Code:      1,
			ErrorCode: "008000023",
			Message:   c.errMsg.Serviceunit["008000023"],
			Detail:    fmt.Sprintf("invalid file format: %+v", err),
		}
	}

	defer file.Close()
	f, err := os.OpenFile(UploadPath+handler.Filename, os.O_RDWR|os.O_CREATE, 0666)
	if _, err = io.Copy(f, file); err != nil {
		klog.Error("import failed,copy file error.")
		return http.StatusInternalServerError, &ImportResponse{
			Code:      1,
			ErrorCode: "008000024",
			Message:   c.errMsg.Serviceunit["008000024"],
			Detail:    fmt.Sprintf("import failed, copy file error: %+v", err),
		}
	}

	return http.StatusOK, &ImportResponse{
		Code:      0,
		ErrorCode: "0",
		Data:      UploadPath + handler.Filename,
	}
}

func (c *controller) GetFnLogs(req *restful.Request, response *restful.Response) (int, *GetFnLogsResponse) {
	fnName := req.QueryParameter("fnName")
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &GetFnLogsResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	nameSpace := authUser.Namespace
	logs, err := c.service.GetLogs(fnName, nameSpace)
	if err != nil {
		return http.StatusInternalServerError, &GetFnLogsResponse{
			Code:      1,
			ErrorCode: "008000025",
			Message:   c.errMsg.Serviceunit["008000025"],
			Detail:    fmt.Sprintf("get function logs error: %+v", err),
		}
	}
	return http.StatusOK, &GetFnLogsResponse{
		Code: 0,
		Logs: logs,
	}
}

//函数调式
func (c *controller) TestFn(req *restful.Request, response *restful.Response) (int, *TestFnResponse) {
	body := &service.TestFunction{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &TestFnResponse{
			Code:      1,
			ErrorCode: "008000001",
			Message:   c.errMsg.Serviceunit["008000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &TestFnResponse{
			Code:      1,
			ErrorCode: "008000003",
			Message:   c.errMsg.Serviceunit["008000003"],
			Detail:    fmt.Sprintf("auth model error: %+v", err),
		}
	}
	namespace := authUser.Namespace
	data, err := c.service.TestFn(body, namespace)
	if err != nil {
		return http.StatusInternalServerError, &TestFnResponse{
			Code:      1,
			ErrorCode: "008000026",
			Message:   c.errMsg.Serviceunit["008000026"],
			Detail:    fmt.Sprintf("test function error: %+v", err),
		}
	}
	return http.StatusOK, &TestFnResponse{
		Code: 0,
		Data: data,
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
