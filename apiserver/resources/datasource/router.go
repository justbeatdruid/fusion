package datasource

import (
	"github.com/chinamobile/nlpt/apiserver/cmd/apiserver/app/config"

	"github.com/emicklei/go-restful"
)

type router struct {
	controller *controller
}

func NewRouter(cfg *config.Config) *router {
	return &router{newController(cfg)}
}

func (r *router) Install(ws *restful.WebService) {
	ws.Route(ws.POST("/datasource/create").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new datasource").
		To(r.createDatasource).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/dataSource/update").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("update a dataSource").
		To(r.updateDatasource).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/datasource/{id}/get").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an datasource by id").
		To(r.getDatasource).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/datasource/{id}/delete").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an datasource by id").
		To(r.deleteDatasource).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/datasource/list").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all datasources").
		To(r.listDatasource).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
}

func (r *router) createDatasource(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateDatasource(request)
	response.WriteHeaderAndEntity(code, result)
}
func (r *router) updateDatasource(request *restful.Request, response *restful.Response) {
	code, result := r.controller.UpdateDatasource(request)
	response.WriteHeaderAndEntity(code, result)
}
func (r *router) getDatasource(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetDatasource(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteDatasource(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteDatasource(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listDatasource(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListDatasource(request)
	response.WriteHeaderAndEntity(code, result)
}
