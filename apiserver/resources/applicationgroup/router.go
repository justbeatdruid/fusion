package applicationgroup

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
	ws.Route(ws.POST("/applicationgroup/create").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new app").
		To(r.createApplicationGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/applicationgroup/{id}/get").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an app by id").
		To(r.getApplicationGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/applicationgroup/{id}/delete").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an app by id").
		To(r.deleteApplicationGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/applicationgroup/list").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all apps").
		To(r.listApplicationGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
}

func (r *router) createApplicationGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateApplicationGroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getApplicationGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetApplicationGroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteApplicationGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteApplicationGroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listApplicationGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListApplicationGroup(request)
	response.WriteHeaderAndEntity(code, result)
}
