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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	buildv1 "github.com/openshift/api/build/v1"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_decoratorForSourceBuilder_enableIncrementalBuild_Test(t *testing.T) {
	kogitoBuild := &v1alpha1.KogitoBuild{
		ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: "test"},
		Spec: v1alpha1.KogitoBuildSpec{
			DisableIncremental: false,
			Type:               "LocalSource",
		},
	}
	bc := &buildv1.BuildConfig{
		ObjectMeta: v12.ObjectMeta{
			Namespace: kogitoBuild.Namespace,
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{Resources: kogitoBuild.Spec.Resources},
		},
	}
	decoratorForSourceBuilder()(kogitoBuild, bc)

	assert.Equal(t, true, *bc.Spec.CommonSpec.Strategy.SourceStrategy.Incremental)
}
func Test_decoratorForSourceBuilder_disableIncrementalBuild_Test(t *testing.T) {
	kogitoBuild := &v1alpha1.KogitoBuild{
		ObjectMeta: v12.ObjectMeta{Name: "test", Namespace: "test"},
		Spec: v1alpha1.KogitoBuildSpec{
			DisableIncremental: true,
			Type:               "LocalSource",
		},
	}
	bc := &buildv1.BuildConfig{
		ObjectMeta: v12.ObjectMeta{
			Namespace: kogitoBuild.Namespace,
		},
		Spec: buildv1.BuildConfigSpec{
			CommonSpec: buildv1.CommonSpec{Resources: kogitoBuild.Spec.Resources},
		},
	}
	decoratorForSourceBuilder()(kogitoBuild, bc)

	assert.Equal(t, false, *bc.Spec.CommonSpec.Strategy.SourceStrategy.Incremental)
}
