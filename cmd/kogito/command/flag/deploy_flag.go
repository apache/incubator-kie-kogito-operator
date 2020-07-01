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
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	"github.com/spf13/cobra"
)

const (
	defaultDeployReplicas = 1
)

// DeployFlags is the base structure for resources that can be deployed in the cluster
type DeployFlags struct {
	OperatorFlags
	PodResourceFlags
	Project  string
	Replicas int32
	Env      []string
	HTTPPort int32
}

// AddDeployFlags adds the common deploy flags to the given command
func AddDeployFlags(command *cobra.Command, flags *DeployFlags) {
	AddOperatorFlags(command, &flags.OperatorFlags)
	AddResourceFlags(command, &flags.PodResourceFlags)
	command.Flags().StringVarP(&flags.Project, "project", "p", "", "The project name where the service will be deployed")
	command.Flags().Int32Var(&flags.Replicas, "replicas", defaultDeployReplicas, "Number of pod replicas that should be deployed.")
	command.Flags().StringArrayVarP(&flags.Env, "env", "e", nil, "Key/Pair value environment variables that will be set to the service runtime. For example 'MY_VAR=my_value'. Can be set more than once.")
	command.Flags().Int32Var(&flags.HTTPPort, "http-port", framework.DefaultExposedPort, "Define port on which service will listen internally")
}

// CheckDeployArgs checks the default deploy flags
func CheckDeployArgs(flags *DeployFlags) error {
	if err := CheckOperatorArgs(&flags.OperatorFlags); err != nil {
		return err
	}
	if err := CheckResourceArgs(&flags.PodResourceFlags); err != nil {
		return err
	}
	if err := util.ParseStringsForKeyPair(flags.Env); err != nil {
		return fmt.Errorf("environment variables are in the wrong format. Valid are key pairs like 'env=value', received %s", flags.Env)
	}
	if flags.Replicas <= 0 {
		return fmt.Errorf("valid replicas are non-zero, positive numbers, received %v", flags.Replicas)
	}
	return nil
}
