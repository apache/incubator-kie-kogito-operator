// Copyright 2019 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package deploy

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/common"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/url"

	"github.com/kiegroup/kogito-cloud-operator/pkg/util"

	"github.com/spf13/cobra"
)

const defaultDeployRuntime = string(v1alpha1.QuarkusRuntimeType)

var deployRuntimeValidEntries = []string{string(v1alpha1.QuarkusRuntimeType), string(v1alpha1.SpringbootRuntimeType)}

type deployFlags struct {
	CommonFlags
	common.ChannelFlags
	name              string
	runtime           string
	serviceLabels     []string
	incrementalBuild  bool
	buildEnv          []string
	reference         string
	contextDir        string
	source            string
	imageS2I          string
	imageRuntime      string
	native            bool
	buildLimits       []string
	buildRequests     []string
	imageVersion      string
	mavenMirrorURL    string
	enableIstio       bool
	enablePersistence bool
	enableEvents      bool
}

type deployCommand struct {
	context.CommandContext
	command *cobra.Command
	flags   deployFlags
	Parent  *cobra.Command
}

// initDeployCommand is the constructor for the deploy command
func initDeployCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &deployCommand{CommandContext: *ctx, Parent: parent}
	cmd.RegisterHook()
	cmd.InitHook()
	return cmd
}

func (i *deployCommand) Command() *cobra.Command {
	return i.command
}

func (i *deployCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "deploy-service NAME [SOURCE]",
		Short:   "Deploys a new Kogito Runtime Service into the given Project",
		Aliases: []string{"deploy"},
		Long: `deploy-service will create a new Kogito Runtime Service in the Project context. 
        If the source is provided, the build will take place on the cluster, otherwise you must upload the application binaries via "oc start-build [NAME-binary] --from-dir=target". 
		Project context is the namespace (Kubernetes) or project (OpenShift) where the Service will be deployed.
		To know what's your context, use "kogito project". To set a new Project in the context use "kogito use-project NAME".

		Please note that this command requires the Kogito Operator installed in the cluster.
		For more information about the Kogito Operator installation please refer to https://github.com/kiegroup/kogito-cloud-operator#kogito-operator-installation.
		`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		// Args validation
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 2 {
				return fmt.Errorf("requires 1 arg, received %v", len(args))
			}
			if len(args) > 1 {
				if _, err := url.ParseRequestURI(args[1]); err != nil {
					return fmt.Errorf("source is not a valid URL, received %s", args[1])
				}
			}
			if err := util.ParseStringsForKeyPair(i.flags.buildEnv); err != nil {
				return fmt.Errorf("build environment variables are in the wrong format. Valid are key pairs like 'env=value', received %s", i.flags.buildEnv)
			}
			if err := util.ParseStringsForKeyPair(i.flags.serviceLabels); err != nil {
				return fmt.Errorf("service labels are in the wrong format. Valid are key pairs like 'service=myservice', received %s", i.flags.serviceLabels)
			}
			if err := util.ParseStringsForKeyPair(i.flags.buildLimits); err != nil {
				return fmt.Errorf("build-limits are in the wrong format. Valid are key pairs like 'cpu=1', received %s", i.flags.buildLimits)
			}
			if err := util.ParseStringsForKeyPair(i.flags.buildRequests); err != nil {
				return fmt.Errorf("build-requests are in the wrong format. Valid are key pairs like 'cpu=1', received %s", i.flags.buildRequests)
			}
			if !util.Contains(i.flags.runtime, deployRuntimeValidEntries) {
				return fmt.Errorf("runtime not valid. Valid runtimes are %s. Received %s", deployRuntimeValidEntries, i.flags.runtime)
			}
			if err := CheckImageTag(i.flags.imageRuntime); err != nil {
				return err
			}
			if err := CheckImageTag(i.flags.imageS2I); err != nil {
				return err
			}
			if err := CheckDeployArgs(&i.flags.CommonFlags); err != nil {
				return err
			}
			if err := common.CheckChannelArgs(&i.flags.ChannelFlags); err != nil {
				return err
			}
			return nil
		},
	}
}

