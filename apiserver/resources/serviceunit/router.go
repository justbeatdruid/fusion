package serviceunit

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
	ws.Route(ws.POST("/serviceunit/create").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new serviceunit").
		To(r.createServiceunit).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/serviceunit/{id}/get").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an serviceunit by id").
		To(r.getServiceunit).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/serviceunit/{id}/delete").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an serviceunit by id").
		To(r.deleteServiceunit).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/serviceunit/list").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all serviceunits").
		To(r.listServiceunit).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
}

func (r *router) createServiceunit(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateServiceunit(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getServiceunit(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetServiceunit(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteServiceunit(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteServiceunit(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listServiceunit(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListServiceunit(request)
	response.WriteHeaderAndEntity(code, result)
}
