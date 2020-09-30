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

type installTrustyUIFlags struct {
	flag.InstallFlags
}

type installTrustyUICommand struct {
	context.CommandContext
	command *cobra.Command
	flags   installTrustyUIFlags
	Parent  *cobra.Command
}

func initInstallTrustyUICommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &installTrustyUICommand{
		CommandContext: *ctx,
		Parent:         parent,
	}

	cmd.RegisterHook()
	cmd.InitHook()

	return cmd
}

func (i *installTrustyUICommand) Command() *cobra.Command {
	return i.command
}

func (i *installTrustyUICommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "trusty-ui [flags]",
		Aliases: []string{"trusty-ui"},
		Short:   "Installs the Kogito Trusty UI in the given Project",
		Example: "trusty-ui -p my-project",
		Long: `'install trusty-ui' deploys the Trusty UI to enable the audit UI for Kogito Services deployed within the same project.

Please note that Trusty UI relies on Trusty (https://github.com/kiegroup/kogito-apps/tree/master/trusty) to retrieve the information to be displayed.
You won't be able to use the Trusty UI if Trusty is not deployed in the same project either using Kogito CLI or the Kogito Operator. 
In addition to that, it is mandatory to set the environment variable KOGITO_TRUSTY_ENDPOINT in the trusty-ui service. The value of that variable should be the endpoint of the trusty service.`,
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

func (i *installTrustyUICommand) InitHook() {
	i.flags = installTrustyUIFlags{}
	i.Parent.AddCommand(i.command)
	flag.AddInstallFlags(i.command, &i.flags.InstallFlags)
}

func (i *installTrustyUICommand) Exec(cmd *cobra.Command, args []string) error {
	var err error
	if i.flags.Project, err = shared.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}

	kogitoTrustyUI := v1alpha1.KogitoTrustyUI{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultTrustyUIName, Namespace: i.flags.Project},
		Spec: v1alpha1.KogitoTrustyUISpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Replicas:              &i.flags.Replicas,
				Env:                   converter.FromStringArrayToEnvs(i.flags.Env, i.flags.SecretEnv),
				Image:                 i.flags.ImageFlags.Image,
				Resources:             converter.FromPodResourceFlagsToResourceRequirement(&i.flags.PodResourceFlags),
				HTTPPort:              i.flags.HTTPPort,
				InsecureImageRegistry: i.flags.ImageFlags.InsecureImageRegistry,
			},
		},
		Status: v1alpha1.KogitoTrustyUIStatus{
			KogitoServiceStatus: v1alpha1.KogitoServiceStatus{
				ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}},
			},
		},
	}

	return shared.
		ServicesInstallationBuilder(i.Client, i.flags.Project).
		SilentlyInstallOperatorIfNotExists(shared.KogitoChannelType(i.flags.Channel)).
		InstallTrustyUI(&kogitoTrustyUI).
		GetError()
}
