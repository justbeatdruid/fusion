package service

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/chinamobile/nlpt/apiserver/database"
	dwtype "github.com/chinamobile/nlpt/apiserver/resources/datasource/datawarehouse"
	mongodriver "github.com/chinamobile/nlpt/apiserver/resources/datasource/mongo/driver"
	"github.com/chinamobile/nlpt/apiserver/resources/datasource/rdb"
	"github.com/chinamobile/nlpt/apiserver/resources/datasource/rdb/driver"
	"github.com/chinamobile/nlpt/crds/datasource/api/v1"
	dwv1 "github.com/chinamobile/nlpt/crds/datasource/datawarehouse/api/v1"
	tgv1 "github.com/chinamobile/nlpt/crds/topicgroup/api/v1"
	dw "github.com/chinamobile/nlpt/pkg/datawarehouse"
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
	client   dynamic.NamespaceableResourceInterface
	tgClient dynamic.NamespaceableResourceInterface

	tenantEnabled bool
	dataService   dw.Connector

	db *database.DatabaseConnection
}

func NewService(client dynamic.Interface, supported []string, dsConnector dw.Connector, tenantEnabled bool, db *database.DatabaseConnection) *Service {
	Supported = supported
	return &Service{
		client:   client.Resource(v1.GetOOFSGVR()),
		tgClient: client.Resource(tgv1.GetOOFSGVR()),

		tenantEnabled: tenantEnabled,
		dataService:   dsConnector,

		db: db,
	}
}

