package apiplugin

import (
	"fmt"
	"net/http"
	"time"

	"github.com/chinamobile/nlpt/apiserver/resources/apiplugin/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/names"
	//"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.TenantEnabled, cfg.Database),
	}
}

type Wrapped struct {
	Code      int                `json:"code"`
	ErrorCode string             `json:"errorCode"`
	Message   string             `json:"message"`
	Detail    string             `json:"detail"`
	Data      *service.ApiPlugin `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
	Detail    string `json:"detail"`
}
type GetResponse = Wrapped
type ListResponse = struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Message   string      `json:"message"`
	Detail    string      `json:"detail"`
	Data      interface{} `json:"data"`
}

type BindRequest struct {
	Data service.BindReq `json:"data"`
}

func (c *controller) CreateApiPlugin(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		code := "000000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "000000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    "auth model error",
		}
	}
	body.Data.Id = names.NewID()
	body.Data.Namespace = authuser.Namespace
	body.Data.User = authuser.Name
	body.Data.CreatedAt = time.Now()
	body.Data.ReleasedAt = time.Now()
	if apl, err := c.service.CreateApiPlugin(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: "",
			Message:   "",
			Detail:    fmt.Errorf("create apigroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: apl,
		}
	}
}

func (c *controller) GetApiPlugin(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if apl, err := c.service.GetApiPlugin(id); err != nil {
		code := "000000007"
		return http.StatusInternalServerError, &GetResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    fmt.Errorf("get apigroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: apl,
		}
	}
}

func (c *controller) DeleteApiPlugin(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if err := c.service.DeleteApiPlugin(id); err != nil {
		code := "000000002"
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    fmt.Errorf("delete apigroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:   0,
			Detail: "",
		}
	}
}

func (c *controller) ListApiPlugin(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	condition := service.ApiPlugin{
		Name:   req.QueryParameter("name"),
		Status: req.QueryParameter("status"),
		Type:   req.QueryParameter("type"),
	}

	authuser, err := auth.GetAuthUser(req)
	if len(req.QueryParameter("tenant")) > 0 {
		condition.Namespace = authuser.Namespace
	}
	//查询某个租户下的api分组，当前不需要提供接口查询所有
	condition.Namespace = authuser.Namespace
	if err != nil {
		code := "000000006"
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    "auth model error",
		}
	}
	if pl, err := c.service.ListApiPlugin(condition); err != nil {
		code := "000000002"
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    fmt.Errorf("list apigroups error: %+v", err).Error(),
		}
	} else {
		var pls ProductList = pl
		data, err := util.PageWrap(pls, page, size)
		if err != nil {
			code := "000000005"
			return http.StatusInternalServerError, &ListResponse{
				Code:      1,
				ErrorCode: code,
				Message:   "",
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: data,
		}
	}
}

func (c *controller) UpdateApiPlugin(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		code := "000000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "000000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    "auth model error",
		}
	}
	body.Data.Namespace = authuser.Namespace
	body.Data.User = authuser.Name
	id := req.PathParameter("id")
	if apl, err := c.service.UpdateApiPlugin(body.Data, id); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: "",
			Message:   "",
			Detail:    fmt.Errorf("update apigroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: apl,
		}
	}
}

func (c *controller) UpdateApiPluginStatus(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		code := "000000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "000000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    "auth model error",
		}
	}
	body.Data.Namespace = authuser.Namespace
	body.Data.User = authuser.Name
	if apl, err := c.service.UpdateApiPluginStatus(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: "",
			Message:   "",
			Detail:    fmt.Errorf("update apigroup error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: apl,
		}
	}
}

func (c *controller) BindOrUnbindApis(req *restful.Request) (int, *CreateResponse) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		code := "000000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data.Operation != "bind" && body.Data.Operation != "unbind" {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "read entity error: data operation is invaild",
		}
	}
	if body.Data.Type != service.ApiType && body.Data.Type != service.ServiceunitType {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "read entity error: data type is invaild",
		}
	}

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "000000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    "auth model error",
		}
	}
	groupID := req.PathParameter("id")
	if err := c.service.BatchBindOrRelease(groupID, body.Data, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: "001000013",
			Message:   "",
			Detail:    fmt.Errorf("bind api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: "0",
			//Data:      api,
		}
	}
}
func (c *controller) ListRelationsByApiPlugin(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	types := req.QueryParameter("type")
	id := req.PathParameter("id")

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			Detail:    "auth model error",
			ErrorCode: "",
			Message:   "",
		}
	}
	if api, err := c.service.ListRelationsByApiPlugin(id, types, util.WithUser(authuser.Name),
		util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			Detail:    fmt.Errorf("list api error: %+v", err).Error(),
			ErrorCode: "",
			Message:   "",
		}
	} else {
		var apis ApiResList = api
		data, err := util.PageWrap(apis, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      1,
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
				ErrorCode: "",
				Message:   "",
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      data,
		}
	}
}

type ProductList []*service.ApiPlugin
type ApiResList []*service.ApiRes

func (apls ProductList) Len() int {
	return len(apls)
}

func (apls ProductList) GetItem(i int) (interface{}, error) {
	if i >= len(apls) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return apls[i], nil
}

func (apls ProductList) Less(i, j int) bool {
	return apls[i].ReleasedAt.After(apls[j].ReleasedAt)
}

func (apls ProductList) Swap(i, j int) {
	apls[i], apls[j] = apls[j], apls[i]
}

func (apis ApiResList) Len() int {
	return len(apis)
}

func (apis ApiResList) GetItem(i int) (interface{}, error) {
	if i >= len(apis) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return apis[i], nil
}

func (apis ApiResList) Swap(i, j int) {
	apis[i], apis[j] = apis[j], apis[i]
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
