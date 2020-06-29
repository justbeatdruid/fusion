package trafficcontrol

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
	ws.Route(ws.POST("/trafficcontrols").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new trafficcontrol").
		To(r.createTrafficcontrol).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/trafficcontrols/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an trafficcontrol by id").
		To(r.getTrafficcontrol).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/trafficcontrols/{id}/apis").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("bind/unbind apis to trafficcontrol").
		To(r.bindOrUnbindApis).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/trafficcontrols/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an trafficcontrol by id").
		To(r.deleteTrafficcontrol).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/trafficcontrols").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("batch delete trafficcontrol").
		To(r.batchDeleteTrafficcontrol).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/trafficcontrols").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all trafficcontrols").
		To(r.listTrafficcontrol).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	// + update
	ws.Route(ws.PATCH("/trafficcontrols/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("update trafficcontrol").
		To(r.updateTrafficcontrol).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

}

func (r *router) createTrafficcontrol(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateTrafficcontrol(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getTrafficcontrol(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetTrafficcontrol(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteTrafficcontrol(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteTrafficcontrol(request)
	response.WriteHeaderAndEntity(code, result)
}
func (r *router) batchDeleteTrafficcontrol(request *restful.Request, response *restful.Response) {
	code, result := r.controller.BatchDeleteTrafficcontrol(request)
	response.WriteHeaderAndEntity(code, result)
}
func (r *router) listTrafficcontrol(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListTrafficcontrol(request)
	response.WriteHeaderAndEntity(code, result)
}

// + update_
func (r *router) updateTrafficcontrol(request *restful.Request, response *restful.Response) {
	code, result := r.controller.UpdateTrafficcontrol(request)
	response.WriteHeaderAndEntity(code, result)
}

func process(f func(*restful.Request) (int, interface{}), request *restful.Request, response *restful.Response) {
	code, result := f(request)
	response.WriteHeaderAndEntity(code, result)
}
func (r *router) bindOrUnbindApis(request *restful.Request, response *restful.Response) {
	process(r.controller.BindOrUnbindApis, request, response)
}
