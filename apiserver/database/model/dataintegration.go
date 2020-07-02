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
	Id        int `orm:"pk;auto"`
	DagId     string
	ExecDate  time.Time
	StartDate time.Time
	EndDate   time.Time
	DagStatus int
	Remark    string `orm:"size(5000)"`
}

//TbMetadata ...
type TbMetadata struct {
	ColId           string `orm:"pk;unique"`
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

// DeleteTbMetadataByDagId ...
func DeleteTbMetadataByDagId(dagId string) (err error) {
	o := orm.NewOrm()
	_, err = o.QueryTable("TbMetadata").Filter("ColId", dagId).Delete()
	return
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
func GetTaskByDagId(dagId, namespace string) (v *Task, err error) {
	o := orm.NewOrm()

	v = &Task{DagId: dagId}
	err = o.QueryTable("Task").Filter("DagId", dagId).Filter("Namespace", namespace).One(v)
	return v, err
}

// GetTaskByName ...
func GetTaskByName(name, namespace string) (v *Task, err error) {
	o := orm.NewOrm()

	v = &Task{
		Name:      name,
		Namespace: namespace,
	}
	err = o.QueryTable("Task").Filter("Name", name).Filter("Namespace", namespace).One(v)
	return v, err
}

// DeleteTaskByDagId ...
func DeleteTaskByDagId(dagId, namespace string) (v *Task, err error) {
	o := orm.NewOrm()
	v = &Task{DagId: dagId}
	err = o.QueryTable("Task").Filter("DagId", dagId).Filter("Namespace", namespace).One(v)
	if err != nil {
		return
	}
	_, err = o.QueryTable("Task").Filter("DagId", dagId).Filter("Namespace", namespace).Delete()
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
func GetTasks(offet, limit int, name, namespace, taskType string, status int, userID []string, createTime []string, createUser string) (tasks []Task, num, total int64, err error) {
	o := orm.NewOrm()

	qs := o.QueryTable("Task").Filter("Namespace", namespace)
	if name != "" {
		qs = qs.Filter("Name__icontains", name)
	}
	if taskType != "all" {
		qs = qs.Filter("Type", taskType)
	}
	if status == 0 {
		qs = qs.Filter("Status", false)
	}
	if status == 1 {
		qs = qs.Filter("Status", true)
	}
	if len(createTime) == 2 {
		start, err1 := time.Parse("2006-01-02 15:04:05", createTime[0])
		end, err2 := time.Parse("2006-01-02 15:04:05", createTime[1])
		if err1 == nil && err2 == nil {
			qs = qs.Filter("StartTime__gte", start).Filter("StartTime__lte", end)
		} else {
			klog.Errorf("createTime:%v,err1:%v, err2:%v", createTime, err1, err2)
		}

	}
	if createUser != "" {
		if len(userID) == 0 {
			return
		}
		qs = qs.Filter("UserId__in", userID)
	}
	num, err = qs.Offset(offet).Limit(limit).OrderBy("-Id").All(&tasks)

	total, err = qs.Count()
	return
}

// GetTbDagRun ...
func GetTbDagRun(offet, limit int, dagID string, execTime []string) (dagRun []TbDagRun, num int64, total int64, err error) {

	o := orm.NewOrm()
	qs := o.QueryTable("TbDagRun").Filter("DagId", dagID)
	if len(execTime) == 2 {
		start, err1 := time.Parse("2006-01-02 15:04:05", execTime[0])
		end, err2 := time.Parse("2006-01-02 15:04:05", execTime[1])
		if err1 == nil && err2 == nil {
			qs = qs.Filter("ExecDate__gte", start).Filter("ExecDate__lte", end)
		} else {
			klog.Errorf("execTime:%v,err1:%v, err2:%v", execTime, err1, err2)
		}

	}
	num, err = qs.Offset(offet).Limit(limit).OrderBy("-ExecDate").All(&dagRun)
	total, err = qs.Count()
	return
}

// OperationTaskStatus ...
func OperationTaskStatus(operation, namespace string, ids []string) (task []Task, err error) {
	o := orm.NewOrm()

	if operation == "delete" {
		_, err = o.QueryTable("Task").Filter("DagId__in", ids).Filter("Namespace", namespace).All(&task)
		if err != nil {
			return task, err
		}

		p, err := o.Raw("DELETE  FROM task  WHERE dag_id = ?  and namespace = ?").Prepare()

		for i := range ids {
			_, err = p.Exec(ids[i], namespace)

			if err != nil {
				klog.Errorf("exec sql failed,err: %v", err)
			}

		}
		p.Close()
		return task, err
	}
	if operation == "stop" {
		_, err = o.QueryTable("Task").Filter("DagId__in", ids).Filter("Namespace", namespace).Filter("Status", true).All(&task)
		if err != nil {
			return task, err
		}

	} else {
		_, err = o.QueryTable("Task").Filter("DagId__in", ids).Filter("Namespace", namespace).Filter("Status", false).All(&task)
		if err != nil {
			return task, err
		}

	}

	p, err := o.Raw("UPDATE task SET status = ? , job = ? WHERE dag_id = ? and namespace = ? and status = ?").Prepare()

	for i := range ids {
		if operation == "stop" {
			_, err = p.Exec(false, "", ids[i], namespace, true)

		} else {
			_, err = p.Exec(true, "", ids[i], namespace, false)

		}
		if err != nil {
			klog.Errorf("exec sql failed,err: %v", err)
		}

	}
	p.Close()
	return task, err
}

// UpdateTaskJob ...
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

// GetStatisticsDataservices ...
func GetStatisticsDataservices(namespace string) (open, total, success int64, err error) {
	o := orm.NewOrm()

	qs := o.QueryTable("Task").Filter("Namespace", namespace)

	total, err = qs.Count()
	qs = qs.Filter("Status", true)
	open, err = qs.Count()

	var dagID orm.ParamsList

	num, err := o.Raw("select dag_id from task where namespace = ?", namespace).ValuesFlat(&dagID)
	if err != nil {
		klog.Errorf("exec get sql failed,err: %v", err)
		return
	}
	klog.Errorf("dagID: %v,num:%v", dagID, num)
	var list orm.ParamsList
	success, err = o.QueryTable("TbDagRun").Distinct().Filter("DagId__in", dagID).Filter("DagStatus", 0).ValuesFlat(&list, "DagId")
	klog.Errorf("list: %v, success:%v, err:%v", list, len(list), success, err)

	return
}
