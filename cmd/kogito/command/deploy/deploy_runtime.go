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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/converter"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deployRuntimeFlags struct {
	flag.DeployFlags
	flag.InfinispanFlags
	flag.KafkaFlags
	flag.RuntimeFlags
	name              string
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
		Use:     "deploy-service NAME [IMAGE]",
		Short:   "Deploys a new Kogito Service into the given Project",
		Aliases: []string{"deploy"},
		Long: `deploy-service will create a new Kogito Service using provided [IMAGE] in the Project context. 
			
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
			if err := flag.CheckDeployArgs(&i.flags.DeployFlags); err != nil {
				return err
			}
			if err := flag.CheckInfinispanArgs(&i.flags.InfinispanFlags); err != nil {
				return err
			}
			if err := flag.CheckKafkaArgs(&i.flags.KafkaFlags); err != nil {
				return err
			}
			if err := flag.CheckRuntimeArgs(&i.flags.RuntimeFlags); err != nil {
				return err
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
		DeployFlags: flag.DeployFlags{
			OperatorFlags:    flag.OperatorFlags{},
			PodResourceFlags: flag.PodResourceFlags{},
		},
		InfinispanFlags: flag.InfinispanFlags{},
		KafkaFlags:      flag.KafkaFlags{},
		RuntimeFlags:    flag.RuntimeFlags{},
	}

	flag.AddDeployFlags(i.command, &i.flags.DeployFlags)
	flag.AddInfinispanFlags(i.command, &i.flags.InfinispanFlags)
	flag.AddKafkaFlags(i.command, &i.flags.KafkaFlags)
	flag.AddRuntimeFlags(i.command, &i.flags.RuntimeFlags)
	i.command.Flags().BoolVar(&i.flags.enableIstio, "enable-istio", false, "Enable Istio integration by annotating the Kogito service pods with the right value for Istio controller to inject sidecars on it. Defaults to false")
	i.command.Flags().BoolVar(&i.flags.enablePersistence, "enable-persistence", false, "If set to true, deployed Kogito service will support integration with Infinispan server for persistence. Default to false")
	i.command.Flags().BoolVar(&i.flags.enableEvents, "enable-events", false, "If set to true, deployed Kogito service will support integration with Kafka cluster for events. Default to false")
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
	infinispanMeta, err := converter.FromInfinispanFlagsToInfinispanMeta(i.Client, i.flags.Project, &i.flags.InfinispanFlags, i.flags.enablePersistence)
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
			Runtime:     converter.FromRuntimeFlagsToRuntimeType(&i.flags.RuntimeFlags),
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas:              &i.flags.Replicas,
				Envs:                  converter.FromStringArrayToEnvs(i.flags.Env),
				Image:                 converter.FromImageTagToImage(i.flags.Image),
				Resources:             converter.FromPodResourceFlagsToResourceRequirement(&i.flags.PodResourceFlags),
				ServiceLabels:         util.FromStringsKeyPairToMap(i.flags.serviceLabels),
				HTTPPort:              i.flags.HTTPPort,
				InsecureImageRegistry: i.flags.InsecureImageRegistry,
			},
			InfinispanMeta: infinispanMeta,
			KafkaMeta:      converter.FromKafkaFlagsToKafkaMeta(&i.flags.KafkaFlags, i.flags.enableEvents),
		},
		Status: v1alpha1.KogitoRuntimeStatus{
			KogitoServiceStatus: v1alpha1.KogitoServiceStatus{
				ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}},
			},
		},
	}

	log.Debugf("Trying to deploy Kogito Service '%s'", kogitoRuntime.Name)
	// Create the Kogito application
	err = shared.
		ServicesInstallationBuilder(i.Client, i.flags.Project).
		SilentlyInstallOperatorIfNotExists(shared.KogitoChannelType(i.flags.Channel)).
		WarnIfDependenciesNotReady(i.flags.InfinispanFlags.UseKogitoInfra, i.flags.KafkaFlags.UseKogitoInfra).
		DeployService(&kogitoRuntime).
		GetError()
	if err != nil {
		return err
	}

	endpoint, err := infrastructure.GetManagementConsoleEndpoint(i.Client, i.flags.Project)
	if err != nil {
		return err
	}
	if endpoint.IsEmpty() {
		log.Info(message.RuntimeServiceMgmtConsole)
	} else {
		log.Infof(message.RuntimeServiceMgmtConsoleEndpoint, endpoint.HTTPRouteURI)
	}

	return nil
}
