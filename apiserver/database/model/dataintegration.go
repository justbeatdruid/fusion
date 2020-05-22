package model

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
)

//Task ...
type Task struct {
	Id               int    `orm:"pk;auto"`
	DagId            string `orm:"unique"`
	Name             string `orm:"unique"`
	Description      string
	Type             string
	SchedualPlan     string `orm:"size(5000)"`
	DataSourceConfig string `orm:"size(5000)"`
	DataTargetConfig string `orm:"size(5000)"`
	CreatedTime      time.Time
}

//TbDagRun ...
type TbDagRun struct {
	DagId     string `orm:"pk;unique"`
	ExecDate  time.Time
	StartDate time.Time
	EndDate   time.Time
	DagStatus int
	Remark    string `orm:"size(5000)"`
}

// AddTask ...
func AddTask(m *Task) (id int64, err error) {
	o := orm.NewOrm()
	v := Task{DagId: m.DagId}

	err = o.QueryTable("Task").Filter("DagId", m.DagId).One(&v)
	if err == nil {
		v.Name = m.Name
		_, err = o.Update(&v)
		id = int64(v.Id)
	} else {
		id, err = o.Insert(m)

	}
	return
}

// GetTaskByID ...
func GetTaskByID(id int) (v *Task, err error) {
	o := orm.NewOrm()
	v = &Task{Id: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetTaskByDagId ...
func GetTaskByDagId(dagId string) (v *Task, err error) {
	o := orm.NewOrm()

	v = &Task{DagId: dagId}
	err = o.QueryTable("Task").Filter("DagId", dagId).One(v)
	return v, err
}

// GetTaskByName ...
func GetTaskByName(name string) (v *Task, err error) {
	o := orm.NewOrm()

	v = &Task{Name: name}
	err = o.QueryTable("Task").Filter("Name", name).One(v)
	return v, err
}

// DeleteTaskByDagId ...
func DeleteTaskByDagId(dagId string) (v *Task, err error) {
	o := orm.NewOrm()
	v = &Task{DagId: dagId}
	err = o.QueryTable("Task").Filter("DagId", dagId).One(v)
	if err != nil {
		return
	}
	_, err = o.QueryTable("Task").Filter("DagId", dagId).Delete()
	return
}

// UpdateTaskByID ...
func UpdateTaskByID(m *Task) (err error) {
	o := orm.NewOrm()
	v := Task{Id: m.Id}
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			fmt.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteTaskByID ...
func DeleteTaskByID(id int) (err error) {
	o := orm.NewOrm()
	if err = o.Read(&Task{Id: id}); err == nil {
		var num int64
		if num, err = o.Delete(&Task{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}

// GetTasks ...
func GetTasks(offet, limit int) (tasks []Task, err error) {
	o := orm.NewOrm()
	_, err = o.QueryTable("Task").Offset(offet).Limit(limit).OrderBy("-Id").All(&tasks)
	return
}
