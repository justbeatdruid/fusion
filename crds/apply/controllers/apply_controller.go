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
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	appv1 "github.com/chinamobile/nlpt/crds/application/api/v1"
	nlptv1 "github.com/chinamobile/nlpt/crds/apply/api/v1"
)

// ApplyReconciler reconciles a Apply object
type ApplyReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Operator *Operator
}

// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=applies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=applies/status,verbs=get;update;patch

func (r *ApplyReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("apply", req.NamespacedName)

	apply := &nlptv1.Apply{}
	klog.V(4).Infof("get apply %s", req.NamespacedName)
	if err := r.Get(ctx, req.NamespacedName, apply); err != nil {
		klog.Errorf("cannot get apply of ctrl req: %+v", err)
		return ctrl.Result{}, nil
	}
	switch apply.Status.Status {
	case nlptv1.Admited:
		if apply.Spec.TargetType == nlptv1.Api && apply.Spec.SourceType == nlptv1.Application {
			if apply.Status.OperationDone {
				return ctrl.Result{}, nil
			} else if apply.Status.Retry <= 0 {
				return ctrl.Result{}, nil
			}

			appID := apply.Spec.SourceID
			api := &apiv1.Api{}
			apiNamespacedName := types.NamespacedName{
				Namespace: req.NamespacedName.Namespace,
				Name:      apply.Spec.TargetID,
			}

			app := &appv1.Application{}
			appNamespacedName := types.NamespacedName{
				Namespace: req.NamespacedName.Namespace,
				Name:      apply.Spec.SourceID,
			}

			if err := r.Get(ctx, apiNamespacedName, api); err != nil {
				klog.Errorf("cannot get api with name %s: %+v", apiNamespacedName, err)
				//TODO always retry if get api error?
				//r.UpdateApply(ctx, apply, nlptv1.Error, "api not found")
				goto failed
			}
			klog.V(4).Infof("get api %+v", *api)
			if err := r.Get(ctx, appNamespacedName, app); err != nil {
				klog.Errorf("cannot get api with name %s: %+v", appNamespacedName, err)
				r.UpdateApi(ctx, api, appID, false, "application not found")
				goto failed
			}

			// check if application already bound to api
			for _, existedapp := range api.Spec.Applications {
				if existedapp.ID == apply.Spec.SourceID {
					goto succeeded
				}
			}

			//application add acl whitelist (api id)
			aclId, err := r.Operator.AddConsumerToAcl(apply, api);
			if err != nil {
				r.UpdateApi(ctx, api, appID, false, "add consumer to acl error")
				goto failed
			}
			// bind api to application
			api.Spec.Applications = append(api.Spec.Applications, apiv1.Application{
				ID: app.ObjectMeta.Name, AclID: aclId,
			})
			api.ObjectMeta.Labels[apiv1.ApplicationLabel(apply.Spec.SourceID)] = "true"
			if err := r.UpdateApi(ctx, api, appID, true, ""); err != nil {
				goto failed
			}

		succeeded:
			apply.Status.OperationDone = true
		failed:
			apply.Status.Retry = apply.Status.Retry - 1
			if err := r.Update(ctx, apply); err != nil {
				klog.Errorf("cannot update apply %s: %+v", apiNamespacedName, err)
			}
			return ctrl.Result{}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *ApplyReconciler) UpdateApi(ctx context.Context, api *apiv1.Api, appid string, succeeded bool, message string) error {
	api.Status.Applications[appid] = apiv1.ApiApplicationStatus{
		BindingSucceeded: succeeded,
		Message:          message,
	}
	if err := r.Update(ctx, api); err != nil {
		klog.Errorf("update api error: %+v", err)
		return err
	}
	return nil
}

func (r *ApplyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Apply{}).
		Complete(r)
}
