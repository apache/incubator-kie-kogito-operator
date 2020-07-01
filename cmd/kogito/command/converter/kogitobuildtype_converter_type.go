// Copyright 2019 Red Hat, Inc. and/or its affiliates
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

package converter

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/util"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_FromResourceTypeToKogitoBuildType(t *testing.T) {
	assert.Equal(t, v1alpha1.LocalSourceBuildType, FromResourceTypeToKogitoBuildType(util.LocalFileResource))
	assert.Equal(t, v1alpha1.LocalSourceBuildType, FromResourceTypeToKogitoBuildType(util.LocalDirectoryResource))
	assert.Equal(t, v1alpha1.LocalSourceBuildType, FromResourceTypeToKogitoBuildType(util.GitFileResource))
	assert.Equal(t, v1alpha1.RemoteSourceBuildType, FromResourceTypeToKogitoBuildType(util.GitRepositoryResource))
	assert.Equal(t, v1alpha1.BinaryBuildType, FromResourceTypeToKogitoBuildType(util.BinaryResource))
}
