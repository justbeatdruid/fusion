package service

import (
	"encoding/json"
	"fmt"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/errors"
	"regexp"
	"strings"

	v1 "github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const (
	NameReg = "^[a-zA-Z\u4e00-\u9fa5][a-zA-Z0-9_\u4e00-\u9fa5]{2,64}$"
)

type Trafficcontrol struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace"`
	Type        v1.LimitType  `json:"type"`
	Config      v1.ConfigInfo `json:"config"`
	Apis        []v1.Api      `json:"apis"`
	Description string        `json:"description"`
	Users       user.Users    `json:"users"`
	CreatedAt   util.Time     `json:"createdAt"`

	Status    v1.Status `json:"status"`
	UpdatedAt util.Time `json:"time"`
	APICount  int       `json:"apiCount"`
	Published bool      `json:"published"`
}

// only used in creation options
func ToAPI(app *Trafficcontrol) *v1.Trafficcontrol {
	crd := &v1.Trafficcontrol{}
	crd.TypeMeta.Kind = "Trafficcontrol"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version
	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = app.Namespace
	crd.ObjectMeta.Labels = make(map[string]string)
	crd.ObjectMeta.Labels[app.ID] = app.ID
	crd.Spec = v1.TrafficcontrolSpec{
		Name:        app.Name,
		Type:        app.Type,
		Config:      app.Config,
		Apis:        app.Apis,
		Description: app.Description,
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Init
	}
	crd.Status = v1.TrafficcontrolStatus{
		Status:    status,
		UpdatedAt: metav1.Now(),
		APICount:  0,
		Published: false,
	}
	if crd.Spec.Apis == nil {
		crd.Spec.Apis = make([]v1.Api, 0)
	}

	if crd.Spec.Config.Special == nil {
		crd.Spec.Config.Special = make([]v1.Special, 0)
	}
	// add user labels
	crd.ObjectMeta.Labels = user.AddUsersLabels(app.Users, crd.ObjectMeta.Labels)
	return crd
}

func ToModel(obj *v1.Trafficcontrol) *Trafficcontrol {
	for index, value := range obj.Spec.Apis {
		switch value.Result {
		case v1.BINDING:
			(*obj).Spec.Apis[index].DisplayStatus = v1.ApiBinding
		case v1.UNBINDING, v1.UPDATING, v1.SUCCESS:
			(*obj).Spec.Apis[index].DisplayStatus = v1.BindedSuccess
		case v1.UNBINDFAILED:
			(*obj).Spec.Apis[index].DisplayStatus = v1.UnBindFail
		case v1.BINDFAILED, v1.UPDATEFAILED:
			(*obj).Spec.Apis[index].DisplayStatus = v1.BindedFail
		}
	}
	for _, value := range obj.Spec.Apis {
		klog.V(5).Infof("get api config : %+v", value)
	}
	traffic := &Trafficcontrol{
		ID:          obj.ObjectMeta.Name,
		Name:        obj.Spec.Name,
		Namespace:   obj.ObjectMeta.Namespace,
		Type:        obj.Spec.Type,
		Config:      obj.Spec.Config,
		Apis:        obj.Spec.Apis,
		CreatedAt:   util.NewTime(obj.ObjectMeta.CreationTimestamp.Time),
		Description: obj.Spec.Description,

		Status:    obj.Status.Status,
		UpdatedAt: util.NewTime(obj.Status.UpdatedAt.Time),
		APICount:  obj.Status.APICount,
		Published: obj.Status.Published,
	}
	traffic.Users = user.GetUsersFromLabels(obj.ObjectMeta.Labels)
	return traffic
}
func ToListModel(items *v1.TrafficcontrolList, opts ...util.OpOption) []*Trafficcontrol {
	if len(opts) > 0 {
		var apps []*Trafficcontrol = make([]*Trafficcontrol, 0)
		nameLike := util.OpList(opts...).NameLike()
		if len(nameLike) > 0 {
			for _, item := range items.Items {
				if !strings.Contains(item.Spec.Name, nameLike) {
					continue
				}
				app := ToModel(&item)
				apps = append(apps, app)
			}
			return apps
		}
	}
	var apps []*Trafficcontrol = make([]*Trafficcontrol, len(items.Items))
	for i := range items.Items {
		apps[i] = ToModel(&items.Items[i])
	}
	return apps
}

