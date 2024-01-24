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

package project

import (
	"fmt"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/context"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/shared"
	"github.com/apache/incubator-kie-kogito-operator/core/client/kubernetes"
	"github.com/spf13/cobra"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deleteProjectFlags struct {
	name string
}

type deleteProjectCommand struct {
	context.CommandContext
	flags   deleteProjectFlags
	command *cobra.Command
	Parent  *cobra.Command
}

func initDeleteProjectCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := deleteProjectCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}
	cmd.RegisterHook()
	cmd.InitHook()
	return &cmd
}

func (i *deleteProjectCommand) Command() *cobra.Command {
	return i.command
}

func (i *deleteProjectCommand) RegisterHook() {
	i.command = &cobra.Command{
		Example: "delete-project kogito",
		Use:     "delete-project NAME",
		Short:   "Deletes a Kogito Project - i.e., the Kubernetes/OpenShift project",
		Long:    `delete-project will exclude the project/project entirely, including all deployed services and infrastructure.`,
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

func (i *deleteProjectCommand) InitHook() {
	i.flags = deleteProjectFlags{}
	i.Parent.AddCommand(i.command)
}

func (i *deleteProjectCommand) Exec(_ *cobra.Command, args []string) error {
	log := context.GetDefaultLogger()
	i.flags.name = args[0]
	var err error
	if i.flags.name, err = shared.EnsureProject(i.Client, i.flags.name); err != nil {
		return err
	}

	log.Debugf("About to delete project %s", i.flags.name)
	if err := kubernetes.ResourceC(i.Client).Delete(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: i.flags.name}}); err != nil {
		return err
	}

	log.Infof("Successfully deleted Kogito Project %s", i.flags.name)

	return nil
}
