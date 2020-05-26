package service

import (
	"encoding/json"
	"fmt"
	"strings"

	datav1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/errors"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Serviceunit struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Namespace    string             `json:"namespace"`
	Type         v1.ServiceType     `json:"type"`
	DatasourceID *v1.Datasource     `json:"datasources,omitempty"`
	KongSevice   v1.KongServiceInfo `json:"kongService"`
	Users        user.Users         `json:"users"`
	Description  string             `json:"description"`

	Status    v1.Status `json:"status"`
	UpdatedAt util.Time `json:"time"`
	APICount  int       `json:"apiCount"`
	Published bool      `json:"published"`
	CreatedAt util.Time `json:"createdAt"`

	Writable bool `json:"writable"`

	Group     string `json:"group"`
	GroupName string `json:"groupName"`
	//sunyu+
	Result        v1.Result    `json:"result"`
	DisplayStatus v1.DisStatus `json:"disStatus"`
}

// only used in creation options
func ToAPI(app *Serviceunit) *v1.Serviceunit {
	crd := &v1.Serviceunit{}
	crd.TypeMeta.Kind = "Serviceunit"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = app.ID
	crd.ObjectMeta.Namespace = app.Namespace
	crd.Spec = v1.ServiceunitSpec{
		Name:         app.Name,
		Type:         app.Type,
		DatasourceID: app.DatasourceID,
		//Datasource:   app.Datasource,
		KongService:   app.KongSevice,
		Description:   app.Description,
		Result:        app.Result,
		DisplayStatus: app.DisplayStatus,
	}
	status := app.Status
	if len(status) == 0 {
		status = v1.Init
	}
	if crd.Spec.APIs == nil {
		crd.Spec.APIs = make([]v1.Api, 0)
	}
	crd.Status = v1.ServiceunitStatus{
		Status:    status,
		UpdatedAt: metav1.Now(),
		APICount:  0,
		Published: false,
	}
	if len(app.Group) > 0 {
		if crd.ObjectMeta.Labels == nil {
			crd.ObjectMeta.Labels = make(map[string]string)
		}
		crd.ObjectMeta.Labels[v1.GroupLabel] = app.Group
	}
	// add user labels
	crd.ObjectMeta.Labels = user.AddUsersLabels(app.Users, crd.ObjectMeta.Labels)
	return crd
}

// +update_sunyu
func ToAPIUpdate(su *Serviceunit, crd *v1.Serviceunit) *v1.Serviceunit {
	id := crd.Spec.KongService.ID
	crd.Spec = v1.ServiceunitSpec{
		Name:         su.Name,
		Type:         su.Type,
		DatasourceID: su.DatasourceID,
		//Datasource:   su.Datasource,
		KongService:   su.KongSevice,
		Description:   su.Description,
		Result:        su.Result,
		DisplayStatus: su.DisplayStatus,
	}
	crd.Spec.KongService.ID = id
	status := su.Status
	if len(status) == 0 {
		status = v1.Update
	}
	if crd.Spec.APIs == nil {
		crd.Spec.APIs = make([]v1.Api, 0)
	}
	crd.Status = v1.ServiceunitStatus{
		Status:    status,
		UpdatedAt: metav1.Now(),
		APICount:  0,
		Published: false,
	}
	return crd
}