// check create parameters
func (s *Service) Validate(a *Trafficcontrol) error {
	for k, v := range map[string]string{
		"name":        a.Name,
		"description": a.Description,
	} {
		if k == "name" {
			if len(v) == 0 {
				return fmt.Errorf("%s is null", k)
			} else if ok, _ := regexp.MatchString(NameReg, v); !ok {
				return fmt.Errorf("name is illegal: %v", v)
			}
		}
		if k == "description" {
			if len([]rune(v)) > 255 {
				return fmt.Errorf("%s cannot exceed 255 characters", k)
			}
		}
	}
	trafficList, errs := s.List(util.WithNamespace(a.Namespace))
	if errs != nil {
		return fmt.Errorf("cannot list trafficcontrol object: %+v", errs)
	}
	for _, t := range trafficList.Items {
		if t.Spec.Name == a.Name {
			return errors.NameDuplicatedError("trafficcontrol name duplicated: %+v", errs)
		}
	}
	if len(a.Type) == 0 {
		return fmt.Errorf("type is null")
	}

	if len(a.Users.Owner.ID) == 0 {
		return fmt.Errorf("owner not set")
	}

	switch a.Type {
	case v1.APIC, v1.APPC, v1.IPC, v1.USERC, v1.SPECAPPC:
	default:
		return fmt.Errorf("wrong type: %s.", a.Type)
	}

	switch a.Type {
	case v1.APIC, v1.APPC, v1.IPC, v1.USERC:
		if (a.Config.Year + a.Config.Month + a.Config.Day + a.Config.Hour + a.Config.Minute + a.Config.Second) == 0 {
			return fmt.Errorf("at least one limit config must exist.")
		} else {
			//list存秒分时日月年，list2存list的index
			var list []int
			list = append(list, a.Config.Second)
			list = append(list, a.Config.Minute)
			list = append(list, a.Config.Hour)
			list = append(list, a.Config.Day)
			list = append(list, a.Config.Month)
			list = append(list, a.Config.Year)
			var list2 []int
			for i, _ := range list {
				if list[i] != 0 {
					list2 = append(list2, i)
				}
			}
			var atime [6]string
			atime[0] = "秒"
			atime[1] = "分"
			atime[2] = "时"
			atime[3] = "日"
			atime[4] = "月"
			atime[5] = "年"
			for i := 0; i < len(list2)-1; i++ {
				if list[list2[i]] > list[list2[i+1]] {
					return fmt.Errorf("每%s的次数必须小于每%s的次数", atime[list2[i]], atime[list2[i+1]])
				}
			}
		}

	case v1.SPECAPPC:
		if len(a.Config.Special) == 0 {
			return fmt.Errorf("at least one special config must exist.")
		}
		if len(a.Config.Special) > v1.MAXNUM {
			return fmt.Errorf("special config maxnum limit exceeded.")
		}
		for _, spec := range a.Config.Special {
			if _, err := s.getApplication(spec.ID, a.Namespace); err != nil {
				return fmt.Errorf("get application for create traffic error: %+v", err)
			}
			if (spec.Year + spec.Month + spec.Day + spec.Hour + spec.Minute + spec.Second) == 0 {
				return fmt.Errorf("at least one limit config must exist.")
			} else {
				//list存秒分时日月年，list2存list的index
				var list []int
				list = append(list, spec.Second)
				list = append(list, spec.Minute)
				list = append(list, spec.Hour)
				list = append(list, spec.Day)
				list = append(list, spec.Month)
				list = append(list, spec.Year)
				var list2 []int
				for i, _ := range list {
					if list[i] != 0 {
						list2 = append(list2, i)
					}
				}
				var atime [6]string
				atime[0] = "秒"
				atime[1] = "分"
				atime[2] = "时"
				atime[3] = "日"
				atime[4] = "月"
				atime[5] = "年"
				for i := 0; i < len(list2)-1; i++ {
					if list[list2[i]] > list[list2[i+1]] {
						return fmt.Errorf("每%s的次数必须小于每%s的次数", atime[list2[i]], atime[list2[i+1]])
					}
				}
			}

		}
	default:
		return fmt.Errorf("wrong type for create traffic: %s.", a.Type)
	}

	a.ID = names.NewID()
	return nil
}

