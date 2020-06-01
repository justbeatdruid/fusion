package application

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
	ws.Route(ws.POST("/applications").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new app").
		To(r.createApplication).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/applications/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an app by id").
		To(r.getApplication).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/applications/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an app by id").
		To(r.deleteApplication).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PATCH("/applications/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("patch an app by id").
		To(r.patchApplication).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/applications").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all apps").
		To(r.listApplication).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/applications/{id}/users").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an app by id").
		To(r.getUsers).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/applications/{id}/users").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("add user").
		To(r.addUser).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/applications/{id}/users/{userid}").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("remove user").
		To(r.removeUser).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/applications/{id}/users/{userid}").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("change user role").
		To(r.changeUser).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/applications/{id}/owner").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("change ower").
		To(r.changeOwner).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//app统计接口
	ws.Route(ws.GET("/applications/statistics").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("statistic apps").
		To(r.doStatisticsOnApps).
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

func (r *router) patchApplication(request *restful.Request, response *restful.Response) {
	code, result := r.controller.PatchApplication(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listApplication(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListApplication(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) addUser(request *restful.Request, response *restful.Response) {
	code, result := r.controller.AddUser(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) removeUser(request *restful.Request, response *restful.Response) {
	code, result := r.controller.RemoveUser(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) changeUser(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ChangeUser(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) changeOwner(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ChangeOwner(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getUsers(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetUsers(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) doStatisticsOnApps(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DoStatisticsOncApps(request)
	response.WriteHeaderAndEntity(code, result)
}
