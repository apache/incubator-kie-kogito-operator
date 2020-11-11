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
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	"reflect"
)

func (m *sourceManager) getBuilderDecorator() decorator {
	if v1beta1.LocalSourceBuildType == m.kogitoBuild.Spec.Type {
		return decoratorForLocalSourceBuilder()
	}
	return decoratorForRemoteSourceBuilder()
}

func (m *sourceManager) GetRequestedResources() (map[reflect.Type][]resource.KubernetesResource, error) {
	resources := make(map[reflect.Type][]resource.KubernetesResource)
	builderBC := newBuildConfig(m.kogitoBuild, decoratorForSourceBuilder(), m.getBuilderDecorator())
	runtimeBC := newBuildConfig(m.kogitoBuild, decoratorForRuntimeBuilder(), decoratorForSourceRuntimeBuilder())
	builderIS := newOutputImageStreamForBuilder(&builderBC)
	runtimeIS, err := newOutputImageStreamForRuntime(&runtimeBC, m.kogitoBuild, m.client)
	if err != nil {
		return resources, err
	}
	if err := framework.SetOwner(m.kogitoBuild, m.scheme, &builderBC, &runtimeBC, &builderIS); err != nil {
		return resources, err
	}
	// the runtime ImageStream is a shared resource among other KogitoBuild instances and KogitoRuntime, we can't own it
	if err := framework.AddOwnerReference(m.kogitoBuild, m.scheme, runtimeIS); err != nil {
		return resources, err
	}
	resources[reflect.TypeOf(imgv1.ImageStream{})] = []resource.KubernetesResource{&builderIS, runtimeIS}
	resources[reflect.TypeOf(buildv1.BuildConfig{})] = []resource.KubernetesResource{&builderBC, &runtimeBC}
	return resources, nil
}

type sourceManager struct {
	manager
}
