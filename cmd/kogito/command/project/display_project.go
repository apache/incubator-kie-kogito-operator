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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"

	"github.com/spf13/cobra"
)

type displayProjectFlags struct {
}

type displayProjectCommand struct {
	context.CommandContext
	flags   displayProjectFlags
	command *cobra.Command
	Parent  *cobra.Command
}

func initDisplayProjectCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := displayProjectCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}
	cmd.RegisterHook()
	cmd.InitHook()
	return &cmd
}

func (i *displayProjectCommand) Command() *cobra.Command {
	return i.command
}

func (i *displayProjectCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "project",
		Aliases: []string{"ns"},
		Short:   "Display the current used project",
		Long:    `project will print the Kubernetes Namespace where the Kogito services will be deployed. It's the Namespace/Project on Kubernetes/OpenShift world.`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		Args:    func(cmd *cobra.Command, args []string) error { return nil },
	}
}

func (i *displayProjectCommand) InitHook() {
	i.flags = displayProjectFlags{}
	i.Parent.AddCommand(i.command)
}

func (i *displayProjectCommand) Exec(cmd *cobra.Command, args []string) error {
	log := context.GetDefaultLogger()
	currentProject := shared.GetCurrentNamespaceFromKubeConfig()
	if len(currentProject) > 0 {
		log.Infof(message.ProjectUsingProject, currentProject)
	} else {
		log.Info(message.ProjectNoProjectConfigured)
	}

	return nil
}
