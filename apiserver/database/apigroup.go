package database

import (
	"fmt"

	"github.com/chinamobile/nlpt/apiserver/database/model"
)

func (d *DatabaseConnection) AddApiGroup(p model.ApiGroup, ss []model.ApiRelation) (err error) {
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

//删除apigroup时，需要关联删除 api关系表
func (d *DatabaseConnection) RemoveApiGroup(id string) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	p := model.ApiGroup{Id: id}
	if _, err = d.Delete(&p); err != nil {
		d.Rollback()
		return err
	}
	ss := []model.ApiRelation{}
	if _, err := d.QueryTable("ApiRelation").Filter("ApiGroupId", id).All(&ss); err != nil {
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

func (d *DatabaseConnection) QueryApiGroup(p model.ApiGroup) ([]model.ApiGroup, error) {
	result := []model.ApiGroup{}
	q := d.QueryTable("ApiGroup")
	if len(p.Status) > 0 {
		q = q.Filter("Status", p.Status)
	}
	if len(p.Namespace) > 0 {
		q = q.Filter("Namespace", p.Namespace)
	}
	if len(p.User) > 0 {
		q = q.Filter("User", p.User)
	}
	if _, err := q.All(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *DatabaseConnection) GetApiGroup(id string) (model.ApiGroup, []model.ApiRelation, error) {
	ApiGroup := model.ApiGroup{Id: id}
	err := d.Read(&ApiGroup)
	if err != nil {
		return model.ApiGroup{}, nil, err
	}
	ss := []model.ApiRelation{}
	if _, err := d.QueryTable("ApiRelation").Filter("ApiGroupId", id).All(&ss); err != nil {
		d.Rollback()
		return model.ApiGroup{}, nil, err
	}
	return ApiGroup, ss, nil
}

func (d *DatabaseConnection) UpdateApiGroup(p model.ApiGroup, ss []model.ApiRelation) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	_, err = d.Update(&p)
	if err != nil {
		d.Rollback()
		return err
	}
	if ss != nil {
		os := []model.ApiRelation{}
		if _, err := d.QueryTable("ApiRelation").Filter("ApiGroupId", p.Id).All(&os); err != nil {
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
