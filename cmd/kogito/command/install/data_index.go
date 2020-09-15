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

package install

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/converter"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
)

type installDataIndexFlags struct {
	flag.InstallFlags
}

type installDataIndexCommand struct {
	context.CommandContext
	command *cobra.Command
	flags   installDataIndexFlags
	Parent  *cobra.Command
}

func initInstallDataIndexCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &installDataIndexCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}

	cmd.RegisterHook()
	cmd.InitHook()

	return cmd
}

func (i *installDataIndexCommand) Command() *cobra.Command {
	return i.command
}

func (i *installDataIndexCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "data-index [flags]",
		Short:   "Installs the Kogito Data Index Service in the given Project",
		Example: "data-index -p my-project",
		Long: `'install data-index' will deploy the Data Index service to enable capturing and indexing data produced by one or more Kogito services.

If kafka-url is provided, it will be used to connect to the external Kafka server that is deployed in other project or infrastructure.
If kafka-instance is provided instead, the value will be used as the Strimzi Kafka instance name to locate the Kafka server deployed in the Data Index service's project.
Otherwise, the operator will try to deploy a Kafka instance via Strimzi operator for you using Kogito Infrastructure in the given project.

If infinispan-url is not provided, a new Infinispan server will be deployed for you using Kogito Infrastructure, if no one exists in the given project.
Only use infinispan-url if you plan to connect to an external Infinispan server that is already provided in other project or infrastructure.

For more information on Kogito Data Index Service see: https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service`,
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

func (i *installDataIndexCommand) InitHook() {
	i.flags = installDataIndexFlags{}
	i.Parent.AddCommand(i.command)
	flag.AddInstallFlags(i.command, &i.flags.InstallFlags)
}

func (i *installDataIndexCommand) Exec(cmd *cobra.Command, args []string) error {
	var err error
	if i.flags.Project, err = shared.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}

	kogitoDataIndex := v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultDataIndexName, Namespace: i.flags.Project},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas:              &i.flags.Replicas,
				Envs:                  converter.FromStringArrayToEnvs(i.flags.Env, i.flags.SecretEnv),
				Image:                 converter.FromImageFlagToImage(&i.flags.ImageFlags),
				Resources:             converter.FromPodResourceFlagsToResourceRequirement(&i.flags.PodResourceFlags),
				HTTPPort:              i.flags.HTTPPort,
				InsecureImageRegistry: i.flags.ImageFlags.InsecureImageRegistry,
			},
		},
		Status: v1alpha1.KogitoDataIndexStatus{
			KogitoServiceStatus: v1alpha1.KogitoServiceStatus{
				ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}},
			},
		},
	}

	return shared.
		ServicesInstallationBuilder(i.Client, i.flags.Project).
		SilentlyInstallOperatorIfNotExists(shared.KogitoChannelType(i.flags.Channel)).
		InstallDataIndex(&kogitoDataIndex).
		GetError()
}
