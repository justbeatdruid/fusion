package service

import (
	"fmt"
	"strings"

	dw "github.com/chinamobile/nlpt/apiserver/resources/datasource/datawarehouse"
	"github.com/chinamobile/nlpt/apiserver/resources/datasource/rdb"
	"github.com/chinamobile/nlpt/apiserver/resources/datasource/rdb/driver"
	"github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/pkg/util"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

var defaultNamespace = "default"

var Supported = []string{}

type Service struct {
	client dynamic.NamespaceableResourceInterface
	//apiService *api.Service
	//suService *serviceunit.Service
}

func NewService(client dynamic.Interface, supported []string) *Service {
	Supported = supported
	return &Service{
		client: client.Resource(v1.GetOOFSGVR()),
		//apiService: api.NewService(client),
		//suService: serviceunit.NewService(client),
	}
}

func (s *Service) CreateDatasource(model *Datasource) (*Datasource, error) {
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	ds, err := s.Create(ToAPI(model, false))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(ds), nil
}
func (s *Service) UpdateDatasource(model *Datasource, opts ...util.OpOption) (*Datasource, error) {
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	ds, err := s.Get(model.ID, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	api := ToAPI(model, true)
	ds.Spec = api.Spec
	ds.Status.UpdatedAt = metav1.Now()
	ds, err = s.Update(ds)
	if err != nil {
		return nil, fmt.Errorf("cannot update object: %+v", err)
	}
	return ToModel(ds), nil
}
func (s *Service) ListDatasource(opts ...util.OpOption) ([]*Datasource, error) {
	dss, err := s.List(opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(dss, opts...), nil
}

func (s *Service) GetDatasource(id string, opts ...util.OpOption) (*Datasource, error) {
	ds, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(ds), nil
}

func (s *Service) DeleteDatasource(id string, opts ...util.OpOption) error {
	ds, err := s.Get(id, opts...)
	if err != nil {
		return fmt.Errorf("get datasource error: %+v", err)
	}
	if ds.Spec.Type == v1.DataWarehouseType {
		return fmt.Errorf("cannot delete datawarehouse datasource")
	}
	return s.Delete(id, opts...)
}

func (s *Service) Create(ds *v1.Datasource) (*v1.Datasource, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ds)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)
	crd, err = s.client.Namespace(ds.ObjectMeta.Namespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ds); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasource of creating: %+v", ds)
	return ds, nil
}

func (s *Service) Update(ds *v1.Datasource, opts ...util.OpOption) (*v1.Datasource, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ds)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(ds.ObjectMeta.Namespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error updateing crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ds); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasource of creating: %+v", ds)
	return ds, nil
}

func (s *Service) List(opts ...util.OpOption) (*v1.DatasourceList, error) {
	ops := util.OpList(opts...)
	crdNamespace := ops.Namespace()
	typpe := ops.Type()
	listOptions := metav1.ListOptions{}
	if len(typpe) > 0 {
		listOptions = metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", v1.TypeLabel, typpe),
		}
	}
	crd, err := s.client.Namespace(crdNamespace).List(listOptions)
	if err != nil {
		return nil, fmt.Errorf("error list crd: %+v", err)
	}
	dss := &v1.DatasourceList{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), dss); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasourceList: %+v", dss)
	return dss, nil
}

func (s *Service) Get(id string, opts ...util.OpOption) (*v1.Datasource, error) {
	crdNamespace := util.OpList(opts...).Namespace()
	if len(crdNamespace) == 0 {
		crdNamespace = "default"
	}
	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("error get crd: %+v", err)
	}
	ds := &v1.Datasource{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ds); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasource: %+v", ds)
	return ds, nil
}

