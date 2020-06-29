// Copyright 2020 Red Hat, Inc. and/or its affiliates
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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	buildutil "github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/util"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deployRuntimeFlags struct {
	common.DeployFlags
	common.InfinispanFlags
	common.KafkaFlags
	name              string
	image             string
	runtime           string
	enableIstio       bool
	enablePersistence bool
	enableEvents      bool
	serviceLabels     []string
}

type deployRuntimeCommand struct {
	context.CommandContext
	command *cobra.Command
	flags   deployRuntimeFlags
	Parent  *cobra.Command
}

// initDeployCommand is the constructor for the deploy command
func initDeployRuntimeCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &deployRuntimeCommand{CommandContext: *ctx, Parent: parent}
	cmd.RegisterHook()
	cmd.InitHook()
	return cmd
}

func (i *deployRuntimeCommand) Command() *cobra.Command {
	return i.command
}

func (i *deployRuntimeCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "deploy-runtime NAME [IMAGE]",
		Short:   "Deploys a new Kogito Runtime Service into the given Project",
		Aliases: []string{"deploy"},
		Long: `deploy-runtime will create a new Kogito Runtime Service using provided [IMAGE] in the Project context. 
			
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
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg, received %v", len(args))
			}
			if err := common.CheckDeployArgs(&i.flags.DeployFlags); err != nil {
				return err
			}
			if err := common.CheckInfinispanArgs(&i.flags.InfinispanFlags); err != nil {
				return err
			}
			if err := common.CheckKafkaArgs(&i.flags.KafkaFlags); err != nil {
				return err
			}
			if err := buildutil.CheckImageTag(i.flags.image); err != nil {
				return err
			}
			if !util.Contains(i.flags.runtime, deployRuntimeValidEntries) {
				return fmt.Errorf("runtime not valid. Valid runtimes are %s. Received %s", deployRuntimeValidEntries, i.flags.runtime)
			}
			if err := util.ParseStringsForKeyPair(i.flags.serviceLabels); err != nil {
				return fmt.Errorf("service labels are in the wrong format. Valid are key pairs like 'service=myservice', received %s", i.flags.serviceLabels)
			}
			return nil
		},
	}
}

func (i *deployRuntimeCommand) InitHook() {
	i.Parent.AddCommand(i.command)
	i.flags = deployRuntimeFlags{
		DeployFlags: common.DeployFlags{
			OperatorFlags: common.OperatorFlags{},
		},
		InfinispanFlags: common.InfinispanFlags{},
		KafkaFlags:      common.KafkaFlags{},
	}

	common.AddDeployFlags(i.command, &i.flags.DeployFlags)
	common.AddInfinispanFlags(i.command, &i.flags.InfinispanFlags)
	common.AddKafkaFlags(i.command, &i.flags.KafkaFlags)
	i.command.Flags().StringVarP(&i.flags.image, "image", "i", "", "The image which should be used to run Service.")
	i.command.Flags().StringVarP(&i.flags.runtime, "runtime", "r", defaultDeployRuntime, "The runtime which should be used to build the Service. Valid values are 'quarkus' or 'springboot'. Default to '"+defaultDeployRuntime+"'.")
	i.command.Flags().BoolVar(&i.flags.enableIstio, "enable-istio", false, "Enable Istio integration by annotating the Kogito service pods with the right value for Istio controller to inject sidecars on it. Defaults to false")
	i.command.Flags().BoolVar(&i.flags.enablePersistence, "enable-persistence", false, "If set to true, deployed runtime service will support integration with Infinispan server for persistence. Default to false")
	i.command.Flags().BoolVar(&i.flags.enableEvents, "enable-events", false, "If set to true, deployed runtime service will support integration with Kafka cluster for events. Default to false")
}

func (i *deployRuntimeCommand) Exec(cmd *cobra.Command, args []string) (err error) {
	log := context.GetDefaultLogger()
	i.flags.name = args[0]
	if i.flags.Project, err = shared.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}
	if err := shared.CheckKogitoRuntimeNotExists(i.Client, i.flags.name, i.flags.Project); err != nil {
		return err
	}
	infinispanProperties, err := common.FromInfinispanFlagsToInfinispanProperties(i.Client, i.flags.Project, &i.flags.InfinispanFlags, i.flags.enablePersistence)
	if err != nil {
		return err
	}

	kogitoRuntime := v1alpha1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{
			Name:      i.flags.name,
			Namespace: i.flags.Project,
		},
		Spec: v1alpha1.KogitoRuntimeSpec{
			EnableIstio: i.flags.enableIstio,
			Runtime:     v1alpha1.RuntimeType(i.flags.runtime),
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas: &i.flags.Replicas,
				Envs:     shared.FromStringArrayToEnvs(i.flags.Env),
				Image:    framework.ConvertImageTagToImage(i.flags.image),
				Resources: corev1.ResourceRequirements{
					Limits:   shared.FromStringArrayToResources(i.flags.Limits),
					Requests: shared.FromStringArrayToResources(i.flags.Requests),
				},
				ServiceLabels: util.FromStringsKeyPairToMap(i.flags.serviceLabels),
				HTTPPort:      i.flags.HTTPPort,
			},
			InfinispanMeta: v1alpha1.InfinispanMeta{
				InfinispanProperties: infinispanProperties,
			},
			KafkaMeta: v1alpha1.KafkaMeta{
				KafkaProperties: common.FromKafkaFlagsToKafkaProperties(&i.flags.KafkaFlags, i.flags.enableEvents),
			},
		},
		Status: v1alpha1.KogitoRuntimeStatus{
			KogitoServiceStatus: v1alpha1.KogitoServiceStatus{
				ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}},
			},
		},
	}

	log.Debugf("Trying to deploy Kogito runtime Service '%s'", kogitoRuntime.Name)
	// Create the Kogito application
	return shared.
		ServicesInstallationBuilder(i.Client, i.flags.Project).
		SilentlyInstallOperatorIfNotExists(shared.KogitoChannelType(i.flags.Channel)).
		WarnIfDependenciesNotReady(i.flags.InfinispanFlags.UseKogitoInfra, i.flags.KafkaFlags.UseKogitoInfra).
		InstallRuntime(&kogitoRuntime).
		GetError()
}
