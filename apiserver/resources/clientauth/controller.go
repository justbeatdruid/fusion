package clientauth

import (
	"fmt"
	"github.com/chinamobile/nlpt/pkg/auth"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/util"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/chinamobile/nlpt/apiserver/resources/clientauth/service"
	"github.com/chinamobile/nlpt/cmd/apiserver/app/config"

	"github.com/chinamobile/nlpt/pkg/go-restful"
)

type controller struct {
	service *service.Service
	errMsg  config.ErrorConfig
}

func newController(cfg *config.Config) *controller {
	return &controller{
		service.NewService(cfg.GetDynamicClient(), cfg.Client),
		cfg.LocalConfig,
	}
}

const (
	success = iota
	fail
)

type Wrapped struct {
	Code      int                 `json:"code"`
	ErrorCode string              `json:"errorCode"`
	Detail    string              `json:"detail"`
	Message   string              `json:"message"`
	Data      *service.Clientauth `json:"data,omitempty"`
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
	Code      int         `json:"code"`
	ErrorCode string      `json:"errorCode"`
	Detail    string      `json:"detail"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
}
type PingResponse = DeleteResponse

type ClientauthList []*service.Clientauth

func (cas ClientauthList) Len() int {
	return len(cas)
}

func (cas ClientauthList) Swap(i, j int) {
	cas[i], cas[j] = cas[j], cas[i]
}
func (cas ClientauthList) Less(i, j int) bool {
	return cas[j].CreatedAt < cas[i].CreatedAt
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
			Code:      fail,
			ErrorCode: "013000004",
			Message:   c.errMsg.ClientAuth["013000004"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      fail,
			ErrorCode: "013000003",
			Message:   c.errMsg.ClientAuth["013000003"],
			Detail:    fmt.Errorf("auth model error: %+v", err).Error(),
		}
	}
	body.CreateUser = user.InitWithOwner(authUser.Name)
	body.Tenant = authUser.Namespace
	body.Namespace = authUser.Namespace
	if ca, err := c.service.CreateClientauth(body); err != nil {
		if strings.Contains(err.Error(), "username already exists") {
			return http.StatusInternalServerError, &CreateResponse{
				Code:      fail,
				ErrorCode: "013000011",
				Message:   c.errMsg.ClientAuth["013000011"],
				Detail:    fmt.Errorf("create database error: %+v", err).Error(),
			}
		}
		if strings.Contains(err.Error(), "token expire time") {
			return http.StatusInternalServerError, &CreateResponse{
				Code:      fail,
				ErrorCode: "013000012",
				Message:   c.errMsg.ClientAuth["013000012"],
				Detail:    fmt.Errorf("create database error: %+v", err).Error(),
			}
		}
		return http.StatusInternalServerError, &CreateResponse{
			Code:      fail,
			ErrorCode: "013000002",
			Message:   c.errMsg.ClientAuth["013000002"],
			Detail:    fmt.Errorf("create database error: %+v", err).Error(),
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
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      fail,
			ErrorCode: "013000014",
			Message:   c.errMsg.ClientAuth["013000014"],
			Detail:    fmt.Errorf("auth model error: %+v", err).Error(),
		}
	}
	if ca, err := c.service.GetClientauth(id, util.WithNamespace(authUser.Namespace)); err != nil {
		return http.StatusInternalServerError, &GetResponse{
			Code:      fail,
			ErrorCode: "013000005",
			Message:   c.errMsg.ClientAuth["013000005"],
			Detail:    fmt.Errorf("get database error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &GetResponse{
			Code: success,
			Data: ca,
		}
	}
}

//批量删除clientauths
/*func (c *controller) DeleteClientauths(req *restful.Request) (int, *ListResponse) {
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
}*/
func (c *controller) DeleteClientauth(req *restful.Request) (int, *DeleteResponse) {
	id := req.PathParameter("id")
	authUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      fail,
			ErrorCode: "013000014",
			Message:   c.errMsg.ClientAuth["013000014"],
			Detail:    fmt.Errorf("auth model error: %+v", err).Error(),
		}
	}
	if ca, err := c.service.DeleteClientauth(id, util.WithNamespace(authUser.Namespace)); err != nil {
		if strings.Contains(err.Error(), "cannot delete authorized client auth user") {
			return http.StatusInternalServerError, &DeleteResponse{
				Code:      fail,
				ErrorCode: "013000013",
				Message:   c.errMsg.ClientAuth["013000013"],
				Detail:    fmt.Errorf("delete clientauth error: %+v", err).Error(),
			}
		}
		return http.StatusInternalServerError, &DeleteResponse{
			Code:      fail,
			ErrorCode: "013000006",
			Message:   c.errMsg.ClientAuth["013000006"],
			Detail:    fmt.Errorf("delete clientauth error: %+v", err).Error(),
		}
	} else {
		return http.StatusOK, &DeleteResponse{
			Code: success,
			Data: ca,
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
	AuthUser, err := auth.GetAuthUser(req)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      fail,
			ErrorCode: "013000014",
			Message:   c.errMsg.ClientAuth["013000014"],
			Detail:    fmt.Errorf("auth model error: %+v", err).Error(),
		}
	}

	//先查出所有用户信息
	ca, err := c.service.ListClientauth(util.WithNamespace(AuthUser.Namespace))
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      fail,
			ErrorCode: "013000007",
			Message:   c.errMsg.ClientAuth["013000007"],
			Detail:    fmt.Errorf("list database error: %+v", err).Error(),
		}
	}
	//创建用户筛选
	if len(createUser) > 0 {
		ca = c.ListTopicByCreateUser(createUser, ca)
	}
	//租户筛选
	if len(tenant) > 0 {
		ca = c.ListTopicByTenant(tenant, ca)
	}
	//authUser筛选
	if len(authUser) > 0 {
		//通过ca字段来匹配
		ca = c.ListTopicByauthUser(authUser, ca)
	}
	//token失效时间段筛选
	if len(expireAtSta) > 0 && len(expireAtEnd) > 0 {
		//通过expireAtSta字段来匹配
		eas, err := strconv.ParseInt(expireAtSta, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      fail,
				ErrorCode: "013000009",
				Message:   c.errMsg.ClientAuth["013000009"],
				Detail:    fmt.Errorf("expireAt string to int64 error: %+v", err).Error(),
			}
		}
		eae, err := strconv.ParseInt(expireAtEnd, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      fail,
				ErrorCode: "013000009",
				Message:   c.errMsg.ClientAuth["013000009"],
				Detail:    fmt.Errorf("expireAt string to int64 error: %+v", err).Error(),
			}
		}
		ca = c.ListTopicBytokenExp(eas, eae, ca)
	}
	//创建时间筛选
	if len(createTimeSta) > 0 && len(createTimeEnd) > 0 {
		cts, err := strconv.ParseInt(createTimeSta, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      fail,
				ErrorCode: "013000009",
				Message:   c.errMsg.ClientAuth["013000009"],
				Detail:    fmt.Errorf("createTime string to int64 error: %+v", err).Error(),
			}
		}
		cte, err := strconv.ParseInt(createTimeEnd, 10, 64)
		if err != nil {
			return http.StatusInternalServerError, &ListResponse{
				Code:      fail,
				ErrorCode: "013000009",
				Message:   c.errMsg.ClientAuth["013000009"],
				Detail:    fmt.Errorf("createTime string to int64 error: %+v", err).Error(),
			}
		}
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
	sort.Sort(cas)
	data, err := util.PageWrap(cas, page, size)
	if err != nil {
		return http.StatusInternalServerError, &ListResponse{
			Code:      fail,
			ErrorCode: "013000008",
			Message:   c.errMsg.ClientAuth["013000008"],
			Detail:    fmt.Sprintf("page parameter error: %+v", err),
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
	caName = strings.ToLower(caName)
	for _, ca := range cas {
		if strings.Contains(strings.ToLower(ca.Name), caName) {
			casResult = append(casResult, ca)
		}
	}
	return casResult
}

//通过createUser匹配
func (c *controller) ListTopicByCreateUser(ctUser string, cas []*service.Clientauth) []*service.Clientauth {
	var casResult []*service.Clientauth
	ctUser = strings.ToLower(ctUser)
	for _, ca := range cas {
		if strings.Contains(strings.ToLower(ca.CreateUser.Owner.Name), ctUser) {
			casResult = append(casResult, ca)
		}
	}
	return casResult
}

//通过tenant匹配
func (c *controller) ListTopicByTenant(tenant string, cas []*service.Clientauth) []*service.Clientauth {
	var casResult []*service.Clientauth
	tenant = strings.ToLower(tenant)
	for _, ca := range cas {
		if strings.Contains(strings.ToLower(ca.Tenant), tenant) {
			casResult = append(casResult, ca)
		}
	}
	return casResult
}

//通过tokenExp匹配
func (c *controller) ListTopicBytokenExp(expireAtSta int64, expireAtEnd int64, cas []*service.Clientauth) []*service.Clientauth {
	var casResult []*service.Clientauth
	for _, ca := range cas {
		if ca.ExpireAt <= expireAtEnd && ca.ExpireAt >= expireAtSta {
			casResult = append(casResult, ca)
		}
	}
	return casResult
}

//根据createTime匹配
func (c *controller) ListTopicBycreateTime(createTimeSta int64, createTimeEnd int64, cas []*service.Clientauth) []*service.Clientauth {
	var casResult []*service.Clientauth
	for _, ca := range cas {
		if ca.CreatedAt <= createTimeEnd && ca.CreatedAt >= createTimeSta {
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
			Code:      fail,
			ErrorCode: "013000005",
			Message:   c.errMsg.ClientAuth["013000005"],
			Detail:    fmt.Errorf("cannot read entity: %+v", err).Error(),
		}
	}
	body.ID = id
	ca, err := c.service.RegenerateToken(body)
	if err != nil {
		return http.StatusInternalServerError, &CreateResponse{
			Code:      fail,
			ErrorCode: "013000010",
			Message:   c.errMsg.ClientAuth["013000010"],
			Detail:    fmt.Sprintf("regenerate token error: %+v", err),
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
