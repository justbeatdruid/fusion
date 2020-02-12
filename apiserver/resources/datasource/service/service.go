package service

import (
	"database/sql"
	"fmt"

	api "github.com/chinamobile/nlpt/apiserver/resources/api/service"
	dw "github.com/chinamobile/nlpt/apiserver/resources/datasource/datawarehouse"
	serviceunit "github.com/chinamobile/nlpt/apiserver/resources/serviceunit/service"
	"github.com/chinamobile/nlpt/crds/datasource/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

var crdNamespace = "default"

var Supported = []string{}

type Service struct {
	client     dynamic.NamespaceableResourceInterface
	apiService *api.Service
	suService  *serviceunit.Service
}

func NewService(client dynamic.Interface, supported []string) *Service {
	Supported = supported
	return &Service{
		client:     client.Resource(v1.GetOOFSGVR()),
		apiService: api.NewService(client),
		suService:  serviceunit.NewService(client),
	}
}

func (s *Service) CreateDatasource(model *Datasource) (*Datasource, error) {
	dealType := "create"
	if err := model.Validate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	ds, err := s.Create(ToAPI(model, dealType))
	if err != nil {
		return nil, fmt.Errorf("cannot create object: %+v", err)
	}
	return ToModel(ds), nil
}
func (s *Service) UpdateDatasource(model *Datasource) (*Datasource, error) {
	dealType := "update"
	if err := model.ValidateForUpdate(); err != nil {
		return nil, fmt.Errorf("bad request: %+v", err)
	}
	ds, err := s.Update(ToAPI(model, dealType))
	if err != nil {
		return nil, fmt.Errorf("cannot update object: %+v", err)
	}
	return ToModel(ds), nil
}
func (s *Service) ListDatasource() ([]*Datasource, error) {
	dss, err := s.List()
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListModel(dss), nil
}

func (s *Service) GetDatasource(id string) (*Datasource, error) {
	ds, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToModel(ds), nil
}

func (s *Service) DeleteDatasource(id string) error {
	return s.Delete(id)
}

func (s *Service) Create(ds *v1.Datasource) (*v1.Datasource, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ds)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error creating crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ds); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasource of creating: %+v", ds)
	return ds, nil
}

func (s *Service) Update(ds *v1.Datasource) (*v1.Datasource, error) {
	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ds)
	if err != nil {
		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
	}
	crd := &unstructured.Unstructured{}
	crd.SetUnstructuredContent(content)

	crd, err = s.client.Namespace(crdNamespace).Update(crd, metav1.UpdateOptions{})
	if err != nil {
		return nil, fmt.Errorf("error updateing crd: %+v", err)
	}

	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ds); err != nil {
		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasource of creating: %+v", ds)
	return ds, nil
}

func (s *Service) List() (*v1.DatasourceList, error) {
	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
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

func (s *Service) Get(id string) (*v1.Datasource, error) {
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

func (s *Service) Delete(id string) error {
	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error delete crd: %+v", err)
	}
	return nil
}

func (s *Service) GetTables(id string) (*Tables, error) {
	result := &Tables{}
	ds, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot datasource: %+v", err)
	}
	switch ds.Spec.Type {
	case v1.RDBType:
		if ds.Spec.RDB == nil {
			return nil, fmt.Errorf("datasource %s in type rdb has no rdb instance", ds.ObjectMeta.Name)
		}
	case v1.DataWarehouseType:
		if ds.Spec.DataWarehouse == nil {
			return nil, fmt.Errorf("datasource %s in type datawarehouse has no datawarehouse instance", ds.ObjectMeta.Name)
		}
		result.DataWarehouseTables = make([]dw.Table, 0)
		for _, t := range ds.Spec.DataWarehouse.Tables {
			result.DataWarehouseTables = append(result.DataWarehouseTables, dw.FromApiTable(t))
		}
	default:
		return nil, fmt.Errorf("wrong datasource type: %s", ds.Spec.Type)
	}
	return result, nil
}

func (s *Service) GetFields(id, table string) (*Fields, error) {
	result := &Fields{}
	ds, err := s.Get(id)
	if err != nil {
		return nil, fmt.Errorf("cannot datasource: %+v", err)
	}
	switch ds.Spec.Type {
	case v1.RDBType:
		if ds.Spec.RDB == nil {
			return nil, fmt.Errorf("datasource %s in type rdb has no rdb instance", ds.ObjectMeta.Name)
		}
	case v1.DataWarehouseType:
		if ds.Spec.DataWarehouse == nil {
			return nil, fmt.Errorf("datasource %s in type datawarehouse has no datawarehouse instance", ds.ObjectMeta.Name)
		}
		result.DataWarehouseFields = make([]dw.Property, 0)
		for _, apiTable := range ds.Spec.DataWarehouse.Tables {
			if apiTable.Name == table {
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

func (s *Service) GetDataSourceByApiId(apiId string, parames string) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	/*
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
	*/
	return result, nil
}

/**
golang连接查询mysql
*/
func GetMySQLDbData(db *sql.DB, querySql string) ([]map[string]string, error) {
	rows, err := db.Query(querySql)
	if err != nil {
		fmt.Println("Query fail 。。。")
		return nil, fmt.Errorf("error query api: %+v", err)
	}
	//获取列名
	columns, _ := rows.Columns()

	//定义一个切片,长度是字段的个数,切片里面的元素类型是sql.RawBytes
	values := make([]sql.RawBytes, len(columns))
	//定义一个切片,元素类型是interface{} 接口
	scanArgs := make([]interface{}, len(values))
	for i := range values {
		//把sql.RawBytes类型的地址存进去了
		scanArgs[i] = &values[i]
	}
	//获取字段值
	var result []map[string]string
	for rows.Next() {
		res := make(map[string]string)
		rows.Scan(scanArgs...)
		for i, col := range values {
			res[columns[i]] = string(col)
		}
		result = append(result, res)
	}

	fmt.Println("Query success")
	rows.Close()
	return result, nil
}
