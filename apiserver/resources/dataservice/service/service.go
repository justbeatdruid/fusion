package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/chinamobile/nlpt/apiserver/database/model"
	k8s "github.com/chinamobile/nlpt/apiserver/kubernetes"
	clientset "k8s.io/client-go/kubernetes"
	"k8s.io/klog"
)

var crdNamespace = "default"

// var oofsGVR = schema.GroupVersionResource{
// 	Group:    v1.GroupVersion.Group,
// 	Version:  v1.GroupVersion.Version,
// 	Resource: "dataservices",
// }

//Service ...
type Service struct {
	//client     dynamic.NamespaceableResourceInterface
	kubeClient *clientset.Clientset
}

//NewService ...
func NewService(kubeClient *clientset.Clientset) *Service {
	return &Service{
		// client:     client.Resource(oofsGVR),
		kubeClient: kubeClient,
	}
}

// func NewService(client dynamic.Interface, kubeClient *clientset.Clientset) *Service {
// 	return &Service{
// 		client:     client.Resource(oofsGVR),
// 		kubeClient: kubeClient,
// 	}
// }

//CreateDataservice ...
func (s *Service) CreateDataservice(dataService *Dataservice) (*Dataservice, error) {
	_, err := model.GetTaskByName(dataService.Name)
	if err == nil {
		return dataService, fmt.Errorf("create error err: %+v,name: %v exist", err, dataService.Name)
	}
	task := ToModel(dataService)
	dataService.DagID = task.DagId
	id, err := model.AddTask(task)
	dataService.ID = int(id)
	if err == nil {
		errJob := s.CreateJob(dataService.DagID, dataService.Name+"-"+dataService.DagID, dataService.SchedualPlanConfig.QuartzCronExpression, "registry.cmcc.com/library/smallcurl:1.0")
		klog.Errorf("errJob:%v", errJob)
	}

	klog.Errorf("DeleteCronJob errJob:%v", err)

	return ToAPI(task), err
}

//ListDataservice ...
func (s *Service) ListDataservice(offet, limit int) ([]*Dataservice, error) {
	dss, err := model.GetTasks(offet, limit)
	if err != nil {
		return nil, fmt.Errorf("cannot list object: %+v", err)
	}
	return ToListAPI(dss), nil
}

//GetDataservice ...
func (s *Service) GetDataservice(id string) (*Dataservice, error) {
	ds, err := model.GetTaskByDagId(id)
	if err != nil {
		return nil, fmt.Errorf("cannot get object: %+v", err)
	}
	return ToAPI(ds), nil
}

//DeleteDataservice ...
func (s *Service) DeleteDataservice(id string) error {
	task, err := model.DeleteTaskByDagId(id)
	if err == nil {
		errCronjob := k8s.DeleteCronJob(s.kubeClient, task.Name+"-"+task.DagId, crdNamespace)
		klog.Errorf("errJob:%v", errCronjob)
	}
	return err
}

// func (s *Service) Create(ds *v1.Dataservice) (*v1.Dataservice, error) {
// 	klog.V(5).Infof("creating dataservice: %+v", *ds)
// 	content, err := runtime.DefaultUnstructuredConverter.ToUnstructured(ds)
// 	if err != nil {
// 		return nil, fmt.Errorf("convert crd to unstructured error: %+v", err)
// 	}
// 	crd := &unstructured.Unstructured{}
// 	crd.SetUnstructuredContent(content)

// 	crd, err = s.client.Namespace(crdNamespace).Create(crd, metav1.CreateOptions{})
// 	if err != nil {
// 		return nil, fmt.Errorf("error creating crd: %+v", err)
// 	}

// 	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ds); err != nil {
// 		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
// 	}
// 	klog.V(5).Infof("get v1.dataservice of creating: %+v", ds)
// 	return ds, nil
// }

// func (s *Service) List() (*v1.DataserviceList, error) {
// 	crd, err := s.client.Namespace(crdNamespace).List(metav1.ListOptions{})
// 	if err != nil {
// 		return nil, fmt.Errorf("error list crd: %+v", err)
// 	}
// 	dss := &v1.DataserviceList{}
// 	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), dss); err != nil {
// 		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
// 	}
// 	klog.V(5).Infof("get v1.dataservicelist: %+v", dss)
// 	return dss, nil
// }

// func (s *Service) Get(id string) (*v1.Dataservice, error) {
// 	crd, err := s.client.Namespace(crdNamespace).Get(id, metav1.GetOptions{})
// 	if err != nil {
// 		return nil, fmt.Errorf("error get crd: %+v", err)
// 	}
// 	ds := &v1.Dataservice{}
// 	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(crd.UnstructuredContent(), ds); err != nil {
// 		return nil, fmt.Errorf("convert unstructured to crd error: %+v", err)
// 	}
// 	klog.V(5).Infof("get v1.dataservice: %+v", ds)
// 	return ds, nil
// }

// func (s *Service) Delete(id string) error {
// 	err := s.client.Namespace(crdNamespace).Delete(id, &metav1.DeleteOptions{})
// 	if err != nil {
// 		return fmt.Errorf("error delete crd: %+v", err)
// 	}
// 	return nil
// }

//GetCmd ...
func (s *Service) GetCmd(jobID string) ([]string, error) {
	options, err := s.GetFlinkxBody(jobID)
	if err != nil {
		return []string{}, err
	}
	url := "http://flinkx-rest:9080/flinkx/task/start?options=" + options + "&isAll=0&maxValue=100&metaId=asdfasdfasdf"
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
	return []string{"curl", url}, nil

}

//GetFlinkxBody ...
func (s *Service) GetFlinkxBody(jobID string) (string, error) {
	flinkJob := NewFlinkxReq()
	job, err := json.Marshal(&flinkJob)
	if err != nil {
		klog.Errorf("json Marshal failed,err:%v, flinkJob:%v", err, flinkJob)
		return "", err
	}
	flinkxBody := map[string]interface{}{
		"-jobid":      jobID,
		"-mode":       "local",
		"-pluginRoot": "/root/flinkx-rest/plugins/",
		"-job":        string(job),
	}
	klog.Infof("flinkxBody:%v", flinkxBody)
	flinkxBytes, err := json.Marshal(flinkxBody)
	if err != nil {
		klog.Errorf("Marshal flinkxBody failed:err:%v", err)
		return "", err
	}
	return base64.StdEncoding.EncodeToString(flinkxBytes), nil
}

//CreateJob ...
func (s *Service) CreateJob(jobID, name, schedule, imageName string) error {
	cmd, err := s.GetCmd(jobID)
	if err != nil {
		klog.Errorf("get cmd failed,err:%v", err)
		return err
	}
	err = k8s.CreateCronJob(s.kubeClient, name, schedule, imageName, jobID, crdNamespace, cmd)
	if err != nil {
		klog.Errorf("Create Job failed,err:%v", err)
		return err
	}
	return nil
}
