package serviceunitgroup

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
	ws.Route(ws.POST("/serviceunitgroup/create").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new serviceunit gruop").
		To(r.createServiceunitGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/serviceunitgroup/{id}/get").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an serviceunit gruop by id").
		To(r.getServiceunitGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/serviceunitgroup/{id}/delete").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an serviceunit gruop by id").
		To(r.deleteServiceunitGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/serviceunitgroup/list").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all serviceunit gruops").
		To(r.listServiceunitGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
}

func (r *router) createServiceunitGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateServiceunitGroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getServiceunitGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetServiceunitGroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteServiceunitGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteServiceunitGroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listServiceunitGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListServiceunitGroup(request)
	response.WriteHeaderAndEntity(code, result)
}
