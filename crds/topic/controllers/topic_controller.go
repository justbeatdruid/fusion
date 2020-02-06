/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nlptv1 "github.com/chinamobile/nlpt/crds/topic/api/v1"
)

// TopicReconciler reconciles a Topic object
type TopicReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Operator *Operator
}

// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=topics,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=topics/status,verbs=get;update;patch

//Reconcile topic
func (r *TopicReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("topic", req.NamespacedName)

	topic := &nlptv1.Topic{}
	if err := r.Get(ctx, req.NamespacedName, topic); err != nil {
		klog.Errorf("cannot get topic of ctrl req: %+v", err)
		return ctrl.Result{}, nil
	}
	klog.Infof("get new topic event: %+v", *topic)
	klog.Infof("Status:%s", topic.Status.Status)

	if topic.Status.Status == nlptv1.Init {
		klog.Info("Current status is Init")
		topic.Status.Status = nlptv1.Creating
		if error := r.Operator.CreateTopic(topic); error != nil {
			topic.Spec.Url = topic.GetUrl(topic)
			topic.Status.Status = nlptv1.Error
			topic.Status.Message = error.Error()
		} else {
			topic.Spec.Url = topic.GetUrl(topic)
			topic.Status.Status = nlptv1.Created
			topic.Status.Message = "success"
		}

		//更新数据库的状态
		klog.Infof("Final Topic: %+v", *topic)
		r.Update(ctx, topic)

	}

	if topic.Status.Status == nlptv1.Delete {
		topic.Status.Status = nlptv1.Deleting
		if error := r.Operator.DeleteTopic(topic); error != nil {
			topic.Status.Status = nlptv1.Error
			topic.Status.Message = error.Error()
		} else {
			//更新数据库的状态
			r.Delete(ctx, topic)
		}

	}
	/* for example
	topic = r.Get()
	if topic.Status == "init" {
		topicID, err := kafka.Create(topic)
		if err != nil {
			topic.Status = "error"
		} else {
			topic.TopicID = topicID
			topic.Status = "success"
		}
		r.Update(topic)
	}
	*/
	// your logic here

	return ctrl.Result{}, nil
}

//SetupWithManager test
func (r *TopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Topic{}).
		Complete(r)
}
