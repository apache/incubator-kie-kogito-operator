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

package main

import (
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/spf13/cobra"
)

type newProjectFlags struct {
	name string
}

var newProjectCmd *cobra.Command
var newProjectCmdFlags = newProjectFlags{}

var _ = RegisterCommandVar(func() {
	newProjectCmd = &cobra.Command{
		Use:     "new-project NAME",
		Aliases: []string{"new-ns"},
		Short:   "Creates a new Kogito Project for your Kogito Services",
		Long: `new-project will create a Kubernetes Namespace with the provided name where your Kogito Services will be deployed. This project then will be used to deploy all infrastructure
				bits needed for the deployed Kogito Services to run.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return newProjectExec(cmd, args)
		},
		PreRun:  preRunF,
		PostRun: posRunF,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(newProjectCmdFlags.name) == 0 {
				if len(args) == 0 {
					return fmt.Errorf("Please set a name for new-project")
				}
				newProjectCmdFlags.name = args[0]
			}
			return nil
		},
	}
})

var _ = RegisterCommandInit(func() {
	rootCmd.AddCommand(newProjectCmd)
	newProjectCmd.Flags().StringVarP(&newProjectCmdFlags.name, "name", "n", "", "The project name")
})

func newProjectExec(cmd *cobra.Command, args []string) error {
	ns, err := kubernetes.NamespaceC(kubeCli).Fetch(newProjectCmdFlags.name)
	if err != nil {
		return err
	}
	if ns == nil {
		ns, err := kubernetes.NamespaceC(kubeCli).Create(newProjectCmdFlags.name)
		if err != nil {
			return err
		}
		config.Namespace = ns.Name
		log.Infof("Project '%s' created successfully", ns.Name)
	} else {
		log.Infof("Project '%s' already exists", newProjectCmdFlags.name)
	}
	return nil
}
