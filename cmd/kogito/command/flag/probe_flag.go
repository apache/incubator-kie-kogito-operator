// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package flag

import (
	"github.com/spf13/cobra"
)

// ProbeFlags is common properties used to configure probe for kogito service
type ProbeFlags struct {
	LivenessInitialDelay  int32
	ReadinessInitialDelay int32
}

// AddProbeFlags adds the ProbeFlags flags to the given command
func AddProbeFlags(command *cobra.Command, flags *ProbeFlags) {
	command.Flags().Int32Var(&flags.LivenessInitialDelay, "liveness-initial-delay", 0, "Number of seconds after the container has started before liveness probes are initiated. Default is 0")
	command.Flags().Int32Var(&flags.ReadinessInitialDelay, "readiness-initial-delay", 0, "Number of seconds after the container has started before readiness probes are initiated. Default is 0")
}

// CheckProbeArgs validates the ProbeFlags flags
func CheckProbeArgs(_ *ProbeFlags) error {
	return nil
}
