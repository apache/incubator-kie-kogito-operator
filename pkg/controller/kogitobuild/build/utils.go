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

package build

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1beta1"
	buildv1 "github.com/openshift/api/build/v1"
	"strconv"
	"strings"
)

const (
	// builderSuffix suffix added to builders objects, that's those responsible to build the service from source
	builderSuffix = "-builder"
)

// GetApplicationName gets the Kogito service name
func GetApplicationName(build *v1beta1.KogitoBuild) string {
	if len(build.Spec.TargetKogitoRuntime) == 0 {
		return build.Name
	}
	return build.Spec.TargetKogitoRuntime
}

// GetBuildBuilderName gets the BuildConfig builder name
func GetBuildBuilderName(build *v1beta1.KogitoBuild) string {
	return strings.Join([]string{build.Name, builderSuffix}, "")
}

// getBuilderLimitsAsIntString gets the string representation for the resource limits defined in the given bc
func getBuilderLimitsAsIntString(bc *buildv1.BuildConfig) (limitCPU, limitMemory string) {
	limitCPU = ""
	limitMemory = ""
	if bc.Spec.Resources.Limits == nil {
		return "", ""
	}
	limitMemoryInt, possible := bc.Spec.Resources.Limits.Memory().AsInt64()
	if !possible {
		limitMemoryInt = bc.Spec.Resources.Limits.Memory().ToDec().AsDec().UnscaledBig().Int64()
	}
	if limitMemoryInt > 0 {
		limitMemory = strconv.FormatInt(limitMemoryInt, 10)
	}
	limitCPU = bc.Spec.Resources.Limits.Cpu().String()
	return limitCPU, limitMemory
}
