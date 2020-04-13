package apply

import (
	"fmt"
	"net/http"

	"github.com/chinamobile/nlpt/apiserver/resources/apply/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/util"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
	errCode map[string]string
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.GetKubeClient(), cfg.TenantEnabled),
		cfg.LocalConfig.Apply,
	}
}

type Wrapped struct {
	Code      int            `json:"code"`
	ErrorCode string         `json:"errorCode"`
	Message   string         `json:"message"`
	Detail    string         `json:"detail"`
	Data      *service.Apply `json:"data,omitempty"`
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
type PingResponse = DeleteResponse
type ApproveRequest struct {
	Data *struct {
		Admitted bool   `json:"admitted"`
		Reason   string `json:"reason"`
	} `json:"data"`
	Code      int    `json:"code"`
	ErrorCode string `json:"errorCode"`
	Message   string `json:"message"`
	Detail    string `json:"detail"`
}
type ApproveResponse = ApproveRequest

func (c *controller) CreateApply(req *restful.Request) (int, *CreateResponse) {
	body := &CreateRequest{}
	if err := req.ReadEntity(body); err != nil {
		code := "004000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
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
		code := "004000005"
		return http.StatusInternalServerError, &CreateResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	body.Data.Users = user.InitWithApplicant(authuser.Name)
	if apl, err := c.service.CreateApply(body.Data); err != nil {
		code := "004000002"
		if errors.IsAlreadyBound(err) {
			code = "004000003"
		}
		if errors.IsPermissionDenied(err) {
			code = "004000004"
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:      2,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    fmt.Errorf("create apply error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: 0,
			Data: apl,
		}
	}
}

func (c *controller) GetApply(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if apl, err := c.service.GetApply(id); err != nil {
		code := "004000007"
		return http.StatusInternalServerError, &GetResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    fmt.Errorf("get apply error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: 0,
			Data: apl,
		}
	}
}

func (c *controller) DeleteApply(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if err := c.service.DeleteApply(id); err != nil {
		code := "004000002"
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    fmt.Errorf("delete apply error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code:   0,
			Detail: "",
		}
	}
}

func (c *controller) ListApply(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	role := req.QueryParameter("role")
	if len(role) == 0 {
		code := "004000005"
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "need role in query parameter: applicant or approver",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "004000006"
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	if apl, err := c.service.ListApply(role, util.WithUser(authuser.Name)); err != nil {
		code := "004000002"
		return http.StatusInternalServerError, &ListResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    fmt.Errorf("list applies error: %+v", err).Error(),
		}
	} else {
		var apls ApplyList = apl
		data, err := util.PageWrap(apls, page, size)
		if err != nil {
			code := "004000005"
			return http.StatusInternalServerError, &ListResponse{
				Code:      1,
				ErrorCode: code,
				Message:   c.errCode[code],
				Detail:    fmt.Sprintf("page parameter error: %+v", err),
			}
		}
		return http.StatusOK, &ListResponse{
			Code: 0,
			Data: data,
		}
	}
}

type ApplyList []*service.Apply

func (apls ApplyList) Len() int {
	return len(apls)
}

func (apls ApplyList) GetItem(i int) (interface{}, error) {
	if i >= len(apls) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return apls[i], nil
}

func (apls ApplyList) Less(i, j int) bool {
	return apls[i].AppliedAt.Time.After(apls[j].AppliedAt.Time)
}

func (apls ApplyList) Swap(i, j int) {
	apls[i], apls[j] = apls[j], apls[i]
}

func (c *controller) ApproveApply(req *restful.Request) (int, *ApproveResponse) {
	id := req.PathParameter("id")
	body := &ApproveRequest{}
	if err := req.ReadEntity(body); err != nil {
		code := "004000005"
		return http.StatusInternalServerError, &ApproveResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if body.Data == nil {
		code := "004000005"
		return http.StatusInternalServerError, &ApproveResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "read entity error: data is null",
		}
	}
	authuser, err := auth.GetAuthUser(req)
	if err != nil {
		code := "004000006"
		return http.StatusInternalServerError, &ApproveResponse{
			Code:      1,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    "auth model error",
		}
	}
	if _, err := c.service.ApproveApply(id, body.Data.Admitted, body.Data.Reason, util.WithUser(authuser.Name)); err != nil {
		code := "004000007"
		if errors.IsPermissionDenied(err) {
			code = "004000004"
		}
		return http.StatusInternalServerError, &ApproveResponse{
			Code:      2,
			ErrorCode: code,
			Message:   c.errCode[code],
			Detail:    fmt.Errorf("create apply error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &ApproveResponse{
			Code: 0,
			Data: nil,
		}
	}
}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
