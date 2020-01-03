package application

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
	ws.Route(ws.POST("/application/create").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new app").
		To(r.createApplication).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/{id}/get").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an app by id").
		To(r.getApplication).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("{id}/delete").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an app by id").
		To(r.deleteApplication).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/list").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all apps").
		To(r.listApplication).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
}

func (r *router) createApplication(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateApplication(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getApplication(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetApplication(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteApplication(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteApplication(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listApplication(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListApplication(request)
	response.WriteHeaderAndEntity(code, result)
}
