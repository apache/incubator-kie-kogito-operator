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

package project

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/spf13/cobra"
)

type newProjectFlags struct {
	project          string
	installDataIndex bool
}

type newProjectCommand struct {
	context.CommandContext
	flags   newProjectFlags
	command *cobra.Command
	Parent  *cobra.Command
}

func initNewProjectCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := newProjectCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}
	cmd.RegisterHook()
	cmd.InitHook()
	return &cmd
}

func (i *newProjectCommand) Command() *cobra.Command {
	return i.command
}

func (i *newProjectCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "new-project NAME",
		Aliases: []string{"new-ns"},
		Short:   "Creates a new Kogito Project for your Kogito Services",
		Long: `new-project will create a Kubernetes Namespace with the provided project where your Kogito Services will be deployed. This project then will be used to deploy all infrastructure
				bits needed for the deployed Kogito Services to run.`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(i.flags.project) == 0 {
				if len(args) == 0 {
					return fmt.Errorf("Please set a project for new-project")
				}
				i.flags.project = args[0]
			}
			return nil
		},
	}
}

func (i *newProjectCommand) InitHook() {
	i.flags = newProjectFlags{}
	i.Parent.AddCommand(i.command)
	i.command.Flags().StringVarP(&i.flags.project, "project", "n", "", "The project project")
	i.command.Flags().BoolVar(&i.flags.installDataIndex, "install-data-index", false, "Installs the default instance of Data Index being provisioned by the Kogito Operator in the new project")
}

func (i *newProjectCommand) Exec(cmd *cobra.Command, args []string) error {
	log := context.GetDefaultLogger()
	ns, err := kubernetes.NamespaceC(i.Client).Fetch(i.flags.project)
	if err != nil {
		return err
	}
	if ns == nil {
		ns, err := kubernetes.NamespaceC(i.Client).Create(i.flags.project)
		if err != nil {
			return err
		}

		config := context.ReadConfig()
		config.Namespace = ns.Name
		config.Save()
		log.Infof("Project '%s' created successfully", ns.Name)

		install := shared.ServicesInstallationBuilder(i.Client, ns.Name).SilentlyInstallOperator()
		if i.flags.installDataIndex {
			install.InstallDataIndex()
		}
		return install.GetError()
	}
	log.Infof("Project '%s' already exists", i.flags.project)
	return nil
}
