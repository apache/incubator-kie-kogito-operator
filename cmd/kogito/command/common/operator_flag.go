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

package common

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/spf13/cobra"
)

// OperatorFlags describes the command lines flags for the operator features
type OperatorFlags struct {
	Channel string
}

// AddOperatorFlags adds the OperatorFlags to the given command
func AddOperatorFlags(command *cobra.Command, oFlags *OperatorFlags) {
	command.Flags().StringVar(&oFlags.Channel, "channel", string(shared.GetDefaultChannel()), "Install Kogito operator from Operator hub using provided channel, e.g. (alpha/dev-preview)")
}

// CheckOperatorArgs verifies the given arguments to the OperatorFlags
func CheckOperatorArgs(oFlags *OperatorFlags) error {
	ch := oFlags.Channel
	if !shared.IsChannelValid(ch) {
		return fmt.Errorf("Invalid Kogito channel type %s, only alpha/dev-preview channels are allowed ", ch)
	}
	return nil
}