func (s *Service) CreateDatasource(model *Datasource) (*Datasource, error) {
	if err := s.Validate(model); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	var ds *v1.Datasource
	var err error
	var create func(*v1.Datasource) (*v1.Datasource, error)
	switch model.Type {
	case v1.DataWarehouseType:
		create = s.CreateDatawarehouse
	default:
		create = s.Create
	}
	ds, err = create(ToAPI(model, false))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(ds), nil
}
func (s *Service) UpdateDatasource(model *Datasource, opts ...util.OpOption) (*Datasource, error) {
	if err := s.Validate(model); err != nil {
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
	if s.tenantEnabled {
		return s.ListDatasourceFromDatabase(opts...)
	}
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
	if !s.tenantEnabled {
		if ds.Spec.Type == v1.DataWarehouseType {
			return fmt.Errorf("cannot delete datawarehouse datasource")
		}
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

func (s *Service) CreateDatawarehouse(ds *v1.Datasource) (*v1.Datasource, error) {
	var err error
	ds.Spec.DataWarehouse, err = s.GetDataWareshouse(ds.ObjectMeta.Annotations)
	if err != nil {
		return nil, fmt.Errorf("cannot get datawarehouse: %+v", err)
	}
	if ds.Spec.DataWarehouse.Tables == nil {
		ds.Spec.DataWarehouse.Tables = make([]dwv1.Table, 0)
	}
	for i := range ds.Spec.DataWarehouse.Tables {
		if ds.Spec.DataWarehouse.Tables[i].Properties == nil {
			ds.Spec.DataWarehouse.Tables[i].Properties = make([]dwv1.Property, 0)
		}
	}
	ds.Status.Status = v1.Normal
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

func (s *Service) GetDataWareshouse(annotations map[string]string) (*dwv1.Database, error) {
	if annotations == nil {
		return nil, fmt.Errorf("annotation not set")
	}
	var name, subject string
	var ok bool
	name, ok = annotations["name"]
	if !ok {
		return nil, fmt.Errorf("name not set in annotation")
	}
	subject, ok = annotations["subject"]
	if !ok {
		return nil, fmt.Errorf("subject not set in annotation")
	}
	datawarehouse, err := s.dataService.GetExampleDatawarehouse() //查询新的数据
	if err != nil {
		return nil, fmt.Errorf("cannot get datawarehouse: %+v", err)
	}
	for _, d := range datawarehouse.Databases {
		db := dwv1.FromApiDatabase(d)
		if db.Name == name && db.SubjectName == subject {
			return &db, nil
		}
	}
	return nil, fmt.Errorf("name [%s] and subject [%s] not found", name, subject)
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
			querySql := "SELECT * FROM pg_tables WHERE schemaname = '" + ds.Spec.RDB.Schema + "'"
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
		result.DataWarehouseTables = make([]dwtype.Table, 0)
		for _, t := range ts {
			result.DataWarehouseTables = append(result.DataWarehouseTables, dwtype.FromApiTable(t))
		}
	case v1.HiveType:
		if ds.Spec.Hive == nil || ds.Spec.Hive.MetadataStore.Validate() != nil {
			return nil, fmt.Errorf("hive metadata store information invalid")
		}
		querySql := "SELECT `TBL_NAME`, `TBL_TYPE` FROM `TBLS` WHERE `DB_ID` = (SELECT `DB_ID` FROM `DBS` WHERE `NAME` = '" + ds.Spec.Hive.Database + "');"
		mysqlData, err := driver.GetRDBDatabaseData(&ds.Spec.Hive.MetadataStore, querySql)
		if err != nil {
			return nil, err
		}
		for _, v := range mysqlData {
			table := rdb.Table{}
			table.Name = v["TBL_NAME"]
			table.Type = v["TBL_TYPE"]
			result.RDBTables = append(result.RDBTables, table)
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
				info := dwtype.FromApiTable(t).Info
				return &Table{DataWarehouseTable: &info}, nil
			}
		}
		return nil, fmt.Errorf("table %s not found in database %s", table, ds.Spec.Name)
	default:
		return nil, fmt.Errorf("wrong datasource type: %s", ds.Spec.Type)
	}
}

func (s *Service) GetFields(id, table string, opts ...util.OpOption) (*Fields, error) {
	if table == "" {
		return nil, fmt.Errorf("table id or name is null")
	}
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
			querySql := `SELECT
    COLUMN_NAME 'column_name',
    COLUMN_TYPE 'column_type',
    IF(EXTRA='auto_increment',CONCAT(COLUMN_KEY,
                                     '(', IF(EXTRA='auto_increment','auto_increment',EXTRA),')'),
                                     COLUMN_KEY) 'primary_key',
    IS_NULLABLE 'nullable',
    COLUMN_COMMENT 'description' 
FROM information_schema.columns
WHERE table_name='%s'
    AND table_schema='%s';`
			querySql = fmt.Sprintf(querySql, table, ds.Spec.RDB.Database)
			mysqlData, err := driver.GetRDBData(ds, querySql)
			if err != nil {
				return nil, err
			}
			for _, v := range mysqlData {
				table := rdb.Field{}
				table.Name = v["column_name"]
				table.DataType = v["column_type"]
				table.Description = v["description"]
				if len(v["primary_key"]) == 0 {
					table.PrimaryKey = false
				} else {
					table.PrimaryKey = true
				}
				table.IsNullAble = strings.ToUpper(v["nullable"]) == "YES"
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
		result.DataWarehouseFields = make([]dwtype.Property, 0)
		for _, apiTable := range ds.Spec.DataWarehouse.Tables {
			if apiTable.Info.ID == table {
				for _, p := range apiTable.Properties {
					result.DataWarehouseFields = append(result.DataWarehouseFields, dwtype.FromApiProperty(p))
				}
				goto rt
			}
		}
		return nil, fmt.Errorf("cannot find table %s in datasource %s", table, ds.ObjectMeta.Name)
	case v1.HiveType:
		if ds.Spec.Hive == nil || ds.Spec.Hive.MetadataStore.Validate() != nil {
			return nil, fmt.Errorf("hive metadata store information invalid")
		}
		querySql := "SELECT `COLUMN_NAME`, `TYPE_NAME`, `INTEGER_IDX` FROM `COLUMNS_V2` WHERE `CD_ID` = (SELECT `TBL_ID` FROM `TBLS` WHERE `DB_ID` = (SELECT `DB_ID` FROM `DBS` WHERE `NAME` = '" + ds.Spec.Hive.Database + "') AND `TBL_NAME` = '" + table + "');"
		mysqlData, err := driver.GetRDBDatabaseData(&ds.Spec.Hive.MetadataStore, querySql)
		if err != nil {
			return nil, err
		}
		for _, v := range mysqlData {
			field := rdb.Field{}
			field.Name = v["COLUMN_NAME"]
			field.DataType = v["TYPE_NAME"]
			field.Index, _ = strconv.Atoi(v["INTEGER_IDX"])
			result.RDBFields = append(result.RDBFields, field)
		}
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
						dwp := dwtype.FromApiProperty(p)
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
	if ds == nil {
		return fmt.Errorf("datasource is null")
	}
	switch ds.Type {
	case v1.RDBType:
		if err := ds.ValidateConnection(); err != nil {
			return err
		}
		return driver.PingRDB(ToAPI(ds, true))
	case v1.DataWarehouseType:
		_, err := s.GetDataWareshouse(ToAPI(ds, true).ObjectMeta.Annotations)
		if err != nil {
			return fmt.Errorf("cannot get datawarehouse: %+v", err)
		}
		return err
	case v1.TopicType:
		return s.CheckTopic(ds.Namespace, ds.MessageQueue)
	case v1.MongoType:
		return s.CheckMongo(ds.Mongo)
	case v1.HiveType:
		return s.CheckHive(ds.Hive)
	default:
		return fmt.Errorf("not supported for %s", ds.Type)
	}
}

func (s *Service) Query(id, querySql string, opts ...util.OpOption) (interface{}, error) {
	ds, err := s.Get(id, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot datasource: %+v", err)
	}
	switch ds.Spec.Type {
	case v1.RDBType:
	default:
		return nil, fmt.Errorf("unsupported datasource type: %s", ds.Spec.Type)
	}
	return driver.GetRDBData(ds, querySql)
}

func (s *Service) Match(sId, sTable, tId, tTable string, opts ...util.OpOption) ([]FieldsTuple, error) {
	sfields, err := s.GetFields(sId, sTable, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get source datasource fields: %+v", err)
	}
	tfields, err := s.GetFields(tId, tTable, opts...)
	if err != nil {
		return nil, fmt.Errorf("cannot get target datasource fields: %+v", err)
	}
	klog.Infof("srouce fields=%+v", sfields.RDBFields)
	klog.Infof("target fields=%+v", tfields.RDBFields)
	if sfields.RDBFields == nil || len(sfields.RDBFields) == 0 ||
		tfields.RDBFields == nil || len(tfields.RDBFields) == 0 {
		return []FieldsTuple{}, nil
	}
	result := make([]FieldsTuple, 0)
	find := func(name string) (string, string, bool) {
		for _, f := range tfields.RDBFields {
			if name == f.Name {
				return f.Name, f.DataType, true
			}
		}
		return "", "", false
	}
	for _, f := range sfields.RDBFields {
		if n, t, ok := find(f.Name); ok {
			result = append(result, FieldsTuple{
				SourceFieldName: f.Name,
				SourceFieldType: f.DataType,
				TargetFieldName: n,
				TargetFieldType: t,
			})
		} else {
			return []FieldsTuple{}, nil
		}
	}
	return result, nil
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

func (s *Service) CheckTopic(crdNamespace string, mq *v1.MessageQueue) error {
	if mq == nil {
		return fmt.Errorf("message queue is null")
	}
	switch mq.Type {
	case "public":
		if mq.Outter == nil {
			return fmt.Errorf("public message queue connection info is null")
		}
		if mq.Outter.Namespace == "" {
			mq.Outter.Namespace = "pulsar/default"
		}
		if mq.Outter.AuthEnabled && mq.Outter.NamespaceToken == "" {
			return fmt.Errorf("auth enabled but namespace token not priveded")
		}
		endpoints := strings.Split(mq.Outter.Address, ",")
		for _, ep := range endpoints {
			a, err := net.ResolveTCPAddr("tcp", ep)
			if err != nil {
				return fmt.Errorf("cannot resolve address: %+v", err)
			}
			c, err := net.DialTCP("tcp", nil, a)
			if err != nil {
				return fmt.Errorf("cannot dial tcp: %+v", err)
			}
			c.Close()
		}
		return nil
	default:
		if mq.InnerID == nil {
			return fmt.Errorf("topic id not set")
		}
		crd, err := s.tgClient.Namespace(crdNamespace).Get(*mq.InnerID, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error get crd: %+v", err)
		}
		t := &tgv1.Topicgroup{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), t); err != nil {
			return fmt.Errorf("convert unstructured to crd error: %+v", err)
		}
		return nil
	}
}

func (s *Service) CheckMongo(mongo *v1.Mongo) error {
	if mongo == nil {
		return fmt.Errorf("mongo is null")
	}
	if len(mongo.Host) == 0 {
		return fmt.Errorf("host is null")
	}
	if mongo.Port < 1 || mongo.Port > 65536 {
		return fmt.Errorf("invalid port")
	}
	if len(mongo.Database) == 0 {
		return fmt.Errorf("database is null")
	}
	return mongodriver.Ping(mongo)
}

func (s *Service) CheckHive(h *v1.Hive) error {
	if h == nil {
		return fmt.Errorf("mongo is null")
	}
	if len(h.Host) == 0 {
		return fmt.Errorf("host is null")
	}
	if h.Port < 1 || h.Port > 65536 {
		return fmt.Errorf("invalid port")
	}
	if len(h.Database) == 0 {
		return fmt.Errorf("database is null")
	}
	return driver.PingRDBDatabase(&h.MetadataStore)
}
