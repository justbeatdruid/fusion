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

	nlptv1 "github.com/chinamobile/nlpt/crds/trafficcontrol/api/v1"
)

// TrafficcontrolReconciler reconciles a Trafficcontrol object
type TrafficcontrolReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Operator *Operator
}

// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=trafficcontrols,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=trafficcontrols/status,verbs=get;update;patch

func (r *TrafficcontrolReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("trafficcontrol", req.NamespacedName)

	trafficcontrol := &nlptv1.Trafficcontrol{}
	if err := r.Get(ctx, req.NamespacedName, trafficcontrol); err != nil {
		klog.Errorf("cannot get trafficcontrol of ctrl req: %+v", err)
		return ctrl.Result{}, nil
	}
	klog.Infof("get new trafficcontrol event: %+v", *trafficcontrol)

	if trafficcontrol.Status.Status == nlptv1.Bind {
		trafficcontrol.Status.Status = nlptv1.Binding
		klog.Infof("trafficcontrol is binding")
		if err := r.Operator.AddRouteRatelimitByKong(trafficcontrol); err != nil {
			klog.Infof("trafficcontrol bind err")
			trafficcontrol.Status.Status = nlptv1.Error
			trafficcontrol.Status.Message = err.Error()
		} else {
			klog.Infof("trafficcontrol bind sunccess")
			trafficcontrol.Status.Status = nlptv1.Binded
			trafficcontrol.Status.Message = "success"
		}
		r.Update(ctx, trafficcontrol)
	}

	if trafficcontrol.Status.Status == nlptv1.UnBind {
		trafficcontrol.Status.Status = nlptv1.UnBinding
		klog.Infof("trafficcontrol is unbinding")
		if err := r.Operator.DeleteRouteLimitByKong(trafficcontrol); err != nil {
			klog.Infof("trafficcontrol unbind err")
			trafficcontrol.Status.Status = nlptv1.Error
			trafficcontrol.Status.Message = err.Error()
		} else {
			klog.Infof("trafficcontrol unbind sunccess")
			trafficcontrol.Status.Status = nlptv1.UnBinded
			trafficcontrol.Status.Message = "success"
		}
	}
	return ctrl.Result{}, nil
}

func (r *TrafficcontrolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Trafficcontrol{}).
		Complete(r)
}
