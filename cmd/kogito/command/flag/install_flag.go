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

package flag

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/spf13/cobra"
)

const (
	defaultDeployReplicas = 1
)

// InstallFlags is the base structure for resources that can be deployed in the cluster
type InstallFlags struct {
	OperatorFlags
	PodResourceFlags
	ImageFlags
	EnvVarFlags
	MonitoringFlags
	ConfigFlags
	Project  string
	Replicas int32
	HTTPPort int32
	Infra    []string
}

// AddInstallFlags adds the common deploy flags to the given command
func AddInstallFlags(command *cobra.Command, flags *InstallFlags) {
	AddOperatorFlags(command, &flags.OperatorFlags)
	AddPodResourceFlags(command, &flags.PodResourceFlags, "")
	AddImageFlags(command, &flags.ImageFlags)
	AddEnvVarFlags(command, &flags.EnvVarFlags, "env", "e")
	AddMonitoringFlags(command, &flags.MonitoringFlags)
	AddConfigFlags(command, &flags.ConfigFlags)
	command.Flags().StringVarP(&flags.Project, "project", "p", "", "The project name where the service will be deployed")
	command.Flags().Int32Var(&flags.Replicas, "replicas", defaultDeployReplicas, "Number of pod replicas that should be deployed.")
	command.Flags().Int32Var(&flags.HTTPPort, "http-port", framework.DefaultExposedPort, "Define port on which service will listen internally")
	command.Flags().StringArrayVar(&flags.Infra, "infra", nil, "Dependent KogitoInfra objects. Can be set more than once.")
}

// CheckInstallArgs checks the default deploy flags
func CheckInstallArgs(flags *InstallFlags) error {
	if err := CheckOperatorArgs(&flags.OperatorFlags); err != nil {
		return err
	}
	if err := CheckPodResourceArgs(&flags.PodResourceFlags); err != nil {
		return err
	}
	if err := CheckImageArgs(&flags.ImageFlags); err != nil {
		return err
	}
	if err := CheckEnvVarArgs(&flags.EnvVarFlags); err != nil {
		return err
	}
	if err := CheckMonitoringArgs(&flags.MonitoringFlags); err != nil {
		return err
	}
	if err := CheckConfigFlags(&flags.ConfigFlags); err != nil {
		return err
	}
	if flags.Replicas <= 0 {
		return fmt.Errorf("valid replicas are non-zero, positive numbers, received %v", flags.Replicas)
	}
	return nil
}
