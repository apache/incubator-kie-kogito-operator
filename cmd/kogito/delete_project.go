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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deleteProjectFlags struct {
	name string
}

var deleteProjectCommand *cobra.Command
var deleteProjectCmdFlags = deleteProjectFlags{}

var _ = RegisterCommandVar(func() {
	deleteProjectCommand = &cobra.Command{
		Example: "delete-project kogito",
		Use:     "delete-project NAME",
		Short:   "Deletes a Kogito Project - i.e., the Kubernetes/OpenShift namespace",
		Long:    `delete-project will exclude the namespace/project entirely, including all deployed services and infrastructure.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return deleteProjectExec(cmd, args)
		},
		PreRun:  preRunF,
		PostRun: posRunF,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg, received %v", len(args))
			}
			return nil
		},
	}
})

var _ = RegisterCommandInit(func() {
	rootCmd.AddCommand(deleteProjectCommand)
})

func deleteProjectExec(cmd *cobra.Command, args []string) error {
	deleteProjectCmdFlags.name = args[0]
	var err error
	if deleteProjectCmdFlags.name, err = checkProjectLocally(deleteProjectCmdFlags.name); err != nil {
		return err
	}
	if err := checkProjectExists(deleteProjectCmdFlags.name); err != nil {
		return err
	}
	log.Debugf("Using project %s", deleteProjectCmdFlags.name)
	log.Debugf("About to delete namespace %s", deleteProjectCmdFlags.name)
	if err := kubernetes.ResourceC(kubeCli).Delete(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: deleteProjectCmdFlags.name}}); err != nil {
		return err
	}

	log.Infof("Successfully deleted Kogito Project %s", deleteProjectCmdFlags.name)

	return nil
}
