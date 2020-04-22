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

package flags

const (
	// InstallDataIndex --install-data-index
	InstallDataIndex = `Installs the default instance of Data Index being provisioned by the Kogito Operator in the project. 
For a more customized Data Index installation, use 'kogito install data-index [OPTIONS]'`
	// InstallJobsService --install-jobs-service
	InstallJobsService = `Installs the default instance of Jobs Service being provisioned by the Kogito Operator in the project.
For a more customized Jobs Service installation, use 'kogito install jobs-service [OPTIONS]'`
	// InstallMgmtConsole --install-mgmt-console
	InstallMgmtConsole = `Installs the default instance of Jobs Service being provisioned by the Kogito Operator in the project.
For a more customized Jobs Service installation, use 'kogito install jobs-service [OPTIONS]'`
	// InstallAllServices --install-all
	InstallAllServices = `Installs the default instance of all Kogito Support Services being provisioned by the Kogito Operator in the project.
Avoid installing the default services on production environments. Prefer the command 'kogito install [SERVICE]' since it can be customized.`
	// ProjectCurrentContext --project / -p
	ProjectCurrentContext = "The project to be used in the current context"
)
