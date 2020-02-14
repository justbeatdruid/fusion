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

	nlptv1 "github.com/chinamobile/nlpt/crds/namespace/api/v1"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	Operator *Operator
}

// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=namespaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=namespaces/status,verbs=get;update;patch

func (r *NamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("namespace", req.NamespacedName)

	namespace := &nlptv1.Namespace{}
	if err := r.Get(ctx, req.NamespacedName, namespace); err != nil {
		klog.Errorf("cannot get namespace of ctrl req: %+v", err)

		return ctrl.Result{}, nil
	}

	if namespace.Status.Status == nlptv1.Init {
		namespace.Status.Status = nlptv1.Creating
		if err := r.Operator.CreateNamespace(namespace); err != nil {
			namespace.Status.Status = nlptv1.Error
			namespace.Status.Message = err.Error()
		} else {
			namespace.Status.Status = nlptv1.Created
			namespace.Status.Message = "success"
		}

		klog.Infof("Final Namespace: %+v", *namespace)
		if err := r.Update(ctx, namespace); err != nil {
			klog.Errorf("update namespace error: %+v", namespace)
		}
	}


	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Namespace{}).
		Complete(r)
}
