package topic

import (
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/go-restful"
	"net/http"
	"os"
)

type router struct {
	controller *controller "github.com/chinamobile/nlpt/pkg/go-restful"
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
	//根据时间和id查询topic信息
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
		To(r.deleteTopics).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/topics/{id}/permissions/{auth-user-id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("modify permissions ").
		To(r.deleteTopics).
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

func (r *router) listMessages(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListMessages(request)
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

}
func (r *router) deletePermissions(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeletePermissions(request)
	response.WriteHeaderAndEntity(code, result)
}