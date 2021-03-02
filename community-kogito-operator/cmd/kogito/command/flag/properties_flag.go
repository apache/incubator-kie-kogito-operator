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

	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/cmd/kogito/command/util"
	"github.com/spf13/cobra"
)

// PropertiesFlag is common properties
type PropertiesFlag struct {
	Properties []string
}

// AddPropertiesFlags adds the Properties flags to the given command
func AddPropertiesFlags(command *cobra.Command, flags *PropertiesFlag) {
	command.Flags().StringArrayVarP(&flags.Properties, "property", "", nil, "Key/Pair value data that will be taken as properties. For example 'MY_VAR=my_value'. Can be set more than once.")
}

// CheckPropertiesArgs validates the PropertiesFlag flags
func CheckPropertiesArgs(flags *PropertiesFlag) error {
	if err := util.CheckKeyPair(flags.Properties); err != nil {
		return fmt.Errorf("property data is in the wrong format. Valid are key pairs like 'env=value', received %s", flags.Properties)
	}
	return nil
}
