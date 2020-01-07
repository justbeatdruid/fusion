/*
Copyright 2019 nlpt.

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

package serviceunit

import (
	"flag"
	"fmt"

	nlptv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"github.com/chinamobile/nlpt/crds/serviceunit/controllers"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	_ = clientgoscheme.AddToScheme(scheme)

	_ = nlptv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func Run() error {
	var metricsAddr string
	var enableLeaderElection bool
	var operatorHost string
	var operatorPort int
	var operatorCAFile string
	flag.StringVar(&metricsAddr, "metrics-addr", ":8002", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&operatorHost, "operator-host", "127.0.0.1", "Host of database warehose service.")
	flag.IntVar(&operatorPort, "operator-port", 80, "Port of database warehose service.")
	flag.StringVar(&operatorCAFile, "operator-cafile", "", "Certificate for TLS communication with database warehose service.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		return fmt.Errorf("unable to start manager: %+v", err)
	}

	operator, err := controllers.NewOperator(operatorHost, operatorPort, operatorCAFile)
	if err != nil {
		setupLog.Error(err, "unable to create operator")
		return fmt.Errorf("unable to create operator: %+v", err)
	}

	if err = (&controllers.ServiceunitReconciler{
		Client:   mgr.GetClient(),
		Operator: operator,
		Log:      ctrl.Log.WithName("controllers").WithName("Database"),
		Scheme:   mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Database")
		return fmt.Errorf("unable to create controller: %+v", err)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting serviceunit manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return fmt.Errorf("problem running manager: %+v", err)
	}
	return nil
}
