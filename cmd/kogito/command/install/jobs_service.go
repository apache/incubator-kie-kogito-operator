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
	"github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/converter"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type installJobsServiceFlags struct {
	flag.InstallFlags
	flag.InfinispanFlags
	flag.KafkaFlags
	backOffRetryMillis            int64
	maxIntervalLimitToRetryMillis int64
	enablePersistence             bool
	enableEvents                  bool
}

type installJobsServiceCommand struct {
	context.CommandContext
	command *cobra.Command
	flags   installJobsServiceFlags
	Parent  *cobra.Command
}

func initInstallJobsServiceCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &installJobsServiceCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}

	cmd.RegisterHook()
	cmd.InitHook()

	return cmd
}

func (i *installJobsServiceCommand) Command() *cobra.Command {
	return i.command
}

func (i *installJobsServiceCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "jobs-service [flags]",
		Short:   "Installs the Kogito Jobs Service in the given project",
		Example: "jobs-service -p my-project",
		Long: `'install jobs-service' deploys the Jobs Service to enable scheduling jobs that aim to be fired at a given time for Kogito services.

If 'enable-persistence' flag is set and 'infinispan-url' is not provided, a new Infinispan server will be deployed for you using Kogito Infrastructure.
Use 'infinispan-url' and set 'enable-persistence' flag if you plan to connect to an external Infinispan server that is already provided 
in other project or infrastructure.

For more information on Kogito Jobs Service see: https://github.com/kiegroup/kogito-runtimes/wiki/Jobs-Service`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := flag.CheckInstallArgs(&i.flags.InstallFlags); err != nil {
				return err
			}
			if err := flag.CheckInfinispanArgs(&i.flags.InfinispanFlags); err != nil {
				return err
			}
			if err := flag.CheckKafkaArgs(&i.flags.KafkaFlags); err != nil {
				return err
			}
			return nil
		},
	}
}

func (i *installJobsServiceCommand) InitHook() {
	i.flags = installJobsServiceFlags{}
	i.Parent.AddCommand(i.command)
	flag.AddInstallFlags(i.command, &i.flags.InstallFlags)
	flag.AddInfinispanFlags(i.command, &i.flags.InfinispanFlags)
	flag.AddKafkaFlags(i.command, &i.flags.KafkaFlags)

	i.command.Flags().BoolVar(&i.flags.enableEvents, "enable-events", false, "Enable persistence using Kafka. Set also 'kafka-url' to specify an instance URL. If left in blank the operator will provide one for you")
	i.command.Flags().BoolVar(&i.flags.enablePersistence, "enable-persistence", false, "Enable persistence using Infinispan. Set also 'infinispan-url' to specify an instance URL. If left in blank the operator will provide one for you")
	i.command.Flags().Int64Var(&i.flags.backOffRetryMillis, "backoff-retry-millis", 0, "Sets the internal property 'kogito.jobs-service.backoffRetryMillis'")
	i.command.Flags().Int64Var(&i.flags.maxIntervalLimitToRetryMillis, "max-internal-limit-retry-millis", 0, "Sets the internal property 'kogito.jobs-service.maxIntervalLimitToRetryMillis'")
}

func (i *installJobsServiceCommand) Exec(cmd *cobra.Command, args []string) error {
	var err error
	if i.flags.Project, err = shared.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}
	infinispanMeta, err := converter.FromInfinispanFlagsToInfinispanMeta(i.Client, i.flags.Project, &i.flags.InfinispanFlags, i.flags.enablePersistence)
	if err != nil {
		return err
	}

	kogitoJobsService := v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultJobsServiceName, Namespace: i.flags.Project},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas:              &i.flags.Replicas,
				Envs:                  converter.FromStringArrayToEnvs(i.flags.Env, i.flags.SecretEnv),
				Image:                 converter.FromImageFlagToImage(&i.flags.ImageFlags),
				Resources:             converter.FromPodResourceFlagsToResourceRequirement(&i.flags.PodResourceFlags),
				HTTPPort:              i.flags.HTTPPort,
				InsecureImageRegistry: i.flags.ImageFlags.InsecureImageRegistry,
			},
			BackOffRetryMillis:            i.flags.backOffRetryMillis,
			MaxIntervalLimitToRetryMillis: i.flags.maxIntervalLimitToRetryMillis,
			InfinispanMeta:                infinispanMeta,
			KafkaMeta:                     converter.FromKafkaFlagsToKafkaMeta(&i.flags.KafkaFlags, i.flags.enableEvents),
		},
		Status: v1alpha1.KogitoJobsServiceStatus{
			KogitoServiceStatus: v1alpha1.KogitoServiceStatus{
				ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}},
			},
		},
	}

	return shared.
		ServicesInstallationBuilder(i.Client, i.flags.Project).
		SilentlyInstallOperatorIfNotExists(shared.KogitoChannelType(i.flags.Channel)).
		WarnIfDependenciesNotReady(i.flags.InfinispanFlags.UseKogitoInfra, i.flags.KafkaFlags.UseKogitoInfra).
		InstallJobsService(&kogitoJobsService).
		GetError()
}
