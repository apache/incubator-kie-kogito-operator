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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/util"
	"github.com/spf13/cobra"
)

// EnvVarFlags is common properties used to configure Env var
type EnvVarFlags struct {
	Env       []string
	SecretEnv []string
}

// AddEnvVarFlags adds the common env var flags to the given command
func AddEnvVarFlags(command *cobra.Command, flags *EnvVarFlags, envVarName, enVarShorthand string) {
	secretEnvName := fmt.Sprintf("%s-%s", "secret", envVarName)
	command.Flags().StringArrayVarP(&flags.Env, envVarName, enVarShorthand, nil, "Key/Pair value environment variables that will be set to the service. For example 'MY_VAR=my_value'. Can be set more than once.")
	command.Flags().StringArrayVar(&flags.SecretEnv, secretEnvName, nil, "Secret Key/Pair value environment variables that will be set to the service. For example 'MY_VAR=secretName#secretKey'. Can be set more than once.")
}

// CheckEnvVarArgs checks the EnvVarFlags flags
func CheckEnvVarArgs(flags *EnvVarFlags) error {
	if err := util.CheckKeyPair(flags.Env); err != nil {
		return fmt.Errorf("environment variables are in the wrong format. Valid are key pairs like 'env=value', received %s", flags.Env)
	}
	if err := util.CheckSecretKeyPair(flags.SecretEnv); err != nil {
		return fmt.Errorf("secret environment variables are in the wrong format. Valid are key pairs like 'env=secretName#secretKey', received %s", flags.SecretEnv)
	}
	return nil
}
