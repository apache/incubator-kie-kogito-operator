package main

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"net/url"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/util"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deployFlags struct {
	deployCommonFlags
	name             string
	runtime          string
	serviceLabels    []string
	incrementalBuild bool
	buildenv         []string
	reference        string
	contextdir       string
	source           string
	imageS2I         string
	imageRuntime     string
	native           bool
}

const (
	defaultDeployRuntime = string(v1alpha1.QuarkusRuntimeType)
)

var (
	deployCmd      *cobra.Command
	deployCmdFlags = deployFlags{
		deployCommonFlags: deployCommonFlags{},
	}
	deployRuntimeValidEntries = []string{string(v1alpha1.QuarkusRuntimeType), string(v1alpha1.SpringbootRuntimeType)}
)

var _ = RegisterCommandVar(func() {
	deployCmd = &cobra.Command{
		Use:   "deploy-service NAME SOURCE",
		Short: "Deploys a new Kogito Runtime Service into the given Project",
		Long: `deploy-service will create a new Kogito Runtime Service from source in the Project context.
		Project context is the namespace (Kubernetes) or project (OpenShift) where the Service will be deployed. 
		To know what's your context, use "kogito use-project". To set a new Project in the context use "kogito use-project NAME".
		
		Please note that this command requires the Kogito Operator installed in the cluster.
		For more information about the Kogito Operator installation please refer to https://github.com/kiegroup/kogito-cloud-operator#installation.
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
			if err := util.ParseStringsForKeyPair(deployCmdFlags.serviceLabels); err != nil {
				return fmt.Errorf("service labels are in the wrong format. Valid are key pairs like 'service=myservice', received %s", deployCmdFlags.serviceLabels)
			}
			if !util.Contains(deployCmdFlags.runtime, deployRuntimeValidEntries) {
				return fmt.Errorf("runtime not valid. Valid runtimes are %s. Received %s", deployRuntimeValidEntries, deployCmdFlags.runtime)
			}
			if err := commonCheckImageTag(deployCmdFlags.imageRuntime); err != nil {
				return err
			}
			if err := commonCheckImageTag(deployCmdFlags.imageS2I); err != nil {
				return err
			}
			if err := commonCheckDeployArgs(&deployCmdFlags.deployCommonFlags); err != nil {
				return err
			}
			return nil
		},
	}
})

var _ = RegisterCommandInit(func() {
	rootCmd.AddCommand(deployCmd)
	commonAddDeployFlags(deployCmd, &deployCmdFlags.deployCommonFlags)
	deployCmd.Flags().StringVarP(&deployCmdFlags.runtime, "runtime", "r", defaultDeployRuntime, "The runtime which should be used to build the Service. Valid values are 'quarkus' or 'springboot'. Default to '"+defaultDeployRuntime+"'.")
	deployCmd.Flags().StringVarP(&deployCmdFlags.reference, "branch", "b", "", "Git branch to use in the git repository")
	deployCmd.Flags().StringVarP(&deployCmdFlags.contextdir, "context-dir", "c", "", "Context/subdirectory where the code is located, relatively to repository root")
	deployCmd.Flags().StringSliceVar(&deployCmdFlags.serviceLabels, "svc-labels", nil, "Labels that should be applied to the internal endpoint of the Kogito Service. Used by the service discovery engine. Example: 'label=value'. Can be set more than once.")
	deployCmd.Flags().BoolVar(&deployCmdFlags.incrementalBuild, "incremental-build", true, "Build should be incremental?")
	deployCmd.Flags().BoolVar(&deployCmdFlags.native, "native", false, "Use native builds? Be aware that native builds takes more time and consume much more resources from the cluster. Defaults to false")
	deployCmd.Flags().StringSliceVar(&deployCmdFlags.buildenv, "build-env", nil, "Key/pair value environment variables that will be set during the build. For example 'MAVEN_URL=http://myinternalmaven.com'. Can be set more than once.")
	deployCmd.Flags().StringVar(&deployCmdFlags.imageS2I, "image-s2i", "", "Image tag (namespace/name:tag) for using during the s2i build, e.g: openshift/kogito-quarkus-ubi8-s2i:latest")
	deployCmd.Flags().StringVar(&deployCmdFlags.imageRuntime, "image-runtime", "", "Image tag (namespace/name:tag) for using during service runtime, e.g: openshift/kogito-quarkus-ubi8:latest")
})

func deployExec(cmd *cobra.Command, args []string) error {
	deployCmdFlags.name = args[0]
	deployCmdFlags.source = args[1]
	var err error
	if deployCmdFlags.project, err = ensureProject(deployCmdFlags.project); err != nil {
		return err
	}
	if err := checkKogitoAppCRDExists(); err != nil {
		return err
	}
	if err := checkKogitoAppNotExists(deployCmdFlags.name, deployCmdFlags.project); err != nil {
		return err
	}

	log.Debugf("About to deploy a new kogito service: %s, runtime %s source %s on namespace %s",
		deployCmdFlags.name,
		deployCmdFlags.runtime,
		deployCmdFlags.source,
		deployCmdFlags.project,
	)

	// build the application
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployCmdFlags.name,
			Namespace: deployCmdFlags.project,
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
				Native:       deployCmdFlags.native,
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
	log.Infof("You can see the deployment status by using 'oc describe %s %s -n %s'", "kogitoapp", deployCmdFlags.name, deployCmdFlags.project)
	log.Infof("Your Kogito Runtime Service should be deploying. To see its logs, run 'oc logs -f bc/%s-builder -n %s'", deployCmdFlags.name, deployCmdFlags.project)

	return nil
}