func (s *Service) assignment(target *v1.Trafficcontrol, reqData interface{}) error {
	data, ok := reqData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("reqData type is error,req data: %v", reqData)
	}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err)
	}
	var source Trafficcontrol
	if err = json.Unmarshal(b, &source); err != nil {
		return fmt.Errorf("json.Unmarshal error,: %v", err)
	}
	klog.V(5).Infof("get update data : %+v", data)
	if _, ok = data["name"]; ok {
		if ok, _ := regexp.MatchString(NameReg, source.Name); !ok {
			return fmt.Errorf("name is illegal: %v", source.Name)
		}
		if target.Spec.Name != source.Name {
			trafficList, errs := s.List(util.WithNamespace(target.ObjectMeta.Namespace))
			if errs != nil {
				return fmt.Errorf("cannot list trafficcontrol object: %+v", errs)
			}
			for _, t := range trafficList.Items {
				if t.Spec.Name == source.Name {
					return errors.NameDuplicatedError("trafficcontrol name duplicated: %+v", errs)
				}
			}
		}
		target.Spec.Name = source.Name
	}
	if _, ok = data["namespace"]; ok {
		target.ObjectMeta.Namespace = source.Namespace
	}
	if _, ok = data["type"]; ok {
		//target.Spec.Type = source.Type
		switch source.Type {
		case v1.APIC, v1.APPC, v1.IPC, v1.USERC:
			if target.Spec.Type != v1.SPECAPPC {
				target.Spec.Type = source.Type
			} else {
				return fmt.Errorf("special application type cannot be changed to other types")
			}
			if reqConfig, ok := data["config"]; ok {
				klog.V(5).Infof("get config : %+v", reqConfig)
				if config, ok := reqConfig.(map[string]interface{}); ok {
					if (source.Config.Year + source.Config.Month + source.Config.Day + source.Config.Hour + source.Config.Minute + source.Config.Second) == 0 {
						return fmt.Errorf("at least one limit config must exist.")
					}
					if _, ok = config["year"]; ok {
						target.Spec.Config.Year = source.Config.Year
					}
					if _, ok = config["month"]; ok {
						target.Spec.Config.Month = source.Config.Month
					}
					if _, ok = config["day"]; ok {
						target.Spec.Config.Day = source.Config.Day
					}
					if _, ok = config["hour"]; ok {
						target.Spec.Config.Hour = source.Config.Hour
					}
					if _, ok = config["minute"]; ok {
						target.Spec.Config.Minute = source.Config.Minute
					}
					if _, ok = config["second"]; ok {
						target.Spec.Config.Second = source.Config.Second
					}
				}
				target.Spec.Config.Special = make([]v1.Special, 0)
			}
		case v1.SPECAPPC:
			if target.Spec.Type != v1.SPECAPPC {
				return fmt.Errorf("other types cannot be changed to special application type")
			}
			if reqConfig, ok := data["config"]; ok {
				klog.V(5).Infof("get special config : %+v", reqConfig)
				if config, ok := reqConfig.(map[string]interface{}); ok {
					/*
					if _, ok = config["special"]; ok {
						target.Spec.Config.Special = source.Config.Special
						klog.V(5).Infof("get special config : %+v", target.Spec.Config.Special)
					}
					 */
					if _, ok := config["operation"]; ok {
						if _, ok = config["special"]; !ok {
							return fmt.Errorf("special application do not exist")
						}
						target.Spec.Config.Operation = source.Config.Operation
						switch target.Spec.Config.Operation {
						case v1.AddSpecApp:
							if target, err = s.addSpecApp(target, source); err != nil {
								return fmt.Errorf("add special application failed: +%v", err)
							}
						case v1.DelSpecApp:
							if target, err = s.delSpecApp(target, source); err != nil {
								return fmt.Errorf("delete special application failed: +%v", err)
							}
						case v1.UpdateSpecApp:
							if target, err = s.updateSpecApp(target, source); err != nil {
								return fmt.Errorf("update special application failed: +%v", err)
							}
						}
					}
				}
			}
		}
	}

	if _, ok = data["apiCount"]; ok {
		target.Status.APICount = source.APICount
	}
	if _, ok = data["description"]; ok {
		if len([]rune(source.Description)) > 255 {
			return fmt.Errorf("%s cannot exceed 255 characters", source.Description)
		}
		target.Spec.Description = source.Description
	}
	if _, ok := data["apis"]; ok {
		target.Spec.Apis = source.Apis
	}

	target.Status.UpdatedAt = metav1.Now()
	return nil
}

