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
	"k8s.io/klog"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	nlptv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
)

// ServiceunitReconciler reconciles a Serviceunit object
type ServiceunitReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	//初始化kong信息
	Operator *Operator
}

// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=serviceunits,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=serviceunits/status,verbs=get;update;patch

func (r *ServiceunitReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("serviceunit", req.NamespacedName)

	//
	serviceunit := &nlptv1.Serviceunit{}
	if err := r.Get(ctx, req.NamespacedName, serviceunit); err != nil {
		klog.Errorf("cannot get serviceunit of ctrl req: %+v", err)
		return ctrl.Result{}, nil
	}
	klog.Infof("get new serviceunit event: %+v", *serviceunit)
	if serviceunit.Status.Status == nlptv1.Init {
		// call kong api create
		serviceunit.Status.Status = nlptv1.Creating
		if err := r.Operator.CreateServiceByKong(serviceunit); err != nil {
			serviceunit.Spec.Result = nlptv1.CREATEFAILED
			serviceunit.Status.Status = nlptv1.Error
			serviceunit.Status.Message = err.Error()
		} else {
			serviceunit.Spec.Result = nlptv1.CREATESUCCESS
			serviceunit.Status.Status = nlptv1.Created
			serviceunit.Status.Message = "success"
		}
		r.Update(ctx, serviceunit)
	}
	if serviceunit.Status.Status == nlptv1.Delete {
		// call kong api delete
		serviceunit.Status.Status = nlptv1.Deleting
		if err := r.Operator.DeleteServiceByKong(serviceunit); err != nil {
			serviceunit.Spec.Result = nlptv1.DELETEFAILED
			serviceunit.Status.Status = nlptv1.Error
			serviceunit.Status.Message = err.Error()
			klog.Infof("delete service by kong failed: %s", serviceunit.Status.Message)
			r.Update(ctx, serviceunit)
		} else {
			r.Delete(ctx, serviceunit)
		}
	}

	// + update_sunyu
	if serviceunit.Status.Status == nlptv1.Update {
		klog.Infof("serviceunit update: %+v", *serviceunit)
		// call kong api update
		serviceunit.Status.Status = nlptv1.Updating
		klog.Infof("service unit updatinging: %+v", *serviceunit)
		id := serviceunit.Spec.KongService.ID
		if len(id) == 0 {
			klog.Infof("service unit update need create")
			if err := r.Operator.CreateServiceByKong(serviceunit); err != nil {
				serviceunit.Spec.Result = nlptv1.UPDATEFAILED
				klog.Infof("serviceunit update error: %+v", *serviceunit)
				serviceunit.Status.Status = nlptv1.Error
				serviceunit.Status.Message = err.Error()
				r.Update(ctx, serviceunit)
				return ctrl.Result{}, nil
			}
		} else {
			if err := r.Operator.UpdateServiceByKong(serviceunit); err != nil {
				serviceunit.Spec.Result = nlptv1.UPDATEFAILED
				klog.Infof("serviceunit update error: %+v", *serviceunit)
				serviceunit.Status.Status = nlptv1.Error
				serviceunit.Status.Message = err.Error()
				r.Update(ctx, serviceunit)
				return ctrl.Result{}, nil
			}
		}
		serviceunit.Spec.Result = nlptv1.UPDATESUCCESS
		klog.Infof("serviceunit updateded: %+v", *serviceunit)
		serviceunit.Status.Status = nlptv1.Updated
		serviceunit.Status.Message = "success"
		r.Update(ctx, serviceunit)
	}

	//TODO 后续处理异常 或者更新状态
	/*if serviceunit.Status.Status == nlptv1.Error {
		// call kong api create
		serviceunit.Status.Status = nlptv1.Creating
		if err := r.Operator.CreateServiceByKong(serviceunit); err != nil {
			serviceunit.Status.Status = nlptv1.Error
			serviceunit.Status.Message = err.Error()
		} else {
			serviceunit.Status.Status = nlptv1.Created
		}
		r.Update(ctx, serviceunit)
	}*/

	// your logic here

	return ctrl.Result{}, nil
}

func (r *ServiceunitReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Serviceunit{}).
		Complete(r)
}
