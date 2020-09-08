package database

import (
	"fmt"

	"github.com/chinamobile/nlpt/apiserver/database/model"
)

func (d *DatabaseConnection) AddApiPlugin(p model.ApiPlugin, ss []model.ApiPluginRelation) (err error) {
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

func (d *DatabaseConnection) AddApiPluginRelation(ss []model.ApiPluginRelation) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
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

//需要关联关系表
func (d *DatabaseConnection) RemoveApiPluginRelation(ss []model.ApiPluginRelation) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
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

func (d *DatabaseConnection) UpdateApiPluginRelation(ss []model.ApiPluginRelation) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	for _, s := range ss {
		if _, err := d.Update(&s); err != nil {
			d.Rollback()
			return err
		}
	}
	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}

func (d *DatabaseConnection) QueryApiPluginRelation(p model.ApiPlugin) ([]model.ApiPluginRelation, error) {
	if err := d.Begin(); err != nil {
		return nil, fmt.Errorf("begin txn error: %+v", err)
	}

	ss := []model.ApiPluginRelation{}
	if _, err := d.QueryTable("ApiPluginRelation").Filter("ApiPluginId", p.Id).All(&ss); err != nil {
		d.Rollback()
		return nil, err
	}

	return ss, nil
}

func (d *DatabaseConnection) QueryApiPluginRelationByApi(p model.ApiPlugin, apiId string) ([]model.ApiPluginRelation, error) {
	if err := d.Begin(); err != nil {
		return nil, fmt.Errorf("begin txn error: %+v", err)
	}

	ss := []model.ApiPluginRelation{}
	if _, err := d.QueryTable("ApiPluginRelation").Filter("ApiPluginId", p.Id).Filter("ApiId", apiId).All(&ss); err != nil {
		d.Rollback()
		return nil, err
	}
	return ss, nil
}

//删除apigroup时，需要关联删除 api关系表
func (d *DatabaseConnection) RemoveApiPlugin(id string) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	p := model.ApiPlugin{Id: id}
	if _, err = d.Delete(&p); err != nil {
		d.Rollback()
		return err
	}
	ss := []model.ApiPluginRelation{}
	if _, err := d.QueryTable("ApiPluginRelation").Filter("ApiPluginId", id).All(&ss); err != nil {
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

func (d *DatabaseConnection) QueryApiPlugin(p model.ApiPlugin) ([]model.ApiPlugin, error) {
	result := []model.ApiPlugin{}
	q := d.QueryTable("ApiPlugin")
	if len(p.Status) > 0 {
		q = q.Filter("Status", p.Status)
	}
	if len(p.Namespace) > 0 {
		q = q.Filter("Namespace", p.Namespace)
	}
	if len(p.User) > 0 {
		q = q.Filter("User", p.User)
	}
	if len(p.Type) > 0 {
		q = q.Filter("Type", p.Type)
	}
	if len(p.Name) > 0 {
		q = q.Filter("Name__icontains", p.Name)
	}
	if _, err := q.All(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *DatabaseConnection) GetApiPlugin(id string) (model.ApiPlugin, []model.ApiPluginRelation, error) {
	ApiPlugin := model.ApiPlugin{Id: id}
	err := d.Read(&ApiPlugin)
	if err != nil {
		return model.ApiPlugin{}, nil, err
	}
	ss := []model.ApiPluginRelation{}
	if _, err := d.QueryTable("ApiPluginRelation").Filter("ApiPluginId", id).All(&ss); err != nil {
		d.Rollback()
		return model.ApiPlugin{}, nil, err
	}
	return ApiPlugin, ss, nil
}

func (d *DatabaseConnection) UpdateApiPlugin(p model.ApiPlugin, ss []model.ApiPluginRelation) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	_, err = d.Update(&p)
	if err != nil {
		d.Rollback()
		return err
	}
	if ss != nil {
		os := []model.ApiPluginRelation{}
		if _, err := d.QueryTable("ApiPluginRelation").Filter("ApiPluginId", p.Id).All(&os); err != nil {
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
