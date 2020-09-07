package database

import (
	"fmt"

	"github.com/chinamobile/nlpt/apiserver/database/model"
)

func (d *DatabaseConnection) AddTask(p model.Task) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	_, err = d.Insert(&p)
	if err != nil {
		d.Rollback()
		return err
	}
	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}

func (d *DatabaseConnection) RemoveTask(id int) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	p := model.Task{Id: id}
	if _, err = d.Delete(&p); err != nil {
		d.Rollback()
		return err
	}
	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}

func (d *DatabaseConnection) QueryTask(p model.Task) ([]model.Task, error) {
	result := []model.Task{}
	q := d.QueryTable("Task")
	if len(p.SourceId) > 0 {
		q = q.Filter("source_id", p.SourceId)
	}
	if len(p.TargetId) > 0 {
		q = q.Filter("target_id", p.SourceId)
	}
	if _, err := q.All(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func (d *DatabaseConnection) GetTask(id int) (model.Task, error) {
	task := model.Task{Id: id}
	err := d.Read(&task)
	if err != nil {
		return model.Task{}, err
	}
	return task, nil
}

func (d *DatabaseConnection) UpdateTask(p model.Task) (err error) {
	if err = d.Begin(); err != nil {
		return fmt.Errorf("begin txn error: %+v", err)
	}
	_, err = d.Update(&p)
	if err != nil {
		d.Rollback()
		return err
	}
	if err = d.Commit(); err != nil {
		return fmt.Errorf("commit txn error: %+v", err)
	}
	return nil
}

func (d *DatabaseConnection) DatasourceOccupiedByTask(did string) (bool, error) {
	sql := fmt.Sprintf(`SELECT * FROM %s WHERE source_id = '%s' OR target_id = '%s'`,
		"task", did, did)
	result := make([]model.Task, 0)
	if n, err := d.Raw(sql).QueryRows(&result); err != nil {
		return false, fmt.Errorf("query db error: %+v", err)
	} else if n > 0 {
		return true, nil
	}
	return false, nil
}
