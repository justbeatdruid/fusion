package restriction

import (
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/restriction/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/errors"
	"k8s.io/klog"
	"net/http"
	rserror "github.com/chinamobile/nlpt/apiserver/resources/restriction/error"
	"github.com/chinamobile/nlpt/apiserver/resources/restriction/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.GetKubeClient(), cfg.TenantEnabled, cfg.LocalConfig),
		cfg.LocalConfig,
	}
}

type Wrapped struct {
	Code      int                  `json:"code"`
	ErrorCode string               `json:"errorCode"`
	Detail    string               `json:"detail"`
	Message   string               `json:"message"`
	Data      *service.Restriction `json:"data,omitempty"`
}

type RequestWrapped struct {
	Data *service.Restriction `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = RequestWrapped
type DeleteResponse = Wrapped
type BatchDeleteResponse = Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Detail    string      `json:"detail"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}
type PingResponse = DeleteResponse

// + update_sunyu
type UpdateRequest = Wrapped
type UpdateResponse = Wrapped

type BindRequest struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Operation string   `json:"operation"`
		Apis      []v1.Api `json:"apis"`
		Restrictions []v1.RestrictionBind `json:"restrictions"`
	} `json:"data,omitempty"`
}
type BindResponse = Wrapped

func (c *controller) CreateRestriction(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: rserror.FailedToReadMessageContent,
			Message:   c.errMsg.Restriction[rserror.FailedToReadMessageContent],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: rserror.MessageBodyIsEmpty,
			Message:   c.errMsg.Restriction[rserror.MessageBodyIsEmpty],
			Detail:    "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: rserror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Trafficcontrol["012000010"],
			Detail:    "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	body.Data.Namespace = authuser.Namespace
	if db, err, code := c.service.CreateRestriction(body.Data); err != nil {
		if errors.IsNameDuplicated(err) {
			code = rserror.DuplicateAccessControlName
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: code,
			Message:   c.errMsg.Restriction[code],
			Detail:    fmt.Errorf("create restriction error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: rserror.Success,
			Data:      db,
		}
	}
}

func (c *controller) GetRestriction(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      1,
			ErrorCode: rserror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Api[rserror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if db, err := c.service.GetRestriction(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      2,
			ErrorCode: rserror.QueryingASingleAccessControlFailedById,
			Message:   c.errMsg.Restriction[rserror.QueryingASingleAccessControlFailedById],
			Detail:    fmt.Errorf("get restriction error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code:      0,
			ErrorCode: rserror.Success,
			Data:      db,
		}
	}
}

func (c *controller) DeleteRestriction(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      1,
			ErrorCode: rserror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Api[rserror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if err := c.service.DeleteRestriction(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      2,
			ErrorCode: rserror.FailedToDeleteAccessControl,
			Message:   c.errMsg.Restriction[rserror.FailedToDeleteAccessControl],
			Detail:    fmt.Errorf("delete restriction error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      0,
			ErrorCode: rserror.Success,
		}
	}
}

func (c *controller)BatchDeleteRestriction(req *restful.Request)(int, *BatchDeleteResponse){
	body := &BindRequest{}
	if err :=req.ReadEntity(body);err!=nil{
		return http.StatusInternalServerError,&BatchDeleteResponse{
			Code:      1,
			ErrorCode: rserror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Restriction[rserror.IncorrectAuthenticationInformation],
			Detail:   fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data.Operation != "delete" {
		return http.StatusInternalServerError, &BatchDeleteResponse{
			Code:      1,
			ErrorCode: rserror.RequestParameterError,
			Message:   c.errMsg.Restriction[rserror.RequestParameterError],
			Detail:    "operation params error",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &BatchDeleteResponse{
			Code:      1,
			ErrorCode: rserror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Restriction[rserror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}

	if err:=c.service.BatchDeleteRestriction(body.Data.Restrictions,util.WithUser(authuser.Name),util.WithNamespace(authuser.Namespace));err!=nil{
		return http.StatusInternalServerError, &BatchDeleteResponse{
			Code:      2,
			ErrorCode: rserror.FailedToDeleteAccessControl,
			Message:   c.errMsg.Restriction[rserror.FailedToDeleteAccessControl],
			Detail:    fmt.Errorf("delete restriction error: %+v", err).Error(),
		}
	}else {
		return http.StatusOK, &BatchDeleteResponse{
			Code:      0,
			ErrorCode: rserror.Success,
		}
	}
}

func (c *controller) ListRestriction(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	name := req.QueryParameter("name")
	apiId:= req.QueryParameter("apiId")

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: rserror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Trafficcontrol[rserror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}

	if tc, err := c.service.ListRestriction(util.WithNameLike(name), util.WithUser(authuser.Name),
		util.WithNamespace(authuser.Namespace),util.WithId(apiId)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			ErrorCode: rserror.QueryAccessControlListFailed,
			Message:   c.errMsg.Restriction[rserror.QueryAccessControlListFailed],
			Detail:    fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var tcs RestrictionList = tc
		data, err := util.PageWrap(tcs, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      3,
				ErrorCode: rserror.QueryAccessControlPagingParameterError,
				Message:   c.errMsg.Restriction[rserror.QueryAccessControlPagingParameterError],
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: rserror.Success,
			Data:      data,
		}
	}
}

type RestrictionList []*service.Restriction

func (tcs RestrictionList) Len() int {
	return len(tcs)
}

func (tcs RestrictionList) GetItem(i int) (interface{}, error) {
	if i >= len(tcs) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return tcs[i], nil
}

// +update
func (c *controller) UpdateRestriction(req *restful.Request) (int, *UpdateResponse) {
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: rserror.FailedToReadMessageContent,
			Message:   c.errMsg.Restriction[rserror.FailedToReadMessageContent],
			Detail:    fmt.Errorf("cannot read entity: %+v, reqbody:%v, req:%v", err, reqBody, req).Error(),
		}
	}
	klog.Infof("get body restriction of updating: %+v", reqBody)
	data, ok := reqBody["data"]
	if !ok {
		klog.Infof("get body restriction of updating: %+v", reqBody)
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: rserror.MessageBodyIsEmpty,
			Message:   c.errMsg.Restriction[rserror.MessageBodyIsEmpty],
			Detail:    "read entity error: data is null",
		}
	}

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:      1,
			ErrorCode: rserror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Restriction[rserror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}

	if db, err := c.service.UpdateRestriction(req.PathParameter("id"), data,
		util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		code := rserror.UpdateAccessControlFailed
		if errors.IsNameDuplicated(err) {
			code = rserror.DuplicateAccessControlName
		}
		return http.StatusInternalServerError, &UpdateResponse{
			Code:      2,
			ErrorCode: code,
			Message:   c.errMsg.Restriction[code],
			Detail:    fmt.Errorf("update restriction error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &UpdateResponse{
			Code:      0,
			ErrorCode: rserror.Success,
			Data:      db,
		}
	}
}

func (c *controller) BindOrUnbindApis(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      1,
			ErrorCode: rserror.FailedToReadMessageContent,
			Message:   c.errMsg.Restriction[rserror.FailedToReadMessageContent],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      1,
			ErrorCode: rserror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Restriction[rserror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}

	if api, err := c.service.BindOrUnbindApis(body.Data.Operation, id, body.Data.Apis,
		util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      2,
			ErrorCode: rserror.BindingOrUnbindingAPIFailed,
			Message:   c.errMsg.Restriction[rserror.BindingOrUnbindingAPIFailed],
			Detail:    fmt.Errorf("bind api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &BindResponse{
			Code:      0,
			ErrorCode: rserror.Success,
			Data:      api,
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
