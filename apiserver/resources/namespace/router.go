package namespace

import (
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

	"github.com/emicklei/go-restful"
)

type router struct {
	controller *controller
}

func NewRouter(cfg *config.Config) *router {
	return &router{newController(cfg)}
}

func (r *router) Install(ws *restful.WebService) {
	ws.Route(ws.POST("/namespaces").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new namespace").
		To(r.createNamespace).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/namespaces/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an topic by id").
		To(r.getNamespace).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/namespaces/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an namespace by id").
		To(r.deleteNamespace).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/namespaces").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all namespaces").
		To(r.listNamespace).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//删除所有topic
	ws.Route(ws.DELETE("/namespaces").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete all namespaces").
		To(r.deleteAllNamespaces).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
}

func (r *router) createNamespace(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateNamespace(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getNamespace(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetNamespace(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteNamespace(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteNamespace(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listNamespace(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListNamespace(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteAllNamespaces(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteAllNamespaces(request)
	response.WriteHeaderAndEntity(code, result)
}
