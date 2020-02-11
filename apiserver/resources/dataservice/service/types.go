package service

import (
	"fmt"
	"time"

	"github.com/chinamobile/nlpt/crds/dataservice/api/v1"
	"github.com/chinamobile/nlpt/pkg/names"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Dataservice struct {
	ID         string        `json:"id"`
	Namespace  string        `json:"namespace"`
	Type       v1.TaskType   `json:"type"`
	CreatedAt  time.Time     `json:"createdAt"`
	CronConfig string        `json:"cronConfig"`
	Status     v1.TaskStatus `json:"status"`
	StartedAt  time.Time     `json:"startedAt"`
	StoppedAt  time.Time     `json:"stoppedAt"`
}

// only used in creation options
func ToAPI(ds *Dataservice) *v1.Dataservice {
	crd := &v1.Dataservice{}
	crd.TypeMeta.Kind = "Dataservice"
	crd.TypeMeta.APIVersion = v1.GroupVersion.Group + "/" + v1.GroupVersion.Version

	crd.ObjectMeta.Name = ds.ID
	crd.ObjectMeta.Namespace = crdNamespace
	crd.Spec = v1.DataserviceSpec{
		Task: v1.Task{
			Type: ds.Type,
		},
	}
	switch ds.Type {
	case v1.Realtime:
		crd.Spec.Task.RealtimeTask = &v1.RealtimeTask{}
	case v1.Periodic:
		crd.Spec.Task.PeriodicTask = &v1.PeriodicTask{
			CronConfig: ds.CronConfig,
		}
	}
	crd.Status = v1.DataserviceStatus{
		Status:    v1.Created,
		StartedAt: metav1.Unix(0, 0),
		StoppedAt: metav1.Unix(0, 0),
	}
	return crd
}

func ToModel(obj *v1.Dataservice) *Dataservice {
	ds := &Dataservice{
		ID:        obj.ObjectMeta.Name,
		Namespace: obj.ObjectMeta.Namespace,
		CreatedAt: obj.ObjectMeta.CreationTimestamp.Time,
		Status:    obj.Status.Status,
		Type:      obj.Spec.Task.Type,
		StartedAt: obj.Status.StartedAt.Time,
		StoppedAt: obj.Status.StoppedAt.Time,
	}
	switch ds.Type {
	case v1.Realtime:
	case v1.Periodic:
		ds.CronConfig = obj.Spec.Task.PeriodicTask.CronConfig
	}
	return ds
}

func ToListModel(items *v1.DataserviceList) []*Dataservice {
	var ds []*Dataservice = make([]*Dataservice, len(items.Items))
	for i := range items.Items {
		ds[i] = ToModel(&items.Items[i])
	}
	return ds
}

func (ds *Dataservice) Validate() error {
	for k, v := range map[string]string{
		"type": string(ds.Type),
	} {
		if len(v) == 0 {
			return fmt.Errorf("%s is null", k)
		}
	}
	switch ds.Type {
	case v1.Realtime:
	case v1.Periodic:
		if err := v1.ValidateCronConfig(ds.CronConfig); err != nil {
			return fmt.Errorf("cron config error: %+v", err)
		}
	default:
		return fmt.Errorf("wrong task type: %s", ds.Type)
	}
	ds.ID = names.NewID()
	return nil
}
