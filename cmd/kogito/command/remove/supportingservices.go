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

package remove

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	"github.com/spf13/cobra"
)

type removeSupportingServiceFlags struct {
	namespace string
}

type removableSupportingService struct {
	cmdName     string
	serviceType api.ServiceType
	aliases     []string
}

type removeSupportingServiceCommand struct {
	context.CommandContext
	flags             removeSupportingServiceFlags
	command           *cobra.Command
	supportingService removableSupportingService
	Parent            *cobra.Command
}

var removableSupportingServices = []removableSupportingService{
	{
		cmdName:     "data-index",
		serviceType: api.DataIndex,
	},
	{
		cmdName:     "explainability",
		serviceType: api.Explainability,
	},
	{
		cmdName:     "jobs-service",
		serviceType: api.JobsService,
	},
	{
		cmdName:     "mgmt-console",
		serviceType: api.MgmtConsole,
		aliases:     []string{"management-console"},
	},
	{
		cmdName:     "task-console",
		serviceType: api.TaskConsole,
	},
	{
		cmdName:     "trusty",
		serviceType: api.TrustyAI,
	},
	{
		cmdName:     "trusty-ui",
		serviceType: api.TrustyUI,
	},
}

func initRemoveSupportingServiceCommands(ctx *context.CommandContext, parent *cobra.Command) []context.KogitoCommand {
	var commands []context.KogitoCommand
	for _, removable := range removableSupportingServices {
		cmd := &removeSupportingServiceCommand{
			CommandContext:    *ctx,
			supportingService: removable,
			Parent:            parent,
		}
		cmd.RegisterHook()
		cmd.InitHook()
		commands = append(commands, cmd)
	}
	return commands
}

func (r *removeSupportingServiceCommand) RegisterHook() {
	r.command = &cobra.Command{
		Use:     r.supportingService.cmdName,
		Aliases: r.supportingService.aliases,
		Short:   fmt.Sprintf("removes installed %s instance from the OpenShift/Kubernetes cluster", r.supportingService.cmdName),
		Example: fmt.Sprintf("remove %s -p my-project", r.supportingService.cmdName),
		Long:    fmt.Sprintf(`removes installed %s instance via custom Kubernetes resources.`, r.supportingService.cmdName),
		RunE:    r.Exec,
		PreRun:  r.CommonPreRun,
		PostRun: r.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func (r *removeSupportingServiceCommand) InitHook() {
	r.flags = removeSupportingServiceFlags{}
	r.Parent.AddCommand(r.command)
	r.command.Flags().StringVarP(&r.flags.namespace, "project", "p", "", fmt.Sprintf("The project use where the %s service is deployed", r.supportingService.cmdName))
}

func (r *removeSupportingServiceCommand) Command() *cobra.Command {
	return r.command
}

func (r *removeSupportingServiceCommand) Exec(cmd *cobra.Command, args []string) error {
	log := context.GetDefaultLogger()
	var err error
	if r.flags.namespace, err = shared.EnsureProject(r.Client, r.flags.namespace); err != nil {
		return err
	}
	supportingServiceList := &v1beta1.KogitoSupportingServiceList{}
	err = kubernetes.ResourceC(r.Client).ListWithNamespace(r.flags.namespace, supportingServiceList)
	if err != nil {
		return err
	}

	var targetServiceItems []v1beta1.KogitoSupportingService
	for _, supportingService := range supportingServiceList.Items {
		if supportingService.Spec.ServiceType == r.supportingService.serviceType {
			targetServiceItems = append(targetServiceItems, supportingService)
		}
	}

	if len(targetServiceItems) == 0 {
		log.Warnf("There's no service %s in the namespace %s", r.supportingService.cmdName, r.flags.namespace)
		return nil
	}
	for _, targetService := range targetServiceItems {
		if err = kubernetes.ResourceC(r.Client).Delete(&targetService); err != nil {
			return fmt.Errorf("error occurs while delete Service %s from namespace %s. Error %s", r.supportingService.cmdName, targetService.Name, err)
		}
		log.Infof("Service %s named %s from namespace %s has been successfully removed", r.supportingService.cmdName, targetService.Name, targetService.Namespace)
	}
	return nil
}
