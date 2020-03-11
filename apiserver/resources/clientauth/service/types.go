package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/crds/clientauth/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"
	"time"
)

type Clientauth struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	CreatedAt int64     `json:"createdAt"`
	TokenIat  int64     `json:"tokenIat"`
	TokenExp  int64     `json:"tokenExp"`
	Token     string    `json:"token"`
	Status    v1.Status `json:"status"`
	Message   string    `json:"message"`
}

// only used in creation options
func ToAPI(app *Clientauth) *v1.Clientauth {
	crd := &v1.Clientauth{}
	crd.TypeMeta.Kind = "Clientauth"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = crdNamespace

	crd.Spec = v1.ClientauthSpec{
		Name:     app.Name,
		TokenExp: app.TokenExp,
		TokenIat: app.TokenIat,
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Init
	}
	crd.Status = v1.ClientauthStatus{
		Status:  status,
		Message: app.Message,
	}
	return crd
}

func ToModel(obj *v1.Clientauth) *Clientauth {
	return &Clientauth{
		ID:        obj.ObjectMeta.Name,
		Name:      obj.Spec.Name,
		Namespace: obj.Spec.Namespace,
		CreatedAt: util.NewTime(obj.ObjectMeta.CreationTimestamp.Time).Unix(),
		TokenIat:  obj.Spec.TokenIat,
		TokenExp:  obj.Spec.TokenExp,
		Token:     obj.Spec.Token,
		Status:    obj.Status.Status,
		Message:   obj.Status.Message,
	}
}

func ToListModel(items *v1.ClientauthList) []*Clientauth {
	var app []*Clientauth = make([]*Clientauth, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (a *Clientauth) Validate() error {
	for k, v := range map[string]string{
		"name": a.Name,
		"iat":  string(a.TokenIat),
		"exp":  string(a.TokenExp),
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}

	//签发时间必须不小于当前时间
	now := time.Now().Unix()
	if a.TokenIat < now {
		return fmt.Errorf("token issued time:%d must be greater than now.", a.TokenIat)
	}

	//校验时间，token的过期时间必须大于签发时间
	if a.TokenExp <= a.TokenIat {
		return fmt.Errorf("token expire time:%d must be greater than issued time:%d.", a.TokenExp, a.TokenIat)
	}

	a.ID = names.NewID()
	return nil
}
