package dataservice

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/chinamobile/nlpt/apiserver/resources/dataservice/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetKubeClient()),
	}
}

//Wrapped ...
type Wrapped struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    *service.Dataservice `json:"data,omitempty"`
}

//CreateResponse ...
type CreateResponse = Wrapped

//CreateRequest ...
type CreateRequest = Wrapped

//DeleteResponse ...
type DeleteResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

//GetResponse ...
type GetResponse = Wrapped

//ListResponse ...
type ListResponse = struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

//PingResponse ...
type PingResponse = DeleteResponse

//CreateDataservice ...
func (c *controller) CreateDataservice(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	ds, err := c.service.CreateDataservice(body.Data)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create database error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &CreateResponse{
		Code: 0,
		Data: ds,
	}

}

//GetDataservice ...
func (c *controller) GetDataservice(req *restful.Request) (int, *GetResponse) {
	ds, err := c.service.GetDataservice(req.PathParameter("id"))
	if err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get database error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &GetResponse{
		Code: 0,
		Data: ds,
	}

}

//DeleteDataservice ...
func (c *controller) DeleteDataservice(req *restful.Request) (int, *DeleteResponse) {
	err := c.service.DeleteDataservice(req.PathParameter("id"))
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete database error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &DeleteResponse{
		Code:    0,
		Message: "",
	}

}

//ListDataservice ...
func (c *controller) ListDataservice(req *restful.Request) (int, *ListResponse) {
	offsetStr := req.QueryParameter("offset")
	limitStr := req.QueryParameter("limit")
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("offset para error, offset: %+v, err :%v", offsetStr, err).Error(),
		}
	}
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("offset para error, limit: %+v, err :%v", limitStr, err).Error(),
		}
	}
	ds, err := c.service.ListDataservice(offset, limit)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	}
	var dss DataserviceList = ds
	data, err := util.PageWrap(dss, offsetStr, limitStr)
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

type DataserviceList []*service.Dataservice

func (dss DataserviceList) Len() int {
	return len(dss)
}

func (dss DataserviceList) GetItem(i int) (interface{}, error) {
	if i >= len(dss) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return dss[i], nil
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
