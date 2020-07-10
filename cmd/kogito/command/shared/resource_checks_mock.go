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

package shared

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/stretchr/testify/mock"
)

// ResourceCheckServiceMock is a mock struct for IResourceCheckService interface
type ResourceCheckServiceMock struct {
	mock.Mock
}

// EnsureProject is mock for IResourceCheckService#EnsureProject
func (r *ResourceCheckServiceMock) EnsureProject(kubeCli *client.Client, project string) (string, error) {
	args := r.Called(kubeCli, project)
	return args.String(0), args.Error(1)
}

// CheckKogitoRuntimeExists is mock for IResourceCheckService#CheckKogitoRuntimeExists
func (r *ResourceCheckServiceMock) CheckKogitoRuntimeExists(kubeCli *client.Client, name string, namespace string) error {
	args := r.Called(kubeCli, name, namespace)
	return args.Error(0)
}

// CheckKogitoRuntimeNotExists is mock for IResourceCheckService#CheckKogitoRuntimeNotExists
func (r *ResourceCheckServiceMock) CheckKogitoRuntimeNotExists(kubeCli *client.Client, name string, namespace string) error {
	args := r.Called(kubeCli, name, namespace)
	return args.Error(0)
}

// CheckKogitoBuildExists is mock for IResourceCheckService#CheckKogitoBuildExists
func (r *ResourceCheckServiceMock) CheckKogitoBuildExists(kubeCli *client.Client, name string, project string) error {
	args := r.Called(kubeCli, name, project)
	return args.Error(0)
}

// CheckKogitoBuildNotExists is mock for IResourceCheckService#CheckKogitoBuildNotExists
func (r *ResourceCheckServiceMock) CheckKogitoBuildNotExists(kubeCli *client.Client, name string, project string) error {
	args := r.Called(kubeCli, name, project)
	return args.Error(0)
}
