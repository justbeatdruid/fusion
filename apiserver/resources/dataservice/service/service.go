package service

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/chinamobile/nlpt/apiserver/concurrency"
	"github.com/chinamobile/nlpt/apiserver/database/model"
	k8s "github.com/chinamobile/nlpt/apiserver/kubernetes"
	dsv1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"
	"github.com/chinamobile/nlpt/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var crdNamespace = "default"

//Service ...
type Service struct {
	dsClient   dynamic.NamespaceableResourceInterface
	kubeClient *clientset.Clientset
	elector    concurrency.Elector
}

//NewService ...
func NewService(kubeClient *clientset.Clientset, client dynamic.Interface, elector concurrency.Elector) *Service {
	service := &Service{
		dsClient:   client.Resource(dsv1.GetOOFSGVR()),
		kubeClient: kubeClient,
		elector:    elector,
	}

	go elector.Campaign("data-intergration", service.dealIntegrationTask)

	return service
}

//CreateDataservice ...
func (s *Service) CreateDataservice(dataService *Dataservice, userId, namespace string) (*Dataservice, error) {
	_, err := model.GetTaskByName(dataService.Name, userId, namespace)
	if err == nil {
		return dataService, fmt.Errorf("name duplicated")
	}
	err = dataService.Validate(s, util.WithUser(userId), util.WithNamespace(namespace))
	if err != nil {
		return dataService, err
	}

	task := ToModel(dataService)
	task.DagId = names.NewID()
	dataService.DagID = task.DagId
	task.Namespace = namespace
	task.UserId = userId
	task.Status = false
	dataService.CreatedAt = task.CreatedTime.Format(TimeStr)
	id, err := model.AddTask(task)
	if err != nil {
		return nil, err

	}
	dataService.ID = int(id)

	// if err == nil {
	// 	errJob := s.CreateJob(dataService.DagID, dataService.Name+"-"+dataService.DagID, dataService.SchedualPlanConfig.QuartzCronExpression, "registry.cmcc.com/library/smallcurl:1.0")
	// 	klog.Errorf("errJob:%v", errJob)
	// }

	// klog.Errorf("DeleteCronJob errJob:%v", err)

	return dataService, err
}

//CreateDataservice ...
func (s *Service) OperationDataservice(operationReq *OperationReq, userId, namespace string) error {

	if err := operationReq.Validate(); err != nil {
		return err
	}
	task, err := model.OperationTaskStatus(operationReq.Operation, userId, namespace, operationReq.DagID)

	if operationReq.Operation == "delete" || operationReq.Operation == "stop" {
		go func(task []model.Task) {
			for i := range task {
				if errCronjob := k8s.DeleteCronJob(s.kubeClient, task[i].Job, crdNamespace); errCronjob != nil {
					klog.Errorf("errJob:%v, task:%v", errCronjob, task[i])
				}
			}
		}(task)

	}
	return err
}

//ListDataservice ...
func (s *Service) ListDataservice(offet, limit int, name, userId, namespace string) (interface{}, error) {
	dss, num, total, err := model.GetTasks((offet-1)*limit, limit, name, userId, namespace)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	body := map[string]interface{}{}

	body["items"] = ToListAPI(dss, s, util.WithUser(userId), util.WithNamespace(namespace))
	body["count"] = num
	body["total"] = total
	body["page"] = offet
	body["limit"] = limit

	return body, nil
}
func (s *Service) GetDataSource(id string, opts ...util.OpOption) (*dsv1.Datasource, string, error) {
	crdNamespace := util.OpList(opts...).Namespace()
	if len(crdNamespace) == 0 {
		crdNamespace = "default"
	}
	crd, err := s.dsClient.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
	if err != nil {
		return nil, "", fmt.Errorf("error get crd: %+v", err)
	}
	ds := &dsv1.Datasource{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ds); err != nil {
		return nil, "", fmt.Errorf("convert unstructured to crd error: %+v", err)
	}
	klog.V(5).Infof("get v1.datasource: %+v", ds)
	return ds, ds.Spec.Name, nil

}

//GetDataservice ...
func (s *Service) GetDataservice(id, userId, namespace string) (*Dataservice, error) {
	ds, err := model.GetTaskByDagId(id, userId, namespace)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}

	return ToAPI(ds, s, util.WithUser(userId), util.WithNamespace(namespace)), nil
}