func (s *Service) Delete(id string, opts ...util.OpOption) error {
	crdNamespace := util.OpList(opts...).Namespace()
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) GetTables(id, associationID string, opts ...util.OpOption) (*Tables, error) {
	result := &Tables{}
	ds, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot datasource: %+v", err)
	}
	switch ds.Spec.Type {
	case v1.RDBType:
		if ds.Spec.RDB == nil {
			return nil, fmt.Errorf("datasource %s in type rdb has no rdb instance", ds.ObjectMeta.Name)
		}
		if len(ds.Spec.RDB.Type) == 0 {
			return nil, fmt.Errorf("datasource %s in type rdb has no databasetype", ds.ObjectMeta.Name)
		}
		if ds.Spec.RDB.Type == "mysql" {
			//mysql类型
			querySql := "SELECT distinct table_name FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = '" + ds.Spec.RDB.Database + "'"
			mysqlData, err := driver.GetRDBData(ds, querySql)
			if err != nil {
				return nil, err
			}
			for _, v := range mysqlData {
				table := rdb.Table{}
				table.Name = v["table_name"]
				result.RDBTables = append(result.RDBTables, table)
			}
			//fmt.Println(result)
			return result, nil
		} else if ds.Spec.RDB.Type == "postgres" || ds.Spec.RDB.Type == "postgresql" {
			if ds.Spec.RDB.Schema == "" {
				ds.Spec.RDB.Schema = "public"
			}
			querySql := "select * from pg_tables where schemaname = '" + ds.Spec.RDB.Schema + "'"
			mysqlData, err := driver.GetRDBData(ds, querySql)
			if err != nil {
				return nil, err
			}
			for _, v := range mysqlData {
				table := rdb.Table{}
				table.Name = v["tablename"]
				result.RDBTables = append(result.RDBTables, table)
			}
			//fmt.Println(result)
			return result, nil
		} else {
			return nil, fmt.Errorf("unsupported rdb type %s", ds.Spec.RDB.Type)
		}
	case v1.DataWarehouseType:
		if ds.Spec.DataWarehouse == nil {
			return nil, fmt.Errorf("datasource %s in type datawarehouse has no datawarehouse instance", ds.ObjectMeta.Name)
		}
		ts := ds.Spec.DataWarehouse.GetTables(associationID)
		klog.V(5).Infof("get tables: %+v", ts)
		result.DataWarehouseTables = make([]dw.Table, 0)
		for _, t := range ts {
			result.DataWarehouseTables = append(result.DataWarehouseTables, dw.FromApiTable(t))
		}
	default:
		return nil, fmt.Errorf("wrong datasource type: %s", ds.Spec.Type)
	}
	return result, nil
}

func (s *Service) GetTable(id, table string, opts ...util.OpOption) (*Table, error) {
	ds, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot datasource: %+v", err)
	}
	switch ds.Spec.Type {
	case v1.RDBType:
		return nil, fmt.Errorf("not suppurted for rdb")
	case v1.DataWarehouseType:
		if ds.Spec.DataWarehouse == nil {
			return nil, fmt.Errorf("datasource %s in type datawarehouse has no datawarehouse instance", ds.ObjectMeta.Name)
		}
		for _, t := range ds.Spec.DataWarehouse.Tables {
			if t.Info.ID == table {
				info := dw.FromApiTable(t).Info
				return &Table{DataWarehouseTable: &info}, nil
			}
		}
		return nil, fmt.Errorf("table %s not found in database %s", table, ds.Spec.Name)
	default:
		return nil, fmt.Errorf("wrong datasource type: %s", ds.Spec.Type)
	}
}

func (s *Service) GetFields(id, table string, opts ...util.OpOption) (*Fields, error) {
	result := &Fields{}
	ds, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot datasource: %+v", err)
	}
	switch ds.Spec.Type {
	case v1.RDBType:
		if ds.Spec.RDB == nil {
			return nil, fmt.Errorf("datasource %s in type rdb has no rdb instance", ds.ObjectMeta.Name)
		}
		//mysql类型
		if ds.Spec.RDB.Type == "mysql" {
			//查询表结构sql
			querySql := "select COLUMN_NAME '字段名称',COLUMN_TYPE '字段类型长度',IF(EXTRA='auto_increment',CONCAT(COLUMN_KEY," +
				"'(', IF(EXTRA='auto_increment','自增长',EXTRA),')'),COLUMN_KEY) '主外键',IS_NULLABLE '空标识',COLUMN_COMMENT " +
				"'字段说明' from information_schema.columns where table_name='" + table + "'" +
				" and table_schema='" + ds.Spec.RDB.Database + "'"
			mysqlData, err := driver.GetRDBData(ds, querySql)
			if err != nil {
				return nil, err
			}
			for _, v := range mysqlData {
				table := rdb.Field{}
				table.Name = v["字段名称"]
				table.DataType = v["字段类型长度"]
				table.Description = v["字段说明"]
				if len(v["主外键"]) == 0 {
					table.PrimaryKey = false
				} else {
					table.PrimaryKey = true
				}
				table.IsNullAble = strings.ToUpper(v["空标识"]) == "YES"
				result.RDBFields = append(result.RDBFields, table)
			}
			//查询索引sql
			querySql = "show keys  from " + table
			data, err := driver.GetRDBData(ds, querySql)
			if err != nil {
				return nil, err
			}
			for _, v := range data {
				for i := 0; i < len(result.RDBFields); i++ {
					//表索引字段==表字段 则索引值为true
					if v["Column_name"] == result.RDBFields[i].Name {
						result.RDBFields[i].Unique = true
					} else {
						result.RDBFields[i].Unique = false
					}
				}
			}
			return result, nil
		} else if ds.Spec.RDB.Type == "postgres" || ds.Spec.RDB.Type == "postgresql" {
			if ds.Spec.RDB.Schema == "" {
				ds.Spec.RDB.Schema = "public"
			}
			querySql := `SELECT
    "pg_attribute".attname                                                    as "Column",
    pg_catalog.format_type("pg_attribute".atttypid, "pg_attribute".atttypmod) as "Datatype",
    not("pg_attribute".attnotnull) AS "Nullable"
FROM
    pg_catalog.pg_attribute "pg_attribute"
WHERE
    "pg_attribute".attnum > 0
    AND NOT "pg_attribute".attisdropped
    AND "pg_attribute".attrelid = (
        SELECT "pg_class".oid
        FROM pg_catalog.pg_class "pg_class"
            LEFT JOIN pg_catalog.pg_namespace "pg_namespace" ON "pg_namespace".oid = "pg_class".relnamespace
        WHERE
            "pg_namespace".nspname = '%s'
            AND "pg_class".relname = '%s'
    );`
			querySql = fmt.Sprintf(querySql, ds.Spec.RDB.Schema, table)
			data, err := driver.GetRDBData(ds, querySql)
			if err != nil {
				return nil, err
			}
			for _, v := range data {
				table := rdb.Field{}
				table.Name = v["Column"]
				table.DataType = v["Datatype"]
				table.IsNullAble = strings.ToLower(v["Nullable"]) == "true"
				result.RDBFields = append(result.RDBFields, table)
			}
			return result, nil
		} else {
			return nil, fmt.Errorf("unsupported rdb type %s", ds.Spec.RDB.Type)
		}
	case v1.DataWarehouseType:
		if ds.Spec.DataWarehouse == nil {
			return nil, fmt.Errorf("datasource %s in type datawarehouse has no datawarehouse instance", ds.ObjectMeta.Name)
		}
		result.DataWarehouseFields = make([]dw.Property, 0)
		for _, apiTable := range ds.Spec.DataWarehouse.Tables {
			if apiTable.Info.ID == table {
				for _, p := range apiTable.Properties {
					result.DataWarehouseFields = append(result.DataWarehouseFields, dw.FromApiProperty(p))
				}
				goto rt
			}
		}
		return nil, fmt.Errorf("cannot find table %s in datasource %s", table, ds.ObjectMeta.Name)
	default:
		return nil, fmt.Errorf("wrong datasource type: %s", ds.Spec.Type)
	}
