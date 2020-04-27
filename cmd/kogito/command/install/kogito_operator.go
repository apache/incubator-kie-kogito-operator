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
	"errors"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/spf13/cobra"
)

type installKogitoOperatorFlags struct {
	namespace          string
	image              string
	installDataIndex   bool
	installJobsService bool
	installMgmtConsole bool
	installAllServices bool
	force              bool
}

type installKogitoOperatorCommand struct {
	context.CommandContext
	flags   installKogitoOperatorFlags
	command *cobra.Command
	Parent  *cobra.Command
}

func initInstallKogitoOperatorCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	command := installKogitoOperatorCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}

	command.RegisterHook()
	command.InitHook()

	return &command
}

func (i *installKogitoOperatorCommand) Command() *cobra.Command {
	return i.command
}

func (i *installKogitoOperatorCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "operator [flags]",
		Short:   "Installs the Kogito Operator into the OpenShift/Kubernetes cluster",
		Example: "install operator -p my-project",
		Long:    `Installs the Kogito Operator via custom Kubernetes resources. This feature won't create custom subscriptions with the OLM.`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func (i *installKogitoOperatorCommand) InitHook() {
	i.flags = installKogitoOperatorFlags{}
	i.Parent.AddCommand(i.command)
	i.command.Flags().StringVarP(&i.flags.namespace, "project", "p", "", "The project name where the operator will be deployed")
	i.command.Flags().StringVarP(&i.flags.image, "image", "i", shared.DefaultOperatorImageNameTag, "The operator image")
	i.command.Flags().BoolVar(&i.flags.installDataIndex, "install-data-index", false, "Installs the default instance of Data Index being provisioned by the Kogito Operator in the project")
	i.command.Flags().BoolVar(&i.flags.installJobsService, "install-jobs-service", false, "Installs the default instance of Jobs Service being provisioned by the Kogito Operator in the project")
	i.command.Flags().BoolVar(&i.flags.installMgmtConsole, "install-mgmt-console", false, "Installs the default instance of Management Console being provisioned by the Kogito Operator in the project")
	i.command.Flags().BoolVar(&i.flags.installAllServices, "install-all-services", false, "Installs the default instance of every Kogito Support services (Data Index, Jobs Service, etc.) being provisioned by the Kogito Operator in the project")
	i.command.Flags().BoolVarP(&i.flags.force, "force", "f", false, "When set, the operator will be installed in the current namespace using a custom image, e.g. quay.io/kiegroup/kogito-cloud-operator:my-custom-tag")
}

func (i *installKogitoOperatorCommand) Exec(cmd *cobra.Command, args []string) error {
	var err error
	// if force flag is set, then a custom image is required.
	if i.flags.force {
		if i.flags.image == shared.DefaultOperatorImageNameTag {
			return errors.New("force install flag is enabled but the custom operator image is missing")
		}
	}

	if i.flags.namespace, err = shared.EnsureProject(i.Client, i.flags.namespace); err != nil {
		return err
	}

	install := shared.ServicesInstallationBuilder(i.Client, i.flags.namespace).InstallOperator(true, i.flags.image, i.flags.force)
	if i.flags.installDataIndex || i.flags.installAllServices {
		install.InstallDataIndex(nil)
	}
	if i.flags.installJobsService || i.flags.installAllServices {
		install.InstallJobsService(nil)
	}
	if i.flags.installMgmtConsole || i.flags.installAllServices {
		install.InstallMgmtConsole(nil)
	}
	return install.GetError()
}
