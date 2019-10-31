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

type useProjectFlags struct {
	project string
}

type useProjectCommand struct {
	context.CommandContext
	flags   useProjectFlags
	command *cobra.Command
	Parent  *cobra.Command
}

func newUseProjectCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := useProjectCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}
	cmd.RegisterHook()
	cmd.InitHook()
	return &cmd
}

func (i *useProjectCommand) Command() *cobra.Command {
	return i.command
}

func (i *useProjectCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "use-project NAME",
		Aliases: []string{"use-ns"},
		Short:   "Sets the Kogito Project where your Kogito Service will be deployed",
		Long:    `use-project will set the Kubernetes Namespace where the Kogito services will be deployed. It's the Namespace/Project on Kubernetes/OpenShift world.`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(i.flags.project) == 0 {
				if len(args) == 0 {
					log := context.GetDefaultLogger()
					config := context.ReadConfig()
					if len(config.Namespace) == 0 {
						return fmt.Errorf("No Project set in the context. Use 'kogito new-project NAME' to create a new Project")
					}
					log.Debugf("Project in the context is '%s'. Use 'kogito deploy-service NAME SOURCE' to deploy a new Kogito Service.", config.Namespace)
					i.flags.project = config.Namespace
					return nil
				}
				i.flags.project = args[0]
			}
			return nil
		},
	}
}

func (i *useProjectCommand) InitHook() {
	i.flags = useProjectFlags{}
	i.Parent.AddCommand(i.command)
	i.command.Flags().StringVarP(&i.flags.project, "project", "n", "", "The project project")
}

func (i *useProjectCommand) Exec(cmd *cobra.Command, args []string) error {
	log := context.GetDefaultLogger()
	if ns, err := kubernetes.NamespaceC(i.Client).Fetch(i.flags.project); err != nil {
		return fmt.Errorf("Error while trying to look for the project. Are you logged in? %s", err)
	} else if ns != nil {
		config := context.ReadConfig()
		config.Namespace = i.flags.project
		config.Save()

		log.Infof("Project set to '%s'", i.flags.project)

		if err := shared.SilentlyInstallOperatorIfNotExists(i.flags.project, "", i.Client); err != nil {
			return err
		}

		return nil
	}

	return fmt.Errorf("Project '%s' not found. Try running 'kogito new-project %s' to create your Project first", i.flags.project, i.flags.project)
}