func ToModel(obj *v1.Serviceunit, opts ...util.OpOption) *Serviceunit {
	switch obj.Spec.Result {
	case v1.CREATING:
		(*obj).Spec.DisplayStatus = v1.SuCreating
	case v1.CREATESUCCESS:
		(*obj).Spec.DisplayStatus = v1.CreateSuccess
	case v1.CREATEFAILED:
		(*obj).Spec.DisplayStatus = v1.CreateFailed
	case v1.UPDATING:
		(*obj).Spec.DisplayStatus = v1.SuUpdating
	case v1.UPDATESUCCESS:
		(*obj).Spec.DisplayStatus = v1.UpdateSuccess
	case v1.UPDATEFAILED:
		(*obj).Spec.DisplayStatus = v1.UpdateFailed
	case v1.DELETEFAILED:
		(*obj).Spec.DisplayStatus = v1.DeleteFailed
	}
	su := &Serviceunit{
		ID:           obj.ObjectMeta.Name,
		Name:         obj.Spec.Name,
		Namespace:    obj.ObjectMeta.Namespace,
		Type:         obj.Spec.Type,
		DatasourceID: obj.Spec.DatasourceID,
		KongSevice:   obj.Spec.KongService,
		Description:  obj.Spec.Description,

		Status:        obj.Status.Status,
		CreatedAt:     util.NewTime(obj.ObjectMeta.CreationTimestamp.Time),
		UpdatedAt:     util.NewTime(obj.Status.UpdatedAt.Time),
		APICount:      obj.Status.APICount,
		Published:     obj.Status.Published,
		Result:        obj.Spec.Result,
		DisplayStatus: obj.Spec.DisplayStatus,
	}
	u := util.OpList(opts...).User()
	if len(u) > 0 {
		su.Writable = user.WritePermitted(u, obj.ObjectMeta.Labels)
	}

	su.Users = user.GetUsersFromLabels(obj.ObjectMeta.Labels)
	//TODO UserCount
	if group, ok := obj.ObjectMeta.Labels[v1.GroupLabel]; ok {
		su.Group = group
		su.GroupName = obj.Spec.Group.Name
	}
	return su
}

func ToListModel(items *v1.ServiceunitList, groups map[string]string, datas map[string]*v1.Datasource, opts ...util.OpOption) []*Serviceunit {
	if len(opts) > 0 {
		nameLike := util.OpList(opts...).NameLike()
		stype := util.OpList(opts...).Stype()
		/*
			if len(nameLike) > 0 {
				var sus []*Serviceunit = make([]*Serviceunit, 0)
				for _, item := range items.Items {
					if !strings.Contains(item.Spec.Name, nameLike) {
						continue
					}
					if gid, ok := item.ObjectMeta.Labels[v1.GroupLabel]; ok {
						item.Spec.Group.ID = gid
					}
					if gname, ok := groups[item.Spec.Group.ID]; ok {
						item.Spec.Group.Name = gname
					}
					if item.Spec.Type == v1.DataService && item.Spec.DatasourceID != nil {
						if data, ok := datas[item.Spec.DatasourceID.ID]; ok {
							item.Spec.DatasourceID = data
						}
					}
					su := ToModel(&item, opts...)
					sus = append(sus, su)
				}
				return sus
			}

		*/
		var sus []*Serviceunit = make([]*Serviceunit, 0)
		for _, item := range items.Items {
			if len(nameLike) > 0 {
				if !strings.Contains(item.Spec.Name, nameLike) {
					continue
				}
				if gid, ok := item.ObjectMeta.Labels[v1.GroupLabel]; ok {
					item.Spec.Group.ID = gid
				}
				if gname, ok := groups[item.Spec.Group.ID]; ok {
					item.Spec.Group.Name = gname
				}
				if item.Spec.Type == v1.DataService && item.Spec.DatasourceID != nil {
					if data, ok := datas[item.Spec.DatasourceID.ID]; ok {
						item.Spec.DatasourceID = data
					}
				}
			}
			if len(stype) > 0 {
				if string(item.Spec.Type) != stype {
					continue
				}
			}
			su := ToModel(&item, opts...)
			sus = append(sus, su)
		}
		return sus
	}
	var sus []*Serviceunit = make([]*Serviceunit, len(items.Items))
	for i, item := range items.Items {
		if gid, ok := item.ObjectMeta.Labels[v1.GroupLabel]; ok {
			item.Spec.Group.ID = gid
		}
		if gname, ok := groups[item.Spec.Group.ID]; ok {
			item.Spec.Group.Name = gname
		}
		if item.Spec.Type == v1.DataService && item.Spec.DatasourceID != nil {
			if data, ok := datas[item.Spec.DatasourceID.ID]; ok {
				item.Spec.DatasourceID = data
			}
		}
		sus[i] = ToModel(&item, opts...)
	}
	return sus
}

