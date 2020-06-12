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

	nlptv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
)

// ApplicationReconciler reconciles a Application object
type ApplicationReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	//初始化kong信息
	Operator *Operator
}

// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=applications,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=applications/status,verbs=get;update;patch

func (r *ApplicationReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("application", req.NamespacedName)

	//
	application := &nlptv1.Application{}
	if err := r.Get(ctx, req.NamespacedName, application); err != nil {
		klog.Errorf("cannot get application of ctrl req: %+v", err)
		return ctrl.Result{}, nil
	}
	klog.Infof("get new application event: %+v", *application)

	//add topic token
	if application.Status.Status == nlptv1.Created && len(application.Spec.TopicAuth.Token) == 0 {
		klog.Infof("need add topic token : %+v", *application)
		token, err := r.Operator.CreateTopicToken(application.ObjectMeta.Name)
		if err != nil {
			klog.Errorf("cannot create topic token : %+v", err)
		}
		klog.Infof("get topic token : %s", token)
		application.Spec.TopicAuth.Token = token
		r.Update(ctx, application)
		return ctrl.Result{}, nil
	}

	// resume consumer
	if application.Status.Status == nlptv1.Created && len(application.Spec.ConsumerInfo.Token) != 0 {
		klog.Infof("begin to check consumer info: %+v", *application)
		err := r.Operator.ResumeConsumerInfoFromKong(application)
		if err != nil {
			klog.Errorf("cannot resume consumer: %+v", err)
		}
		r.Update(ctx, application)
		return ctrl.Result{}, nil
	}

	// your logic here
	if application.Status.Status == nlptv1.Init {
		// call kong api create
		if err := r.Operator.CreateConsumerByKong(application); err != nil {
			application.Spec.Result = nlptv1.CREATEFAILED
			application.Status.Status = nlptv1.Error
			application.Status.Message = err.Error()
			r.Update(ctx, application)
			return ctrl.Result{}, nil
		}
		if err := r.Operator.CreateConsumerCredentials(application); err != nil {
			application.Spec.Result = nlptv1.CREATEFAILED
			application.Status.Status = nlptv1.Error
			application.Status.Message = err.Error()
			r.Update(ctx, application)
			return ctrl.Result{}, nil
		}
		if err := CreateToken(application); err != nil {
			application.Spec.Result = nlptv1.CREATEFAILED
			application.Status.Status = nlptv1.Error
			application.Status.Message = err.Error()
			r.Update(ctx, application)
			return ctrl.Result{}, nil
		}
		application.Spec.Result = nlptv1.CREATESUCCESS
		application.Status.Status = nlptv1.Created
		application.Status.Message = "success"
		r.Update(ctx, application)
		return ctrl.Result{}, nil

	}
	if application.Status.Status == nlptv1.Delete {
		// call kong api delete
		if err := r.Operator.DeleteConsumerByKong(application); err != nil {
			application.Spec.Result = nlptv1.DELETEFAILED
			application.Status.Status = nlptv1.Error
			application.Status.Message = err.Error()
			r.Update(ctx, application)
			return ctrl.Result{}, nil
		}
		r.Delete(ctx, application)
	}

	return ctrl.Result{}, nil
}

func (r *ApplicationReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Application{}).
		Complete(r)
}
