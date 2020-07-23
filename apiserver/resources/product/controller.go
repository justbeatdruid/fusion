package product

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/product/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/names"
	//"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.TenantEnabled, cfg.Database),
	}
}

type Wrapped struct {
	Code      int              `json:"code"`
	ErrorCode string           `json:"errorCode"`
	Message   string           `json:"message"`
	Detail    string           `json:"detail"`
	Data      *service.Product `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse struct {
	Code      int    `json:"code"`
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
	Detail    string `json:"detail"`
}
type GetResponse = Wrapped
type ListResponse = struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Message   string      `json:"message"`
	Detail    string      `json:"detail"`
	Data      interface{} `json:"data"`
}

func (c *controller) CreateProduct(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		code := "000000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:   1,
			Detail: "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "000000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    "auth model error",
		}
	}
	body.Data.Id = names.NewID()
	body.Data.Tenant = authuser.Namespace
	body.Data.User = authuser.Name
	if apl, err := c.service.CreateProduct(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: "",
			Message:   "",
			Detail:    fmt.Errorf("create product error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: apl,
		}
	}
}

func (c *controller) GetProduct(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if apl, err := c.service.GetProduct(id); err != nil {
		code := "000000007"
		return http.StatusInternalServerError, &GetResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    fmt.Errorf("get product error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: apl,
		}
	}
}

func (c *controller) DeleteProduct(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if err := c.service.DeleteProduct(id); err != nil {
		code := "000000002"
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    fmt.Errorf("delete product error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:   0,
			Detail: "",
		}
	}
}

func (c *controller) ListProduct(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	condition := service.Product{
		Category: req.QueryParameter("category"),
		Status:   req.QueryParameter("status"),
	}
	authuser, err := auth.GetAuthUser(req)
	if len(req.QueryParameter("tenant")) > 0 {
		condition.Tenant = authuser.Namespace
	}
	if err != nil {
		code := "000000006"
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    "auth model error",
		}
	}
	if pl, err := c.service.ListProduct(condition); err != nil {
		code := "000000002"
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: code,
			Message:   "",
			Detail:    fmt.Errorf("list products error: %+v", err).Error(),
		}
	} else {
		var pls ProductList = pl
		data, err := util.PageWrap(pls, page, size)
		if err != nil {
			code := "000000005"
			return http.StatusInternalServerError, &ListResponse{
				Code:      1,
				ErrorCode: code,
				Message:   "",
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: data,
		}
	}
}

type ProductList []*service.Product

func (apls ProductList) Len() int {
	return len(apls)
}

func (apls ProductList) GetItem(i int) (interface{}, error) {
	if i >= len(apls) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return apls[i], nil
}

func (apls ProductList) Less(i, j int) bool {
	return apls[i].ReleasedAt.After(apls[j].ReleasedAt)
}

func (apls ProductList) Swap(i, j int) {
	apls[i], apls[j] = apls[j], apls[i]
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
