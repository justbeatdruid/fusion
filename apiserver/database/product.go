package database

import (
	"fmt"

	"github.com/chinamobile/nlpt/apiserver/database/model"
)

func (d *DatabaseConnection) AddProduct(p model.Product, ss []model.Scenario) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	_, err = d.Insert(&p)
	if err != nil {
		d.Rollback()
		return err
	}
	for _, s := range ss {
		_, err = d.Insert(&s)
		if err != nil {
			d.Rollback()
			return err
		}
	}
	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}

func (d *DatabaseConnection) RemoveProduct(id string) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	p := model.Product{Id: id}
	if _, err = d.Delete(&p); err != nil {
		d.Rollback()
		return err
	}
	ss := []model.Scenario{}
	if _, err := d.QueryTable("Scenario").Filter("ProductId", id).All(&ss); err != nil {
		d.Rollback()
		return err
	}
	for _, s := range ss {
		if _, err := d.Delete(&s); err != nil {
			d.Rollback()
			return err
		}
	}
	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}

func (d *DatabaseConnection) QueryProduct(p model.Product) ([]model.Product, error) {
	result := []model.Product{}
	q := d.QueryTable("Product")
	if len(p.Status) > 0 {
		q = q.Filter("Status", p.Status)
	}
	if len(p.Tenant) > 0 {
		q = q.Filter("Tenant", p.Tenant)
	}
	if len(p.User) > 0 {
		q = q.Filter("User", p.User)
	}
	if len(p.Category) > 0 {
		q = q.Filter("Category", p.Category)
	}
	if _, err := q.All(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *DatabaseConnection) GetProduct(id string) (model.Product, []model.Scenario, error) {
	product := model.Product{Id: id}
	err := d.Read(&product)
	if err != nil {
		return model.Product{}, nil, err
	}
	ss := []model.Scenario{}
	if _, err := d.QueryTable("Scenario").Filter("ProductId", id).All(&ss); err != nil {
		d.Rollback()
		return model.Product{}, nil, err
	}
	return product, ss, nil
}

func (d *DatabaseConnection) UpdateProduct(p model.Product, ss []model.Scenario) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	_, err = d.Update(&p)
	if err != nil {
		d.Rollback()
		return err
	}
	if ss != nil {
		os := []model.Scenario{}
		if _, err := d.QueryTable("Scenario").Filter("ProductId", p.Id).All(&os); err != nil {
			d.Rollback()
			return err
		}
		for _, s := range os {
			if _, err := d.Delete(&s); err != nil {
				d.Rollback()
				return err
			}
		}
		for _, s := range ss {
			_, err = d.Insert(&s)
			if err != nil {
				d.Rollback()
				return err
			}
		}
	}
	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}
