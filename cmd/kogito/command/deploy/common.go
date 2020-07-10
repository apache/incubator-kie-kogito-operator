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

package deploy

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/util"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/common"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/spf13/cobra"
)

const (
	defaultDeployReplicas = 1
)

// CommonFlags is the base structure for resources that can be deployed in the cluster
type CommonFlags struct {
	common.OperatorFlags
	Project               string
	Replicas              int32
	Env                   []string
	SecretEnv             []string
	Limits                []string
	Requests              []string
	HTTPPort              int32
	Image                 string
	InsecureImageRegistry bool
}

// AddDeployFlags adds the common deploy flags to the given command
func AddDeployFlags(command *cobra.Command, flags *CommonFlags) {
	common.AddOperatorFlags(command, &flags.OperatorFlags)
	command.Flags().StringVarP(&flags.Project, "project", "p", "", "The project name where the service will be deployed")
	command.Flags().Int32Var(&flags.Replicas, "replicas", defaultDeployReplicas, "Number of pod replicas that should be deployed.")
	command.Flags().StringArrayVarP(&flags.Env, "env", "e", nil, "Key/Pair value environment variables that will be set to the service runtime. For example 'MY_VAR=my_value'. Can be set more than once.")
	command.Flags().StringArrayVar(&flags.SecretEnv, "secret-env", nil, "Secret Key/Pair value environment variables that will be set to the service runtime. For example 'MY_VAR=secretName#secretKey'. Can be set more than once.")
	command.Flags().StringSliceVar(&flags.Limits, "limits", nil, "Resource limits for the Service pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
	command.Flags().StringSliceVar(&flags.Requests, "requests", nil, "Resource requests for the Service pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
	command.Flags().Int32Var(&flags.HTTPPort, "http-port", framework.DefaultExposedPort, "Define port on which service will listen internally")
	command.Flags().StringVarP(&flags.Image, "image", "i", "", "Image tag for the Service, example: quay.io/kiegroup/kogito-data-index:latest")
	command.Flags().BoolVar(&flags.InsecureImageRegistry, "insecure-image-registry", false, "Indicates that the Service image points to insecure image registry")
}

// CheckDeployArgs checks the default deploy flags
func CheckDeployArgs(flags *CommonFlags) error {
	if err := util.CheckKeyPair(flags.Env); err != nil {
		return fmt.Errorf("environment variables are in the wrong format. Valid are key pairs like 'env=value', received %s", flags.Env)
	}
	if err := util.CheckSecretKeyPair(flags.SecretEnv); err != nil {
		return fmt.Errorf("secret environment variables are in the wrong format. Valid are key pairs like 'env=secretName#secretKey', received %s", flags.SecretEnv)
	}
	if err := util.CheckKeyPair(flags.Limits); err != nil {
		return fmt.Errorf("limits are in the wrong format. Valid are key pairs like 'cpu=1', received %s", flags.Limits)
	}
	if err := util.CheckKeyPair(flags.Requests); err != nil {
		return fmt.Errorf("requests are in the wrong format. Valid are key pairs like 'cpu=1', received %s", flags.Requests)
	}
	if flags.Replicas <= 0 {
		return fmt.Errorf("valid replicas are non-zero, positive numbers, received %v", flags.Replicas)
	}
	if err := common.CheckOperatorArgs(&flags.OperatorFlags); err != nil {
		return err
	}
	return nil
}
