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

package main

import (
	"flag"
	"k8s.io/apimachinery/pkg/util/wait"
	"os"
	"time"

	nlptv1 "github.com/chinamobile/nlpt/crds/api/api/v1"
	"github.com/chinamobile/nlpt/crds/api/controllers"
	suv1 "github.com/chinamobile/nlpt/crds/serviceunit/api/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/klog"
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
	_ = suv1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
	klog.InitFlags(nil)
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	var operatorHost string
	var operatorPort int
	var operatorCAFile string
	var portalPort int
	var prometheusHost string
	var prometheusPort int
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.StringVar(&operatorHost, "operator-host", "127.0.0.1", "Host of kong service.")
	flag.IntVar(&operatorPort, "operator-port", 800, "Port of kong admin service.")
	flag.IntVar(&portalPort, "portal-port", 8443, "Port of kong portal service.")
	flag.StringVar(&prometheusHost, "prometheus-host", "127.0.0.1", "Host of prometheus service.")
	flag.IntVar(&prometheusPort, "prometheus-port", 32008, "Port of prometheus service.")
	flag.StringVar(&operatorCAFile, "operator-cafile", "", "Certificate for TLS communication with database warehose service.")
	flag.Parse()

	ctrl.SetLogger(zap.New(func(o *zap.Options) {
		o.Development = true
	}))

	namespace := os.Getenv("MY_POD_NAMESPACE")
	if len(namespace) == 0 {
		namespace = "default"
	}
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		MetricsBindAddress:      metricsAddr,
		LeaderElection:          enableLeaderElection,
		LeaderElectionNamespace: namespace,
		LeaderElectionID:        "fusion-api-controller-manager",
		Port:                    9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	operator, err := controllers.NewOperator(operatorHost, operatorPort, portalPort, operatorCAFile, prometheusHost, prometheusPort)
	if err != nil {
		setupLog.Error(err, "unable to create operator")
		os.Exit(1)
	}

	var ar *controllers.ApiReconciler = &controllers.ApiReconciler{
		Client:   mgr.GetClient(),
		Log:      ctrl.Log.WithName("controllers").WithName("Api"),
		Scheme:   mgr.GetScheme(),
		Operator: operator,
	}
	if ar.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Api")
		os.Exit(1)
	}

	// +kubebuilder:scaffold:builder
	stop := make(chan struct{})
	go func() {
		setupLog.Info("starting manager")
		if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
			setupLog.Error(err, "problem running manager")
			os.Exit(1)
		}
		close(stop)
	}()
	// wait for caches up
	time.Sleep(time.Second)
	wait.Until(func() {
		if err := ar.SyncApiCountFromKong(); err != nil {
			klog.Errorf("sync api count error: %+v", err)
		}
		// do not use wait.NerverStop
	}, time.Second*60, stop)
}
