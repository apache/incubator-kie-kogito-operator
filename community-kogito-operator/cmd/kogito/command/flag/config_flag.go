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

// ConfigFlags is common properties used to configure application properties
type ConfigFlags struct {
	Config     []string
	ConfigFile string
}

// AddConfigFlags adds the common config flags to the given command
func AddConfigFlags(command *cobra.Command, flags *ConfigFlags) {
	command.Flags().StringArrayVar(&flags.Config, "config", nil, "Custom application properties that will be set to the service. For example 'MY_VAR=my_value'. Can be set more than once.")
	command.Flags().StringVar(&flags.ConfigFile, "config-file", "", "Path for custom application properties file to be deployed with the service. This file will be mounted as an external ConfigMap to the service Deployment.")
}

// CheckConfigFlags checks the ConfigFlags flags
func CheckConfigFlags(flags *ConfigFlags) error {
	if err := util.CheckKeyPair(flags.Config); err != nil {
		return fmt.Errorf("config are in the wrong format. Valid are key pairs like 'key=value', received %s", flags.Config)
	}
	if err := checkConfigFile(flags.ConfigFile); err != nil {
		return err
	}
	return nil
}

func checkConfigFile(configFilePath string) error {
	if len(configFilePath) > 0 {
		if exists, err := util.CheckFileExists(configFilePath); err != nil {
			return err
		} else if !exists {
			return fmt.Errorf("configuration file does not exist in the given path: %s", configFilePath)
		}
	}
	return nil
}
