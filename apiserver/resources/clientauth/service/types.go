package service

import (
	"fmt"
	"github.com/chinamobile/nlpt/crds/clientauth/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"
	"time"
)

const (
	DefaultNamespace = "default"
)

type Clientauth struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Namespace     string         `json:"namespace"`
	CreatedAt     int64          `json:"createdAt"`
	IssuedAt      int64          `json:"issuedAt"`
	ExpireAt      int64          `json:"expireAt"`
	Token         string         `json:"token"`
	Effective     bool           `json:"effective"`
	AuthorizedMap map[string]int `json:"authorizedMap"` //已授权信息，key：topic id，value：1
	Status        v1.Status      `json:"status"`
	Message       string         `json:"message"`
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
		Token:    app.Token,
		ExipreAt: app.ExpireAt,
		IssuedAt: app.IssuedAt,
	}
	if len(crd.Spec.Namespace) == 0 {
		crd.Spec.Namespace = DefaultNamespace
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Created
	}
	crd.Status = v1.ClientauthStatus{
		Status:  status,
		Message: "success",
	}
	return crd
}

func ToModel(obj *v1.Clientauth) *Clientauth {
	return &Clientauth{
		ID:            obj.ObjectMeta.Name,
		Name:          obj.Spec.Name,
		Namespace:     obj.Spec.Namespace,
		CreatedAt:     util.NewTime(obj.ObjectMeta.CreationTimestamp.Time).Unix(),
		IssuedAt:      obj.Spec.IssuedAt,
		ExpireAt:      obj.Spec.ExipreAt,
		Token:         obj.Spec.Token,
		Status:        obj.Status.Status,
		Message:       obj.Status.Message,
		AuthorizedMap: obj.Spec.AuthorizedMap,
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
		"exp":  string(a.ExpireAt),
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}

	//校验时间，token的过期时间必须大于当前时间
	if a.ExpireAt <= time.Now().Unix() {
		return fmt.Errorf("token expire time:%d must be greater than now", a.ExpireAt)
	}

	a.ID = names.NewID()
	return nil
}
