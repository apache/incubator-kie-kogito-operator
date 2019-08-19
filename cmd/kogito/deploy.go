package main

import (
	"fmt"
	"net/url"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"

	"github.com/kiegroup/kogito-cloud-operator/pkg/util"

	"github.com/spf13/cobra"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deployFlags struct {
	name             string
	namespace        string
	runtime          string
	replicas         int32
	env              []string
	limits           []string
	requests         []string
	serviceLabels    []string
	incrementalBuild bool
	buildenv         []string
	reference        string
	contextdir       string
	source           string
}

const (
	defaultDeployReplicas             = 1
	defaultDeployRuntime              = "quarkus"
	kogitoOperatorInstallationInstruc = "https://github.com/kiegroup/kogito-cloud-operator/blob/master/README.md#installation"
)

var (
	deployCmd                 *cobra.Command
	deployCmdFlags            = deployFlags{}
	deployRuntimeValidEntries = []string{"quarkus", "springboot"}
)

var _ = RegisterCommandVar(func() {
	deployCmd = &cobra.Command{
		Use:   "deploy NAME SOURCE",
		Short: "Deploys a new Kogito Service into the application context",
		Long: `'deploy' will create a new Kogito Service from source in the application context.
		Application context is the namespace (Kubernetes) or project (OpenShift) where the service will be deployed. 
		To know what's your context, use "kogito app". To set a new context use "kogito app NAME".
		
		Please note that this command requires the Kogito Operator installed in the cluster.
		For more information about the Kogito Operator installation please refer to https://github.com/kiegroup/kogito-cloud-operator/blob/master/README.md#installation.
		`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deployExec(cmd, args)
		},
		PostRun: saveConfiguration,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("requires 2 args, received %v", len(args))
			}
			if _, err := url.ParseRequestURI(args[1]); err != nil {
				return fmt.Errorf("source is not a valid URL, received %s", args[1])
			}
			if err := util.ParseStringsForKeyPair(deployCmdFlags.buildenv); err != nil {
				return fmt.Errorf("build environment variables are in the wrong format. Valid are key pairs like 'env=value', received %s", deployCmdFlags.buildenv)
			}
			if err := util.ParseStringsForKeyPair(deployCmdFlags.env); err != nil {
				return fmt.Errorf("environment variables are in the wrong format. Valid are key pairs like 'env=value', received %s", deployCmdFlags.env)
			}
			if err := util.ParseStringsForKeyPair(deployCmdFlags.limits); err != nil {
				return fmt.Errorf("limits are in the wrong format. Valid are key pairs like 'cpu=1', received %s", deployCmdFlags.limits)
			}
			if err := util.ParseStringsForKeyPair(deployCmdFlags.requests); err != nil {
				return fmt.Errorf("requests are in the wrong format. Valid are key pairs like 'cpu=1', received %s", deployCmdFlags.requests)
			}
			if err := util.ParseStringsForKeyPair(deployCmdFlags.serviceLabels); err != nil {
				return fmt.Errorf("service labels are in the wrong format. Valid are key pairs like 'service=myservice', received %s", deployCmdFlags.serviceLabels)
			}
			if deployCmdFlags.replicas <= 0 {
				return fmt.Errorf("valid replicas are non-zero, positive numbers, received %v", deployCmdFlags.replicas)
			}
			if !util.Contains(deployCmdFlags.runtime, deployRuntimeValidEntries) {
				return fmt.Errorf("runtime not valid. Valid runtimes are %s. Received %s", deployRuntimeValidEntries, deployCmdFlags.runtime)
			}
			return nil
		},
	}
})

