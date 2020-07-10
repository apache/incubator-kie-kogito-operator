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

package deploy

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/service"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func Test_DeleteServiceCmd_Success_OpenShiftCluster(t *testing.T) {
	ns := "default"
	name := "process-springboot-example"
	kubeCli := test.CreateFakeClientOnOpenShift(nil, nil, nil)
	resourceCheckServiceMock := new(shared.ResourceCheckServiceMock)
	buildService := new(service.BuildServiceMock)
	runtimeService := new(service.RuntimeServiceMock)

	resourceCheckServiceMock.On("EnsureProject", kubeCli, "").Return(ns, nil)
	buildService.On("DeleteBuildService", kubeCli, name, ns).Return(nil)
	runtimeService.On("DeleteRuntimeService", kubeCli, name, ns).Return(nil)

	deleteFlags := &deleteServiceFlags{}
	deleteServiceCmd := &deleteServiceCommand{
		CommandContext:       context.CommandContext{Client: kubeCli},
		flags:                deleteFlags,
		resourceCheckService: resourceCheckServiceMock,
		buildService:         buildService,
		runtimeService:       runtimeService,
	}

	args := []string{
		"process-springboot-example",
	}

	err := deleteServiceCmd.Exec(nil, args)
	assert.NoError(t, err)
	buildService.AssertCalled(t, "DeleteBuildService", kubeCli, name, ns)
	runtimeService.AssertCalled(t, "DeleteRuntimeService", kubeCli, name, ns)
}

func Test_DeleteServiceCmd_WhenProjectDoesNotExist(t *testing.T) {
	kubeCli := test.CreateFakeClient(nil, nil, nil)
	resourceCheckServiceMock := new(shared.ResourceCheckServiceMock)
	buildService := new(service.BuildServiceMock)
	runtimeService := new(service.RuntimeServiceMock)

	resourceCheckServiceMock.On("EnsureProject", kubeCli, "").Return("", fmt.Errorf(""))

	deleteFlags := &deleteServiceFlags{}

	deleteServiceCmd := &deleteServiceCommand{
		CommandContext:       context.CommandContext{Client: kubeCli},
		flags:                deleteFlags,
		resourceCheckService: resourceCheckServiceMock,
		buildService:         buildService,
		runtimeService:       runtimeService,
	}

	args := []string{
		"process-springboot-example",
	}

	err := deleteServiceCmd.Exec(nil, args)
	assert.Error(t, err)
	buildService.AssertNotCalled(t, "DeleteBuildService", mock.Anything, mock.Anything, mock.Anything)
	runtimeService.AssertNotCalled(t, "DeleteRuntimeService", mock.Anything, mock.Anything, mock.Anything)
}

func Test_DeleteServiceCmd_Error_DeleteKogitoRuntimeFailed(t *testing.T) {
	ns := "default"
	name := "process-springboot-example"
	kubeCli := test.CreateFakeClient(nil, nil, nil)
	resourceCheckServiceMock := new(shared.ResourceCheckServiceMock)
	buildService := new(service.BuildServiceMock)
	runtimeService := new(service.RuntimeServiceMock)

	resourceCheckServiceMock.On("EnsureProject", kubeCli, "").Return(ns, nil)
	runtimeService.On("DeleteRuntimeService", kubeCli, name, ns).Return(fmt.Errorf(""))

	deleteFlags := &deleteServiceFlags{}

	deleteServiceCmd := &deleteServiceCommand{
		CommandContext:       context.CommandContext{Client: kubeCli},
		flags:                deleteFlags,
		resourceCheckService: resourceCheckServiceMock,
		buildService:         buildService,
		runtimeService:       runtimeService,
	}

	args := []string{
		"process-springboot-example",
	}

	err := deleteServiceCmd.Exec(nil, args)
	assert.Error(t, err)
	buildService.AssertNotCalled(t, "DeleteBuildService", mock.Anything, mock.Anything, mock.Anything)
	runtimeService.AssertCalled(t, "DeleteRuntimeService", kubeCli, name, ns)
}
