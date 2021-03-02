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
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/api"
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/core/framework/util"
	"github.com/spf13/cobra"
)

const defaultDeployRuntimeType = string(api.QuarkusRuntimeType)

var (
	runtimeTypeValidEntries = []string{string(api.QuarkusRuntimeType), string(api.SpringBootRuntimeType)}
)

// RuntimeTypeFlags is common properties used to configure runtime properties
type RuntimeTypeFlags struct {
	Runtime string
}

// AddRuntimeTypeFlags adds the Runtime flags to the given command
func AddRuntimeTypeFlags(command *cobra.Command, flags *RuntimeTypeFlags) {
	command.Flags().StringVarP(&flags.Runtime, "runtime", "r", defaultDeployRuntimeType, "The runtime which should be used to build the Service. Valid values are 'quarkus' or 'springboot'. Default to '"+defaultDeployRuntimeType+"'.")
}

// CheckRuntimeTypeArgs validates the RuntimeTypeFlags flags
func CheckRuntimeTypeArgs(flags *RuntimeTypeFlags) error {
	if !util.Contains(flags.Runtime, runtimeTypeValidEntries) {
		return fmt.Errorf("runtime not valid. Valid runtimes are %s. Received %s", runtimeTypeValidEntries, flags.Runtime)
	}
	return nil
}
