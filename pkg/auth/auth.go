package auth

import (
	"fmt"

	"github.com/emicklei/go-restful"
)

type AuthUser struct {
	Name      string
	Namespace string
}

func SetAuthUser(req *restful.Request, u AuthUser) {
	req.SetAttribute("user", u)
}

func GetAuthUser(req *restful.Request) (AuthUser, error) {
	u := req.Attribute("user")
	if user, ok := u.(AuthUser); ok {
		return user, nil
	}
	return AuthUser{}, fmt.Errorf("auth user not set by token filter")
}
