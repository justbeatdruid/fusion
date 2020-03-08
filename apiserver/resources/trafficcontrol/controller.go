package trafficcontrol

import (
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/trafficcontrol/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/emicklei/go-restful"
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
	Code    int                     `json:"code"`
	Message string                  `json:"message"`
	Data    *service.Trafficcontrol `json:"data,omitempty"`
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
	Code    int      `json:"code"`
	Message string   `json:"message"`
	Data    []v1.Api `json:"data,omitempty"`
}
type BindResponse = Wrapped

func (c *controller) CreateTrafficcontrol(req *restful.Request) (int, *CreateResponse) {
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
	if db, err := c.service.CreateTrafficcontrol(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create trafficcontrol error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) GetTrafficcontrol(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if db, err := c.service.GetTrafficcontrol(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    1,
			Message: fmt.Errorf("get trafficcontrol error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: db,
		}
	}
}

func (c *controller) DeleteTrafficcontrol(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if err := c.service.DeleteTrafficcontrol(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    1,
			Message: fmt.Errorf("delete trafficcontrol error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: 0,
		}
	}
}

func (c *controller) ListTrafficcontrol(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	if tc, err := c.service.ListTrafficcontrol(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    1,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var tcs TrafficcontrolList = tc
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

type TrafficcontrolList []*service.Trafficcontrol

func (tcs TrafficcontrolList) Len() int {
	return len(tcs)
}

func (tcs TrafficcontrolList) GetItem(i int) (interface{}, error) {
	if i >= len(tcs) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return tcs[i], nil
}

// +update_sunyu
func (c *controller) UpdateTrafficcontrol(req *restful.Request) (int, *UpdateResponse) {
	reqBody := make(map[string]interface{})
	if err := req.ReadEntity(&reqBody); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v, reqbody:%v, req:%v", err, reqBody, req).Error(),
		}
	}
	data, ok := reqBody["data"]
	if !ok {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: "read entity error: data is null",
		}
	}

	if db, err := c.service.UpdateTrafficcontrol(req.PathParameter("id"), data); err != nil {
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

func (c *controller) BindApis(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	trafficID := req.PathParameter("id")
	if api, err := c.service.BindApi(trafficID, body.Data); err != nil {
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
func (c *controller) UnBindApis(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	trafficID := req.PathParameter("id")
	if api, err := c.service.UnBindApi(trafficID, body.Data); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:    2,
			Message: fmt.Errorf("unbind api error: %+v", err).Error(),
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
