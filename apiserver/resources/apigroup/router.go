package apigroup

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
	ws.Route(ws.POST("/apigroups").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new apigroup").
		To(r.createApiGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/apigroups/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an apigroup by id").
		To(r.getApiGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/apigroups/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an apigroups by id").
		To(r.deleteApiGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/apigroups").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all apigroups").
		To(r.listApiGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/apigroups/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("update apigroups").
		To(r.updateApiGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/apigroups/status").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("update apigroups status").
		To(r.updateApiGroupStatus).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

}

func (r *router) createApiGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateApiGroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getApiGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetApiGroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteApiGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteApiGroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listApiGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListApiGroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) updateApiGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.UpdateApiGroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) updateApiGroupStatus(request *restful.Request, response *restful.Response) {
	code, result := r.controller.UpdateApiGroupStatus(request)
	response.WriteHeaderAndEntity(code, result)
}
