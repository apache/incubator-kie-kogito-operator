// Copyright 2021 Red Hat, Inc. and/or its affiliates
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
	"errors"
	"github.com/spf13/cobra"
)

// TrustStoreFlags ...
type TrustStoreFlags struct {
	ConfigMapTrustStore string
	SecretPassword      string
}

// TODO: these flags operations SHOULD BE AN INTERFACE, having each structure to implement it!!

// AddTrustStoreFlags adds the TrustStoreFlags flags to the given command
func AddTrustStoreFlags(command *cobra.Command, flags *TrustStoreFlags) {
	command.Flags().StringVar(&flags.ConfigMapTrustStore, "truststore-configmap", "", "Name of the ConfigMap containing the custom JKS TrustStore for this service")
	command.Flags().StringVar(&flags.SecretPassword, "truststore-secret", "", "Name of the Secret containing the custom TrustStore password")
}

// CheckTrustStoreArgs validates the TrustStoreFlags flags
func CheckTrustStoreArgs(flags *TrustStoreFlags) error {
	if len(flags.ConfigMapTrustStore) == 0 && len(flags.SecretPassword) > 0 {
		return errors.New("If TrustStore Password is provided, the TrustStore ConfigMap (--truststore-configmap) also must be provided ")
	}
	return nil
}