var _ = RegisterCommandInit(func() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().StringVar(&deployCmdFlags.namespace, "app", "", "Application context (namespace) where the Kogito Service will be deployed.")
	deployCmd.Flags().StringVarP(&deployCmdFlags.runtime, "runtime", "r", defaultDeployRuntime, "The runtime which should be used to build the service. Valid values are 'quarkus' or 'springboot'. Default to '"+defaultDeployRuntime+"'.")
	deployCmd.Flags().StringVarP(&deployCmdFlags.reference, "branch", "b", "", "Git branch to use in the git repository")
	deployCmd.Flags().StringVarP(&deployCmdFlags.contextdir, "context", "c", "", "Context/subdirectory where the code is located, relatively to repo root")
	deployCmd.Flags().Int32Var(&deployCmdFlags.replicas, "replicas", defaultDeployReplicas, "Number of pod replicas that should be deployed.")
	deployCmd.Flags().StringSliceVarP(&deployCmdFlags.env, "env", "e", nil, "Key/pair value environment variables that will be set to the Kogito Service. For example 'MY_VAR=my_value'. Can be set more than once.")
	deployCmd.Flags().StringSliceVar(&deployCmdFlags.limits, "limits", nil, "Resource limits for the Kogito Service pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
	deployCmd.Flags().StringSliceVar(&deployCmdFlags.requests, "requests", nil, "Resource requests for the Kogito Service pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
	deployCmd.Flags().StringSliceVarP(&deployCmdFlags.serviceLabels, "svc-labels", "s", nil, "Labels that should be applied to the internal endpoint of the Kogito Service. Used by the service discovery engine. Example: 'label=value'. Can be set more than once.")
	deployCmd.Flags().BoolVar(&deployCmdFlags.incrementalBuild, "incremental-build", true, "Build should be incremental?")
	deployCmd.Flags().StringSliceVarP(&deployCmdFlags.buildenv, "build-env", "u", nil, "Key/pair value environment variables that will be set during the build. For example 'MAVEN_URL=http://myinternalmaven.com'. Can be set more than once.")
})

func deployExec(cmd *cobra.Command, args []string) error {
	log.Debugf("About to deploy a new kogito service application %s", deployCmdFlags)
	deployCmdFlags.name = args[0]
	deployCmdFlags.source = args[1]
	if len(deployCmdFlags.namespace) == 0 {
		if len(config.Namespace) == 0 {
			return fmt.Errorf("Couldn't find any application in the current context. Use 'kogito app NAME' to set the Kogito Application where the service will be deployed or pass '--app NAME' flag to this one")
		}
		deployCmdFlags.namespace = config.Namespace
	}

	if ns, err := kubernetes.NamespaceC(kubeCli).Fetch(deployCmdFlags.namespace); err != nil {
		return fmt.Errorf("Error while trying to fetch for the application context (namespace): %s", err)
	} else if ns == nil {
		return fmt.Errorf("Namespace %s not found. Try setting your application context using 'kogito app NAME'", deployCmdFlags.namespace)
	}

	log.Debugf("Using namespace %s", deployCmdFlags.namespace)

	// check for the KogitoApp CRD
	kogitocrd := &apiextensionsv1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kogitoapps.app.kiegroup.org",
		},
	}
	if exists, err := kubernetes.ResourceC(kubeCli).Fetch(kogitocrd); err != nil {
		return fmt.Errorf("Error while trying to look for the Kogito Operator: %s", err)
	} else if !exists {
		return fmt.Errorf("Couldn't find the Kogito Operator in your cluster, please follow the instructions in %s to install it", kogitoOperatorInstallationInstruc)
	}

	// check if a kogito service with this name already exists in this namespace
	kogitoapp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployCmdFlags.name,
			Namespace: deployCmdFlags.namespace,
		},
	}
	if exists, err := kubernetes.ResourceC(kubeCli).Fetch(kogitoapp); err != nil {
		return fmt.Errorf("Error while trying to look for the KogitoApp: %s", err)
	} else if exists {
		return fmt.Errorf("Looks like a Kogito App with the name '%s' already exists in this context/namespace. Please try another name", deployCmdFlags.name)
	}

	// build the application

	// create it!

	return nil
}
