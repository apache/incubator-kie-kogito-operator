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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/converter"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type installMgmtConsoleFlags struct {
	flag.InstallFlags
}

type installMgmtConsoleCommand struct {
	context.CommandContext
	command *cobra.Command
	flags   installMgmtConsoleFlags
	Parent  *cobra.Command
}

func initInstallMgmtConsoleCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &installMgmtConsoleCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}

	cmd.RegisterHook()
	cmd.InitHook()

	return cmd
}

func (i *installMgmtConsoleCommand) Command() *cobra.Command {
	return i.command
}

func (i *installMgmtConsoleCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "mgmt-console [flags]",
		Aliases: []string{"management-console"},
		Short:   "Installs the Kogito Management Console in the given Project",
		Example: "mgmt-console -p my-project",
		Long: `'install mgmt-console' deploys the Management Console to enable management for Kogito Services deployed within the same project.

Please note that Management Console relies on Data Index (https://github.com/kiegroup/kogito-runtimes/wiki/Data-Index-Service) to retrieve the processes instances via its GraphQL API.
You won't be able to use the Management Console if Data Index is not deployed in the same project either using Kogito CLI or the Kogito Operator.

For more information on Management Console see: https://github.com/kiegroup/kogito-runtimes/wiki/Process-Instance-Management`,
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

func (i *installMgmtConsoleCommand) InitHook() {
	i.flags = installMgmtConsoleFlags{}
	i.Parent.AddCommand(i.command)
	flag.AddInstallFlags(i.command, &i.flags.InstallFlags)
}

func (i *installMgmtConsoleCommand) Exec(cmd *cobra.Command, args []string) error {
	var err error
	if i.flags.Project, err = shared.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}

	kogitoMgmtConsole := v1alpha1.KogitoMgmtConsole{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultMgmtConsoleName, Namespace: i.flags.Project},
		Spec: v1alpha1.KogitoMgmtConsoleSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas:              &i.flags.Replicas,
				Envs:                  converter.FromStringArrayToEnvs(i.flags.Env, i.flags.SecretEnv),
				Image:                 i.flags.ImageFlags.Image,
				Resources:             converter.FromPodResourceFlagsToResourceRequirement(&i.flags.PodResourceFlags),
				HTTPPort:              i.flags.HTTPPort,
				InsecureImageRegistry: i.flags.ImageFlags.InsecureImageRegistry,
				Infra:                 i.flags.Infra,
			},
		},
		Status: v1alpha1.KogitoMgmtConsoleStatus{
			KogitoServiceStatus: v1alpha1.KogitoServiceStatus{
				ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}},
			},
		},
	}

	return shared.
		ServicesInstallationBuilder(i.Client, i.flags.Project).
		SilentlyInstallOperatorIfNotExists(shared.KogitoChannelType(i.flags.Channel)).
		InstallMgmtConsole(&kogitoMgmtConsole).
		GetError()
}
