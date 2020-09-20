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

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/spf13/cobra"
)

type removeRuntimeServiceFlags struct {
	namespace string
}

type removableRuntimeService struct {
	name    string
	list    v1alpha1.KogitoServiceList
	aliases []string
}

type removeRuntimeServiceCommand struct {
	context.CommandContext
	flags          removeRuntimeServiceFlags
	command        *cobra.Command
	runtimeService removableRuntimeService
	Parent         *cobra.Command
}

var removableRuntimeServices = []removableRuntimeService{
	{
		name: "data-index",
		list: &v1alpha1.KogitoDataIndexList{},
	},
	{
		name:    "mgmt-console",
		list:    &v1alpha1.KogitoMgmtConsoleList{},
		aliases: []string{"management-console"},
	},
	{
		name: "jobs-service",
		list: &v1alpha1.KogitoJobsServiceList{},
	},
	{
		name: "explainability",
		list: &v1alpha1.KogitoExplainabilityList{},
	},
	{
		name: "trusty",
		list: &v1alpha1.KogitoTrustyList{},
	},
	{
		name: "trusty-ui",
		list: &v1alpha1.KogitoTrustyUIList{},
	},
}

func initRemoveRuntimeServiceCommands(ctx *context.CommandContext, parent *cobra.Command) []context.KogitoCommand {
	var commands []context.KogitoCommand
	for _, removable := range removableRuntimeServices {
		cmd := &removeRuntimeServiceCommand{
			CommandContext: *ctx,
			runtimeService: removable,
			Parent:         parent,
		}
		cmd.RegisterHook()
		cmd.InitHook()
		commands = append(commands, cmd)
	}
	return commands
}

func (r *removeRuntimeServiceCommand) RegisterHook() {
	r.command = &cobra.Command{
		Use:     r.runtimeService.name,
		Aliases: r.runtimeService.aliases,
		Short:   fmt.Sprintf("removes installed %s instance from the OpenShift/Kubernetes cluster", r.runtimeService.name),
		Example: fmt.Sprintf("remove %s -p my-project", r.runtimeService.name),
		Long:    fmt.Sprintf(`removes installed %s instance via custom Kubernetes resources.`, r.runtimeService.name),
		RunE:    r.Exec,
		PreRun:  r.CommonPreRun,
		PostRun: r.CommonPostRun,
		Args: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
}

func (r *removeRuntimeServiceCommand) InitHook() {
	r.flags = removeRuntimeServiceFlags{}
	r.Parent.AddCommand(r.command)
	r.command.Flags().StringVarP(&r.flags.namespace, "project", "p", "", fmt.Sprintf("The project name where the %s service is deployed", r.runtimeService.name))
}

func (r *removeRuntimeServiceCommand) Command() *cobra.Command {
	return r.command
}

func (r *removeRuntimeServiceCommand) Exec(cmd *cobra.Command, args []string) error {
	log := context.GetDefaultLogger()
	var err error
	if r.flags.namespace, err = shared.EnsureProject(r.Client, r.flags.namespace); err != nil {
		return err
	}
	err = kubernetes.ResourceC(r.Client).ListWithNamespace(r.flags.namespace, r.runtimeService.list)
	if err != nil {
		return err
	}
	if r.runtimeService.list.GetItemsCount() == 0 {
		log.Warnf("There's no service %s in the namespace %s", r.runtimeService.name, r.flags.namespace)
		return nil
	}
	for i := 0; i < r.runtimeService.list.GetItemsCount(); i++ {
		serviceName := r.runtimeService.list.GetItemAt(i).GetName()
		if err = kubernetes.ResourceC(r.Client).Delete(r.runtimeService.list.GetItemAt(i)); err != nil {
			return err
		}
		log.Infof("Service %s named %s from namespace %s has been successfully removed", r.runtimeService.name, serviceName, r.flags.namespace)
	}
	return nil
}
