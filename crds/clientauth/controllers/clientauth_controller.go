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

	nlptv1 "github.com/chinamobile/nlpt/crds/clientauth/api/v1"
)

// ClientauthReconciler reconciles a Clientauth object
type ClientauthReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=clientauths,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=clientauths/status,verbs=get;update;patch

func (r *ClientauthReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("clientauth", req.NamespacedName)

	// your logic here
	clientAuth := &nlptv1.Clientauth{}
	if err := r.Get(ctx, req.NamespacedName, clientAuth); err != nil{
		klog.Errorf("cannot get clientauth of ctrl req: %+v", err)
		return ctrl.Result{}, nil
	}
	if clientAuth.Status.Status == nlptv1.Init {
		clientAuth.Status.Status = nlptv1.Created
		clientAuth.Status.Message = "success"
	}

	r.Update(ctx, clientAuth)
	return ctrl.Result{}, nil
}

func (r *ClientauthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Clientauth{}).
		Complete(r)
}
