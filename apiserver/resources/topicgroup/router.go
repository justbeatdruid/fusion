package topicgroup

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
	ws.Route(ws.POST("/topicgroups").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new topicgroup").
		To(r.createTopicgroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/topicgroups/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an topicgroup by id").
		To(r.getTopicgroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/topicgroups/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an topicgroup by id").
		To(r.deleteTopicgroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/topicgroups").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all topicgroups").
		To(r.listTopicgroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//删除所有topic
	ws.Route(ws.DELETE("/topicgroups").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete all topicgroup").
		To(r.deleteAllTopicgroups).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
}

func (r *router) createTopicgroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateTopicgroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getTopicgroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetTopicgroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteTopicgroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteTopicgroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listTopicgroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListTopicgroup(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteAllTopicgroups(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteAllTopicgroups(request)
	response.WriteHeaderAndEntity(code, result)
}