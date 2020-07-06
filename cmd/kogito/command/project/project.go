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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/spf13/cobra"
)

type projectFlags struct {
	flag.OperatorFlags
	project                  string
	installDataIndex         bool
	installJobsService       bool
	installManagementConsole bool
	installAll               bool
	enablePersistence        bool
	enableEvents             bool
}

func addProjectFlagsToCommand(command *cobra.Command, pFlags *projectFlags) {
	flag.AddOperatorFlags(command, &pFlags.OperatorFlags)
	command.Flags().StringVarP(&pFlags.project, "project", "n", "", message.ProjectCurrentContext)
	command.Flags().BoolVar(&pFlags.installDataIndex, "install-data-index", false, message.InstallDataIndex)
	command.Flags().BoolVar(&pFlags.installJobsService, "install-jobs-service", false, message.InstallJobsService)
	command.Flags().BoolVar(&pFlags.installManagementConsole, "install-mgmt-console", false, message.InstallMgmtConsole)
	command.Flags().BoolVar(&pFlags.installAll, "install-all", false, message.InstallAllServices)
	command.Flags().BoolVar(&pFlags.enablePersistence, "enable-persistence", false, "If set will install Infinispan in the same namespace and inject the environment variables to configure the service connection to the Infinispan server.")
	command.Flags().BoolVar(&pFlags.enableEvents, "enable-events", false, "If set will install a Kafka cluster via the Strimzi Operator. ")
}

func handleServicesInstallation(pFlags *projectFlags, cli *client.Client) error {
	install := shared.
		ServicesInstallationBuilder(cli, pFlags.project).
		SilentlyInstallOperatorIfNotExists(shared.KogitoChannelType(pFlags.Channel)).
		WarnIfDependenciesNotReady(pFlags.installDataIndex, pFlags.installDataIndex)

	if pFlags.installAll || pFlags.installDataIndex {
		dataIndex := shared.GetDefaultDataIndex(pFlags.project)
		install.InstallDataIndex(&dataIndex)
	}
	if pFlags.installAll || pFlags.installJobsService {
		jobsService := shared.GetDefaultJobsService(pFlags.project, pFlags.enablePersistence, pFlags.enableEvents)
		install.InstallJobsService(&jobsService)
	}
	if pFlags.installAll || pFlags.installManagementConsole {
		mgmtConsole := shared.GetDefaultMgmtConsole(pFlags.project)
		install.InstallMgmtConsole(&mgmtConsole)
	}
	return install.GetError()
}
