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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/converter"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

const (
	defaultTrustyInfinispanSecretName  = infrastructure.DefaultTrustyName + "-infinispan-credentials"
	defaultTrustyInfinispanUsernameKey = "username"
	defaultTrustyInfinispanPasswordKey = "password"
)

type installTrustyFlags struct {
	flag.InstallFlags
	flag.InfinispanFlags
	flag.KafkaFlags
}

type installTrustyCommand struct {
	context.CommandContext
	command *cobra.Command
	flags   installTrustyFlags
	Parent  *cobra.Command
}

func initInstallTrustyCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &installTrustyCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}

	cmd.RegisterHook()
	cmd.InitHook()

	return cmd
}

func (i *installTrustyCommand) Command() *cobra.Command {
	return i.command
}

func (i *installTrustyCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "trusty [flags]",
		Short:   "Installs the Kogito Trusty Service in the given Project",
		Example: "trusty -p my-project",
		Long: `'install trusty' will deploy the Trusty service to enable capturing tracing events produced by one or more Kogito services and provide analysis capabilities on top of the data.

If kafka-url is provided, it will be used to connect to the external Kafka server that is deployed in other namespace or infrastructure.
If kafka-instance is provided instead, the value will be used as the Strimzi Kafka instance name to locate the Kafka server deployed in the Trusty service's namespace.
Otherwise, the operator will try to deploy a Kafka instance via Strimzi operator for you using Kogito Infrastructure in the given namespace.

If infinispan-url is not provided, a new Infinispan server will be deployed for you using Kogito Infrastructure, if no one exists in the given project.
Only use infinispan-url if you plan to connect to an external Infinispan server that is already provided in other namespace or infrastructure.`,
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

func (i *installTrustyCommand) InitHook() {
	i.flags = installTrustyFlags{}
	i.Parent.AddCommand(i.command)
	flag.AddInstallFlags(i.command, &i.flags.InstallFlags)
	flag.AddInfinispanFlags(i.command, &i.flags.InfinispanFlags)
	flag.AddKafkaFlags(i.command, &i.flags.KafkaFlags)
}

func (i *installTrustyCommand) Exec(cmd *cobra.Command, args []string) error {
	var err error
	if i.flags.Project, err = shared.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}
	infinispanMeta, err := converter.FromInfinispanFlagsToInfinispanMeta(i.Client, i.flags.Project, &i.flags.InfinispanFlags, true)
	if err != nil {
		return err
	}

	kogitoTrusty := v1alpha1.KogitoTrusty{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultTrustyName, Namespace: i.flags.Project},
		Spec: v1alpha1.KogitoTrustySpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas:              &i.flags.Replicas,
				Envs:                  converter.FromStringArrayToEnvs(i.flags.Env, i.flags.SecretEnv),
				Image:                 converter.FromImageFlagToImage(&i.flags.ImageFlags),
				Resources:             converter.FromPodResourceFlagsToResourceRequirement(&i.flags.PodResourceFlags),
				HTTPPort:              i.flags.HTTPPort,
				InsecureImageRegistry: i.flags.ImageFlags.InsecureImageRegistry,
			},
			InfinispanMeta: infinispanMeta,
			KafkaMeta:      converter.FromKafkaFlagsToKafkaMeta(&i.flags.KafkaFlags, true),
		},
		Status: v1alpha1.KogitoTrustyStatus{
			KogitoServiceStatus: v1alpha1.KogitoServiceStatus{
				ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}},
			},
		},
	}

	return shared.
		ServicesInstallationBuilder(i.Client, i.flags.Project).
		SilentlyInstallOperatorIfNotExists(shared.KogitoChannelType(i.flags.Channel)).
		WarnIfDependenciesNotReady(i.flags.InfinispanFlags.UseKogitoInfra, i.flags.KafkaFlags.UseKogitoInfra).
		InstallTrusty(&kogitoTrusty).
		GetError()
}
