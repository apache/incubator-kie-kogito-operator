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

package install

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/converter"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type installSupportingServiceFlags struct {
	flag.InstallFlags
}

type installableSupportingService struct {
	cmdName     string
	serviceName string
	aliases     []string
	displayName string
	description string
	serviceType v1beta1.ServiceType
}

type installSupportingServiceCommand struct {
	context.CommandContext
	command           *cobra.Command
	flags             installSupportingServiceFlags
	supportingService installableSupportingService
	Parent            *cobra.Command
}

var installableSupportingServices = []installableSupportingService{
	{
		cmdName:     "data-index",
		serviceName: infrastructure.DefaultDataIndexName,
		displayName: "Data Index",
		serviceType: v1beta1.DataIndex,
		description: `'install data-index' will deploy the Data Index service to enable capturing and indexing data produced by one or more Kogito services.

If kafka-url is provided, it will be used to connect to the external Kafka server that is deployed in other project or infrastructure.
If kafka-instance is provided instead, the value will be used as the Strimzi Kafka instance name to locate the Kafka server deployed in the Data Index service's project.
Otherwise, the operator will try to deploy a Kafka instance via Strimzi operator for you using Kogito Infrastructure in the given project.

If infinispan-url is not provided, a new Infinispan server will be deployed for you using Kogito Infrastructure, if no one exists in the given project.
Only use infinispan-url if you plan to connect to an external Infinispan server that is already provided in other project or infrastructure.

For more information on Kogito Data Index Service see: https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service`,
	},
	{
		cmdName:     "explainability",
		serviceName: infrastructure.DefaultExplainabilityName,
		displayName: "Explainability",
		serviceType: v1beta1.Explainability,
		description: `'install explainability' will deploy the Explainability service to provide analysis on the decisions that have been taken by a kogito runtime application.

If kafka-url is provided, it will be used to connect to the external Kafka server that is deployed in other project or infrastructure.
If kafka-instance is provided instead, the value will be used as the Strimzi Kafka instance name to locate the deployed Kafka server in the Explainability service's project.
Otherwise, the operator will try to deploy a Kafka instance via Strimzi operator for you using Kogito Infrastructure in the given project.`,
	},
	{
		cmdName:     "jobs-service",
		serviceName: infrastructure.DefaultJobsServiceName,
		displayName: "Jobs",
		serviceType: v1beta1.JobsService,
		description: `'install jobs-service' deploys the Jobs Service to enable scheduling jobs that aim to be fired at a given time for Kogito services.

If 'enable-persistence' flag is set and 'infinispan-url' is not provided, a new Infinispan server will be deployed for you using Kogito Infrastructure.
Use 'infinispan-url' and set 'enable-persistence' flag if you plan to connect to an external Infinispan server that is already provided 
in other project or infrastructure.

For more information on Kogito Jobs Service see: https://github.com/kiegroup/kogito-runtimes/wiki/Jobs-Service`,
	},
	{
		cmdName:     "mgmt-console",
		serviceName: infrastructure.DefaultMgmtConsoleName,
		aliases:     []string{"management-console"},
		displayName: "Mgmt Console",
		serviceType: v1beta1.MgmtConsole,
		description: `'install mgmt-console' deploys the Management Console to enable management for Kogito Services deployed within the same project.

Please note that Management Console relies on Data Index (https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service) to retrieve the processes instances via its GraphQL API.
You won't be able to use the Management Console if Data Index is not deployed in the same project either using Kogito CLI or the Kogito Operator.

For more information on Management Console see: https://github.com/kiegroup/kogito-runtimes/wiki/Process-Instance-Management`,
	},
	{
		cmdName:     "task-console",
		serviceName: infrastructure.DefaultTaskConsoleName,
		displayName: "Task Console",
		serviceType: v1beta1.TaskConsole,
		description: `'install task-console' deploys the Task Console to enable management for Kogito Services deployed within the same project.

Please note that Task Console relies on Data Index (https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service) to retrieve the processes instances via its GraphQL API.
You won't be able to use the Task Console if Data Index is not deployed in the same project either using Kogito CLI or the Kogito Operator.

For more information on Task Console see: https://github.com/kiegroup/kogito-runtimes/wiki/Process-Instance-Management`,
	},
	{
		cmdName:     "trusty",
		serviceName: infrastructure.DefaultTrustyName,
		displayName: "Trusty",
		serviceType: v1beta1.TrustyAI,
		description: `'install trusty' will deploy the Trusty service to enable capturing tracing events produced by one or more Kogito services and provide analysis capabilities on top of the data.

If kafka-url is provided, it will be used to connect to the external Kafka server that is deployed in other namespace or infrastructure.
If kafka-instance is provided instead, the value will be used as the Strimzi Kafka instance name to locate the Kafka server deployed in the Trusty service's namespace.
Otherwise, the operator will try to deploy a Kafka instance via Strimzi operator for you using Kogito Infrastructure in the given namespace.

If infinispan-url is not provided, a new Infinispan server will be deployed for you using Kogito Infrastructure, if no one exists in the given project.
Only use infinispan-url if you plan to connect to an external Infinispan server that is already provided in other namespace or infrastructure.

See https://github.com/kiegroup/kogito-apps/tree/master/trusty/README.md for more information about the trusty service.`,
	},
	{
		cmdName:     "trusty-ui",
		serviceName: infrastructure.DefaultTrustyUIName,
		displayName: "Trusty UI",
		serviceType: v1beta1.TrustyUI,
		description: `'install trusty-ui' deploys the Trusty UI to enable the audit UI for Kogito Services deployed within the same project.

Please note that Trusty UI relies on Trusty (https://github.com/kiegroup/kogito-apps/tree/master/trusty) to retrieve the information to be displayed.
You won't be able to use the Trusty UI if Trusty is not deployed in the same project either using Kogito CLI or the Kogito Operator. 
In addition to that, it is mandatory to set the environment variable KOGITO_TRUSTY_ENDPOINT in the trusty-ui service. The value of that variable should be the endpoint of the trusty service.`,
	},
}