func (i *deployCommand) InitHook() {
	i.flags = deployFlags{
		CommonFlags:  CommonFlags{},
		ChannelFlags: common.ChannelFlags{},
	}
	i.Parent.AddCommand(i.command)
	AddDeployFlags(i.command, &i.flags.CommonFlags)
	common.AddChannelFlags(i.command, &i.flags.ChannelFlags)

	i.command.Flags().StringVarP(&i.flags.runtime, "runtime", "r", defaultDeployRuntime, "The runtime which should be used to build the Service. Valid values are 'quarkus' or 'springboot'. Default to '"+defaultDeployRuntime+"'.")
	i.command.Flags().StringVarP(&i.flags.reference, "branch", "b", "", "Git branch to use in the git repository")
	i.command.Flags().StringVarP(&i.flags.contextDir, "context-dir", "c", "", "Context/subdirectory where the code is located, relatively to repository root")
	i.command.Flags().StringSliceVar(&i.flags.serviceLabels, "svc-labels", nil, "Labels that should be applied to the internal endpoint of the Kogito Service. Used by the service discovery engine. Example: 'label=value'. Can be set more than once.")
	i.command.Flags().BoolVar(&i.flags.incrementalBuild, "incremental-build", true, "Build should be incremental?")
	i.command.Flags().BoolVar(&i.flags.native, "native", false, "Use native builds? Be aware that native builds takes more time and consume much more resources from the cluster. Defaults to false")
	i.command.Flags().StringArrayVar(&i.flags.buildEnv, "build-env", nil, "Key/pair value environment variables that will be set during the build. For example 'MY_CUSTOM_ENV=my_custom_value'. Can be set more than once.")
	i.command.Flags().StringSliceVar(&i.flags.buildLimits, "build-limits", nil, "Resource limits for the s2i build pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
	i.command.Flags().StringSliceVar(&i.flags.buildRequests, "build-requests", nil, "Resource requests for the s2i build pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
	i.command.Flags().StringVar(&i.flags.imageS2I, "image-s2i", "", "Custom image tag for the s2i build to build the application binaries, e.g: quay.io/mynamespace/myimage:latest")
	i.command.Flags().StringVar(&i.flags.imageRuntime, "image-runtime", "", "Custom image tag for the s2i build, e.g: quay.io/mynamespace/myimage:latest")
	i.command.Flags().BoolVar(&i.flags.enablePersistence, "enable-persistence", false, "If set to true will install Infinispan in the same namespace and inject the environment variables to configure the service connection to the Infinispan server.")
	i.command.Flags().BoolVar(&i.flags.enableEvents, "enable-events", false, "If set to true will install a Kafka cluster via the Strimzi Operator. The environment variable 'KAFKA_BOOTSTRAP_SERVERS' will be available for the service during runtime.")
	i.command.Flags().StringVar(&i.flags.imageVersion, "image-version", "", "Image version for standard Kogito build images. Ignored if a custom image is set for image-s2i or image-runtime.")
	i.command.Flags().StringVar(&i.flags.mavenMirrorURL, "maven-mirror-url", "", "Internal Maven Mirror to be used during source-to-image builds to considerably increase build speed, e.g: https://my.internal.nexus/content/group/public")
	i.command.Flags().BoolVar(&i.flags.enableIstio, "enable-istio", false, "Enable Istio integration by annotating the Kogito service pods with the right value for Istio controller to inject sidecars on it. Defaults to false")
}

func (i *deployCommand) Exec(cmd *cobra.Command, args []string) (err error) {
	log := context.GetDefaultLogger()
	i.flags.name = args[0]
	hasSource := false
	if len(args) > 1 {
		i.flags.source = args[1]
		hasSource = true
	}

	if len(i.flags.mavenMirrorURL) > 0 {
		if _, err := url.ParseRequestURI(i.flags.mavenMirrorURL); err != nil {
			return err
		}
	}

	if i.flags.Project, err = shared.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}

	installationChannel := shared.KogitoChannelType(i.flags.Channel)
	if installed, err := shared.SilentlyInstallOperatorIfNotExists(i.flags.Project, "", i.Client, installationChannel); err != nil {
		return err
	} else if !installed {
		return nil
	}

	if err := shared.CheckKogitoAppNotExists(i.Client, i.flags.name, i.flags.Project); err != nil {
		return err
	}

	log.Debugf("About to deploy a new kogito service: %s, runtime %s source %s on namespace %s",
		i.flags.name,
		i.flags.runtime,
		i.flags.source,
		i.flags.Project,
	)

	// Build the application
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: v1.ObjectMeta{
			Name:      i.flags.name,
			Namespace: i.flags.Project,
		},
		Spec: v1alpha1.KogitoAppSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas: &i.flags.Replicas,
				Envs:     shared.FromStringArrayToEnvs(i.flags.Env),
				Resources: corev1.ResourceRequirements{
					Limits:   shared.FromStringArrayToResources(i.flags.Limits),
					Requests: shared.FromStringArrayToResources(i.flags.Requests),
				},
			},

			Runtime: v1alpha1.RuntimeType(i.flags.runtime),
			Build: &v1alpha1.KogitoAppBuildObject{
				Envs: shared.FromStringArrayToEnvs(i.flags.buildEnv),
				Resources: corev1.ResourceRequirements{
					Limits:   shared.FromStringArrayToResources(i.flags.buildLimits),
					Requests: shared.FromStringArrayToResources(i.flags.buildRequests),
				},
				Incremental:     i.flags.incrementalBuild,
				ImageS2ITag:     i.flags.imageS2I,
				ImageRuntimeTag: i.flags.imageRuntime,
				ImageVersion:    i.flags.imageVersion,
				Native:          i.flags.native,
				MavenMirrorURL:  i.flags.mavenMirrorURL,
			},
			Service: v1alpha1.KogitoAppServiceObject{
				Labels: util.FromStringsKeyPairToMap(i.flags.serviceLabels),
			},
			EnablePersistence: i.flags.enablePersistence,
			EnableEvents:      i.flags.enableEvents,
			EnableIstio:       i.flags.enableIstio,
		},
		Status: v1alpha1.KogitoAppStatus{
			ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: make([]v1alpha1.Condition, 0)},
		},
	}

	if hasSource {
		kogitoApp.Spec.Build.GitSource = v1alpha1.GitSource{
			URI:        i.flags.source,
			ContextDir: i.flags.contextDir,
			Reference:  i.flags.reference,
		}
	}

	log.Debugf("Trying to deploy Kogito Service '%s'", kogitoApp.Name)
	// Create the Kogito application
	if err := kubernetes.ResourceC(i.Client).Create(kogitoApp); err != nil {
		return fmt.Errorf("Error while creating a new KogitoApp in the context: %v ", err)
	}

	log.Infof(message.KogitoAppSuccessfullyCreated, kogitoApp.Name, kogitoApp.Namespace)
	if hasSource {
		log.Infof(message.KogitoAppViewDeploymentStatus, i.flags.name, i.flags.Project)
		log.Infof(message.KogitoAppViewBuildStatus, i.flags.name, i.flags.Project)
	} else {
		log.Infof(message.KogitoAppUploadBinariesInstruction, i.flags.name, i.flags.Project)
	}

	endpoint, err := infrastructure.GetManagementConsoleEndpoint(i.Client, i.flags.Project)
	if err != nil {
		return err
	}
	if endpoint.IsEmpty() {
		log.Info(message.KogitoAppNoMgmtConsole)
	} else {
		log.Infof(message.KogitoAppMgmtConsoleEndpoint, endpoint.HTTPRouteURI)
	}

	return nil
}
