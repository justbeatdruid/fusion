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

	trafficcontrl := &nlptv1.Trafficcontrol{}
	if err := r.Get(ctx, req.NamespacedName, trafficcontrl); err != nil {
		klog.Errorf("cannot get trafficcontrol of ctrl req: %+v", err)
		return ctrl.Result{}, nil
	}
	klog.Infof("get new trafficcontrol event: %+v", *trafficcontrl)

	if trafficcontrl.Status.Status == nlptv1.Bind {
		trafficcontrl.Status.Status = nlptv1.Binding
		klog.Infof("trafficcontrol is binding")
		if err := r.Operator.AddRouteRatelimitByKong(trafficcontrl); err != nil {
			klog.Infof("trafficcontrol bind err")
			trafficcontrl.Status.Status = nlptv1.Error
			trafficcontrl.Status.Message = err.Error()
		} else {
			klog.Infof("trafficcontrol bind sunccess")
			trafficcontrl.Status.Status = nlptv1.Binded
			trafficcontrl.Status.Message = "success"
		}
		r.Update(ctx, trafficcontrl)
	}

	return ctrl.Result{}, nil
}

func (r *TrafficcontrolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Trafficcontrol{}).
		Complete(r)
}
