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
	"github.com/kiegroup/community-kogito-operator/cmd/kogito/command/message"
	"github.com/kiegroup/community-kogito-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/community-kogito-operator/core/client"
	"github.com/spf13/cobra"
)

type projectFlags struct {
	project                  string
	installDataIndex         bool
	installJobsService       bool
	installManagementConsole bool
}

func addProjectFlagsToCommand(command *cobra.Command, pFlags *projectFlags) {
	command.Flags().StringVarP(&pFlags.project, "project", "n", "", message.ProjectCurrentContext)
	command.Flags().BoolVar(&pFlags.installDataIndex, "install-data-index", false, message.InstallDataIndex)
	command.Flags().BoolVar(&pFlags.installJobsService, "install-jobs-service", false, message.InstallJobsService)
	command.Flags().BoolVar(&pFlags.installManagementConsole, "install-mgmt-console", false, message.InstallMgmtConsole)
}

func handleServicesInstallation(pFlags *projectFlags, cli *client.Client) error {
	install := shared.
		ServicesInstallationBuilder(cli, pFlags.project).
		CheckOperatorCRDs()

	if pFlags.installDataIndex {
		persistenceInfra := shared.GetDefaultPersistenceInfra(pFlags.project)
		install.InstallInfraService(persistenceInfra)
		messagingInfra := shared.GetDefaultMessagingInfra(pFlags.project)
		install.InstallInfraService(messagingInfra)
		dataIndex := shared.GetDefaultDataIndex(pFlags.project)
		install.InstallSupportingService(&dataIndex)
	}
	if pFlags.installJobsService {
		jobsService := shared.GetDefaultJobsService(pFlags.project)
		install.InstallSupportingService(&jobsService)
	}
	if pFlags.installManagementConsole {
		mgmtConsole := shared.GetDefaultMgmtConsole(pFlags.project)
		install.InstallSupportingService(&mgmtConsole)
	}

	return install.GetError()
}
