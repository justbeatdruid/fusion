package datasource

import (
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type router struct {
	controller *controller
}

func NewRouter(cfg *config.Config) *router {
	return &router{newController(cfg)}
}

func (r *router) Install(ws *restful.WebService) {
	ws.Route(ws.POST("/datasources").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new datasource").
		To(r.createDatasource).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/datasources").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("update a dataSource").
		To(r.updateDatasource).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/datasources/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an datasource by id").
		To(r.getDatasource).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/datasources/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an datasource by id").
		To(r.deleteDatasource).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/datasources").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all datasources").
		To(r.listDatasource).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/datasources/connection").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("test datasource connection").
		To(r.testConnection).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/datasources/{id}/sql").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("test datasource if sql correct").
		To(r.testSql).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/datasources/{id}/tables").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list datasource tables").
		To(r.getTables).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/datasources/{id}/tables/{table}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list datasource tables").
		To(r.getTable).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/datasources/{id}/tables/{table}/fields").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list fields of a table").
		To(r.getFields).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/datasources/{id}/tables/{table}/fields/{field}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list fields of a table").
		To(r.getField).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/datasources/match").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("automatically match two rdb datasources fields").
		To(r.match).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	/*
		ws.Route(ws.GET("/datasources/{apiId}/data").
			Consumes(restful.MIME_JSON).
			Produces(restful.MIME_JSON).
			Doc("query data by api").
			To(r.getDataByApi).
			Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
			Do(returns200, returns500))

		ws.Route(ws.GET("/datasources/ConnectMysql").
			Consumes(restful.MIME_JSON).
			Produces(restful.MIME_JSON).
			Doc("Connect  Mysql").
			To(r.getMysqlData).
			Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
			Do(returns200, returns500))
	*/

	ws.Route(ws.GET("/datasources/statistics").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("statistic datasources").
		To(r.doStatistics).
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

func (r *router) getTables(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetTables(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getTable(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetTable(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getFields(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetFields(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getField(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetField(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) testConnection(request *restful.Request, response *restful.Response) {
	code, result := r.controller.Ping(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) testSql(request *restful.Request, response *restful.Response) {
	code, result := r.controller.Query(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) doStatistics(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DoStatistics(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) match(request *restful.Request, response *restful.Response) {
	code, result := r.controller.Match(request)
	response.WriteHeaderAndEntity(code, result)
}
