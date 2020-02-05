package apply

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
	ws.Route(ws.POST("/applies").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new apply").
		To(r.createApply).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/applies/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an apply by id").
		To(r.getApply).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/applies/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an apply by id").
		To(r.deleteApply).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/applies").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all applys").
		To(r.listApply).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/applies/{id}/approval").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("approve applys").
		To(r.approveApply).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
}

func (r *router) createApply(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateApply(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getApply(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetApply(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteApply(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteApply(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listApply(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListApply(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) approveApply(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ApproveApply(request)
	response.WriteHeaderAndEntity(code, result)
}
