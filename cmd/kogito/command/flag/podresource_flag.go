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

package flag

import (
	"fmt"
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/util"
	"github.com/spf13/cobra"
)

// PodResourceFlags is common properties used to configure CPU resources
type PodResourceFlags struct {
	Limits   []string
	Requests []string
}

// AddPodResourceFlags adds the common resource flags to the given command
func AddPodResourceFlags(command *cobra.Command, flags *PodResourceFlags, prefix string) {
	limitName := "limits"
	requestName := "requests"
	if len(prefix) > 0 {
		limitName = fmt.Sprintf("%s-%s", prefix, limitName)
		requestName = fmt.Sprintf("%s-%s", prefix, requestName)
	}
	command.Flags().StringSliceVar(&flags.Limits, limitName, nil, "Resource limits for the pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
	command.Flags().StringSliceVar(&flags.Requests, requestName, nil, "Resource requests for the pod. Valid values are 'cpu' and 'memory'. For example 'cpu=1'. Can be set more than once.")
}

// CheckPodResourceArgs validates the resource flags
func CheckPodResourceArgs(flags *PodResourceFlags) error {
	if err := util.CheckKeyPair(flags.Limits); err != nil {
		return fmt.Errorf("limits are in the wrong format. Valid are key pairs like 'cpu=1', received %s", flags.Limits)
	}
	if err := util.CheckKeyPair(flags.Requests); err != nil {
		return fmt.Errorf("requests are in the wrong format. Valid are key pairs like 'cpu=1', received %s", flags.Requests)
	}
	return nil
}
