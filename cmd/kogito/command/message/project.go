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

package message

const (
	// ProjectUsingProject format in: project name
	ProjectUsingProject = "Using project '%s'"
	// ProjectNoProjectConfigured ...
	ProjectNoProjectConfigured = "No project configured yet..."
	// ProjectErrorGetProject ...
	ProjectErrorGetProject = "Error while trying to look for the project. Are you logged in? %s "
	// ProjectSet format in: project name
	ProjectSet = "Project set to '%s'"
	// ProjectNotFound in: project name, project name
	ProjectNotFound = "Project '%s' not found. Try running 'kogito new-project %s' to create your Project first "
	// ProjectNoContext ...
	ProjectNoContext = "Couldn't find any project in the current context. Make sure to connect to the cluster first or use `kogito use-project NAME` "
	// ProjectCantIdentifyContext ...
	ProjectCantIdentifyContext = "Can't identify the current context "
	// ProjectCurrentContextInfo in: current context
	ProjectCurrentContextInfo = "Project in the context is '%s'. Use 'kogito deploy-service NAME SOURCE' to deploy a new Kogito Service."
	// ProjectAlreadyExists in: project's name
	ProjectAlreadyExists = "Project '%s' already exists"
	// ProjectCreatedSuccessfully in: project's name
	ProjectCreatedSuccessfully = "Project '%s' created successfully"
	// ProjectCurrentContext --project / -p
	ProjectCurrentContext = "The project to be used in the current context"
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
)