//DeleteDataservice ...
func (s *Service) DeleteDataservice(id, userId, namespace string) error {
	task, err := model.DeleteTaskByDagId(id, userId, namespace)
	if err == nil {
		errCronjob := k8s.DeleteCronJob(s.kubeClient, task.Job, crdNamespace)
		klog.Errorf("errJob:%v", errCronjob)
	}
	return err
}

//UpdateDateService ...
func (s *Service) UpdateDateService(reqData map[string]interface{}, dagId, userId, namespace string) (*Dataservice, error) {
	taskdb, err := model.GetTaskByDagId(dagId, userId, namespace)
	if err != nil {
		klog.Errorf("get error:%v", err)
		return nil, fmt.Errorf("not found")
	}
	if taskdb.Status {
		klog.Errorf("task's status (%v) is running, please stop task first. :%v", taskdb.Status)
		return nil, fmt.Errorf("task's status is running, please stop task first")
	}
	if err = s.assignment(taskdb, reqData); err != nil {
		return nil, err
	}

	if _, ok := reqData["Name"]; ok {
		_, err = model.GetTaskByName(taskdb.Name, userId, namespace)
		if err == nil {
			return nil, fmt.Errorf("name duplicated")
		}
	}
	api := ToAPI(taskdb, s, util.WithUser(userId), util.WithNamespace(namespace))
	err = api.Validate(s, util.WithUser(userId), util.WithNamespace(namespace))

	if err != nil {
		return nil, err
	}

	if err = model.UpdateTaskByID(taskdb); err != nil {
		return nil, err
	}
	return api, err
}

//GetCmd ...
func (s *Service) GetCmd(task model.Task) ([]string, error) {
	options, err := s.GetFlinkxBody(task)
	if err != nil {
		return []string{}, err
	}
	//url := "http://flinkx-rest:9080/flinkx/task/start?options=" + options + "&isAll=0&maxValue=100&metaId=asdfasdfasdf"
	//curl localhost:9080/api/v1/apis -H 'content-type:application/json'  \
	//-d'
	// {
	//   "options": ""
	// }'
	// body := map[string]string{
	// 	"options": options,
	// }
	// byetes, _ := json.Marshal(body)

	//return []string{"curl", url, "-H", "'content-type:application/json'"}, nil
	// return []string{"curl", url}, nil
	config, err := json.Marshal(task)
	if err != nil {
		klog.Errorf("json Marshal failed,err:%v, task:%v", err, task)
		//return "", err
	}
	return []string{"./curlflinkx", options, string(config)}, nil

}

