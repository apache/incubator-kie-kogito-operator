package main

import (
	"fmt"
	"net/url"
	"regexp"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/util"

	"github.com/spf13/cobra"
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
	imageRuntime     string
	imageS2I         string
}

const (
	defaultDeployReplicas = 1
	defaultDeployRuntime  = string(v1alpha1.QuarkusRuntimeType)
	// see: https://github.com/docker/distribution/blob/master/reference/regexp.go
	dockerTagRegx = `[\w][\w.-]{0,127}`
)

var (
	deployCmd                 *cobra.Command
	deployCmdFlags            = deployFlags{}
	deployRuntimeValidEntries = []string{string(v1alpha1.QuarkusRuntimeType), string(v1alpha1.SpringbootRuntimeType)}
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
		PreRun:  preRunF,
		PostRun: posRunF,
		// args validation
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
			re := regexp.MustCompile(dockerTagRegx)
			if len(deployCmdFlags.imageRuntime) > 0 && !re.MatchString(deployCmdFlags.imageRuntime) {
				return fmt.Errorf("invalid name for runtime image tag. Received %s", deployCmdFlags.imageRuntime)
			}
			if len(deployCmdFlags.imageS2I) > 0 && !re.MatchString(deployCmdFlags.imageS2I) {
				return fmt.Errorf("invalid name for s2i image tag. Received %s", deployCmdFlags.imageRuntime)
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
	deployCmd.Flags().StringVar(&deployCmdFlags.imageRuntime, "image-runtime", "", "Image tag (namespace/name:tag) for using during service runtime, e.g: openshift/kogito-quarkus-ubi8:latest")
	deployCmd.Flags().StringVar(&deployCmdFlags.imageS2I, "image-s2i", "", "Image tag (namespace/name:tag) for using during the s2i build, e.g: openshift/kogito-quarkus-ubi8-s2i:latest")
})

func deployExec(cmd *cobra.Command, args []string) error {
	deployCmdFlags.name = args[0]
	deployCmdFlags.source = args[1]
	if len(deployCmdFlags.namespace) == 0 {
		if len(config.Namespace) == 0 {
			return fmt.Errorf("Couldn't find any application in the current context. Use 'kogito app NAME' to set the Kogito Application where the service will be deployed or pass '--app NAME' flag to this one")
		}
		deployCmdFlags.namespace = config.Namespace
	}
	log.Debugf("About to deploy a new kogito service application: %s, runtime %s source %s on namespace %s",
		deployCmdFlags.name,
		deployCmdFlags.runtime,
		deployCmdFlags.source,
		deployCmdFlags.namespace,
	)

	if err := checkNamespaceExists(deployCmdFlags.namespace); err != nil {
		return err
	}
	log.Debugf("Using namespace %s", deployCmdFlags.namespace)

	if err := checkKogitoAppCRDExists(); err != nil {
		return err
	}

	if err := checkKogitoAppNotExists(deployCmdFlags.name, deployCmdFlags.namespace); err != nil {
		return err
	}

	// build the application
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployCmdFlags.name,
			Namespace: deployCmdFlags.namespace,
		},
		Spec: v1alpha1.KogitoAppSpec{
			Name:     deployCmdFlags.name,
			Replicas: &deployCmdFlags.replicas,
			Runtime:  v1alpha1.RuntimeType(deployCmdFlags.runtime),
			Build: &v1alpha1.KogitoAppBuildObject{
				Incremental: deployCmdFlags.incrementalBuild,
				GitSource: &v1alpha1.GitSource{
					URI:        &deployCmdFlags.source,
					ContextDir: deployCmdFlags.contextdir,
					Reference:  deployCmdFlags.reference,
				},
				Env:          fromStringArrayToControllerEnvs(deployCmdFlags.buildenv),
				ImageS2I:     fromStringToImage(deployCmdFlags.imageS2I),
				ImageRuntime: fromStringToImage(deployCmdFlags.imageRuntime),
			},
			Env: fromStringArrayToControllerEnvs(deployCmdFlags.env),
			Service: v1alpha1.KogitoAppServiceObject{
				Labels: util.FromStringsKeyPairToMap(deployCmdFlags.serviceLabels),
			},
			Resources: v1alpha1.Resources{
				Limits:   fromStringArrayToControllerResourceMap(deployCmdFlags.limits),
				Requests: fromStringArrayToControllerResourceMap(deployCmdFlags.requests),
			},
		},
		Status: v1alpha1.KogitoAppStatus{
			Conditions: []v1alpha1.Condition{},
		},
	}
	log.Debugf("Trying to deploy Kogito Service '%s'", kogitoApp.Name)
	// create it!
	if err := kubernetes.ResourceC(kubeCli).Create(kogitoApp); err != nil {
		return fmt.Errorf("Error while creating a new KogitoApp in the context: %v", err)
	}
	config.LastKogitoAppCreated = kogitoApp
	log.Infof("KogitoApp '%s' successfully created on namespace '%s'", kogitoApp.Name, kogitoApp.Namespace)
	// TODO: we should provide this info with a -f flag
	log.Infof("You can see the deployment status by using 'oc describe %s %s -n %s'", "kogitoapp", deployCmdFlags.name, deployCmdFlags.namespace)
	log.Infof("Your service should be deploying. To see its logs, run 'oc logs -f bc/%s-builder -n %s'", deployCmdFlags.name, deployCmdFlags.namespace)

	return nil
}
