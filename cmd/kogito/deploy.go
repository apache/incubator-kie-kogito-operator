package main

import (
	"fmt"
	"net/url"

	"github.com/spf13/cobra"
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
}

const (
	defaultDeployReplicas = 1
	defaultDeployRuntime  = "quarkus"
)

var (
	deployCmd                 *cobra.Command
	deployCmdFlags            = deployFlags{}
	deployRuntimeValidEntries = [...]string{"quarkus", "springboot"}
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

	return nil
}
