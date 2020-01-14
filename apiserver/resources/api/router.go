package api

import (
	"fmt"

	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

	"github.com/emicklei/go-restful"
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
	ws.Route(ws.POST("/api/create").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new api").
		To(r.createApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/api/{id}/get").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an api by id").
		To(r.getApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/api/{id}/delete").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an api by id").
		To(r.deleteApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/api/publish").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("publish an api by id").
		To(r.publishApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/api/offline").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("offline an api by id").
		To(r.offlineApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/api/list").Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all apis").
		To(r.listApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/api/bind").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("bind an api to application").
		To(r.bindApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.POST("/api/release").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("bind an api to application").
		To(r.releaseApi).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET(fmt.Sprintf("/apiquery/{%s}", apiidPath)).
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("query api data").
		To(r.query).
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

func (r *router) releaseApi(request *restful.Request, response *restful.Response) {
	process(r.controller.ReleaseApi, request, response)
}

func (r *router) query(request *restful.Request, response *restful.Response) {
	process(r.controller.Query, request, response)
}
