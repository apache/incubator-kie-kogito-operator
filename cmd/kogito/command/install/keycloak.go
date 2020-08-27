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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/spf13/cobra"
)

type installKeycloakFlags struct {
	flag.OperatorFlags
	project string
}

type installKeycloakCommand struct {
	context.CommandContext
	flags   installKeycloakFlags
	command *cobra.Command
	Parent  *cobra.Command
}

func initInstallKeycloakCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	command := installKeycloakCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}

	command.RegisterHook()
	command.InitHook()

	return &command
}

func (i *installKeycloakCommand) Command() *cobra.Command {
	return i.command
}

func (i *installKeycloakCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "keycloak [flags]",
		Short:   "Installs a keycloak instance into the OpenShift/Kubernetes cluster",
		Example: "install keycloak -p my-project",
		Long:    `Installs a keycloak instance via custom Kubernetes resources.
Requires Keycloak Operator to be installed, please make sure that you've installed it before proceeding.
This feature won't create custom subscriptions with the OLM.`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := flag.CheckOperatorArgs(&i.flags.OperatorFlags); err != nil {
				return err
			}
			return nil
		},
	}
}

func (i *installKeycloakCommand) InitHook() {
	i.flags = installKeycloakFlags{}
	i.Parent.AddCommand(i.command)
	flag.AddOperatorFlags(i.command, &i.flags.OperatorFlags)

	i.command.Flags().StringVarP(&i.flags.project, "project", "p", "", "The project name where the Keycloak server will be deployed")
}

func (i *installKeycloakCommand) Exec(cmd *cobra.Command, args []string) error {
	var err error
	if i.flags.project, err = shared.EnsureProject(i.Client, i.flags.project); err != nil {
		return err
	}
	return shared.
		ServicesInstallationBuilder(i.Client, i.flags.project).
		SilentlyInstallOperatorIfNotExists(shared.KogitoChannelType(i.flags.Channel)).
		InstallKeycloak().
		GetError()
}
