package api

import (
	"fmt"

	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type router struct {
	controller *controller
}

func NewRouter(cfg *config.Config) *router {
	return &router{newController(cfg)}
}

const (
	apiidPath = "apiid"
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

	ws.Route(ws.POST("/apis/{id}/applications/{appid}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("bind/release an api to/from application").
		To(r.bindApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET(fmt.Sprintf("/apis/{%s}/data", apiidPath)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("query api data").
		To(r.query).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/api/test").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("test an api function").
		To(r.testApi).
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

func (r *router) query(request *restful.Request, response *restful.Response) {
	process(r.controller.Query, request, response)
}

func (r *router) testApi(request *restful.Request, response *restful.Response) {
	process(r.controller.TestApi, request, response)
}
