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

import "fmt"

var (
	// RuntimeServiceErrCreating ...
	RuntimeServiceErrCreating = fmt.Sprintf("Error while trying to create a new Kogito Service: %s", "%s")
	// RuntimeServiceSuccessfulInstalled ...
	RuntimeServiceSuccessfulInstalled = fmt.Sprintf("Kogito Service successfully installed in the Project %s.", "%s")
	// RuntimeServiceCheckStatus ...
	RuntimeServiceCheckStatus = fmt.Sprintf(serviceCheckStatus, "kogitoruntime", "%s", "%s")
	// RuntimeServiceNotInstalledNoKogitoOperator ...
	RuntimeServiceNotInstalledNoKogitoOperator = "Skipping deploy Kogito Service since Kogito Operator is not available."
	// RuntimeServiceMgmtConsole ...
	RuntimeServiceMgmtConsole = `To more easily manage your Kogito Service install Data Index Service and Process Instance Management. 
For how to install see: https://docs.jboss.org/kogito/release/latest/html_single/#con-kogito-operator-with-data-index-service_kogito-deploying-on-openshift
and https://docs.jboss.org/kogito/release/latest/html_single/#con-management-console_kogito-developing-process-services`
	// RuntimeServiceMgmtConsoleEndpoint ...
	RuntimeServiceMgmtConsoleEndpoint = `You can manage your process using the management console: %s`
)
