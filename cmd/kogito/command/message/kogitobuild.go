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
	// KogitoBuildViewDeploymentStatus ...
	KogitoBuildViewDeploymentStatus = "You can see the deployment status by using 'oc describe kogitobuild %s -n %s'"
	// KogitoViewBuildStatus ...
	KogitoViewBuildStatus = "Your Kogito Build should be deploying. To see its logs, run 'oc logs -f bc/%s-builder -n %s'"
	// KogitoBuildSuccessfullyUploadedFile ...
	KogitoBuildSuccessfullyUploadedFile = "The requested file(s) was successfully uploaded to OpenShift, the build %s with this file(s) should now be running. To see the logs, run 'oc logs -f bc/%s-builder -n %s'"
	// KogitoBuildUploadBinariesInstruction ...
	KogitoBuildUploadBinariesInstruction = "Your Kogito Runtime Service needs the application binaries to proceed. To upload your binaries please run 'oc start-build %s-binary --from-dir=target -n %s' from your project's root"
	// KogitoBuildFoundFile ...
	KogitoBuildFoundFile = "File(s) found: %s."
	// KogitoBuildFoundAsset ...
	KogitoBuildFoundAsset = "Asset found: %s."
	// KogitoBuildProvidedFileIsDir ...
	KogitoBuildProvidedFileIsDir = "The provided source is a directory, packing files."
	// KogitoBuildFileWalkingError ...
	KogitoBuildFileWalkingError = "Error while walking through %s directory: %s"
	// BuildServiceErrCreating ...
	BuildServiceErrCreating = fmt.Sprintf(serviceErrCreating, "Build", "%s")
	// BuildServiceSuccessfulInstalled ...
	BuildServiceSuccessfulInstalled = fmt.Sprintf(serviceSuccessfulInstalled, "Build", "%s")
	// BuildServiceCheckStatus ...
	BuildServiceCheckStatus = fmt.Sprintf(serviceCheckStatus, "kogitobuild", "%s", "%s")
	// BuildServiceNotInstalledNoKogitoOperator ...
	BuildServiceNotInstalledNoKogitoOperator = fmt.Sprintf("Skipping deploy %s since Kogito Operator is not available.", "Build Service")
	// BuildTriggeringNewBuild ...
	BuildTriggeringNewBuild = "Triggering the new build"
)
