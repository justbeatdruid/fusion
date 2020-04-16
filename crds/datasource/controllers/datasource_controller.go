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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/chinamobile/nlpt/apiserver/resources/datasource/rdb/driver"
	nlptv1 "github.com/chinamobile/nlpt/crds/datasource/api/v1"
)

var defaultNamespace = "default"

// DatasourceReconciler reconciles a Datasource object
type DatasourceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=datasources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=nlpt.cmcc.com,resources=datasources/status,verbs=get;update;patch

func (r *DatasourceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("datasource", req.NamespacedName)
	ds := &nlptv1.Datasource{}
	if err := r.Get(ctx, req.NamespacedName, ds); err != nil {
		logger.Error(err, "cannot get datasource")
	}
	if ds.Spec.Type == nlptv1.RDBType {
		lastStatus := ds.Status.Status
		if err := driver.PingRDB(ds); err != nil {
			ds.Status.Status = nlptv1.Abnormal
			ds.Status.Detail = err.Error()
		} else {
			ds.Status.Status = nlptv1.Normal
			ds.Status.Detail = ""
		}
		if ds.Status.Status != lastStatus {
			ds.Status.UpdatedAt = metav1.Now()
		}
		if err := r.Update(ctx, ds); err != nil {
			logger.Error(err, "cannot update datasource")
		}
	}
	return ctrl.Result{}, nil
}

func (r *DatasourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&nlptv1.Datasource{}).
		Complete(r)
}
