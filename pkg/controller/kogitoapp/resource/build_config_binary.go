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

package resource

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	buildv1 "github.com/openshift/api/build/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	binaryBuildEnv = "BINARY_BUILD"
)

func newBuildConfigRuntimeBinary(kogitoApp *v1alpha1.KogitoApp) (buildConfig buildv1.BuildConfig) {
	buildConfig = getBCRuntime(kogitoApp, fmt.Sprintf("%s-%s", kogitoApp.Name, BuildVariantBinary), BuildVariantBinary)
	buildConfig.Spec.Source.Type = buildv1.BuildSourceBinary
	buildConfig.Spec.Strategy.SourceStrategy.Env = append(buildConfig.Spec.Strategy.SourceStrategy.Env, v1.EnvVar{Name: binaryBuildEnv, Value: "true"})
	// The comparator hits reconciliation if this are not set to empty values. TODO: fix on the operator-utils project
	buildConfig.Spec.Source.Binary = &buildv1.BinaryBuildSource{AsFile: ""}
	buildConfig.Spec.Triggers = []buildv1.BuildTriggerPolicy{}

	return buildConfig
}
