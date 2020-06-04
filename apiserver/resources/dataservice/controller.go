package dataservice

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/chinamobile/nlpt/pkg/auth"

	"github.com/chinamobile/nlpt/apiserver/resources/dataservice/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/errors"
	"k8s.io/klog"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetKubeClient(), cfg.GetDynamicClient()),
		cfg.LocalConfig,
	}
}

//Wrapped ...
type Wrapped struct {
	Code      int                  `json:"code"`
	ErrorCode string               `json:"errorCode"`
	Detail    string               `json:"detail"`
	Message   string               `json:"message"`
	Data      *service.Dataservice `json:"data,omitempty"`
}

//CreateResponse ...
type CreateResponse = Wrapped

//CreateResponse ...
type UpdateResponse = Wrapped

//CreateRequest ...
type CreateRequest = Wrapped

//CreateRequest ...
type UpdateRequest = Wrapped

//DeleteResponse ...
type DeleteResponse struct {
	Code      int                  `json:"code"`
	ErrorCode string               `json:"errorCode"`
	Detail    string               `json:"detail"`
	Message   string               `json:"message"`
	Data      *service.Dataservice `json:"data,omitempty"`
}

//GetResponse ...
type GetResponse = Wrapped

//ListResponse ...
type ListResponse = struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Detail    string      `json:"detail"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}

type OpertationRequest = struct {
	Code      int                   `json:"code"`
	ErrorCode string                `json:"errorCode"`
	Detail    string                `json:"detail"`
	Message   string                `json:"message"`
	Data      *service.OperationReq `json:"data,omitempty"`
}

//PingResponse ...
type PingResponse = DeleteResponse

//CreateDataservice ...
func (c *controller) CreateDataservice(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "005000001",
			Message:   c.errMsg.DataService["005000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "005000002",
			Message:   c.errMsg.DataService["005000002"],
			Detail:    "read entity error: data is null",
		}
	}

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "005000003",
			Message:   c.errMsg.DataService["005000003"],
			Detail:    "auth model error",
		}
	}

	ds, err := c.service.CreateDataservice(body.Data, authuser.Name, authuser.Namespace)

	if err == nil {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      ds,
		}
	}
	if errors.IsNameDuplicated(err) {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: "005000004",
			Message:   c.errMsg.DataService["005000004"],
			Detail:    fmt.Errorf("create database error: %+v", err).Error(),
		}
	}

	return http.StatusInternalServerError, &CreateResponse{
		Code:      2,
		ErrorCode: "005000005",
		Message:   c.errMsg.DataService["005000005"],
		Detail:    fmt.Errorf("create database error: %+v", err).Error(),
	}

}

//OperationDataservice ...
func (c *controller) OperationDataservice(req *restful.Request) (int, *CreateResponse) {
	body := &OpertationRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "005000001",
			Message:   c.errMsg.DataService["005000001"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "005000002",
			Message:   c.errMsg.DataService["005000002"],
			Detail:    "read entity error: data is null",
		}
	}

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "005000003",
			Message:   c.errMsg.DataService["005000003"],
			Detail:    "auth model error",
		}
	}

	if err = c.service.OperationDataservice(body.Data, authuser.Name, authuser.Namespace); err == nil {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: "0",
		}
	}

	return http.StatusInternalServerError, &CreateResponse{
		Code:      2,
		ErrorCode: "005000005",
		Message:   c.errMsg.DataService["005000005"],
		Detail:    fmt.Errorf("update database error: %+v", err).Error(),
	}

}

//GetDataservice ...
func (c *controller) GetDataservice(req *restful.Request) (int, *GetResponse) {
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "005000003",
			Message:   c.errMsg.DataService["005000003"],
			Detail:    "auth model error",
		}
	}

	ds, err := c.service.GetDataservice(req.PathParameter("id"), authuser.Name, authuser.Namespace)
	if err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      1,
			ErrorCode: "000000001",
			Message:   c.errMsg.Common["000000001"],
			Detail:    fmt.Errorf("get database error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &GetResponse{
		Code:      0,
		ErrorCode: "0",
		Data:      ds,
	}

}

//DeleteDataservice ...
func (c *controller) DeleteDataservice(req *restful.Request) (int, *DeleteResponse) {
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      1,
			ErrorCode: "005000003",
			Message:   c.errMsg.DataService["005000003"],
			Detail:    "auth model error",
		}
	}

	err = c.service.DeleteDataservice(req.PathParameter("id"), authuser.Name, authuser.Namespace)
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      1,
			ErrorCode: "005000008",
			Message:   c.errMsg.DataService["005000008"],
			Detail:    fmt.Errorf("delete database error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &DeleteResponse{
		Code:      0,
		ErrorCode: "0",
		Message:   "",
	}

}

//DeleteDataservice ...
func (c *controller) UpdateDateService(req *restful.Request) (int, *UpdateResponse) {
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		klog.Errorf("read entity failed, err :%v", err)
		return http.StatusInternalServerError, &UpdateResponse{
			Code:      1,
			ErrorCode: "005000001",
			Message:   c.errMsg.DataService["005000001"],
			Detail:    fmt.Errorf("read entity failed: %+v", err).Error(),
		}
	}

	reqData, ok := reqBody["data"]
	if !ok {
		klog.Errorf("read data failed, reqBody :%v", reqBody)
		return http.StatusInternalServerError, &UpdateResponse{
			Code:      1,
			ErrorCode: "005000002",
			Message:   c.errMsg.DataService["005000002"],
			Detail:    "read entity error: data is null",
		}
	}

	data, ok := reqData.(map[string]interface{})
	if !ok {
		klog.Errorf("reqData type error:%T, reqdata:%v", reqData, reqData)
		return http.StatusInternalServerError, &UpdateResponse{
			Code:      1,
			ErrorCode: "005000001",
			Message:   c.errMsg.DataService["005000001"],
			Detail:    fmt.Errorf("read entity failed").Error(),
		}
	}

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:      1,
			ErrorCode: "005000003",
			Message:   c.errMsg.DataService["005000003"],
			Detail:    "auth model error",
		}
	}
	ds, err := c.service.UpdateDateService(data, req.PathParameter("id"), authuser.Name, authuser.Namespace)
	if err == nil {
		return http.StatusOK, &UpdateResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      ds,
		}
	}
	if errors.IsNameDuplicated(err) {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:      2,
			ErrorCode: "005000004",
			Message:   c.errMsg.DataService["005000004"],
			Detail:    fmt.Errorf("updata data integration error: %+v", err).Error(),
		}
	}

	if errors.IsNotFound(err) {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:      2,
			ErrorCode: "005000006",
			Message:   c.errMsg.DataService["005000006"],
			Detail:    fmt.Errorf("create data integration error: %+v", err).Error(),
		}
	}

	return http.StatusInternalServerError, &UpdateResponse{
		Code:      2,
		ErrorCode: "005000007",
		Message:   c.errMsg.DataService["005000007"],
		Detail:    fmt.Errorf("create data integration error: %+v", err).Error(),
	}

}

//ListDataservice ...
func (c *controller) ListDataservice(req *restful.Request) (int, *ListResponse) {
	pageStr := req.QueryParameter("page")
	limitStr := req.QueryParameter("limit")
	name := req.QueryParameter("name")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		klog.Errorf("offset para error, offset: %+v, err :%v", pageStr, err)
		page = service.DefaultPage
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit > service.Maxlimit || limit <= 0 {
		klog.Errorf("limit para error, offset: %+v, err :%v", limitStr, err)
		limit = service.Maxlimit
	}

	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: "005000003",
			Message:   c.errMsg.DataService["005000003"],
			Detail:    "auth model error",
		}
	}
	ds, err := c.service.ListDataservice(page, limit, name, authuser.Name, authuser.Namespace)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: "000000001",
			Message:   c.errMsg.Common["000000001"],
			Detail:    fmt.Errorf("get database error: %+v", err).Error(),
		}
	}

	return http.StatusOK, &ListResponse{
		Code:      0,
		ErrorCode: "0",
		Data:      ds,
	}

}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
