package datasource

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/datasource/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

	"github.com/emicklei/go-restful"
)

type controller struct {
	service *service.Service
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.DatasourceConfig.Supported),
	}
}

type Wrapped struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    *service.Datasource `json:"data,omitempty"`
}
type QueryDataResponse struct {
	Code    int                    `json:"code"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"queryData"`
}
type CreateResponse = Wrapped
type UpdateResponse = Wrapped
type CreateRequest = Wrapped
type UpdateRequest = Wrapped
type DeleteResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
type GetResponse = Wrapped
type ListResponse = struct {
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	Data    []*service.Datasource `json:"data"`
}
type PingResponse = DeleteResponse

func (c *controller) CreateDatasource(req *restful.Request) (int, *CreateResponse) {
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
	if db, err := c.service.CreateDatasource(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: db,
		}
	}
}
func (c *controller) UpdateDatasource(req *restful.Request) (int, *UpdateResponse) {
	body := &UpdateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if db, err := c.service.UpdateDatasource(body.Data); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    2,
			Message: fmt.Errorf("update database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &UpdateResponse{
			Code: 0,
			Data: db,
		}
	}
}
func (c *controller) GetDatasource(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if db, err := c.service.GetDatasource(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) DeleteDatasource(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if err := c.service.DeleteDatasource(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:    0,
			Message: "",
		}
	}
}

func (c *controller) ListDatasource(req *restful.Request) (int, *ListResponse) {
	if db, err := c.service.ListDatasource(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: db,
		}
	}
}
func (c *controller) getDataByApi(req *restful.Request) (int, *QueryDataResponse) {
	apiId := req.PathParameter("apiId")
	//todo Acquisition parameters in the request body（Provisional use this method）
	parameters, err := req.BodyParameter("params")
	if err != nil {
		return http.StatusInternalServerError, &QueryDataResponse{
			Code:    1,
			Message: fmt.Errorf("get parameters error: %+v", err).Error(),
		}
	}
	result, err := c.service.GetDataSourceByApiId(apiId, parameters)
	if err != nil {
		return http.StatusInternalServerError, &QueryDataResponse{
			Code:    1,
			Message: fmt.Errorf("query data error: %+v", err).Error(),
		}
	}
	return http.StatusOK, &QueryDataResponse{
		Code: 0,
		Data: result,
	}
}
func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
