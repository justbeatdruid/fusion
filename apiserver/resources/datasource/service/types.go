package service

import (
	"fmt"
	"strings"

	dw "github.com/chinamobile/nlpt/apiserver/resources/datasource/datawarehouse"
	rdb "github.com/chinamobile/nlpt/apiserver/resources/datasource/rdb"
	"github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/pkg/auth/user"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

type Datasource struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Namespace string  `json:"namespace"`
	Type      v1.Type `json:"type"`

	Users user.Users `json:"users"`

	RDB *v1.RDB `json:"rdb,omitempty"`

	DataWarehouse *dw.Database `json:"datawarehouse,omitempty"`

	MessageQueue *v1.MessageQueue `json:"mq,omitempty"`

	Mongo *v1.Mongo `json:"mongo,omitempty"`

	Status    string    `json:"status"`
	UpdatedAt util.Time `json:"updatedAt"`
	CreatedAt util.Time `json:"createdAt"`
}

type Tables struct {
	RDBTables           []rdb.Table `json:"rdbTables,omitempty"`
	DataWarehouseTables []dw.Table  `json:"tables,omitempry"`
}

type Table struct {
	RDBTable           *rdb.Table    `json:"rdbTable,omitempty"`
	DataWarehouseTable *dw.TableInfo `json:"table,omitempry"`
}

type Fields struct {
	RDBFields           []rdb.Field   `json:"rdbFields,omitempty"`
	DataWarehouseFields []dw.Property `json:"properties,omitempty"`
}

type Field struct {
	DataWarehouseField *dw.Property `json:"property,omitempty"`
	RDBField           *rdb.Field   `json:"rdbField,omitempty"`
}

type Statistics struct {
	Total      int    `json:"total"`
	Increment  int    `json:"increment"`
	Percentage string `json:"percentage"`
}

// only used in creation or update options
func ToAPI(ds *Datasource, specOnly bool) *v1.Datasource {
	crd := &v1.Datasource{}
	crd.TypeMeta.Kind = "Datasource"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = ds.ID
	crd.ObjectMeta.Namespace = ds.Namespace
	crd.Spec = v1.DatasourceSpec{
		Name: ds.Name,
		Type: ds.Type,
	}
	if ds.Type == v1.RDBType {
		crd.Spec.RDB = ds.RDB
	} else if ds.Type == v1.TopicType {
		crd.Spec.MessageQueue = ds.MessageQueue
	} else if ds.Type == v1.DataWarehouseType {
		crd.ObjectMeta.Annotations = make(map[string]string)
		crd.ObjectMeta.Annotations["name"] = ds.DataWarehouse.Name
		crd.ObjectMeta.Annotations["subject"] = ds.DataWarehouse.SubjectName
	} else if ds.Type == v1.MongoType {
		crd.Spec.Mongo = ds.Mongo
	}
	status := ds.Status
	if len(status) == 0 {
		status = ""
	}
	if !specOnly {
		crd.Status = v1.DatasourceStatus{
			Status:    v1.FromString(status),
			CreatedAt: metav1.Now(),
			UpdatedAt: metav1.Now(),
		}
	}
	crd.ObjectMeta.Labels = user.AddUsersLabels(ds.Users, crd.ObjectMeta.Labels)
	crd.ObjectMeta.Labels[v1.TypeLabel] = string(crd.Spec.Type)
	return crd
}

const opaque = "opaque"

