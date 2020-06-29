package trafficcontrol

import (
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/errors"
	"net/http"
	"strings"
	tcerror "github.com/chinamobile/nlpt/apiserver/resources/trafficcontrol/error"
	"github.com/chinamobile/nlpt/apiserver/resources/trafficcontrol/service"
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
	Code      int                     `json:"code"`
	ErrorCode string                  `json:"errorCode"`
	Detail    string                  `json:"detail"`
	Message   string                  `json:"message"`
	Data      *service.Trafficcontrol `json:"data,omitempty"`
}

type RequestWrapped struct {
	Data *service.Trafficcontrol `json:"data,omitempty"`
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
		Trafficcontrols []v1.TrafficcontrolBind `json:"trafficcontrols"`
	} `json:"data,omitempty"`
}
type BindResponse = Wrapped

func (c *controller) CreateTrafficcontrol(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: tcerror.FailedToReadMessageContent,
			Message:   c.errMsg.Trafficcontrol[tcerror.FailedToReadMessageContent],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: tcerror.MessageBodyIsEmpty,
			Message:   c.errMsg.Trafficcontrol[tcerror.MessageBodyIsEmpty],
			Detail:    "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: tcerror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Trafficcontrol[tcerror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	body.Data.Namespace = authuser.Namespace
	if db, err, code := c.service.CreateTrafficcontrol(body.Data); err != nil {
		if strings.Contains(err.Error(), "必须小于每") {
			comma := strings.Index(err.Error(), "每")
			return http.StatusInternalServerError, &CreateResponse{
				Code:      2,
				ErrorCode: tcerror.FailedToCreateTrafficControl,
				Message:   err.Error()[comma:],
			}
		}
		if errors.IsNameDuplicated(err) {
			code = tcerror.FlowControlWithDuplicateName
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: code,
			Message:   c.errMsg.Trafficcontrol[code],
			Detail:    fmt.Errorf("create trafficcontrol error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: tcerror.Success,
			Data:      db,
		}
	}
}

func (c *controller) GetTrafficcontrol(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: tcerror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Trafficcontrol[tcerror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if db, err := c.service.GetTrafficcontrol(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      2,
			ErrorCode: tcerror.QuerySingleFlowControlFailureBasedOnId,
			Message:   c.errMsg.Trafficcontrol[tcerror.IncorrectAuthenticationInformation],
			Detail:    fmt.Errorf("get trafficcontrol error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code:      0,
			ErrorCode: tcerror.Success,
			Data:      db,
		}
	}
}

func (c *controller) DeleteTrafficcontrol(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: tcerror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Api[tcerror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if err := c.service.DeleteTrafficcontrol(id, util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      2,
			ErrorCode: tcerror.FailedToDeleteFlowControl,
			Message:   c.errMsg.Trafficcontrol[tcerror.FailedToDeleteFlowControl],
			Detail:    fmt.Errorf("delete trafficcontrol error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      0,
			ErrorCode: tcerror.Success,
		}
	}
}
func (c *controller)BatchDeleteTrafficcontrol(req *restful.Request) (int,*BatchDeleteResponse){
	body :=&BindRequest{}
	if err :=req.ReadEntity(body);err!=nil{
		return http.StatusInternalServerError,&BatchDeleteResponse{
			Code:      1,
			ErrorCode: tcerror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Trafficcontrol["007000013"],
			Detail:   fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data.Operation != "delete" {
		return http.StatusInternalServerError, &BatchDeleteResponse{
			Code:      1,
			ErrorCode: tcerror.RequestParameterError,
			Message:   c.errMsg.Trafficcontrol[tcerror.RequestParameterError],
			Detail:    "operation params error",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &BatchDeleteResponse{
			Code:      1,
			ErrorCode: tcerror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Trafficcontrol[tcerror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if err:=c.service.BatchDeleteTrafficcontrol(body.Data.Trafficcontrols,util.WithUser(authuser.Name),util.WithNamespace(authuser.Namespace));err!=nil{
		return http.StatusInternalServerError, &BatchDeleteResponse{
			Code:      2,
			ErrorCode: tcerror.FailedToDeleteFlowControl,
			Message:   c.errMsg.Restriction[tcerror.FailedToDeleteFlowControl],
			Detail:    fmt.Errorf("delete trafficcontrol error: %+v", err).Error(),
		}
	}else {
		return http.StatusOK, &BatchDeleteResponse{
			Code:      0,
			ErrorCode: tcerror.Success,
		}
	}
}
func (c *controller) ListTrafficcontrol(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	name := req.QueryParameter("name")
	apiId := req.QueryParameter("apiId")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: tcerror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Trafficcontrol[tcerror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if tc, err := c.service.ListTrafficcontrol(util.WithNameLike(name), util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace), util.WithId(apiId)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			ErrorCode: tcerror.QueryFlowControlListFailed,
			Message:   c.errMsg.Trafficcontrol[tcerror.QueryFlowControlListFailed],
			Detail:    fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var tcs TrafficcontrolList = tc
		data, err := util.PageWrap(tcs, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      3,
				ErrorCode: tcerror.QueryFlowControlPagingParameterError,
				Message:   c.errMsg.Trafficcontrol[tcerror.QueryFlowControlPagingParameterError],
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: tcerror.Success,
			Data:      data,
		}
	}
}

type TrafficcontrolList []*service.Trafficcontrol

func (tcs TrafficcontrolList) Len() int {
	return len(tcs)
}

func (tcs TrafficcontrolList) GetItem(i int) (interface{}, error) {
	if i >= len(tcs) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return tcs[i], nil
}

// +update_sunyu
func (c *controller) UpdateTrafficcontrol(req *restful.Request) (int, *UpdateResponse) {
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: tcerror.FailedToReadMessageContent,
			Message:   c.errMsg.Trafficcontrol[tcerror.FailedToReadMessageContent],
			Detail:    fmt.Errorf("cannot read entity: %+v, reqbody:%v, req:%v", err, reqBody, req).Error(),
		}
	}
	data, ok := reqBody["data"]
	if !ok {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: tcerror.MessageBodyIsEmpty,
			Message:   c.errMsg.Trafficcontrol[tcerror.MessageBodyIsEmpty],
			Detail:    "read entity error: data is null",
		}
	}

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: tcerror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Trafficcontrol[tcerror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}

	if db, err := c.service.UpdateTrafficcontrol(req.PathParameter("id"), data,
		util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		code := tcerror.UpdateFlowControlFailed
		if strings.Contains(err.Error(), "必须小于每") {
			comma := strings.Index(err.Error(), "每")
			return http.StatusInternalServerError, &CreateResponse{
				Code:      2,
				ErrorCode: tcerror.FailedToCreateTrafficControl,
				Message:   err.Error()[comma:],
			}
		}
		if errors.IsNameDuplicated(err) {
			code = tcerror.FlowControlWithDuplicateName
		}
		return http.StatusInternalServerError, &UpdateResponse{
			Code:      2,
			ErrorCode: code,
			Message:   c.errMsg.Trafficcontrol[code],
			Detail:    fmt.Errorf("update trafficcontrol error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &UpdateResponse{
			Code:      0,
			ErrorCode: tcerror.Success,
			Data:      db,
		}
	}
}

func (c *controller) BindOrUnbindApis(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      1,
			ErrorCode: tcerror.FailedToReadMessageContent,
			Message:   c.errMsg.Trafficcontrol[tcerror.FailedToReadMessageContent],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	trafficID := req.PathParameter("id")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      1,
			ErrorCode: tcerror.IncorrectAuthenticationInformation,
			Message:   c.errMsg.Trafficcontrol[tcerror.IncorrectAuthenticationInformation],
			Detail:    "auth model error",
		}
	}
	if api, err := c.service.BindOrUnbindApis(body.Data.Operation, trafficID, body.Data.Apis,
		util.WithUser(authuser.Name), util.WithNamespace(authuser.Namespace)); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      2,
			ErrorCode: tcerror.BindingOrUnbindingAPIFailed,
			Message:   c.errMsg.Trafficcontrol[tcerror.BindingOrUnbindingAPIFailed],
			Detail:    fmt.Errorf("bind or unbind api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &BindResponse{
			Code:      0,
			ErrorCode: tcerror.Success,
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
