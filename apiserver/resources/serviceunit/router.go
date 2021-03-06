package serviceunit

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
	ws.Route(ws.POST("/serviceunits").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Reads(RequestWrapped{}).
		Writes(Wrapped{}).
		Doc("create new serviceunit").
		To(r.createServiceunit).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/serviceunits/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Writes(Wrapped{}).
		Doc("get an serviceunit by id").
		To(r.getServiceunit).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/serviceunits/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an serviceunit by id").
		To(r.deleteServiceunit).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PATCH("/serviceunits/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Reads(RequestWrapped{}).
		Writes(Wrapped{}).
		Doc("patch a serviceunit by id").
		To(r.patchServiceunit).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/serviceunits/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Reads(RequestWrapped{}).
		Writes(Wrapped{}).
		Doc("patch a serviceunit by id").
		To(r.patchServiceunit).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/serviceunits").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Writes(ListResponse{}).
		Doc("list all serviceunits").
		To(r.listServiceunit).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/serviceunits/fissions").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all  fissions serviceunits").
		To(r.listSuFission).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/serviceunits/{id}/release").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new serviceunit").
		To(r.publishServiceunit).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	/*
		// + update_sunyu
		ws.Route(ws.PUT("/serviceunits/{id}").
			Consumes(restful.MIME_JSON).
			Produces(restful.MIME_JSON).
			Doc("update serviceunit").
			To(r.updateServiceunit).
			Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
			Do(returns200, returns500))

	*/
	ws.Route(ws.GET("/serviceunits/{id}/users").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an serviceunit by id").
		To(r.getUsers).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/serviceunits/{id}/users").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("add user").
		To(r.addUser).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/serviceunits/{id}/users/{userid}").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("remove user").
		To(r.removeUser).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/serviceunits/{id}/users/{userid}").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("change user role").
		To(r.changeUser).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/serviceunits/{id}/owner").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("change ower").
		To(r.changeOwner).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//导入函数
	ws.Route(ws.POST("/serviceunits/import").
		Consumes("multipart/form-data").
		Produces(restful.MIME_JSON).
		Doc("import functions from files").
		To(r.importServiceunits).
		Do(returns200, returns500))

	//获取函数的日志
	ws.Route(ws.GET("/serviceunits/function/logs").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get logs of function").
		To(r.getFnLogs).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//调式函数
	ws.Route(ws.POST("/serviceunits/function/test").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("test function").
		To(r.testFn).
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

func (r *router) patchServiceunit(request *restful.Request, response *restful.Response) {
	code, result := r.controller.PatchServiceunit(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listServiceunit(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListServiceunit(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listSuFission(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListSuFission(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) publishServiceunit(request *restful.Request, response *restful.Response) {
	code, result := r.controller.PublishServiceunit(request)
	response.WriteHeaderAndEntity(code, result)
}

/*
// + update_sunyu
func (r *router) updateServiceunit(request *restful.Request, response *restful.Response) {
	code, result := r.controller.UpdateServiceunit(request)
	response.WriteHeaderAndEntity(code, result)
}

*/

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
	code, headers, result := r.controller.GetUsers(request)
	for k, v := range headers {
		response.AddHeader(k, v)
	}
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) importServiceunits(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ImportServiceunits(request, response)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getFnLogs(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetFnLogs(request, response)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) testFn(request *restful.Request, response *restful.Response) {
	code, result := r.controller.TestFn(request, response)
	response.WriteHeaderAndEntity(code, result)
}
