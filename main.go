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
	"github.com/operatify/operatify/controllers/b"
	"os"

	"github.com/operatify/operatify/controllers/a"
	"github.com/operatify/operatify/controllers/manager"
	"github.com/operatify/operatify/reconciler"

	api "github.com/operatify/operatify/api/v1alpha1"
	testv1alpha1 "github.com/operatify/operatify/api/v1alpha1"
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

	_ = api.AddToScheme(scheme)
	_ = testv1alpha1.AddToScheme(scheme)
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.Logger(true))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		LeaderElection:     enableLeaderElection,
		Port:               9443,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// create controllers
	controllerParams := reconciler.ReconcileParameters{
		RequeueAfter:        5000,
		RequeueAfterSuccess: 15000,
		RequeueAfterFailure: 30000,
	}
	store := manager.CreateManager()
	if err = (&a.ControllerFactory{
		ResourceManagerCreator: a.CreateResourceManager,
		Scheme:                 scheme,
		Manager:                store,
	}).SetupWithManager(mgr, controllerParams, nil); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ATest")
		os.Exit(1)
	}

	if err = (&b.ControllerFactory{
		ResourceManagerCreator: b.CreateResourceManager,
		Scheme:                 scheme,
		Manager:                store,
	}).SetupWithManager(mgr, controllerParams, nil); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "BTest")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
