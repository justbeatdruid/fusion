package clientauth

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
	ws.Route(ws.POST("/clientauths").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new clientauth").
		To(r.createClientauth).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/clientauths/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an clientauth by id").
		To(r.getClientauth).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/clientauths/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an clientauth by id").
		To(r.deleteClientauth).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/clientauths").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all clientauth").
		To(r.listClientauth).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//删除所有clientauths
	ws.Route(ws.DELETE("/clientauths").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete all clientauth").
		To(r.deleteAllClientauths).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	ws.Route(ws.GET("/clientauths/{id}/token").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get token of auth user").
		To(r.getToken).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/clientauths/{id}/token").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("regenerate token of auth user").
		To(r.regenerateToken).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

}

func (r *router) createClientauth(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateClientauth(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getClientauth(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetClientauth(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteClientauth(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteClientauth(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listClientauth(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListClientauth(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteAllClientauths(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteAllClientauths(request)
	response.WriteHeaderAndEntity(code, result)
}

//查看token
func (r *router) getToken(request *restful.Request, response *restful.Response) {

}

//重新生成token
func (r *router) regenerateToken(request *restful.Request, response *restful.Response) {
	code, result := r.controller.RegenerateToken(request)
	response.WriteHeaderAndEntity(code, result)
}
