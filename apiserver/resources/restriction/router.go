package restriction

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
	ws.Route(ws.POST("/restrictions").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new restriction").
		To(r.createRestriction).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/restrictions/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an restriction by id").
		To(r.getRestriction).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/restrictions/{id}/apis").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("bind/unbind apis to restriction").
		To(r.bindOrUnbindApis).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/restrictions/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an restriction by id").
		To(r.deleteRestriction).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/restrictions").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("batch delete restriction").
		To(r.batchDeleteRestriction).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/restrictions").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all restrictions").
		To(r.listRestriction).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	// + update
	ws.Route(ws.PATCH("/restrictions/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("update restriction").
		To(r.updateRestriction).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

}

func (r *router) createRestriction(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateRestriction(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getRestriction(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetRestriction(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteRestriction(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteRestriction(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) batchDeleteRestriction(request *restful.Request, response *restful.Response) {
	code, result := r.controller.BatchDeleteRestriction(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listRestriction(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListRestriction(request)
	response.WriteHeaderAndEntity(code, result)
}

// + update_
func (r *router) updateRestriction(request *restful.Request, response *restful.Response) {
	code, result := r.controller.UpdateRestriction(request)
	response.WriteHeaderAndEntity(code, result)
}

func process(f func(*restful.Request) (int, interface{}), request *restful.Request, response *restful.Response) {
	code, result := f(request)
	response.WriteHeaderAndEntity(code, result)
}
func (r *router) bindOrUnbindApis(request *restful.Request, response *restful.Response) {
	process(r.controller.BindOrUnbindApis, request, response)
}
