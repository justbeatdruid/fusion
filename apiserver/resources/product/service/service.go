package service

import (
	"fmt"

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

func (s *Service) CreateProduct(model *Product) (*Product, error) {
	model.Status = "unpublished"
	p, ss, err := ToModel(*model)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	if err = s.db.AddProduct(p, ss); err != nil {
		return nil, fmt.Errorf("cannot write database: %+v", err)
	}
	return model, nil
}

func (s *Service) ListProduct(p Product) ([]*Product, error) {
	condition, _, err := ToModel(p)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	products, err := s.db.QueryProduct(condition)
	if err != nil {
		return nil, fmt.Errorf("cannot read database: %+v", err)
	}
	result := make([]*Product, len(products))
	for i := range products {
		product, err := FromModel(products[i], nil)
		if err != nil {
			return nil, fmt.Errorf("cannot get model: %+v", err)
		}
		result[i] = &product
	}
	return result, nil
}

func (s *Service) GetProduct(id string) (*Product, error) {
	p, ss, err := s.db.GetProduct(id)
	if err != nil {
		return nil, fmt.Errorf("cannot query database: %+v", err)
	}
	product, err := FromModel(p, ss)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	return &product, nil
}

func (s *Service) DeleteProduct(id string) error {
	if err := s.db.RemoveProduct(id); err != nil {
		return fmt.Errorf("cannot write database: %+v", err)
	}
	return nil
}

func (s *Service) UpdateProduct(model *Product) (*Product, error) {
	existedProduct, err := s.GetProduct(model.Id)
	if err != nil {
		return nil, fmt.Errorf("cannot find product with id %s: %+v", model.Id, err)
	}
	if existedProduct.User != model.User {
		return nil, fmt.Errorf("permission denied: wrong user")
	}
	if existedProduct.Tenant != model.Tenant {
		return nil, fmt.Errorf("permission denied: wrong tenant")
	}
	p, ss, err := ToModel(*model)
	if err != nil {
		return nil, fmt.Errorf("cannot get model: %+v", err)
	}
	if err := s.db.UpdateProduct(p, ss); err != nil {
		return nil, fmt.Errorf("cannot write database: %+v", err)
	}
	return model, nil
}

func (s *Service) UpdateProductStatus(model *Product) (*Product, error) {
	existedProduct, err := s.GetProduct(model.Id)
	if err != nil {
		return nil, fmt.Errorf("cannot find product with id %s: %+v", model.Id, err)
	}
	if existedProduct.User != model.User {
		return nil, fmt.Errorf("permission denied: wrong user")
	}
	if existedProduct.Tenant != model.Tenant {
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
	if err := s.db.UpdateProduct(p, nil); err != nil {
		return nil, fmt.Errorf("cannot write database: %+v", err)
	}
	return model, nil
}
