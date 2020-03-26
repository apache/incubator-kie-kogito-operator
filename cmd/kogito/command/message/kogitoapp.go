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

// Messages for KogitoApp deployment
const (
	KogitoAppSuccessfullyCreated       = "KogitoApp '%s' successfully created on namespace '%s'"
	KogitoAppViewDeploymentStatus      = "You can see the deployment status by using 'oc describe kogitoapp %s -n %s'"
	KogitoAppViewBuildStatus           = "Your Kogito Runtime Service should be deploying. To see its logs, run 'oc logs -f bc/%s-builder -n %s'"
	KogitoAppUploadBinariesInstruction = "Your Kogito Runtime Service needs the application binaries to proceed. To upload your binaries please run 'oc start-build %s-binary --from-dir=target -n %s' from your project's root"

	KogitoAppNoMgmtConsole = `To more easily manage your Kogito Runtime Service install Data Index Service and Process Instance Management. 
For how to install see: https://github.com/kiegroup/kogito-cloud-operator#kogito-data-index-service-deployment
and https://github.com/kiegroup/kogito-cloud-operator#kogito-management-console-install`
	KogitoAppMgmtConsoleEndpoint = `You can manage your process using the management console: %s`
)
