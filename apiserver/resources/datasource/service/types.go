package service

import (
	"fmt"
	"time"

	dw "github.com/chinamobile/nlpt/apiserver/resources/datasource/datawarehouse"
	rdb "github.com/chinamobile/nlpt/apiserver/resources/datasource/rdb"
	"github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

type Datasource struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Namespace string  `json:"namespace"`
	Type      v1.Type `json:"type"`

	RDB *v1.RDB `json:"rdb"`

	DataWarehouse *dw.Database `json:"datawarehouse"`

	Status    v1.Status `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`

	CreateUser v1.User `json:"createdBy"`
	UpdateUser v1.User `json:"updatedBy"`
}

type Tables struct {
	RDBTables           []rdb.Table `json:"rdbTables,omitempty"`
	DataWarehouseTables []dw.Table  `json:"tables,omitempry"`
}

type Fields struct {
	RDBFields           []rdb.Field   `json:"rdbFields,omitempty"`
	DataWarehouseFields []dw.Property `json:"properties,omitempty"`
}

/**
mysql 连接
*/
type Connect struct {
	UserName       string
	Password       string
	Ip             string
	Port           string
	DBName         string
	TableName      string
	QueryCondition map[string]string
	QType          string
}

// only used in creation or update options
func ToAPI(ds *Datasource, dealType string) *v1.Datasource {
	crd := &v1.Datasource{}
	crd.TypeMeta.Kind = "Datasource"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = ds.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.DatasourceSpec{
		Name: ds.Name,
		Type: ds.Type,
	}
	if ds.Type == v1.RDBType {
		crd.Spec.RDB = ds.RDB
	}
	status := ds.Status
	if len(status) == 0 {
		status = v1.Init
	}
	if dealType == "create" {
		crd.Status = v1.DatasourceStatus{
			Status:    status,
			CreatedAt: metav1.Now(),
			UpdatedAt: metav1.Now(),
		}
	} else if dealType == "update" {
		crd.Status = v1.DatasourceStatus{
			Status:    status,
			UpdatedAt: metav1.Now(),
		}
	}
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

		Status:    obj.Status.Status,
		UpdatedAt: obj.Status.UpdatedAt.Time,
		CreatedAt: obj.Status.CreatedAt.Time,
	}
	switch obj.Spec.Type {
	case v1.RDBType:
		if obj.Spec.RDB != nil {
			ds.RDB = obj.Spec.RDB
			ds.RDB.Connect = v1.ConnectInfo{
				Host:     opaque,
				Port:     0,
				Username: opaque,
				Password: opaque,
			}
		} else {
			klog.Errorf("datasource %s in type rdb has no rdb instance", obj.ObjectMeta.Name)
		}
	case v1.DataWarehouseType:
		if obj.Spec.DataWarehouse != nil {
			ds.DataWarehouse = &dw.Database{
				Name: obj.Spec.DataWarehouse.Name,
			}
		} else {
			klog.Errorf("datasource %s in type datawarehouse has no datawarehouse instance", obj.ObjectMeta.Name)
		}
	}
	return ds
}

func ToListModel(items *v1.DatasourceList) []*Datasource {
	var ds []*Datasource = make([]*Datasource, len(items.Items))
	for i := range items.Items {
		ds[i] = ToModel(&items.Items[i])
	}
	return ds
}

func (a *Datasource) Validate() error {
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
		return fmt.Errorf("cannot create datawarehouse datasource")
	default:
		return fmt.Errorf("unknown datasource type: %s", a.Type)
	}

	if !support(a.Type) {
		return fmt.Errorf("type %s not supported", a.Type)
	}
	a.ID = names.NewID()
	return nil
}
func (a *Datasource) ValidateForUpdate() error {
	for k, v := range map[string]string{
		"id":   a.ID,
		"name": a.Name,
		"type": a.Type.String(),
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	if a.Type == v1.RDBType {
		if err := a.RDB.Validate(); err != nil {
			return err
		}
	}

	if !support(a.Type) {
		return fmt.Errorf("type %s not supported", a.Type)
	}
	return nil
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
