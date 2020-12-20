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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	buildv1 "github.com/openshift/api/build/v1"
	imagev1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	appv1beta1 "github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(appv1beta1.AddToScheme(scheme))
	utilruntime.Must(routev1.Install(scheme))
	utilruntime.Must(imagev1.Install(scheme))
	utilruntime.Must(buildv1.Install(scheme))
	// +kubebuilder:scaffold:scheme
}

func main() {
	var metricsAddr string
	var enableLeaderElection bool
	flag.StringVar(&metricsAddr, "metrics-addr", ":8080", "The address the metric endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "enable-leader-election", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	watchNamespace := getWatchNamespace()
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:             scheme,
		MetricsBindAddress: metricsAddr,
		Port:               9443,
		LeaderElection:     enableLeaderElection,
		LeaderElectionID:   "4662f1d5.kiegroup.org",
		Namespace:          watchNamespace,
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.KogitoRuntimeReconciler{
		Client: client.NewForController(mgr.GetConfig()),
		Log:    logger.GetLogger("kogitoruntime_controllers"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KogitoRuntime")
		os.Exit(1)
	}
	if err = (&controllers.KogitoSupportingServiceReconciler{
		Client: client.NewForController(mgr.GetConfig()),
		Log:    logger.GetLogger("kogitoSupportingService-controller"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KogitoSupportingService")
		os.Exit(1)
	}
	if err = (&controllers.KogitoBuildReconciler{
		Client: client.NewForController(mgr.GetConfig()),
		Log:    logger.GetLogger("kogitoBuild-controller"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KogitoBuild")
		os.Exit(1)
	}
	if err = (&controllers.KogitoInfraReconciler{
		Client: client.NewForController(mgr.GetConfig()),
		Log:    logger.GetLogger("KogitoInfra-controller"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KogitoInfra")
		os.Exit(1)
	}
	if err = (&controllers.FinalizeKogitoSupportingService{
		Client: client.NewForController(mgr.GetConfig()),
		Log:    logger.GetLogger("KogitoSupportingService-finalizer"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KogitoSupportingService-finalizer")
		os.Exit(1)
	}
	if err = (&controllers.FinalizeKogitoRuntime{
		Client: client.NewForController(mgr.GetConfig()),
		Log:    logger.GetLogger("KogitoRuntime-finalizer"),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KogitoRuntime-finalizer")
		os.Exit(1)
	}
	// +kubebuilder:scaffold:builder

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}

// getWatchNamespace returns the Namespace the operator should be watching for changes
func getWatchNamespace() string {
	// WatchNamespaceEnvVar is the constant for env variable WATCH_NAMESPACE
	// which specifies the Namespace to watch.
	// An empty value means the operator is running with cluster scope.
	var watchNamespaceEnvVar = "WATCH_NAMESPACE"

	ns, found := os.LookupEnv(watchNamespaceEnvVar)
	if !found {
		setupLog.Info("unable to get WatchNamespace, "+
			"the manager will watch and manage resources in all namespaces",
			"Env Var lookup", watchNamespaceEnvVar)
		return ""
	}
	return ns
}
