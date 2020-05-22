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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/common"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/spf13/cobra"
)

type installInfinispanFlags struct {
	common.ChannelFlags
	namespace string
}

type installInfinispanCommand struct {
	context.CommandContext
	flags   installInfinispanFlags
	command *cobra.Command
	Parent  *cobra.Command
}

func initInstallInfinispanCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	command := installInfinispanCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}

	command.RegisterHook()
	command.InitHook()

	return &command
}

func (i *installInfinispanCommand) Command() *cobra.Command {
	return i.command
}

func (i *installInfinispanCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "infinispan [flags]",
		Short:   "Installs an infinispan instance into the OpenShift/Kubernetes cluster",
		Example: "install infinispan -p my-project",
		Long:    `Installs an infinispan instance via custom Kubernetes resources. This feature won't create custom subscriptions with the OLM.`,
		RunE:    i.Exec,
		PreRun:  i.CommonPreRun,
		PostRun: i.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := common.CheckChannelArgs(&i.flags.ChannelFlags); err != nil {
				return err
			}
			return nil
		},
	}
}

func (i *installInfinispanCommand) InitHook() {
	i.flags = installInfinispanFlags{
		ChannelFlags: common.ChannelFlags{},
	}
	i.Parent.AddCommand(i.command)
	common.AddChannelFlags(i.command, &i.flags.ChannelFlags)

	i.command.Flags().StringVarP(&i.flags.namespace, "project", "p", "", "The project name where the operator will be deployed")
}

func (i *installInfinispanCommand) Exec(cmd *cobra.Command, args []string) error {
	var err error
	if i.flags.namespace, err = shared.EnsureProject(i.Client, i.flags.namespace); err != nil {
		return err
	}

	installationChannel := shared.KogitoChannelType(i.flags.Channel)
	if installed, err := shared.SilentlyInstallOperatorIfNotExists(i.flags.namespace, "", i.Client, installationChannel); err != nil {
		return err
	} else if !installed {
		return nil
	}

	return shared.ServicesInstallationBuilder(i.Client, i.flags.namespace).InstallInfinispan().GetError()
}
