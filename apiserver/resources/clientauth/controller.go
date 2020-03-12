package clientauth

import (
	"fmt"
	"github.com/chinamobile/nlpt/pkg/util"
	"net/http"
	"strconv"
	"strings"

	"github.com/chinamobile/nlpt/apiserver/resources/clientauth/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

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

const (
	success = iota
	fail
)

type Wrapped struct {
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Data    *service.Clientauth `json:"data,omitempty"`
}

type CreateResponse = Wrapped
type CreateRequest = Wrapped
type DeleteResponse = Wrapped

/*type DeleteResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}*/
type GetResponse = Wrapped
type ListResponse = struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
type PingResponse = DeleteResponse

type ClientauthList []*service.Clientauth

func (cas ClientauthList) Len() int {
	return len(cas)
}

func (cas ClientauthList) GetItem(i int) (interface{}, error) {
	if i >= len(cas) {
		return struct{}{}, fmt.Errorf("index overflow")
	}
	return cas[i], nil
}

func (c *controller) CreateClientauth(req *restful.Request) (int, *CreateResponse) {
	body := &service.Clientauth{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	if ca, err := c.service.CreateClientauth(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    2,
			Message: fmt.Errorf("create database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &CreateResponse{
			Code: success,
			Data: ca,
		}
	}
}

func (c *controller) GetClientauth(req *restful.Request) (int, *GetResponse) {
	id := req.PathParameter("id")
	if ca, err := c.service.GetClientauth(id); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:    fail,
			Message: fmt.Errorf("get database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: success,
			Data: ca,
		}
	}
}

//批量删除clientauths
func (c *controller) DeleteClientauths(req *restful.Request) (int, *ListResponse) {
	ids := req.QueryParameters("ids")
	for _, id := range ids {
		if _, err := c.service.Delete(id); err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("delete clientauth error: %+v", err).Error(),
			}
		}
	}
	return http.StatusOK, &ListResponse{
		Code:    success,
		Message: "delete clientauth success",
	}
}
func (c *controller) DeleteClientauth(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	if ca, err := c.service.DeleteClientauth(id); err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:    fail,
			Message: fmt.Errorf("delete clientauth error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: success,
			Data: ca,
		}
	}
}

func (c *controller) ListClientauth(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")

	if ca, err := c.service.ListClientauth(); err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    fail,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	} else {
		var cas ClientauthList = ca

		data, err := util.PageWrap(cas, page, size)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Sprintf("page parameter error: %+v", err),
			}
		} else {
			return http.StatusOK, &ListResponse{
				Code: success,
				Data: data,
			}
		}
	}
}

//模糊查询
func (c *controller) ListClientauths(req *restful.Request) (int, *ListResponse) {
	page := req.QueryParameter("page")
	size := req.QueryParameter("size")
	authUser := req.QueryParameter("authUser")
	createTime := req.QueryParameter("createTime")
	tokenExp := req.QueryParameter("tokenExp")
	//先查出所有用户信息
	ca, err := c.service.ListClientauth()
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    fail,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	}
	//接收参数只有clientAuth
	if len(authUser) > 0 && len(createTime) == 0 && len(tokenExp) == 0 {
		//通过ca字段来匹配
		ca = c.ListTopicByauthUser(authUser, ca)
	} else if len(tokenExp) > 0 && len(authUser) == 0 && len(createTime) == 0 { //接收参数只有tokenExp
		//通过tokenExp字段来匹配
		te, err := strconv.ParseInt(tokenExp, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicBytokenExp(te, ca)
	} else if len(createTime) > 0 && len(tokenExp) == 0 && len(authUser) == 0 { //接收参数只有时间
		ct, err := strconv.ParseInt(createTime, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicBycreateTime(ct, ca)
	} else if len(authUser) > 0 && len(tokenExp) > 0 && len(createTime) == 0 { //接收参数有authUser,tokenExp
		te, err := strconv.ParseInt(tokenExp, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicByauthUser(authUser, ca)
		ca = c.ListTopicBytokenExp(te, ca)
	} else if len(tokenExp) > 0 && len(authUser) == 0 && len(createTime) > 0 { //接收参数有tokenExp和时间
		te, err := strconv.ParseInt(tokenExp, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ct, err := strconv.ParseInt(createTime, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicBytokenExp(te, ca)
		ca = c.ListTopicBycreateTime(ct, ca)
	} else if len(tokenExp) == 0 && len(authUser) > 0 && len(createTime) > 0 { //接收参数有authUser和时间
		ct, err := strconv.ParseInt(createTime, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicByauthUser(authUser, ca)
		ca = c.ListTopicBycreateTime(ct, ca)
	} else if len(authUser) > 0 && len(tokenExp) > 0 && len(createTime) > 0 { //接收参数有authUser、tokenExp、时间
		te, err := strconv.ParseInt(tokenExp, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ct, err := strconv.ParseInt(createTime, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicByauthUser(authUser, ca)
		ca = c.ListTopicBytokenExp(te, ca)
		ca = c.ListTopicBycreateTime(ct, ca)
	}
	var cas ClientauthList = ca
	data, err := util.PageWrap(cas, page, size)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    fail,
			Message: fmt.Sprintf("page parameter error: %+v", err),
		}
	} else {
		return http.StatusOK, &ListResponse{
			Code: success,
			Data: data,
		}
	}
}

//通过authUser匹配
func (c *controller) ListTopicByauthUser(caName string, cas []*service.Clientauth) []*service.Clientauth {
	var casResult []*service.Clientauth
	for _, ca := range cas {
		if strings.Contains(ca.Name, caName) {
			casResult = append(casResult, ca)
		}
	}
	return casResult
}

//通过tokenExp匹配
func (c *controller) ListTopicBytokenExp(tokenExp int64, cas []*service.Clientauth) []*service.Clientauth {
	var casResult []*service.Clientauth
	for _, ca := range cas {
		if ca.TokenExp < tokenExp {
			casResult = append(casResult, ca)
		}
	}
	return casResult
}

//根据createTime匹配
func (c *controller) ListTopicBycreateTime(createTime int64, cas []*service.Clientauth) []*service.Clientauth {
	var casResult []*service.Clientauth
	for _, ca := range cas {
		if ca.CreatedAt < createTime {
			casResult = append(casResult, ca)
		}
	}
	return casResult
}
func (c *controller) RegenerateToken(req *restful.Request) (int, *CreateResponse) {
	id := req.PathParameter("id")
	body := &service.Clientauth{}
	if err := req.ReadEntity(body); err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    fail,
			Message: fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	body.ID = id
	ca, err := c.service.RegenerateToken(body)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:    fail,
			Message: fmt.Sprintf("regenerate token error: %+v", err),
		}
	}

	return http.StatusOK, &CreateResponse{
		Code:    success,
		Data:    ca,
		Message: "success",
	}

}

func returns200(b *restful.RouteBuilder) {
	b.Returns(http.StatusOK, "OK", "success")
}

func returns500(b *restful.RouteBuilder) {
	b.Returns(http.StatusInternalServerError, "internal server error", nil)
}
