// Copyright 2020 Red Hat, Inc. and/or its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not serviceName this file except in compliance with the License.
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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/spf13/cobra"
)

type removeSupportingServiceFlags struct {
	namespace string
}

type removableSupportingService struct {
	serviceName string
	serviceType v1alpha1.ServiceType
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
		serviceName: "data-index",
		serviceType: v1alpha1.DataIndex,
	},
	{
		serviceName: "mgmt-console",
		serviceType: v1alpha1.MgmtConsole,
		aliases:     []string{"management-console"},
	},
	{
		serviceName: "jobs-service",
		serviceType: v1alpha1.JobsService,
	},
	{
		serviceName: "explainability",
		serviceType: v1alpha1.Explainablity,
	},
	{
		serviceName: "trusty",
		serviceType: v1alpha1.TrustyAI,
	},
	{
		serviceName: "trusty-ui",
		serviceType: v1alpha1.TrustyUI,
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
		Use:     r.supportingService.serviceName,
		Aliases: r.supportingService.aliases,
		Short:   fmt.Sprintf("removes installed %s instance from the OpenShift/Kubernetes cluster", r.supportingService.serviceName),
		Example: fmt.Sprintf("remove %s -p my-project", r.supportingService.serviceName),
		Long:    fmt.Sprintf(`removes installed %s instance via custom Kubernetes resources.`, r.supportingService.serviceName),
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
	r.command.Flags().StringVarP(&r.flags.namespace, "project", "p", "", fmt.Sprintf("The project serviceName where the %s service is deployed", r.supportingService.serviceName))
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
	supportingServiceList := &v1alpha1.KogitoSupportingServiceList{}
	err = kubernetes.ResourceC(r.Client).ListWithNamespace(r.flags.namespace, supportingServiceList)
	if err != nil {
		return err
	}

	var targetServiceItems []v1alpha1.KogitoSupportingService
	for _, supportingService := range supportingServiceList.Items {
		if supportingService.Spec.ServiceType == r.supportingService.serviceType {
			targetServiceItems = append(targetServiceItems, supportingService)
		}
	}

	if len(targetServiceItems) == 0 {
		log.Warnf("There's no service %s in the namespace %s", r.supportingService.serviceName, r.flags.namespace)
		return nil
	}
	for _, targetService := range targetServiceItems {
		if err = kubernetes.ResourceC(r.Client).Delete(&targetService); err != nil {
			return fmt.Errorf("error occurs while delete Service %s from namespace %s. Error %s", r.supportingService.serviceName, targetService.Name, err)
		}
		log.Infof("Service %s named %s from namespace %s has been successfully removed", r.supportingService.serviceName, targetService.Name, targetService.Namespace)
	}
	return nil
}
