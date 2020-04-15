package clientauth

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

	//批量删除clientauths
	ws.Route(ws.DELETE("/clientauths").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("batch delete clientauth").
		To(r.deleteClientauths).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	ws.Route(ws.POST("/clientauths/{id}/token").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("regenerate token of auth user").
		To(r.regenerateToken).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//模糊查询
	ws.Route(ws.GET("/clientauths").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all clientauth").
		To(r.listClientauths).
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

func (r *router) listClientauths(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListClientauths(request)
	response.WriteHeaderAndEntity(code, result)
}
func (r *router) deleteClientauths(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteClientauths(request)
	response.WriteHeaderAndEntity(code, result)
}

//重新生成token
func (r *router) regenerateToken(request *restful.Request, response *restful.Response) {
	code, result := r.controller.RegenerateToken(request)
	response.WriteHeaderAndEntity(code, result)
}
