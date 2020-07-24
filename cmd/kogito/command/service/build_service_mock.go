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

package service

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/stretchr/testify/mock"
)

// BuildServiceMock is a mock struct for IBuildService interface
type BuildServiceMock struct {
	mock.Mock
}

// InstallBuildService is mock for IBuildService#InstallBuildService
func (b *BuildServiceMock) InstallBuildService(cli *client.Client, flags *flag.BuildFlags, resource string) (err error) {
	args := b.Called(cli, flags, resource)
	return args.Error(0)
}

// DeleteBuildService is mock for IBuildService#DeleteBuildService
func (b *BuildServiceMock) DeleteBuildService(cli *client.Client, name, project string) (err error) {
	args := b.Called(cli, name, project)
	return args.Error(0)
}
