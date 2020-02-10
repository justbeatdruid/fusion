package topic

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
	ws.Route(ws.POST("/topics").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new topic").
		To(r.createTopic).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/topics/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an topic by id").
		To(r.getTopic).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/topics/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an topic by id").
		To(r.deleteTopic).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/topics").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all topics").
		To(r.listTopic).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//批量删除topic
	ws.Route(ws.DELETE("/topics").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete all topics ").
		To(r.deleteAllTopics).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//查询topic信息
	ws.Route(ws.GET("/messages/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all messages ").
		To(r.listMessages).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
}

func (r *router) createTopic(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateTopic(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getTopic(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetTopic(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteTopic(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteTopic(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listTopic(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListTopic(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteAllTopics(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteAllTopics(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listMessages(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListMessages(request)
	response.WriteHeaderAndEntity(code, result)
}
