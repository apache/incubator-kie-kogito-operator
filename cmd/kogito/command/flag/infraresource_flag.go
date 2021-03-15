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
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"

	"github.com/spf13/cobra"
)

// InfraResourceFlags is common properties used to configure Resource
type InfraResourceFlags struct {
	APIVersion        string
	Kind              string
	ResourceNamespace string
	ResourceName      string
}

// AddInfraResourceFlags adds the Resource flags to the given command
func AddInfraResourceFlags(command *cobra.Command, flags *InfraResourceFlags) {
	command.Flags().StringVar(&flags.APIVersion, "apiVersion", "", "API Version of referred Kubernetes resource for example, "+infrastructure.InfinispanAPIVersion)
	command.Flags().StringVar(&flags.Kind, "kind", "", "kind of referred Kubernetes resource for example, "+infrastructure.InfinispanKind)
	command.Flags().StringVar(&flags.ResourceNamespace, "resource-namespace", "", "Namespace where referred resource exists")
	command.Flags().StringVar(&flags.ResourceName, "resource-name", "", "Name of referred resource.")
}

// CheckInfraResourceArgs validates the InfraResourceFlags flags
func CheckInfraResourceArgs(flags *InfraResourceFlags) error {
	if len(flags.Kind) == 0 {
		return fmt.Errorf("kind can't be empty")
	}
	if len(flags.APIVersion) == 0 {
		return fmt.Errorf("apiVersion can't be empty")
	}
	if len(flags.ResourceName) == 0 {
		return fmt.Errorf("resource-name can't be empty")
	}
	return nil
}