rt:
	return result, nil
}

func (s *Service) GetField(id, table, field string, opts ...util.OpOption) (*Field, error) {
	ds, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot datasource: %+v", err)
	}
	switch ds.Spec.Type {
	case v1.RDBType:
		return nil, fmt.Errorf("rdb not supported")
	case v1.DataWarehouseType:
		if ds.Spec.DataWarehouse == nil {
			return nil, fmt.Errorf("datasource %s in type datawarehouse has no datawarehouse instance", ds.ObjectMeta.Name)
		}
		for _, apiTable := range ds.Spec.DataWarehouse.Tables {
			if apiTable.Info.ID == table {
				for _, p := range apiTable.Properties {
					if p.ID == field {
						dwp := dw.FromApiProperty(p)
						return &Field{DataWarehouseField: &dwp}, nil
					}
				}
				return nil, fmt.Errorf("cannot find field %s in datasource %s and table %s", field, ds.ObjectMeta.Name, apiTable.Info.Name)
			}
		}
		return nil, fmt.Errorf("cannot find table %s in datasource %s", table, ds.ObjectMeta.Name)
	default:
		return nil, fmt.Errorf("wrong datasource type: %s", ds.Spec.Type)
	}
}

func (s *Service) Ping(ds *Datasource) error {
	if err := ds.ValidateConnection(); err != nil {
		return fmt.Errorf("database connection validate error: %+v", err)
	}
	if ds == nil {
		return fmt.Errorf("datasource is null")
	}
	switch ds.Type {
	case v1.RDBType:
		return driver.PingRDB(ToAPI(ds, true))
	default:
		return fmt.Errorf("not supported for %s", ds.Type)
	}
}

/*
func (s *Service) GetDataSourceByApiId(apiId string, parames string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	//get api by apiID
	api, err := s.apiService.Get(apiId)
	if err != nil {
		return nil, fmt.Errorf("error query api: %+v", err)
	}
	//get serviceunitId by api
	serverUnitID := api.Spec.Serviceunit.ID
	//get serviceunit by serviceunitId
	serverUnit, err := s.suService.Get(serverUnitID)
	if err != nil {
		return nil, fmt.Errorf("error query serverUnit: %+v", err)
	}
	//check unit type (single or multi)
	//get dataSources by  multiDateSourceId
	for _, v := range serverUnit.Spec.DatasourceID {
		datasource, err := s.Get(v.ID)
		if err != nil {
			return nil, fmt.Errorf("error query singleDateSourceId: %+v", err)
		}
		//TODO The remaining operation after the query to the data source
		result["Fields"] = datasource.Spec.Fields
	}
	return result, nil
}
*/
