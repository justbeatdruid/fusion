package topicgroup

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
	//创建Topicgroup接口
	ws.Route(ws.POST("/topicgroups").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new topicgroup").
		To(r.createTopicgroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//修改Topicgroup策略
	ws.Route(ws.PUT("/topicgroups/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("modify exist topicgroup").
		To(r.modifyTopicgroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//查询指定Topicgroup的详情
	ws.Route(ws.GET("/topicgroups/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an topicgroup by id").
		To(r.getTopicgroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//删除指定Topicgroup以及其下所有Topic
	ws.Route(ws.DELETE("/topicgroups/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an topicgroup by id and all topics under it").
		To(r.deleteTopicgroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//获取Topicgroup列表
	ws.Route(ws.GET("/topicgroups").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all topicgroups").
		To(r.listTopicgroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//批量删除topicgroup
/*	ws.Route(ws.DELETE("/topicgroups").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete all topicgroup and all topics under topicgroup").
		To(r.deleteTopicgroups).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
*/
	//查询topicgroup下的所有topic
	ws.Route(ws.GET("/topicgroups/{id}/topics").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get topics of topicgroup").
		To(r.getTopics).
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

func (r *router) getTopics(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetTopics(request)
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

/*func (r *router) deleteTopicgroups(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteTopicgroups(request)
	response.WriteHeaderAndEntity(code, result)
}*/

func (r *router) modifyTopicgroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ModifyTopicgroup(request)
	response.WriteHeaderAndEntity(code, result)
}
