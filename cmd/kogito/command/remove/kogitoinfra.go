// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package remove

import (
	"fmt"

	"github.com/apache/incubator-kie-kogito-operator/apis/app/v1beta1"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/context"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/shared"
	"github.com/apache/incubator-kie-kogito-operator/core/client/kubernetes"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deleteKogitoInfraFlags struct {
	name    string
	project string
}

func initDeleteKogitoInfraCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &deleteKogitoInfraServiceCommand{
		CommandContext:       *ctx,
		Parent:               parent,
		resourceCheckService: shared.NewResourceCheckService(),
	}
	cmd.RegisterHook()
	cmd.InitHook()
	return cmd
}

type deleteKogitoInfraServiceCommand struct {
	context.CommandContext
	command              *cobra.Command
	flags                *deleteKogitoInfraFlags
	Parent               *cobra.Command
	resourceCheckService shared.ResourceCheckService
}

func (i *deleteKogitoInfraServiceCommand) RegisterHook() {
	i.command = &cobra.Command{
		Example: "kogito-infra kogito-kafka --project kogito",
		Use:     "kogito-infra NAME [flags]",
		Short:   "remove Kogito infra service deployed in the OpenShift/Kubernetes cluster",
		Long:    `remove kogito-infra will exclude every OpenShift/Kubernetes resource created to deploy the Kogito Infra Service into the namespace.`,
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

func (i *deleteKogitoInfraServiceCommand) Command() *cobra.Command {
	return i.command
}

func (i *deleteKogitoInfraServiceCommand) InitHook() {
	i.flags = &deleteKogitoInfraFlags{}
	i.Parent.AddCommand(i.command)
	i.command.Flags().StringVarP(&i.flags.project, "project", "p", "", "The project context name from where the service needs to be deleted")
}

func (i *deleteKogitoInfraServiceCommand) Exec(_ *cobra.Command, args []string) (err error) {
	log := context.GetDefaultLogger()
	i.flags.name = args[0]
	if i.flags.project, err = i.resourceCheckService.EnsureProject(i.Client, i.flags.project); err != nil {
		return err
	}
	if err := i.resourceCheckService.CheckKogitoInfraExists(i.Client, i.flags.name, i.flags.project); err != nil {
		return err
	}
	log.Debugf("About to delete infra service %s in namespace %s", i.flags.name, i.flags.project)
	if err := kubernetes.ResourceC(i.Client).Delete(&v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      i.flags.name,
			Namespace: i.flags.project,
		},
	}); err != nil {
		return err
	}
	log.Infof("Successfully deleted Kogito Infra Service %s in the Project context %s", i.flags.name, i.flags.project)
	return nil
}
