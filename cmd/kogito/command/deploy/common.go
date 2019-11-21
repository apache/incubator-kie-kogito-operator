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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	"github.com/spf13/cobra"
)

const (
	defaultDeployReplicas = 1
)

// CommonFlags is the base structure for resources that can be deployed in the cluster
type CommonFlags struct {
	Project  string
	Replicas int32
	Env      []string
	Limits   []string
	Requests []string
}

// AddDeployFlags adds the common deploy flags to the given command
func AddDeployFlags(command *cobra.Command, flags *CommonFlags) {
	command.Flags().StringVarP(&flags.Project, "project", "p", "", "The project name where the service will be deployed")
	command.Flags().Int32Var(&flags.Replicas, "replicas", defaultDeployReplicas, "Number of pod replicas that should be deployed.")
	command.Flags().StringSliceVarP(&flags.Env, "env", "e", nil, "Key/Pair value environment variables that will be set to the service runtime. For example 'MY_VAR=my_value'. Can be set more than once.")
	command.Flags().StringSliceVar(&flags.Limits, "limits", nil, "Resource limits for the Service pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
	command.Flags().StringSliceVar(&flags.Requests, "requests", nil, "Resource requests for the Service pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
}

// CheckDeployArgs checks the default deploy flags
func CheckDeployArgs(flags *CommonFlags) error {
	if err := util.ParseStringsForKeyPair(flags.Env); err != nil {
		return fmt.Errorf("environment variables are in the wrong format. Valid are key pairs like 'env=value', received %s", flags.Env)
	}
	if err := util.ParseStringsForKeyPair(flags.Limits); err != nil {
		return fmt.Errorf("limits are in the wrong format. Valid are key pairs like 'cpu=1', received %s", flags.Limits)
	}
	if err := util.ParseStringsForKeyPair(flags.Requests); err != nil {
		return fmt.Errorf("requests are in the wrong format. Valid are key pairs like 'cpu=1', received %s", flags.Requests)
	}
	if flags.Replicas <= 0 {
		return fmt.Errorf("valid replicas are non-zero, positive numbers, received %v", flags.Replicas)
	}
	return nil
}

// CheckImageTag checks the given image tag
func CheckImageTag(image string) error {
	if len(image) > 0 && !shared.DockerTagRegxCompiled.MatchString(image) {
		return fmt.Errorf("invalid name for image tag. Valid format is namespace/image-name:tag. Received %s", image)
	}
	return nil
}
