package auth

import (
	"fmt"

	"github.com/chinamobile/nlpt/pkg/go-restful"
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
		if len(user.Name) == 0 {
			return AuthUser{}, fmt.Errorf("cannot get userId from headers")
		}
		return user, nil
	}
	return AuthUser{}, fmt.Errorf("auth user not set by token filter")
}
