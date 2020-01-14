package service

import (
	"fmt"
	api "github.com/chinamobile/nlpt/apiserver/resources/api/service"
	serviceunit "github.com/chinamobile/nlpt/apiserver/resources/serviceunit/service"
	"github.com/chinamobile/nlpt/crds/datasource/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
)

var crdNamespace = "default"

type Service struct {
	client     dynamic.NamespaceableResourceInterface
	apiService *api.Service
	suService  *serviceunit.Service
}

func NewService(client dynamic.Interface) *Service {
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
	for _, v := range serverUnit.Spec.DatasourcesID {
		datasource, err := s.Get(v.ID)
		if err != nil {
			return nil, fmt.Errorf("error query singleDateSourceId: %+v", err)
		}
		//TODO The remaining operation after the query to the data source
		result["Fields"] = datasource.Spec.Fields
	}
	return result, nil
}
