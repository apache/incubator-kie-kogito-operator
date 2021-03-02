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

package deploy

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/cmd/kogito/command/service"
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/cmd/kogito/command/shared"
	"github.com/spf13/cobra"
)

type deleteServiceFlags struct {
	name    string
	project string
}

func initDeleteServiceCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &deleteServiceCommand{
		CommandContext:       *ctx,
		Parent:               parent,
		resourceCheckService: shared.NewResourceCheckService(),
		buildService:         service.NewBuildService(ctx.Client),
		runtimeService:       service.NewRuntimeService(),
	}
	cmd.RegisterHook()
	cmd.InitHook()
	return cmd
}

type deleteServiceCommand struct {
	context.CommandContext
	command              *cobra.Command
	flags                *deleteServiceFlags
	Parent               *cobra.Command
	resourceCheckService shared.ResourceCheckService
	buildService         service.BuildService
	runtimeService       service.RuntimeService
}

func (i *deleteServiceCommand) RegisterHook() {
	i.command = &cobra.Command{
		Example: "delete-service example-drools --project kogito",
		Use:     "delete-service NAME [flags]",
		Short:   "Deletes a Kogito service deployed in the namespace/project",
		Long:    `delete-service will exclude every OpenShift/Kubernetes resource created to deploy the Kogito Service into the namespace.`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg, received %v", len(args))
			}
			return nil
		},
	}
}

func (i *deleteServiceCommand) Command() *cobra.Command {
	return i.command
}

func (i *deleteServiceCommand) InitHook() {
	i.flags = &deleteServiceFlags{}
	i.Parent.AddCommand(i.command)
	i.command.Flags().StringVarP(&i.flags.project, "project", "p", "", "The project name from where the service needs to be deleted")
}

func (i *deleteServiceCommand) Exec(cmd *cobra.Command, args []string) (err error) {
	i.flags.name = args[0]
	if i.flags.project, err = i.resourceCheckService.EnsureProject(i.Client, i.flags.project); err != nil {
		return err
	}
	if err = i.runtimeService.DeleteRuntimeService(i.Client, i.flags.name, i.flags.project); err != nil {
		return err
	}
	if err = i.buildService.DeleteBuildService(i.flags.name, i.flags.project); err != nil {
		return err
	}
	return nil
}
