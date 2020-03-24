package clientauth

import (
	"fmt"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
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
	authUser,err := auth.GetAuthUser(req)
	if err!=nil{
		return http.StatusInternalServerError, &CreateResponse{
			Code:    1,
			Message: fmt.Errorf("auth model error: %+v", err).Error(),
		}
	}
    body.CreateUser = user.InitWithOwner(authUser.Name)
    body.Tenant = authUser.Namespace
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
	createTimeSta := req.QueryParameter("createTimeSta")
	createTimeEnd := req.QueryParameter("createTimeEnd")
	expireAtSta := req.QueryParameter("expireAtSta")
	expireAtEnd := req.QueryParameter("expireAtEnd")
	createUser := req.QueryParameter("createUser")
	tenant := req.QueryParameter("tenant")
	//先查出所有用户信息
	ca, err := c.service.ListClientauth()
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:    fail,
			Message: fmt.Errorf("list database error: %+v", err).Error(),
		}
	}
	//通过创建用户筛选
	if len(createUser)>0 {
		ca = c.ListTopicByCreateUser(createUser,ca)
	}
	//通过租户筛选
	if len(tenant)>0 {
		ca = c.ListTopicByTenant(tenant,ca)
	}
	//接收参数只有authUser
	if len(authUser) > 0 && len(createTimeSta) == 0 && len(expireAtSta) == 0 {
		//通过ca字段来匹配
		ca = c.ListTopicByauthUser(authUser, ca)
	} else if len(expireAtSta) > 0 && len(authUser) == 0 && len(createTimeSta) == 0 { //接收参数只有expireAtSta
		//通过expireAtSta字段来匹配
		eas, err := strconv.ParseInt(expireAtSta, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		eae, err := strconv.ParseInt(expireAtEnd, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicBytokenExp(eas, eae, ca)
	} else if len(createTimeSta) > 0 && len(expireAtSta) == 0 && len(authUser) == 0 { //接收参数只有时间
		cts, err := strconv.ParseInt(createTimeSta, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		cte, err := strconv.ParseInt(createTimeEnd, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicBycreateTime(cts, cte, ca)
	} else if len(authUser) > 0 && len(expireAtSta) > 0 && len(createTimeSta) == 0 { //接收参数有authUser,expireAtSta
		eas, err := strconv.ParseInt(expireAtSta, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		eae, err := strconv.ParseInt(expireAtEnd, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicByauthUser(authUser, ca)
		ca = c.ListTopicBytokenExp(eas, eae, ca)
	} else if len(expireAtSta) > 0 && len(authUser) == 0 && len(createTimeSta) > 0 { //接收参数有expireAtSta和时间
		eas, err := strconv.ParseInt(expireAtSta, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		eae, err := strconv.ParseInt(expireAtEnd, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		cts, err := strconv.ParseInt(createTimeSta, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		cte, err := strconv.ParseInt(createTimeEnd, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicBytokenExp(eas, eae, ca)
		ca = c.ListTopicBycreateTime(cts, cte, ca)
	} else if len(expireAtSta) == 0 && len(authUser) > 0 && len(createTimeSta) > 0 { //接收参数有authUser和时间
		cts, err := strconv.ParseInt(createTimeSta, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		cte, err := strconv.ParseInt(createTimeEnd, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicByauthUser(authUser, ca)
		ca = c.ListTopicBycreateTime(cts, cte, ca)
	} else if len(authUser) > 0 && len(expireAtSta) > 0 && len(createTimeSta) > 0 { //接收参数有authUser、expireAtSta、时间
		eas, err := strconv.ParseInt(expireAtSta, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		eae, err := strconv.ParseInt(expireAtEnd, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		cts, err := strconv.ParseInt(createTimeSta, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		cte, err := strconv.ParseInt(createTimeEnd, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:    fail,
				Message: fmt.Errorf("string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicByauthUser(authUser, ca)
		ca = c.ListTopicBytokenExp(eas, eae, ca)
		ca = c.ListTopicBycreateTime(cts, cte, ca)
	}
	//判断token是否有效
	for _, cla := range ca {
		if cla.ExpireAt > util.Now().Unix() {
			cla.Effective = true
		} else {
			cla.Effective = false
		}
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
//通过createUser匹配
func (c *controller) ListTopicByCreateUser(ctUser string, cas []*service.Clientauth) []*service.Clientauth {
	var casResult []*service.Clientauth
	for _, ca := range cas {
		if strings.Contains(ca.CreateUser.Owner.Name, ctUser) {
			casResult = append(casResult, ca)
		}
	}
	return casResult
}

//通过tenant匹配
func (c *controller) ListTopicByTenant(tenant string, cas []*service.Clientauth) []*service.Clientauth {
	var casResult []*service.Clientauth
	for _, ca := range cas {
		if strings.Contains(ca.Tenant, tenant) {
			casResult = append(casResult, ca)
		}
	}
	return casResult
}
//通过tokenExp匹配
func (c *controller) ListTopicBytokenExp(expireAtSta int64, expireAtEnd int64, cas []*service.Clientauth) []*service.Clientauth {
	var casResult []*service.Clientauth
	for _, ca := range cas {
		if ca.ExpireAt < expireAtEnd && ca.ExpireAt > expireAtSta {
			casResult = append(casResult, ca)
		}
	}
	return casResult
}

//根据createTime匹配
func (c *controller) ListTopicBycreateTime(createTimeSta int64, createTimeEnd int64, cas []*service.Clientauth) []*service.Clientauth {
	var casResult []*service.Clientauth
	for _, ca := range cas {
		if ca.CreatedAt < createTimeEnd && ca.CreatedAt > createTimeSta {
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
