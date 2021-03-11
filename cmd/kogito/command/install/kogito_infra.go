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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/converter"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type infraFlags struct {
	flag.InfraResourceFlags
	flag.PropertiesFlag
	Name    string
	Project string
}

type infraCommand struct {
	context.CommandContext
	command              *cobra.Command
	flags                *infraFlags
	Parent               *cobra.Command
	resourceCheckService shared.ResourceCheckService
}

// initDeployCommand is the constructor for the deploy command
func initInfraCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &infraCommand{
		CommandContext:       *ctx,
		Parent:               parent,
		resourceCheckService: shared.NewResourceCheckService(),
	}

	cmd.RegisterHook()
	cmd.InitHook()

	return cmd
}

func (i *infraCommand) Command() *cobra.Command {
	return i.command
}

func (i *infraCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:   "infra NAME",
		Short: "Installs the Kogito Infra Service in the given Project",
		Long: `install infra will create a new Kogito infra service in the Project context. 
	Resource Namespace & Resource Name MUST be provided then Kogito Infra can refer to provided resources.
	
	Project context is the namespace (Kubernetes) or project (OpenShift) where the Service will be deployed.
	To know what's your context, use "kogito project". To set a new Project in the context use "kogito use-project NAME".
	Please note that this command requires the Kogito Operator installed in the cluster.
	For more information about the Kogito Operator installation please refer to https://github.com/kiegroup/kogito-cloud-operator#kogito-operator-installation.
		`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		// Args validation
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) > 1 {
				return fmt.Errorf("requires 1 arg, received %v", len(args))
			}
			if len(args) == 0 {
				return fmt.Errorf("the kogito infra service requires a name ")
			}
			if err := flag.CheckInfraResourceArgs(&i.flags.InfraResourceFlags); err != nil {
				return err
			}
			if err := flag.CheckPropertiesArgs(&i.flags.PropertiesFlag); err != nil {
				return err
			}
			return nil
		},
	}
}

func (i *infraCommand) InitHook() {
	i.Parent.AddCommand(i.command)
	i.flags = &infraFlags{}
	flag.AddInfraResourceFlags(i.command, &i.flags.InfraResourceFlags)
	flag.AddPropertiesFlags(i.command, &i.flags.PropertiesFlag)
	i.command.Flags().StringVarP(&i.flags.Project, "project", "p", "", "The project name where the service will be deployed")
}

func (i *infraCommand) Exec(cmd *cobra.Command, args []string) (err error) {
	log := context.GetDefaultLogger()
	log.Debugf("Installing Kogito Infra : %s", i.flags.Name)

	if i.flags.Project, err = i.resourceCheckService.EnsureProject(i.Client, i.flags.Project); err != nil {
		return err
	}

	kogitoInfra := v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      args[0],
			Namespace: i.flags.Project,
		},
		Spec: v1beta1.KogitoInfraSpec{
			Resource:        converter.FromInfraResourceFlagsToResource(&i.flags.InfraResourceFlags),
			InfraProperties: converter.FromPropertiesFlagToStringMap(&i.flags.PropertiesFlag),
		},
		Status: v1beta1.KogitoInfraStatus{
			Condition: v1beta1.KogitoInfraCondition{},
		},
	}

	log.Debugf("Trying to install Kogito Infra Service '%s'", kogitoInfra.Name)

	// Create the Kogito infra application
	return shared.
		ServicesInstallationBuilder(i.Client, i.flags.Project).
		CheckOperatorCRDs().
		InstallInfraResource(&kogitoInfra).
		GetError()
}
