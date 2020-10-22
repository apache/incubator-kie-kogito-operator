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

// RuntimeFlags is common properties used to configure Runtime service
type RuntimeFlags struct {
	InstallFlags
	RuntimeTypeFlags
	Name              string
	EnableIstio       bool
	EnablePersistence bool
	EnableEvents      bool
	ServiceLabels     []string
}

// AddRuntimeFlags adds the RuntimeFlags to the given command
func AddRuntimeFlags(command *cobra.Command, flags *RuntimeFlags) {
	AddInstallFlags(command, &flags.InstallFlags)
	command.Flags().BoolVar(&flags.EnableIstio, "enable-istio", false, "Enable Istio integration by annotating the Kogito service pods with the right value for Istio controller to inject sidecars on it. Defaults to false")
	command.Flags().BoolVar(&flags.EnablePersistence, "enable-persistence", false, "If set to true, deployed Kogito service will support integration with Infinispan server for persistence. Default to false")
	command.Flags().BoolVar(&flags.EnableEvents, "enable-events", false, "If set to true, deployed Kogito service will support integration with Kafka cluster for events. Default to false")
	command.Flags().StringSliceVar(&flags.ServiceLabels, "svc-labels", nil, "Labels that should be applied to the internal endpoint of the Kogito Service. Used by the service discovery engine. Example: 'label=value'. Can be set more than once.")
}

// CheckRuntimeArgs validates the RuntimeFlags flags
func CheckRuntimeArgs(flags *RuntimeFlags) error {
	if err := CheckInstallArgs(&flags.InstallFlags); err != nil {
		return err
	}
	if err := util.CheckKeyPair(flags.ServiceLabels); err != nil {
		return fmt.Errorf("service labels are in the wrong format. Valid are key pairs like 'service=myservice', received %s", flags.ServiceLabels)
	}
	return nil
}
