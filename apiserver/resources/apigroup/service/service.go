package service

import (
	"fmt"
	"k8s.io/klog"

	"github.com/chinamobile/nlpt/apiserver/database"
	// "k8s.io/klog"
)

var crdNamespace = "default"

type Service struct {
	tenantEnabled bool
	db            *database.DatabaseConnection
}

func NewService(tenantEnabled bool, db *database.DatabaseConnection) *Service {
	return &Service{
		tenantEnabled: tenantEnabled,
		db:            db,
	}
}

func (s *Service) CreateApiGroup(model *ApiGroup) (*ApiGroup, error) {
	model.Status = "unpublished"
	p, ss, err := ToModel(*model)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	if err = s.db.AddApiGroup(p, ss); err != nil {
		return nil, fmt.Errorf("cannot write database: %+v", err)
	}
	klog.Infof("create apigroup success :%+v", model)
	return model, nil
}

func (s *Service) ListApiGroup(p ApiGroup) ([]*ApiGroup, error) {
	condition, _, err := ToModel(p)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	apigroups, err := s.db.QueryApiGroup(condition)
	if err != nil {
		return nil, fmt.Errorf("cannot read database: %+v", err)
	}
	result := make([]*ApiGroup, len(apigroups))
	for i := range apigroups {
		apigroup, err := FromModel(apigroups[i], nil)
		if err != nil {
			return nil, fmt.Errorf("cannot get model: %+v", err)
		}
		result[i] = &apigroup
	}
	return result, nil
}

func (s *Service) GetApiGroup(id string) (*ApiGroup, error) {
	p, ss, err := s.db.GetApiGroup(id)
	if err != nil {
		return nil, fmt.Errorf("cannot query database: %+v", err)
	}
	product, err := FromModel(p, ss)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	return &product, nil
}

func (s *Service) DeleteApiGroup(id string) error {
	if err := s.db.RemoveApiGroup(id); err != nil {
		return fmt.Errorf("cannot write database: %+v", err)
	}
	return nil
}

func (s *Service) UpdateApiGroup(model *ApiGroup, id string) (*ApiGroup, error) {
	existed, err := s.GetApiGroup(id)
	if err != nil {
		return nil, fmt.Errorf("cannot find apigroups with id %s: %+v", id, err)
	}
	if existed.User != model.User {
		return nil, fmt.Errorf("permission denied: wrong user")
	}
	if existed.Namespace != model.Namespace {
		return nil, fmt.Errorf("permission denied: wrong tenant")
	}
	//当前支持更新名称和描述信息
	if len(model.Name) != 0 {
		existed.Name = model.Name
	}
	if len(model.Description) != 0 {
		existed.Description = model.Description
	}
	//p, _, err := ToModel(*model)
	p, _, err := ToModel(*existed)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	//更新时只能更新名称描述信息，关联关系通过其他接口更新
	if err := s.db.UpdateApiGroup(p, nil); err != nil {
		return nil, fmt.Errorf("cannot write database: %+v", err)
	}
	return model, nil
}

func (s *Service) UpdateApiGroupStatus(model *ApiGroup) (*ApiGroup, error) {
	existedProduct, err := s.GetApiGroup(model.Id)
	if err != nil {
		return nil, fmt.Errorf("cannot find product with id %s: %+v", model.Id, err)
	}
	if existedProduct.User != model.User {
		return nil, fmt.Errorf("permission denied: wrong user")
	}
	if existedProduct.Namespace != model.Namespace {
		return nil, fmt.Errorf("permission denied: wrong tenant")
	}
	switch model.Status {
	case "online", "offline":
	default:
		return nil, fmt.Errorf("wrong status: %s", model.Status)
	}
	existedProduct.Status = model.Status
	p, _, err := ToModel(*existedProduct)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	if err := s.db.UpdateApiGroup(p, nil); err != nil {
		return nil, fmt.Errorf("cannot write database: %+v", err)
	}
	return model, nil
}
