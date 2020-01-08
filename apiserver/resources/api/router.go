package api

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
	ws.Route(ws.POST("/api/create").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new api").
		To(r.createApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/api/{id}/get").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an api by id").
		To(r.getApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/api/{id}/delete").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an api by id").
		To(r.deleteApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/api/list").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all apis").
		To(r.listApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
}

func process(f func(*restful.Request) (int, interface{}), request *restful.Request, response *restful.Response) {
	code, result := f(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) createApi(request *restful.Request, response *restful.Response) {
	process(r.controller.CreateApi, request, response)
}

func (r *router) getApi(request *restful.Request, response *restful.Response) {
	process(r.controller.GetApi, request, response)
}

func (r *router) deleteApi(request *restful.Request, response *restful.Response) {
	process(r.controller.DeleteApi, request, response)
}

func (r *router) listApi(request *restful.Request, response *restful.Response) {
	process(r.controller.ListApi, request, response)
}

func (r *router) bindApi(request *restful.Request, response *restful.Response) {
	process(r.controller.BindApi, request, response)
}