// check create parameters
func (s *Service) Validate(a *Serviceunit) error {
	for k, v := range map[string]string{
		"name": a.Name,
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	suList, errs := s.List(util.WithNamespace(a.Namespace))
	if errs != nil {
		return fmt.Errorf("cannot list serviceunit object: %+v", errs)
	}
	for _, p := range suList.Items {
		if p.Spec.Name == a.Name {
			return errors.NameDuplicatedError("serviceunit name duplicated: %+v", errs)
		}
	}

	if len(a.Users.Owner.ID) == 0 {
		return fmt.Errorf("owner not set")
	}

	switch a.Type {
	case v1.DataService:
		if len(a.DatasourceID.ID) == 0 {
			return fmt.Errorf("datasource is null")
		} else {
			if _, err := s.checkDatasource(a.Namespace, a.DatasourceID); err != nil {
				return fmt.Errorf("error datasource: %+v", err)
			} else {
				//a.Datasource = *ds
			}
		}
	case v1.WebService:
		if len(a.KongSevice.Host) == 0 || len(a.KongSevice.Protocol) == 0 {
			return fmt.Errorf("webservice is null")
		}
	default:
		return fmt.Errorf("wrong type: %s", a.Type)
	}

	a.ID = names.NewID()
	return nil
}

func (s *Service) checkDatasource(namespace string, d *v1.Datasource) (*datav1.DatasourceSpec, error) {
	if d == nil {
		return nil, fmt.Errorf("datasource is null")
	}
	data, err := s.getDatasource(namespace, d.ID)
	if err != nil {
		return nil, fmt.Errorf("cannot get datasource: %+v", err)
	}
	ds := &data.Spec

	for k, v := range map[string]string{
		"name": ds.Name,
		"type": string(ds.Type),
	} {
		if len(v) == 0 {
			return nil, fmt.Errorf("%s is null", k)
		}
	}

	if ds.Type == datav1.RDBType {
		if ds.RDB == nil {
			return nil, fmt.Errorf("cannot find rdb info")
		}
		if ds.RDB.Connect.Port < 1 || ds.RDB.Connect.Port > 65535 {
			return nil, fmt.Errorf("invalid port: %d", ds.RDB.Connect.Port)
		}

		// TODO move to api
		/*
			for i, field := range ds.Fields {
				if err := field.Validate(); err != nil {
					return nil, fmt.Errorf("%dth field invalide: %+v", i, err)
				}
			}
		*/
	}
	return ds, nil
}

func (s *Service) assignment(target *v1.Serviceunit, reqData interface{}) error {
	data, ok := reqData.(map[string]interface{})
	if !ok {
		return fmt.Errorf("reqData type is error,req data: %v", reqData)
	}
	b, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("json.Marshal error,: %v", err)
	}
	var source Serviceunit
	if err = json.Unmarshal(b, &source); err != nil {
		return fmt.Errorf("json.Unmarshal error,: %v", err)
	}
	if _, ok := data["name"]; ok {
		if target.Spec.Name != source.Name {
			suList, errs := s.List(util.WithNamespace(target.ObjectMeta.Namespace))
			if errs != nil {
				return fmt.Errorf("cannot list servieunit object: %+v", errs)
			}
			for _, p := range suList.Items {
				if p.Spec.Name == source.Name {
					return errors.NameDuplicatedError("serviceunit name duplicated: %+v", errs)
				}
			}
		}
		target.Spec.Name = source.Name
	}
	if _, ok := data["description"]; ok {
		target.Spec.Description = source.Description
	}
	if _, ok := data["group"]; ok {
		if target.ObjectMeta.Labels == nil {
			target.ObjectMeta.Labels = make(map[string]string)
		}
		target.ObjectMeta.Labels[v1.GroupLabel] = source.Group
	}
	if _, ok := data["type"]; ok {
		if target.Spec.Type != source.Type {
			return fmt.Errorf("type err")
		}
	}
	if _, ok := data["kongService"]; ok {
		target.Spec.KongService.Host = source.KongSevice.Host
		target.Spec.KongService.Port = source.KongSevice.Port
		target.Spec.KongService.Protocol = source.KongSevice.Protocol
	}
	if _, ok := data["datasources"]; ok {
		target.Spec.DatasourceID.ID = source.DatasourceID.ID
	}
	if target.Spec.APIs == nil {
		target.Spec.APIs = make([]v1.Api, 0)
	}
	return nil
}
