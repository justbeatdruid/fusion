package restriction

import (
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/restriction/api/v1"
	"k8s.io/klog"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/restriction/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient()),
	}
}

type Wrapped struct {
	Code    int                  `json:"code"`
	Message string               `json:"message"`
	Data    *service.Restriction `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse = Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type PingResponse = DeleteResponse

// + update_sunyu
type UpdateRequest = Wrapped
type UpdateResponse = Wrapped

type BindRequest struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Operation string   `json:"operation"`
		Apis      []v1.Api `json:"apis"`
	} `json:"data,omitempty"`
}
type BindResponse = Wrapped

func (c *controller) CreateRestriction(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}
	if db, err := c.service.CreateRestriction(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create restriction error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) GetRestriction(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if db, err := c.service.GetRestriction(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get restriction error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) DeleteRestriction(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if err := c.service.DeleteRestriction(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete restriction error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: 0,
		}
	}
}

func (c *controller) ListRestriction(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	if tc, err := c.service.ListRestriction(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var tcs RestrictionList = tc
		data, err := util.PageWrap(tcs, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    1,
				Message: fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: data,
		}
	}
}

type RestrictionList []*service.Restriction

func (tcs RestrictionList) Len() int {
	return len(tcs)
}

func (tcs RestrictionList) GetItem(i int) (interface{}, error) {
	if i >= len(tcs) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return tcs[i], nil
}

// +update
func (c *controller) UpdateRestriction(req *restful.Request) (int, *UpdateResponse) {
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v, reqbody:%v, req:%v", err, reqBody, req).Error(),
		}
	}
	klog.Infof("get body restriction of updating: %+v", reqBody)
	data, ok := reqBody["data"]
	if !ok {
		klog.Infof("get body restriction of updating: %+v", reqBody)
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}

	if db, err := c.service.UpdateRestriction(req.PathParameter("id"), data); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:    2,
			Message: fmt.Errorf("update database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &UpdateResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) BindOrUnbindApis(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	id := req.PathParameter("id")
	if api, err := c.service.BindOrUnbindApis(body.Data.Operation, id, body.Data.Apis); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:    2,
			Message: fmt.Errorf("bind api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &BindResponse{
			Code: 0,
			Data: api,
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
