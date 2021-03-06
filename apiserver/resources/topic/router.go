package topic

import (
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	"net/http"
	"os"
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
		To(r.deleteTopics).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//查询topic信息
	ws.Route(ws.GET("/topics/data").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all messages ").
		To(r.listMessages).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//导出topics信息
	ws.Route(ws.GET("/topics/export").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("export information of topics ").
		To(r.exportTopics).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//导入topics
	ws.Route(ws.POST("/topics/import").
		Consumes("multipart/form-data").
		Produces(restful.MIME_JSON).
		Doc("import topics from excel files").
		To(r.importTopics).
		Do(returns200, returns500))

	ws.Route(ws.POST("/topics/{id}/permissions/{auth-user-id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("grant permissions ").
		To(r.grantPermissions).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/topics/{id}/permissions/{auth-user-id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("modify permissions ").
		To(r.modifyPermissions).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//删除用户授权
	ws.Route(ws.DELETE("/topics/{id}/permissions/{auth-user-id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete permissions ").
		To(r.deletePermissions).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//查询topic授权用户
	ws.Route(ws.GET("/topics/{id}/permissions").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all permissions").
		To(r.listUsers).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//Topic统计接口
	ws.Route(ws.GET("/topics/statistics").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("statistic topics").
		To(r.doStatisticsOnTopics).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//增加分区topic的分区数
	ws.Route(ws.PUT("/topics/{id}/partitions/{partitions}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("Increment partitions of an existing partitioned topic.").
		To(r.addPartitionsOfTopic).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/topics/applications/{app-id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("batch bind topics to application").
		To(r.batchBindTopics).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/topics/{id}/applications/{app-id}/permission").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("modify application permission").
		To(r.modifyApplicationPermission).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/topics/{id}/subscriptions").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get subscriptions of topics").
		To(r.getSubscriptionsOfTopic).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//多分区Topic的订阅列表
	ws.Route(ws.GET("/topics/{id}/subscriptions/partition").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get subscriptions of partitioned topics").
		To(r.getPartitionedSubscritionsOfTopic).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//手动刷新消费者页面
	ws.Route(ws.GET("/topics/{id}/subscriptions/{subName}/consumers").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("refresh consumers info").
		To(r.getPartitionedSubscritionsOfTopic).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//发送消息
	ws.Route(ws.POST("/topics/messages").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("send message to topic").
		To(r.sendMessages).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//重置消费者订阅位置
	ws.Route(ws.POST("/topics/messagePosition").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("reset position").
		To(r.resetPosition).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//从最新位置开始消费接口开发
	ws.Route(ws.POST("/topics/{id}/subscription/{subName}/skip_all").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("Completely clears the backlog on the subscription.").
		To(r.skipAllMessages).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//批量设置用户权限
	ws.Route(ws.POST("/topics/{id}/permissions").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("batch grant permissions ").
		To(r.batchGrantPermissions).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//从最新位置开始消费接口开发
	ws.Route(ws.POST("/topics/{id}/subscription/{subName}/skip/{numMessages}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("Skipping messages on a topic subscription").
		To(r.skipMessages).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//监控与统计接口
	ws.Route(ws.GET("/topics/statistics/{query}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("monitoring and statistics").
		To(r.skipMessages).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//修改Topic描述/备注
	ws.Route(ws.PUT("/topics/{id}/description").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("modify topic description").
		To(r.modifyDescription).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//终止Topic
	ws.Route(ws.POST("/topics/{id}/terminate").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("terminate a topic").
		To(r.terminateTopic).
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

//批量删除topics
func (r *router) deleteTopics(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteTopics(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) doStatisticsOnTopics(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DoStatisticsOnTopics(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listMessages(request *restful.Request, response *restful.Response) {
	//code, result := r.controller.ListMessages(request)
	code, result := r.controller.QueryMessage(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listUsers(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListUsers(request)
	response.WriteHeaderAndEntity(code, result)
}

//导出关于topics的信息
func (r *router) exportTopics(request *restful.Request, response *restful.Response) {
	r.controller.ExportTopics(request)
	response.Header().Add("Content-Disposition", "attachment;filename=topics.xlsx")
	response.Header().Add("Content-Type", "application/vnd.ms-excel")
	http.ServeFile(response.ResponseWriter, request.Request, "/tmp/topics.xlsx")
	os.Remove("/tmp/topics.xlsx")

}

func (r *router) importTopics(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ImportTopics(request, response)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) grantPermissions(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GrantPermissions(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) modifyPermissions(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ModifyPermissions(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deletePermissions(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeletePermissions(request)
	response.WriteHeaderAndEntity(code, result)
}
func (r *router) addPartitionsOfTopic(request *restful.Request, response *restful.Response) {
	code, result := r.controller.AddPartitionsOfTopic(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) batchBindTopics(request *restful.Request, response *restful.Response) {
	code, result := r.controller.BatchBindOrReleaseApi(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getSubscriptionsOfTopic(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetSubscriptionsOfTopic(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getPartitionedSubscritionsOfTopic(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetPartitionedSubscritionsOfTopic(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) sendMessages(request *restful.Request, response *restful.Response) {
	code, result := r.controller.SendMessages(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) resetPosition(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ResetPosition(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) skipAllMessages(request *restful.Request, response *restful.Response) {
	code, result := r.controller.SkipAllMessages(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) batchGrantPermissions(request *restful.Request, response *restful.Response) {
	code, result := r.controller.BatchGrantPermissions(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) skipMessages(request *restful.Request, response *restful.Response) {
	code, result := r.controller.SkipMessages(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) refreshConsumers(request *restful.Request, response *restful.Response) {
	code, result := r.controller.Refresh(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) modifyDescription(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ModifyDescription(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) modifyApplicationPermission(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ModifyApplicationPermission(request)
	response.WriteHeaderAndEntity(code, result)
}


func (r *router) terminateTopic(request *restful.Request, response *restful.Response) {
	code, result := r.controller.TerminateTopic(request)
	response.WriteHeaderAndEntity(code, result)
}