func initInstallSupportingServiceCommands(ctx *context.CommandContext, parent *cobra.Command) []context.KogitoCommand {
	var commands []context.KogitoCommand
	for _, installable := range installableSupportingServices {
		cmd := &installSupportingServiceCommand{
			CommandContext:    *ctx,
			supportingService: installable,
			Parent:            parent,
		}
		cmd.RegisterHook()
		cmd.InitHook()
		commands = append(commands, cmd)
	}
	return commands
}

func (i *installSupportingServiceCommand) Command() *cobra.Command {
	return i.command
}

func (i *installSupportingServiceCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     i.supportingService.cmdName,
		Aliases: i.supportingService.aliases,
		Short:   fmt.Sprintf("Installs the Kogito %s Service in the given Project", i.supportingService.displayName),
		Example: fmt.Sprintf("install %s -p my-project", i.supportingService.cmdName),
		Long:    i.supportingService.description,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := flag.CheckInstallArgs(&i.flags.InstallFlags); err != nil {
				return err
			}
			return nil
		},
	}
}

func (i *installSupportingServiceCommand) InitHook() {
	i.flags = installSupportingServiceFlags{}
	i.Parent.AddCommand(i.command)
	flag.AddInstallFlags(i.command, &i.flags.InstallFlags)
}

func (i *installSupportingServiceCommand) Exec(cmd *cobra.Command, args []string) error {
	var err error
	if i.flags.Project, err = shared.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}
	configMap, err := converter.CreateConfigMapFromFile(i.Client, i.supportingService.serviceName, i.flags.Project, &i.flags.ConfigFlags)
	if err != nil {
		return err
	}
	supportingService := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      i.supportingService.serviceName,
			Namespace: i.flags.Project,
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: i.supportingService.serviceType,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas:              &i.flags.Replicas,
				Env:                   converter.FromStringArrayToEnvs(i.flags.Env, i.flags.SecretEnv),
				Image:                 i.flags.ImageFlags.Image,
				Resources:             converter.FromPodResourceFlagsToResourceRequirement(&i.flags.PodResourceFlags),
				InsecureImageRegistry: i.flags.ImageFlags.InsecureImageRegistry,
				Infra:                 i.flags.Infra,
				PropertiesConfigMap:   configMap,
				Config:                converter.FromConfigFlagsToMap(&i.flags.ConfigFlags),
				Probes:                converter.FromProbeFlagToKogitoProbe(&i.flags.ProbeFlags),
			},
		},
		Status: v1beta1.KogitoSupportingServiceStatus{
			KogitoServiceStatus: v1beta1.KogitoServiceStatus{
				ConditionsMeta: v1beta1.ConditionsMeta{Conditions: []v1beta1.Condition{}},
			},
		},
	}

	return shared.
		ServicesInstallationBuilder(i.Client, i.flags.Project).
		CheckOperatorCRDs().
		InstallSupportingService(supportingService).
		GetError()
}
