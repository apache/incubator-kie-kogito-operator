// Copyright 2019 Red Hat, Inc. and/or its affiliates
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

package main

import (
	"github.com/kiegroup/kogito-cloud-operator/version"
	"github.com/spf13/cobra"
)

var versionCmd *cobra.Command

var _ = RegisterCommandVar(func() {
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Prints the kogito CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			return versionExec(cmd, args)
		},
		PreRun: preRunF,
	}
})

var _ = RegisterCommandInit(func() {
	rootCmd.AddCommand(versionCmd)
})

func versionExec(cmd *cobra.Command, args []string) error {
	log.Infof("Kogito CLI version: %s", version.Version)
	return nil
}