func (s *Service) addSpecApp(target *v1.Trafficcontrol, source Trafficcontrol) (*v1.Trafficcontrol, error) {
	if len (source.Config.Special) > 1 {
		return target, fmt.Errorf("only one application can be added at a time")
	}
	for _, spec := range target.Spec.Config.Special {
		if spec.ID == source.Config.Special[0].ID {
			return target, fmt.Errorf("Applications already exist ")
		}
	}
	if (source.Config.Special[0].Year + source.Config.Special[0].Month + source.Config.Special[0].Day +
		source.Config.Special[0].Hour + source.Config.Special[0].Minute + source.Config.Special[0].Second) == 0 {
		return target, fmt.Errorf("at least one limit config must exist for special.")
	} else {
		//list存秒分时日月年，list2存list的index
		var list []int
		list = append(list, source.Config.Special[0].Second)
		list = append(list, source.Config.Special[0].Minute)
		list = append(list, source.Config.Special[0].Hour)
		list = append(list, source.Config.Special[0].Day)
		list = append(list, source.Config.Special[0].Month)
		list = append(list, source.Config.Special[0].Year)
		var list2 []int
		for i, _ := range list {
			if list[i] != 0 {
				list2 = append(list2, i)
			}
		}
		var atime [6]string
		atime[0] = "秒"
		atime[1] = "分"
		atime[2] = "时"
		atime[3] = "日"
		atime[4] = "月"
		atime[5] = "年"
		for i := 0; i < len(list2)-1; i++ {
			if list[list2[i]] > list[list2[i+1]] {
				return target, fmt.Errorf("每%s的次数必须小于每%s的次数", atime[list2[i]], atime[list2[i+1]])
			}
		}
	}

	target.Spec.Config.Special = append(target.Spec.Config.Special, source.Config.Special[0])
	return target, nil
}

func (s *Service) delSpecApp(target *v1.Trafficcontrol, source Trafficcontrol) (*v1.Trafficcontrol, error) {
	if len (source.Config.Special) > 1 {
		return target, fmt.Errorf("only one application can be delete at a time")
	}
	if len(target.Spec.Config.Special) == 1 {
		return target, fmt.Errorf("at least one special application")
	}
	for k, spec := range target.Spec.Config.Special {
		if spec.ID == source.Config.Special[0].ID {
			target.Spec.Config.Special = append(target.Spec.Config.Special[:k], target.Spec.Config.Special[k+1:]...)
			return target, nil
		}
	}
	return target, fmt.Errorf("application not available")
}

func (s *Service) updateSpecApp(target *v1.Trafficcontrol, source Trafficcontrol) (*v1.Trafficcontrol, error) {
	if len (source.Config.Special) > 1 {
		return target, fmt.Errorf("only one application can be update at a time")
	}
	if (source.Config.Special[0].Year + source.Config.Special[0].Month + source.Config.Special[0].Day +
		source.Config.Special[0].Hour + source.Config.Special[0].Minute + source.Config.Special[0].Second) == 0 {
		return target, fmt.Errorf("at least one limit config must exist for special.")
	} else {
		//list存秒分时日月年，list2存list的index
		var list []int
		list = append(list, source.Config.Special[0].Second)
		list = append(list, source.Config.Special[0].Minute)
		list = append(list, source.Config.Special[0].Hour)
		list = append(list, source.Config.Special[0].Day)
		list = append(list, source.Config.Special[0].Month)
		list = append(list, source.Config.Special[0].Year)
		var list2 []int
		for i, _ := range list {
			if list[i] != 0 {
				list2 = append(list2, i)
			}
		}
		var atime [6]string
		atime[0] = "秒"
		atime[1] = "分"
		atime[2] = "时"
		atime[3] = "日"
		atime[4] = "月"
		atime[5] = "年"
		for i := 0; i < len(list2)-1; i++ {
			if list[list2[i]] > list[list2[i+1]] {
				return target, fmt.Errorf("每%s的次数必须小于每%s的次数", atime[list2[i]], atime[list2[i+1]])
			}
		}
	}
	for k, spec := range target.Spec.Config.Special {
		if spec.ID == source.Config.Special[0].ID {
			target.Spec.Config.Special[k] = source.Config.Special[0]
			return target, nil
		}
	}
	return target, fmt.Errorf("application not available")
}