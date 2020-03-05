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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"k8s.io/klog"
	nlptv1 "github.com/chinamobile/nlpt/crds/restriction/api/v1"
)

// RestrictionReconciler reconciles a Restriction object
type RestrictionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Operator *Operator
}

// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=restrictions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=restrictions/status,verbs=get;update;patch

func (r *RestrictionReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("restriction", req.NamespacedName)

	restriction := &nlptv1.Restriction{}
	if err := r.Get(ctx, req.NamespacedName, restriction); err != nil {
		klog.Errorf("cannot get restriction of ctrl req: %+v", err)
		return ctrl.Result{}, nil
	}
	klog.Infof("get new restriction event: %+v", *restriction)

	if restriction.Status.Status == nlptv1.Bind {
		restriction.Status.Status = nlptv1.Binding
		klog.Infof("restriction is binding")
		if restriction.Spec.Type == nlptv1.IP {
			if err := r.Operator.AddRestrictionByKong(restriction); err != nil {
				klog.Infof("restriction bind err")
				restriction.Status.Status = nlptv1.Error
				restriction.Status.Message = err.Error()
			} else {
				klog.Infof("restriction bind sunccess")
				restriction.Status.Status = nlptv1.Binded
				restriction.Status.Message = "success"
			}
		} else if restriction.Spec.Type == nlptv1.USER {
			//ToDo
		}
		// update status
		r.Update(ctx, restriction)
	}

	if restriction.Status.Status == nlptv1.UnBind {
		restriction.Status.Status = nlptv1.UnBinding
		//r.Update(ctx, trafficcontrol)
		klog.Infof("restriction is unbinding")
		if restriction.Spec.Type == nlptv1.IP {
			if err := r.Operator.DeleteRestrictionByKong(restriction); err != nil {
				klog.Infof("restriction unbind err")
				restriction.Status.Status = nlptv1.Error
				restriction.Status.Message = err.Error()
			} else {
				klog.Infof("restriction unbind sunccess")
				restriction.Status.Status = nlptv1.UnBinded
				restriction.Status.Message = "success"
			}
		} else  if  restriction.Spec.Type == nlptv1.USER {
			//ToDo
		}
		// update status
		r.Update(ctx, restriction)
	}
	return ctrl.Result{}, nil
}

func (r *RestrictionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Restriction{}).
		Complete(r)
}
