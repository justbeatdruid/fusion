package apiplugin

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
	ws.Route(ws.POST("/apiplugins").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new apiplugins").
		To(r.createApiPlugin).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/apiplugins/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an apiplugins by id").
		To(r.getApiPlugin).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/apiplugins/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an apiplugins by id").
		To(r.deleteApiPlugin).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/apiplugins").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all apiplugins").
		To(r.listApiPlugin).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/apiplugins/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("update apiplugins").
		To(r.updateApiPlugin).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/apiplugins/status").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("update apiplugins status").
		To(r.updateApiPluginStatus).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/apiplugins/{id}/apis").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("bind or unbind apis to apigroup").
		To(r.bindOrUnbindApis).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/apiplugins/{id}/relations").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all objects by apiplugins").
		To(r.listRelationsByApiPlugin).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

}

func (r *router) createApiPlugin(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateApiPlugin(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getApiPlugin(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetApiPlugin(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteApiPlugin(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteApiPlugin(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listApiPlugin(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListApiPlugin(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) updateApiPlugin(request *restful.Request, response *restful.Response) {
	code, result := r.controller.UpdateApiPlugin(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) updateApiPluginStatus(request *restful.Request, response *restful.Response) {
	code, result := r.controller.UpdateApiPluginStatus(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) bindOrUnbindApis(request *restful.Request, response *restful.Response) {
	code, result := r.controller.BindOrUnbindApis(request)
	response.WriteHeaderAndEntity(code, result)
}
func (r *router) listRelationsByApiPlugin(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListRelationsByApiPlugin(request)
	response.WriteHeaderAndEntity(code, result)
}