func ToModel(obj *v1.Datasource) *Datasource {
	ds := &Datasource{
		ID:        obj.ObjectMeta.Name,
		Namespace: obj.ObjectMeta.Namespace,
		Name:      obj.Spec.Name,
		Type:      obj.Spec.Type,

		//RDB: obj.Spec.RDB,

		Status:    v1.ToString(obj.Status.Status),
		UpdatedAt: util.NewTime(obj.Status.UpdatedAt.Time),
		CreatedAt: util.NewTime(obj.Status.CreatedAt.Time),
	}
	switch obj.Spec.Type {
	case v1.RDBType:
		if obj.Spec.RDB != nil {
			ds.RDB = obj.Spec.RDB
			/*
				ds.RDB.Connect = v1.ConnectInfo{
					Host:     opaque,
					Port:     0,
					Username: opaque,
					Password: opaque,
				}
			*/
		} else {
			klog.Errorf("datasource %s in type rdb has no rdb instance", obj.ObjectMeta.Name)
		}
	case v1.DataWarehouseType:
		if obj.Spec.DataWarehouse != nil {
			ds.DataWarehouse = &dw.Database{
				Name:               obj.Spec.DataWarehouse.Name,
				DisplayName:        obj.Spec.DataWarehouse.DisplayName,
				SubjectName:        obj.Spec.DataWarehouse.SubjectName,
				SubjectDisplayName: obj.Spec.DataWarehouse.SubjectDisplayName,
			}
		} else {
			klog.Errorf("datasource %s in type datawarehouse has no datawarehouse instance", obj.ObjectMeta.Name)
		}
	case v1.TopicType:
		ds.MessageQueue = obj.Spec.MessageQueue
	case v1.MongoType:
		ds.Mongo = obj.Spec.Mongo
	}
	ds.Users = user.GetUsersFromLabels(obj.ObjectMeta.Labels)
	return ds
}

func ToListModel(items *v1.DatasourceList, opts ...util.OpOption) []*Datasource {
	if len(opts) > 0 {
		nameLike := util.OpList(opts...).NameLike()
		if len(nameLike) > 0 {
			var dss []*Datasource = make([]*Datasource, 0)
			for _, item := range items.Items {
				if !strings.Contains(item.Spec.Name, nameLike) {
					continue
				}
				ds := ToModel(&item)
				dss = append(dss, ds)
			}
			return dss
		}
	}
	var ds []*Datasource = make([]*Datasource, len(items.Items))
	for i := range items.Items {
		ds[i] = ToModel(&items.Items[i])
	}
	return ds
}

func (s *Service) Validate(a *Datasource) error {
	for k, v := range map[string]string{
		"name": a.Name,
		"type": a.Type.String(),
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	switch a.Type {
	case v1.RDBType:
		if err := a.RDB.Validate(); err != nil {
			return err
		}
	case v1.DataWarehouseType:
		if !s.tenantEnabled {
			return fmt.Errorf("cannot create or update datawarehouse datasource")
		} else {
			if a.DataWarehouse == nil {
				return fmt.Errorf("cannot find datawarehouse in request body")
			}
			if a.DataWarehouse.Name == "" || a.DataWarehouse.SubjectName == "" {
				return fmt.Errorf("name or subject name is null")
			}
		}
	case v1.TopicType:
		if err := s.CheckTopic(a.Namespace, a.MessageQueue); err != nil {
			return fmt.Errorf("topic validate error: %+v", err)
		}
	case v1.MongoType:
		if err := s.CheckMongo(a.Mongo); err != nil {
			return fmt.Errorf("mongo validate error: %+v", err)
		}
	default:
		return fmt.Errorf("unknown datasource type: %s", a.Type)
	}

	if !support(a.Type) {
		return fmt.Errorf("type %s not supported", a.Type)
	}
	if len(a.ID) == 0 {
		a.ID = names.NewID()
	}
	return nil
}

func (a *Datasource) ValidateConnection() error {
	switch a.Type {
	case v1.RDBType:
		return a.RDB.Validate()
	default:
		return nil
	}
}

func support(tp v1.Type) bool {
	for _, t := range Supported {
		if t == "*" {
			return true
		}
		if tp.String() == t {
			return true
		}
	}
	return false
}

type FieldsTuple struct {
	SourceFieldName string `json:"sourceFieldName"`
	SourceFieldType string `json:"sourceFieldType"`
	TargetFieldName string `json:"targetFieldName"`
	TargetFieldType string `json:"targetFieldType"`
}
