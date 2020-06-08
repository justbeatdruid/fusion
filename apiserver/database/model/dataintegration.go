package model

import (
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"k8s.io/klog"
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
	Namespace        string
	UserId           string
	Status           bool
	Job              string
	StartTime        time.Time
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

//TbMetadata ...
type TbMetadata struct {
	DagId           string `orm:"pk;unique"`
	DbId            string
	Owner           string
	TableName       string `orm:"size(128)"`
	ColumnName      string `orm:"size(128)"`
	OrdinalPosition int
	ColumnType      string
	Charset         string
	Comments        string
	IsAdd           int
	AddValue        string
	StatusFlag      int
	PartitionInfo   string
	Remark          string
	CreateTime      time.Time
	UpdataTime      time.Time
}

func AddTbMetadata(md *TbMetadata) error {
	o := orm.NewOrm()
	_, err := o.Insert(md)
	return err

}

// AddTask ...
func AddTask(m *Task) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)

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
func GetTaskByDagId(dagId, userId, namespace string) (v *Task, err error) {
	o := orm.NewOrm()

	v = &Task{DagId: dagId}
	err = o.QueryTable("Task").Filter("DagId", dagId).Filter("UserId", userId).Filter("Namespace", namespace).One(v)
	return v, err
}

// GetTaskByName ...
func GetTaskByName(name, userId, namespace string) (v *Task, err error) {
	o := orm.NewOrm()

	v = &Task{
		Name:      name,
		UserId:    userId,
		Namespace: namespace,
	}
	err = o.QueryTable("Task").Filter("Name", name).Filter("UserId", userId).Filter("Namespace", namespace).One(v)
	return v, err
}

// DeleteTaskByDagId ...
func DeleteTaskByDagId(dagId, userId, namespace string) (v *Task, err error) {
	o := orm.NewOrm()
	v = &Task{DagId: dagId}
	err = o.QueryTable("Task").Filter("DagId", dagId).Filter("UserId", userId).Filter("Namespace", namespace).One(v)
	if err != nil {
		return
	}
	_, err = o.QueryTable("Task").Filter("DagId", dagId).Filter("UserId", userId).Filter("Namespace", namespace).Delete()
	return
}

// UpdateTaskByID ...
func UpdateTaskByID(m *Task) error {
	o := orm.NewOrm()

	num, err := o.Update(m)
	if err == nil {
		klog.Errorf("Number of records updated in database:%v", num)

	}
	return err
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
func GetTasks(offet, limit int, name, userId, namespace string) (tasks []Task, num, total int64, err error) {
	o := orm.NewOrm()
	num, err = o.QueryTable("Task").Filter("Name__contains", name).Filter("UserId", userId).Filter("Namespace", namespace).Offset(offet).Limit(limit).OrderBy("-Id").All(&tasks)

	total, err = o.QueryTable("Task").Filter("Name__contains", name).Filter("UserId", userId).Filter("Namespace", namespace).Count()
	return
}

// GetTbDagRun ...
func GetTbDagRun(dagID string) (dagRun []TbDagRun, num int64, err error) {
	o := orm.NewOrm()
	num, err = o.QueryTable("TbDagRun").Filter("DagId", dagID).OrderBy("-ExecDate").All(&dagRun)
	return
}

// OperationTaskStatus ...
func OperationTaskStatus(operation, userId, namespace string, ids []string) (task []Task, err error) {
	o := orm.NewOrm()

	if operation == "delete" {
		_, err = o.QueryTable("Task").Filter("DagId__in", ids).Filter("UserId", userId).Filter("Namespace", namespace).All(&task)
		if err != nil {
			return task, err
		}

		p, err := o.Raw("DELETE  FROM task  WHERE dag_id = ? and user_id = ? and namespace = ?").Prepare()

		for i := range ids {
			_, err = p.Exec(ids[i], userId, namespace)

			if err != nil {
				klog.Errorf("exec sql failed,err: %v", err)
			}

		}
		p.Close()
		return task, err
	}
	if operation == "stop" {
		_, err = o.QueryTable("Task").Filter("DagId__in", ids).Filter("UserId", userId).Filter("Namespace", namespace).Filter("Status", true).All(&task)
		if err != nil {
			return task, err
		}

	}
	_, err = o.QueryTable("Task").Filter("DagId__in", ids).Filter("UserId", userId).Filter("Namespace", namespace).Filter("Status", false).All(&task)
	if err != nil {
		return task, err
	}

	p, err := o.Raw("UPDATE task SET status = ? , job = ? WHERE dag_id = ? and user_id = ? and namespace = ? and status = ?").Prepare()

	for i := range ids {
		if operation == "stop" {
			_, err = p.Exec(false, "", ids[i], userId, namespace, true)

		} else {
			_, err = p.Exec(true, "", ids[i], userId, namespace, false)

		}
		if err != nil {
			klog.Errorf("exec sql failed,err: %v", err)
		}

	}
	p.Close()
	return task, err
}

// OperationTaskStatus ...
func UpdateTaskJob(dagid, job string) error {

	o := orm.NewOrm()
	_, err := o.Raw("UPDATE task SET job = ? WHERE dag_id = ?", job, dagid).Exec()

	return err
}

//GetTaskByStartTime ...
func GetTaskByStartTime() (tasks []Task, err error) {
	o := orm.NewOrm()
	_, err = o.QueryTable("Task").Filter("StartTime__lte", time.Now()).Filter("Job", "").Filter("Status", true).Limit(10).OrderBy("-StartTime").All(&tasks)

	p, err := o.Raw("UPDATE task SET Job = ? WHERE dag_id = ? ").Prepare()
	for i := range tasks {

		_, err = p.Exec("default", tasks[i].DagId)
		if err != nil {
			klog.Errorf("exec sql failed,err: %v", err)
		}

	}
	p.Close()

	return

}
