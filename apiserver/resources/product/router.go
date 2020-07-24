package product

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
	ws.Route(ws.POST("/products").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("create new product").
		To(r.createProduct).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/products/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("get an product by id").
		To(r.getProduct).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.DELETE("/products/{id}").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("delete an product by id").
		To(r.deleteProduct).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

	ws.Route(ws.GET("/products").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON).
		Doc("list all products").
		To(r.listProduct).
		Param(ws.HeaderParameter("content-type", "content-type").DataType("string")).
		Do(returns200, returns500))

}

func (r *router) createProduct(request *restful.Request, response *restful.Response) {
	code, result := r.controller.CreateProduct(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) getProduct(request *restful.Request, response *restful.Response) {
	code, result := r.controller.GetProduct(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) deleteProduct(request *restful.Request, response *restful.Response) {
	code, result := r.controller.DeleteProduct(request)
	response.WriteHeaderAndEntity(code, result)
}

func (r *router) listProduct(request *restful.Request, response *restful.Response) {
	code, result := r.controller.ListProduct(request)
	response.WriteHeaderAndEntity(code, result)
}
