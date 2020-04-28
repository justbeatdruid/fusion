package cas

import (
	"fmt"

	"github.com/chinamobile/nlpt/pkg/logs"

	"github.com/parnurzeal/gorequest"
	"k8s.io/klog"
)

const getTenantPath = "tenant-manager/sys/support/userInfo/%s"
const listTenantPath = "tenant-manager/sys/support/userList"

type TenantAutoGenerated struct {
	Msg  string     `json:"msg"`
	Code int        `json:"code"`
	User TenantUser `json:"user"`
}

type TenantUserList struct {
	Msg  string       `json:"msg"`
	Code int          `json:"code"`
	Page []TenantUser `json:"page"`
}

type TenantUser struct {
	UserID            int    `json:"userId"`
	Username          string `json:"username"`
	Password          string `json:"password"`
	Salt              string `json:"salt"`
	Email             string `json:"email"`
	Mobile            string `json:"mobile,omitempty"`
	Status            int    `json:"status,omitempty"`
	RoleIDList        string `json:"roleIdList,omitempty"`
	CreateUserID      int    `json:"createUserId"`
	CreateUserAccount string `json:"createUserAccount,omitempty"`
	CreateTime        string `json:"createTime"`
	RoleName          string `json:"roleName,omitempty"`
	GroupName         string `json:"groupName,omitempty"`
}

func FromTenantUser(c TenantUser) User {
	return User{
		UserID:     c.UserID,
		Username:   c.Username,
		Password:   c.Password,
		Salt:       c.Salt,
		Status:     c.Status,
		Email:      c.Email,
		Mobile:     c.Mobile,
		CreateTime: c.CreateTime,
	}
}

func FromTenantUserList(c []TenantUser) []User {
	ul := make([]User, len(c))
	for i, u := range c {
		ul[i] = FromTenantUser(u)
	}
	return ul
}

type tenant struct{}

func NewTenantOperator() Operator {
	return &tenant{}
}

func (*tenant) GetUserByID(id string) (User, error) {
	request := gorequest.New().SetLogger(logs.GetGoRequestLogger(6)).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Get(fmt.Sprintf("%s://%s:%d/%s", schema, casHost, casPort, fmt.Sprintf(getTenantPath, id)))

	responseBody := &TenantAutoGenerated{}
	response, body, errs := request.EndStruct(responseBody)
	if len(errs) > 0 {
		return User{}, fmt.Errorf("request for getting tenant user error: %+v", errs)
	}
	if response.StatusCode/100 != 2 {
		klog.V(5).Infof("create operation failed: %d %s", response.StatusCode, string(body))
		return User{}, fmt.Errorf("request for getting tenant user error: receive wrong status code: %s", string(body))
	}
	if responseBody.Code != 0 {
		return User{}, fmt.Errorf("request for getting tenant user error: received cod is not 200: message: %s", responseBody.Msg)
	}
	return FromTenantUser(responseBody.User), nil
}

func (*tenant) ListUsers() ([]User, error) {
	request := gorequest.New().SetLogger(logs.GetGoRequestLogger(6)).SetDebug(true).SetCurlCommand(true)
	schema := "http"
	request = request.Get(fmt.Sprintf("%s://%s:%d/%s", schema, casHost, casPort, listTenantPath)).Query("page=0").Query("limit=999999")

	responseBody := &TenantUserList{}
	response, body, errs := request.EndStruct(responseBody)
	if len(errs) > 0 {
		return nil, fmt.Errorf("request for getting tenant user error: %+v", errs)
	}
	if response.StatusCode/100 != 2 {
		klog.V(5).Infof("create operation failed: %d %s", response.StatusCode, string(body))
		return nil, fmt.Errorf("request for getting tenant user error: receive wrong status code: %s", string(body))
	}
	if responseBody.Code != 0 {
		return nil, fmt.Errorf("request for getting tenant user error: received cod is not 200: message: %s", responseBody.Msg)
	}
	return FromTenantUserList(responseBody.Page), nil
}