//GetFlinkxBody ...
func (s *Service) GetFlinkxBody(ds model.Task) (string, error) {
	flinkJob := NewFlinkxReq()

	// ds, err := model.GetTaskByDagId(jobID, userId, namespace)
	// if err != nil {
	// 	return "", fmt.Errorf("cannot get object: %+v", err)
	// }
	dataSource := DataSource{}
	err := json.Unmarshal([]byte(ds.DataSourceConfig), &dataSource)
	if err != nil {
		klog.Errorf("unmarshal failed,err:%v, config:%v", err, ds.DataSourceConfig)
		return "", err
	}
	sourceDB, _, err := s.GetDataSource(dataSource.RelationalDb.SourceID, util.WithUser(ds.UserId), util.WithNamespace(ds.Namespace))
	if err != nil {
		klog.Errorf("get source failed,err:%v", err)
		return "", err
	}

	dataTarget := DataTarget{}
	err = json.Unmarshal([]byte(ds.DataTargetConfig), &dataTarget)
	if err != nil {
		klog.Errorf("unmarshal failed,err:%v, config:%v", err, ds.DataTargetConfig)
		return "", err
	}
	targetDB, _, err := s.GetDataSource(dataTarget.RelationalDbTarget.TargetID, util.WithUser(ds.UserId), util.WithNamespace(ds.Namespace))
	if err != nil {
		klog.Errorf("get source failed,err:%v", err)
		return "", err
	}

	flinkJob.Job.Content[0].Reader.Parameter.UserName = sourceDB.Spec.RDB.Connect.Username
	flinkJob.Job.Content[0].Reader.Parameter.Password = sourceDB.Spec.RDB.Connect.Password
	flinkJob.Job.Content[0].Reader.Parameter.Connection[0].Table = []string{dataSource.RelationalDb.SourceTable}

	flinkJob.Job.Content[0].Reader.Parameter.Connection[0].JdbcURL = []string{"jdbc:mysql://" + sourceDB.Spec.RDB.Connect.Host + ":" + strconv.Itoa(sourceDB.Spec.RDB.Connect.Port) + "/" + sourceDB.Spec.RDB.Database}
	flinkJob.Job.Content[0].Writer.Parameter.UserName = targetDB.Spec.RDB.Connect.Username
	flinkJob.Job.Content[0].Writer.Parameter.Password = targetDB.Spec.RDB.Connect.Password

	flinkJob.Job.Content[0].Writer.Parameter.Connection[0].Table = []string{dataTarget.RelationalDbTarget.TargetTable}

	flinkJob.Job.Content[0].Writer.Parameter.Connection[0].JdbcURL = []string{"jdbc:mysql://" + targetDB.Spec.RDB.Connect.Host + ":" + strconv.Itoa(targetDB.Spec.RDB.Connect.Port) + "/" + targetDB.Spec.RDB.Database}

	for _, v := range dataTarget.RelationalDbTarget.MappingRelation {
		flinkJob.Job.Content[0].Reader.Parameter.Column = append(flinkJob.Job.Content[0].Reader.Parameter.Column, v.SourceField)
		flinkJob.Job.Content[0].Writer.Parameter.Column = append(flinkJob.Job.Content[0].Writer.Parameter.Column, v.SourceField)

	}
	klog.Infof("flinkJob:%v", flinkJob)
	job, err := json.Marshal(&flinkJob)
	if err != nil {
		klog.Errorf("json Marshal failed,err:%v, flinkJob:%v", err, flinkJob)
		return "", err
	}

	flinkxBody := map[string]string{
		"-jobid":      ds.DagId,
		"-mode":       "local",
		"-pluginRoot": "/root/flinkx-rest/plugins/",
		"-job":        string(job),
		//	"-config":     string(config),
	}
	klog.Infof("flinkxBody:%v", flinkxBody)
	flinkxBytes, err := json.Marshal(flinkxBody)
	if err != nil {
		klog.Errorf("Marshal flinkxBody failed:err:%v", err)
		return "", err
	}
	//return base64.StdEncoding.EncodeToString(flinkxBytes), nil
	return string(flinkxBytes), nil
}

//CreateJob ...
func (s *Service) CreateJob(imageName string, task model.Task) error {
	cmd, err := s.GetCmd(task)
	if err != nil {
		klog.Errorf("get cmd failed,err:%v", err)
		return err
	}
	schedule, err := s.getSchedule(task)
	if err != nil {
		klog.Errorf("get schedual failed,err:%v", err)
		return err
	}
	err = k8s.CreateCronJob(s.kubeClient, task.Name+"-"+task.DagId, schedule, imageName, task.DagId, crdNamespace, cmd)
	if err != nil {
		klog.Errorf("Create Job failed,err:%v", err)
		return err
	}
	err = model.UpdateTaskJob(task.DagId, task.Name+"-"+task.DagId)
	if err != nil {
		klog.Errorf("update Job failed,err:%v", err)
	}

	return err
}

func (s *Service) getSchedule(task model.Task) (string, error) {
	if task.Type == "realtime" {
		return "*/1 * * * *", nil
	}
	config := SchedualPlan{}
	if err := json.Unmarshal([]byte(task.SchedualPlan), &config); err != nil {
		return "", err
	}
	if config.QuartzCron {
		return config.QuartzCronExpression, nil
	}
	if config.TimeUnit == "minute" {
		return "*/" + strconv.Itoa(config.SchedualPeriod) + " * * * *", nil
	}
	if config.TimeUnit == "hour" {
		return "* */" + strconv.Itoa(config.SchedualPeriod) + " * * *", nil
	}
	if config.TimeUnit == "day" {
		return "* * */" + strconv.Itoa(config.SchedualPeriod) + " * *", nil
	}
	if config.TimeUnit == "month" {
		return "* * * */" + strconv.Itoa(config.SchedualPeriod) + " *", nil
	}
	return "* * * * */" + strconv.Itoa(config.SchedualPeriod), nil

}

func (s *Service) dealIntegrationTask() {
	for {
		time.Sleep(30 * time.Second)
		task, err := model.GetTaskByStartTime()
		if err != nil {
			klog.Errorf("get task falied  Job failed,err:%v", err)
			continue
		}
		for i := range task {
			err = s.CreateJob("registry.cmcc.com/library/smallcurl:1.0", task[i])
			if err != nil {
				klog.Errorf("errJob:%v", err)
				continue
			}
		}
	}
	return
}
