package api

import (
	"fmt"
	"github.com/chinamobile/nlpt/apiserver/resources/api/service"
	"net/http"
	"os"

	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type router struct {
	controller *controller
}
type ImportResponse struct {
	Code      int           `json:"code"`
	ErrorCode string        `json:"errorCode"`
	Message   string        `json:"message"`
	Data      []service.Api `json:"data"`
	Detail    string        `json:"detail"`
}

func NewRouter(cfg *config.Config) *router {
	return &router{newController(cfg)}
}

const (
	apiidPath    = "apiid"
	tenantidPath = "tenantid"
)

func (r *router) Install(ws *restful.WebService) {
	ws.Route(ws.POST("/apis").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new api").
		To(r.createApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/apis/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an api by id").
		To(r.getApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PATCH("/apis/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an api by id").
		To(r.patchApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/apis/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an api by id").
		To(r.deleteApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.PUT("/apis").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("batch delete api").
		To(r.batchDeleteApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/apis/{id}/release").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("publish an api by id").
		To(r.publishApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/apis/{id}/release").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("offline an api by id").
		To(r.offlineApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/apis").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all apis").
		To(r.listApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/apis/applications").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all application apis").
		To(r.listApplicationApis).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/apis/serviceunits").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all serviceunit apis").
		To(r.listServiceunitApis).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/apis/{id}/applications/{appid}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("bind/release an api to/from application").
		To(r.bindApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	ws.Route(ws.POST("/apis/applications/{appid}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("batch bind/release an api to/from application").
		To(r.batchBindApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET(fmt.Sprintf("/apis/{%s}/data", apiidPath)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("query api data").
		To(r.query).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET(fmt.Sprintf("/apis/{%s}/{%s}/data", tenantidPath, apiidPath)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("query api data").
		To(r.kongQuery).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST(fmt.Sprintf("/apis/{%s}/{%s}/data", tenantidPath, apiidPath)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("query api data").
		To(r.kongQuery).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/api/test").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("test an api function").
		To(r.testApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//api统计接口
	ws.Route(ws.GET("/apis/statistics").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("statistic apis").
		To(r.doStatisticsOnApis).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//api export
	ws.Route(ws.POST("/apis/export").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("export information of apis ").
		To(r.exportApis).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	//api import
	ws.Route(ws.POST("/apis/import").
		Consumes("multipart/form-data").
		Produces(restful.MIME_JSON).
		Doc("import apis from excel files").
		To(r.importApis).
		Do(returns200, returns500))
	// add plugins
	ws.Route(ws.POST("/apis/{id}/plugins").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("add an api plugins").
		To(r.addApiPlugins).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//delete plugins
	ws.Route(ws.DELETE("/apis/{api_id}/plugins/{plugin_id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an api plugins").
		To(r.deleteApiPlugins).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
	//update plugins
	ws.Route(ws.PATCH("/apis/{api_id}/plugins/{plugin_id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an api plugins").
		To(r.patchApiPlugins).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/apis/apigroups/{id}").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all apis").
		To(r.listApisByApiGroup).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))
}

func process(f func(*restful.Request) (int, interface{}), request *restful.Request, response *restful.Response) {
	code, result := f(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) createApi(request *restful.Request, response *restful.Response) {
	process(r.controller.CreateApi, request, response)
}

func (r *router) patchApi(request *restful.Request, response *restful.Response) {
	process(r.controller.PatchApi, request, response)
}

func (r *router) getApi(request *restful.Request, response *restful.Response) {
	process(r.controller.GetApi, request, response)
}

func (r *router) deleteApi(request *restful.Request, response *restful.Response) {
	process(r.controller.DeleteApi, request, response)
}

func (r *router) batchDeleteApi(request *restful.Request, response *restful.Response) {
	process(r.controller.BatchDeleteApi, request, response)
}

func (r *router) publishApi(request *restful.Request, response *restful.Response) {
	process(r.controller.PublishApi, request, response)
}

func (r *router) offlineApi(request *restful.Request, response *restful.Response) {
	process(r.controller.OfflineApi, request, response)
}

func (r *router) listApi(request *restful.Request, response *restful.Response) {
	process(r.controller.ListApi, request, response)
}

func (r *router) bindApi(request *restful.Request, response *restful.Response) {
	process(r.controller.BindApi, request, response)
}
func (r *router) batchBindApi(request *restful.Request, response *restful.Response) {
	process(r.controller.BatchBindApi, request, response)
}

func (r *router) query(request *restful.Request, response *restful.Response) {
	process(r.controller.Query, request, response)
}

func (r *router) kongQuery(request *restful.Request, response *restful.Response) {
	process(r.controller.KongQuery, request, response)
}

func (r *router) testApi(request *restful.Request, response *restful.Response) {
	process(r.controller.TestApi, request, response)
}

func (r *router) doStatisticsOnApis(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DoStatisticsOncApis(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listApplicationApis(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListAllApplicationApis(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listServiceunitApis(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListAllServiceunitApis(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) exportApis(request *restful.Request, response *restful.Response) {
	r.controller.ExportApis(request)
	response.Header().Add("Content-Disposition", "attachment;filename=api.xlsx")
	response.Header().Add("Content-Type", "application/vnd.ms-excel")
	http.ServeFile(response.ResponseWriter, request.Request, "./tmp/api.xlsx")
	defer os.Remove("./tmp/api.xlsx")
}
func (r *router) importApis(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ImportApis(request, response)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) addApiPlugins(request *restful.Request, response *restful.Response) {
	process(r.controller.AddApiPlugins, request, response)
}

func (r *router) deleteApiPlugins(request *restful.Request, response *restful.Response) {
	process(r.controller.DeleteApiPlugins, request, response)
}

func (r *router) patchApiPlugins(request *restful.Request, response *restful.Response) {
	process(r.controller.PatchApiPlugins, request, response)
}

func (r *router) listApisByApiGroup(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListApisByApiGroup(request)
	response.WriteHeaderAndEntity(code, result)
}
