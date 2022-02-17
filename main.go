/*
Copyright 2021.

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
	"github.com/kiegroup/kogito-operator/controllers/rhpam"
	"github.com/kiegroup/kogito-operator/core/client"
	"github.com/kiegroup/kogito-operator/core/framework/util"
	"github.com/kiegroup/kogito-operator/core/logger"
	"github.com/kiegroup/kogito-operator/meta"
	"os"
	"strings"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/kiegroup/kogito-operator/controllers/app"
	//+kubebuilder:scaffold:imports
)

var (
	scheme               *runtime.Scheme
	setupLog             = logger.GetLogger("setup")
	metricsAddr          string
	enableLeaderElection bool
	probeAddr            string
)

func init() {
	scheme = meta.GetRegisteredSchema()
	//+kubebuilder:scaffold:scheme
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
}

func main() {
	opts := zap.Options{
		Development: isDebugMode(),
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "d1731e98.kiegroup.org",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	kubeCli := client.NewForController(mgr)

	if !util.IsProductMode() {
		if err = app.NewKogitoRuntimeReconciler(kubeCli, mgr.GetScheme()).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "KogitoRuntime")
			os.Exit(1)
		}
		if err = app.NewKogitoSupportingServiceReconciler(kubeCli, mgr.GetScheme()).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "KogitoSupportingService")
			os.Exit(1)
		}
		if err = app.NewKogitoBuildReconciler(kubeCli, mgr.GetScheme()).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "KogitoBuild")
			os.Exit(1)
		}
		if err = app.NewKogitoInfraReconciler(kubeCli, mgr.GetScheme()).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "KogitoInfra")
			os.Exit(1)
		}
	} else {
		if err = rhpam.NewKogitoRuntimeReconciler(kubeCli, mgr.GetScheme()).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "KogitoRuntime")
			os.Exit(1)
		}
		if err = rhpam.NewKogitoSupportingServiceReconciler(kubeCli, mgr.GetScheme()).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "KogitoSupportingService")
			os.Exit(1)
		}
		if err = rhpam.NewKogitoBuildReconciler(kubeCli, mgr.GetScheme()).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "KogitoBuild")
			os.Exit(1)
		}
		if err = rhpam.NewKogitoInfraReconciler(kubeCli, mgr.GetScheme()).SetupWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", "KogitoInfra")
			os.Exit(1)
		}
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

func isDebugMode() bool {
	var debug = "DEBUG"
	devMode, _ := os.LookupEnv(debug)

	if strings.ToUpper(devMode) == "TRUE" {
		setupLog.Info("Running in Debug Mode")
		return true
	}
	return false

}
