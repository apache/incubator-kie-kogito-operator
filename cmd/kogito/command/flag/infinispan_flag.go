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
	"github.com/spf13/cobra"
)

// InfinispanFlags is common properties used to configure Infinispan
type InfinispanFlags struct {
	URI                string
	AuthRealm          string
	InfinispanSasl     string
	InfinispanUser     string
	InfinispanPassword string
	UseKogitoInfra     bool
}

// AddInfinispanFlags adds the infinispan flags to the given command
func AddInfinispanFlags(command *cobra.Command, flags *InfinispanFlags) {
	command.Flags().StringVar(&flags.URI, "infinispan-url", "", "Set only if enable-persistence is defined. The Infinispan Server URI, example: infinispan-server:11222")
	command.Flags().StringVar(&flags.AuthRealm, "infinispan-authrealm", "", "Set only if enable-persistence is defined. The Infinispan Server Auth Realm for authentication, example: ApplicationRealm")
	command.Flags().StringVar(&flags.InfinispanSasl, "infinispan-sasl", "", "Set only if enable-persistence is defined. The Infinispan Server SASL Mechanism, example: PLAIN")
	command.Flags().StringVar(&flags.InfinispanUser, "infinispan-user", "", "Set only if enable-persistence is defined. The Infinispan Server username")
	command.Flags().StringVar(&flags.InfinispanPassword, "infinispan-password", "", "Set only if enable-persistence is defined. The Infinispan Server password")
}

// CheckInfinispanArgs validates the InfinispanFlags flags
func CheckInfinispanArgs(flags *InfinispanFlags) error {
	if len(flags.InfinispanUser) > 0 && len(flags.InfinispanPassword) == 0 {
		return fmt.Errorf("infinispan-password wasn't provided, please set both infinispan-user and infinispan-password")
	}

	if len(flags.InfinispanUser) == 0 && len(flags.InfinispanPassword) > 0 {
		return fmt.Errorf("infinispan-user wasn't provided, please set both infinispan-user and infinispan-password")
	}

	if len(flags.URI) == 0 {
		if len(flags.InfinispanPassword) > 0 || len(flags.InfinispanUser) > 0 {
			return fmt.Errorf("Credentials given, but infinispan-url not set. Please set infinispan URL when providing credentials ")
		}
	}

	if len(flags.URI) > 0 {
		if len(flags.InfinispanPassword) == 0 || len(flags.InfinispanUser) == 0 {
			return fmt.Errorf("infinispan-url given, but Credentials not set. Please set Credentials when providing infinispan URL ")
		}
	}
	return nil
}
