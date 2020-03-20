package trafficcontrol

import (
	"fmt"
	v1 "github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/trafficcontrol/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.LocalConfig),
		cfg.LocalConfig,
	}
}

type Wrapped struct {
	Code      int                     `json:"code"`
	ErrorCode string                  `json:"errorCode"`
	Msg       string                  `json:"msg"`
	Message   string                  `json:"message"`
	Data      *service.Trafficcontrol `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse = Wrapped
type GetResponse = Wrapped
type ListResponse = struct {
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Msg       string      `json:"msg"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
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

func (c *controller) CreateTrafficcontrol(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "012000001",
			Msg:       c.errMsg.Trafficcontrol["012000001"],
			Message:   fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "012000002",
			Msg:       c.errMsg.Trafficcontrol["012000002"],
			Message:   "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "012000003",
			Msg:       c.errMsg.Trafficcontrol["012000003"],
			Message:   "auth model error",
		}
	}
	body.Data.Users = user.InitWithOwner(authuser.Name)
	if db, err := c.service.CreateTrafficcontrol(body.Data); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: "012000004",
			Msg:       c.errMsg.Trafficcontrol["012000004"],
			Message:   fmt.Errorf("create trafficcontrol error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      db,
		}
	}
}

func (c *controller) GetTrafficcontrol(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if db, err := c.service.GetTrafficcontrol(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      2,
			ErrorCode: "012000005",
			Msg:       c.errMsg.Trafficcontrol["012000005"],
			Message:   fmt.Errorf("get trafficcontrol error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      db,
		}
	}
}

func (c *controller) DeleteTrafficcontrol(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if err := c.service.DeleteTrafficcontrol(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      2,
			ErrorCode: "012000006",
			Msg:       c.errMsg.Trafficcontrol["012000006"],
			Message:   fmt.Errorf("delete trafficcontrol error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:      0,
			ErrorCode: "0",
		}
	}
}

func (c *controller) ListTrafficcontrol(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	name := req.QueryParameter("name")
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: "012000003",
			Msg:       c.errMsg.Trafficcontrol["012000003"],
			Message:   "auth model error",
		}
	}
	if tc, err := c.service.ListTrafficcontrol(util.WithNameLike(name), util.WithUser(authuser.Name)); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      2,
			ErrorCode: "012000007",
			Msg:       c.errMsg.Trafficcontrol["012000007"],
			Message:   fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var tcs TrafficcontrolList = tc
		data, err := util.PageWrap(tcs, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      3,
				ErrorCode: "012000008",
				Msg:       c.errMsg.Trafficcontrol["012000008"],
				Message:   fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      data,
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
			Code:      1,
			ErrorCode: "012000001",
			Msg:       c.errMsg.Trafficcontrol["012000001"],
			Message:   fmt.Errorf("cannot read entity: %+v, reqbody:%v, req:%v", err, reqBody, req).Error(),
		}
	}
	data, ok := reqBody["data"]
	if !ok {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: "012000002",
			Msg:       c.errMsg.Trafficcontrol["012000002"],
			Message:   "read entity error: data is null",
		}
	}

	if db, err := c.service.UpdateTrafficcontrol(req.PathParameter("id"), data); err != nil {
		return http.StatusInternalServerError, &UpdateResponse{
			Code:      2,
			ErrorCode: "012000009",
			Msg:       c.errMsg.Trafficcontrol["012000009"],
			Message:   fmt.Errorf("update trafficcontrol error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &UpdateResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      db,
		}
	}
}

func (c *controller) BindOrUnbindApis(req *restful.Request) (int, interface{}) {
	body := &BindRequest{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      1,
			ErrorCode: "012000001",
			Msg:       c.errMsg.Trafficcontrol["012000001"],
			Message:   fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	trafficID := req.PathParameter("id")
	if api, err := c.service.BindOrUnbindApis(body.Data.Operation, trafficID, body.Data.Apis); err != nil {
		return http.StatusInternalServerError, &BindResponse{
			Code:      2,
			ErrorCode: "012000010",
			Msg:       c.errMsg.Trafficcontrol["012000010"],
			Message:   fmt.Errorf("bind or unbind api error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &BindResponse{
			Code:      0,
			ErrorCode: "0",
			Data:      api,
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
