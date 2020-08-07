package database

import (
	"fmt"

	"github.com/chinamobile/nlpt/apiserver/database/model"
)

func (d *DatabaseConnection) AddDatasource(obj interface{}) error {
	o, err := model.DatasourceGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get topic from obj error: %+v", err)
	}
	return d.AddObject(&o, nil, 0, nil, 0)
}

func (d *DatabaseConnection) UpdateDatasource(old, obj interface{}) error {
	_, err := model.DatasourceGetFromObject(old)
	if err != nil {
		return fmt.Errorf("get users from obj error: %+v", err)
	}
	o, err := model.DatasourceGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get topic from obj error: %+v", err)
	}
	return d.UpdateObject(&o, nil, 0, nil, 0)
}

func (d *DatabaseConnection) DeleteDatasource(obj interface{}) error {
	o, err := model.DatasourceGetFromObject(obj)
	if err != nil {
		return fmt.Errorf("get topic from obj error: %+v", err)
	}
	return d.DeleteObject(&o)
}

func (d *DatabaseConnection) QueryDatasource(md *model.Datasource) ([]model.Datasource, error) {
	conditions := make([]model.Condition, 0)
	if md == nil {
		return nil, fmt.Errorf("model is null")
	}
	if len(md.Namespace) == 0 {
		return nil, fmt.Errorf("namespace not set in model")
	}
	conditions = append(conditions, model.Condition{"namespace", model.Equals, md.Namespace})
	if len(md.User) > 0 {
		conditions = append(conditions, model.Condition{"user", model.Equals, md.User})
	}
	if len(md.Name) > 0 {
		conditions = append(conditions, model.Condition{"name", model.Like, md.Name})
	}
	if len(md.Status) > 0 {
		conditions = append(conditions, model.Condition{"status", model.Equals, md.Status})
	}
	if len(md.Type) > 0 {
		conditions = append(conditions, model.Condition{"type", model.Equals, md.Type})
	}
	result := make([]model.Datasource, 0)
	err := d.query("", md, conditions, &result)
	return result, err
}
