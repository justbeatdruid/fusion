package service

import (
	"errors"
	"fmt"
	"github.com/chinamobile/nlpt/crds/clientauth/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
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
	CreateUser    user.Users     `json:"createUser"` //创建用户
	CreatedAt     int64          `json:"createdAt"`
	IssuedAt      int64          `json:"issuedAt"`
	ExpireAt      int64          `json:"expireAt"`
	Token         string         `json:"token"`
	Effective     bool           `json:"effective"`
	AuthorizedMap map[string]int `json:"authorizedMap"` //已授权信息，key：topic id，value：1
	Status        v1.Status      `json:"status"`
	Message       string         `json:"message"`
	Description   string         `json:"description"` //描述
	IsPermanent   bool           `json:"isPermanent"` //token是否永久有效
}

// only used in creation options
func ToAPI(app *Clientauth) *v1.Clientauth {
	crd := &v1.Clientauth{}
	crd.TypeMeta.Kind = "Clientauth"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = app.Namespace

	crd.Spec = v1.ClientauthSpec{
		Name:        app.Name,
		Token:       app.Token,
		ExipreAt:    app.ExpireAt,
		IssuedAt:    app.IssuedAt,
		Description: app.Description,
	}
	if len(crd.Namespace) == 0 {
		crd.Namespace = DefaultNamespace
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Created
	}
	crd.Status = v1.ClientauthStatus{
		Status:  status,
		Message: "success",
	}
	crd.ObjectMeta.Labels = user.AddUsersLabels(app.CreateUser, crd.ObjectMeta.Labels)
	return crd
}

func ToModel(obj *v1.Clientauth) *Clientauth {
	ca := &Clientauth{
		ID:          obj.ObjectMeta.Name,
		Name:        obj.Spec.Name,
		CreateUser:  user.GetUsersFromLabels(obj.ObjectMeta.Labels),
		Namespace:   obj.Namespace,
		CreatedAt:   util.NewTime(obj.ObjectMeta.CreationTimestamp.Time).Unix(),
		IssuedAt:    obj.Spec.IssuedAt,
		ExpireAt:    obj.Spec.ExipreAt,
		Token:       obj.Spec.Token,
		Status:      obj.Status.Status,
		Message:     obj.Status.Message,
		Description: obj.Spec.Description,
	}

	if obj.Spec.AuthorizedMap != nil {
		ca.AuthorizedMap = *obj.Spec.AuthorizedMap
	}
	return ca
}

func ToListModel(items *v1.ClientauthList) []*Clientauth {
	var app  = make([]*Clientauth, len(items.Items))
	for i := range items.Items {
		app[i] = ToModel(&items.Items[i])
	}
	return app
}

func (a *Clientauth) Validate() error {
	for k, v := range map[string]string{
		"name": a.Name,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	if a.ExpireAt == 0 && a.IsPermanent==false{
		return errors.New("ExpireAt is null")
	}
	//校验时间，token的过期时间必须大于当前时间
	if a.ExpireAt != 0 && a.ExpireAt <= time.Now().Unix() {
		return fmt.Errorf("token expire time:%d must be greater than now", a.ExpireAt)
	}

	a.ID = names.NewID()
	return nil
}
