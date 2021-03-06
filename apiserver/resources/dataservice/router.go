package dataservice

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
	ws.Route(ws.POST("/dataservices").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new dataservice").
		To(r.createDataservice).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/dataservices/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an dataservice by id").
		To(r.getDataservice).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/dataservices/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an dataservice by id").
		To(r.deleteDataservice).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/dataservices").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all dataservices").
		To(r.listDataservice).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PATCH("/dataservices/{id}").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("update dataService").
		To(r.updateDateService).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/dataservices/operation").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("open or stop dataservice").
		To(r.operationDataservice).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/dataservices/taskrunlog/{id}").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list a task all runlog").
		To(r.getTaskRunlog).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/dataservices/statistics").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("statistic dataservices").
		To(r.doStatisticsdataservices).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

}

func (r *router) createDataservice(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateDataservice(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) operationDataservice(request *restful.Request, response *restful.Response) {
	code, result := r.controller.OperationDataservice(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getDataservice(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetDataservice(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteDataservice(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteDataservice(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listDataservice(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListDataservice(request)
	response.WriteHeaderAndEntity(code, result)
}
func (r *router) updateDateService(request *restful.Request, response *restful.Response) {
	code, result := r.controller.UpdateDateService(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getTaskRunlog(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetTaskRunlog(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) doStatisticsdataservices(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DoStatisticsDataservices(request)
	response.WriteHeaderAndEntity(code, result)
}
