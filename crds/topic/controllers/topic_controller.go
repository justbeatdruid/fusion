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
	"fmt"
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
	Operator *Connector
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
	//klog.Infof("get new topic event: %+v", *topic)
	//klog.Infof("Status:%s", topic.Status.Status)

	if topic.Status.Status == nlptv1.Init {
		//klog.Info("Current status is Init")
		topic.Status.Status = nlptv1.Creating
		if err := r.Operator.CreateTopic(topic); err != nil {
			topic.Spec.Url = topic.GetUrl()
			topic.Status.Status = nlptv1.Error
			topic.Status.Message = fmt.Sprintf("create topic error:%+v", err)
			klog.Errorf("create topic failed, err: %+v", err)
		} else {
			topic.Spec.Url = topic.GetUrl()
			topic.Status.Status = nlptv1.Created
			topic.Status.Message = "success"
		}

		//更新数据库的状态
		//klog.Infof("Final Topic: %+v", *topic)
		if err := r.Update(ctx, topic); err != nil {
			klog.Errorf("Update Topic Failed: %+v", *topic)
		}

	}

	if topic.Status.Status == nlptv1.Delete {
		topic.Status.Status = nlptv1.Deleting
		if err := r.Operator.DeleteTopic(topic); err != nil {
			topic.Status.Status = nlptv1.Error
			topic.Status.Message = fmt.Sprintf("delete topic error: %+v", err)
			r.Update(ctx, topic)
		} else {
			//更新数据库的状态
			r.Delete(ctx, topic)
		}

	}

	if topic.Status.Status == nlptv1.Update {
		topic.Status.Status = nlptv1.Updating
		//更新数据库的状态
		klog.Infof("Start Grant Topic: %+v", *topic)
		//授权操作
		for _, p := range topic.Spec.Permissions {
			if p.Status.Status == nlptv1.Grant {
				if err := r.Operator.GrantPermission(topic, &p); err != nil {
					p.Status.Status = "error"
					p.Status.Message = fmt.Sprintf("grant permission error: %+v", err)
				} else {
					p.Status.Status = nlptv1.Granted
					p.Status.Message = "success"
				}

			}
		}

		//删除授权
		for index, p := range topic.Spec.Permissions {
			if p.Status.Status == "delete" {
				p.Status.Status = "deleting"
				if err := r.Operator.DeletePer(topic, &p); err != nil {
					p.Status.Status = "error"
					p.Status.Message = fmt.Sprintf("revoke permission error: %+v", err)

					//删除失败，将标签重置为true
					topic.ObjectMeta.Labels[p.AuthUserID] = "true"
					r.Update(ctx, topic)
				} else {
					pers := topic.Spec.Permissions
					topic.Spec.Permissions = append(pers[:index], pers[index+1:]...)

					//收回权限成功，删除标签
					delete(topic.ObjectMeta.Labels, p.AuthUserID)
				}
			}
		}
		topic.Status.Status = nlptv1.Updated
		//更新数据库的状态
		klog.Infof("Final Topic: %+v", *topic)
		r.Update(ctx, topic)
	}
	return ctrl.Result{}, nil
}

//SetupWithManager test
func (r *TopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Topic{}).
		Complete(r)
}
