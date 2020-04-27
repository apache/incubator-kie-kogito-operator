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

package project

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message/flags"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/spf13/cobra"
)

type projectFlags struct {
	project                  string
	installDataIndex         bool
	installJobsService       bool
	installManagementConsole bool
	installAll               bool
}

func addProjectFlagsToCommand(command *cobra.Command, pFlags *projectFlags) {
	command.Flags().StringVarP(&pFlags.project, "project", "n", "", flags.ProjectCurrentContext)
	command.Flags().BoolVar(&pFlags.installDataIndex, "install-data-index", false, flags.InstallDataIndex)
	command.Flags().BoolVar(&pFlags.installJobsService, "install-jobs-service", false, flags.InstallJobsService)
	command.Flags().BoolVar(&pFlags.installManagementConsole, "install-mgmt-console", false, flags.InstallMgmtConsole)
	command.Flags().BoolVar(&pFlags.installAll, "install-all", false, flags.InstallAllServices)
}

func handleServicesInstallation(pFlags *projectFlags, cli *client.Client) error {
	install := shared.
		ServicesInstallationBuilder(cli, pFlags.project).
		SilentlyInstallOperatorIfNotExists().
		WarnIfDependenciesNotReady(pFlags.installDataIndex, pFlags.installDataIndex)

	if pFlags.installAll || pFlags.installDataIndex {
		install.InstallDataIndex(nil)
	}
	if pFlags.installAll || pFlags.installJobsService {
		install.InstallJobsService(nil)
	}
	if pFlags.installAll || pFlags.installManagementConsole {
		install.InstallMgmtConsole(nil)
	}
	return install.GetError()
}
