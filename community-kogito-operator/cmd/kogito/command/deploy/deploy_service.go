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
	"github.com/kiegroup/community-kogito-operator/cmd/kogito/command/context"
	"github.com/kiegroup/community-kogito-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/community-kogito-operator/cmd/kogito/command/service"
	"github.com/kiegroup/community-kogito-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/community-kogito-operator/core/client"
	"github.com/spf13/cobra"
)

type deployFlags struct {
	flag.BuildFlags
	flag.RuntimeFlags
	flag.RuntimeTypeFlags
}

type deployCommand struct {
	context.CommandContext
	command              *cobra.Command
	flags                *deployFlags
	Parent               *cobra.Command
	resourceCheckService shared.ResourceCheckService
	buildService         service.BuildService
	runtimeService       service.RuntimeService
}

// initDeployCommand is the constructor for the deploy command
func initDeployCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &deployCommand{
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

func (i *deployCommand) Command() *cobra.Command {
	return i.command
}

func (i *deployCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "deploy-service NAME [SOURCE]",
		Short:   "Deploys a new Kogito Service into the given Project",
		Aliases: []string{"deploy"},
		Long: `deploy-service will create a new Kogito service in the Project context.
	Providing a directory containing a pom.xml file in root will upload the whole directory for s2i build on the cluster.
	Providing a dmn/drl/bpmn/bpmn2 file or a directory containing one or more of those files as [SOURCE] will create a s2i build on the cluster.
	Providing a target directory (from mvn package) as [SOURCE] will directly upload the application binaries.
			
	Project context is the namespace (Kubernetes) or project (OpenShift) where the Service will be deployed.
	To know what's your context, use "kogito project". To set a new Project in the context use "kogito use-project NAME".
	Please note that this command requires the Kogito Operator installed in the cluster.
	For more information about the Kogito Operator installation please refer to https://github.com/kiegroup/community-kogito-operator#kogito-operator-installation.
		`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		// Args validation
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 2 {
				return fmt.Errorf("requires 2 args maximum, received %v", len(args))
			}
			if len(args) == 0 {
				return fmt.Errorf("the service requires a name ")
			}
			if err := flag.CheckRuntimeTypeArgs(&i.flags.RuntimeTypeFlags); err != nil {
				return err
			}
			if err := flag.CheckBuildArgs(&i.flags.BuildFlags); err != nil {
				return err
			}
			if err := flag.CheckRuntimeArgs(&i.flags.RuntimeFlags); err != nil {
				return err
			}
			return nil
		},
	}
}

func (i *deployCommand) InitHook() {
	i.Parent.AddCommand(i.command)
	i.flags = &deployFlags{}
	flag.AddBuildFlags(i.command, &i.flags.BuildFlags)
	flag.AddRuntimeFlags(i.command, &i.flags.RuntimeFlags)
	flag.AddRuntimeTypeFlags(i.command, &i.flags.RuntimeTypeFlags)
}

func (i *deployCommand) Exec(cmd *cobra.Command, args []string) (err error) {
	name := args[0]
	project, err := i.resourceCheckService.EnsureProject(i.Client, i.flags.RuntimeFlags.Project)
	if err != nil {
		return err
	}
	if err = i.installBuildService(i.Client, i.flags, name, project, args); err != nil {
		return err
	}
	if err = i.installRuntimeService(i.Client, i.flags, name, project); err != nil {
		return err
	}
	return nil
}

func (i *deployCommand) installBuildService(cli *client.Client, flags *deployFlags, name, project string, args []string) error {
	log := context.GetDefaultLogger()

	if !flags.ImageFlags.IsEmpty() {
		log.Info("Image details are provided, skipping to install kogito build")
		return nil
	}

	resource := ""
	if len(args) == 2 {
		resource = args[1]
	}

	log.Debug("Image details are not provided, going to install kogito build")
	flags.BuildFlags.Name = name
	flags.BuildFlags.Project = project
	flags.BuildFlags.RuntimeTypeFlags = flags.RuntimeTypeFlags
	return i.buildService.InstallBuildService(&flags.BuildFlags, resource)
}

func (i *deployCommand) installRuntimeService(cli *client.Client, flags *deployFlags, name, project string) error {
	flags.RuntimeFlags.Name = name
	flags.RuntimeFlags.Project = project
	flags.RuntimeFlags.RuntimeTypeFlags = flags.RuntimeTypeFlags
	return i.runtimeService.InstallRuntimeService(cli, &flags.RuntimeFlags)
}
