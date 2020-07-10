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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func Test_InstallBuildService_Success_OpenShiftCluster(t *testing.T) {
	ns := t.Name()
	name := "example-quarkus"
	resource := "https://github.com/kiegroup/kogito-examples/"
	buildFlag := flag.BuildFlags{
		Name:    name,
		Project: ns,
		GitSourceFlags: flag.GitSourceFlags{
			ContextDir: "drools-quarkus-example",
		},
		RuntimeTypeFlags: flag.RuntimeTypeFlags{
			Runtime: "springboot",
		},
		BuildImage:                "mydomain.io/mynamespace/builder-image-s2i:1.0",
		RuntimeImage:              "mydomain.io/mynamespace/runnable-image:1.0",
		MavenMirrorURL:            "http://172.18.0.1:8080/repository/local/",
		EnableMavenDownloadOutput: true,
		IncrementalBuild:          true,
	}
	client := test.CreateFakeClientOnOpenShift(nil, nil, nil)
	resourceCheckService := new(shared.ResourceCheckServiceMock)
	resourceCheckService.On("CheckKogitoBuildNotExists", client, name, ns).Return(nil)

	buildService := buildServiceImpl{
		resourceCheckService: resourceCheckService,
	}

	err := buildService.InstallBuildService(client, &buildFlag, resource)
	assert.NoError(t, err)

	// This should be created, given the command above
	kogitoBuild := &v1alpha1.KogitoBuild{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "example-quarkus",
			Namespace: ns,
		},
	}

	exist, err := kubernetes.ResourceC(client).Fetch(kogitoBuild)
	assert.NoError(t, err)
	assert.True(t, exist)
	assert.Equal(t, v1alpha1.RemoteSourceBuildType, kogitoBuild.Spec.Type)
	assert.Equal(t, false, kogitoBuild.Spec.DisableIncremental)
	assert.Equal(t, v1alpha1.SpringbootRuntimeType, kogitoBuild.Spec.Runtime)
	assert.Equal(t, "mydomain.io/mynamespace/builder-image-s2i:1.0", kogitoBuild.Spec.BuildImage.String())
	assert.Equal(t, "mydomain.io/mynamespace/runnable-image:1.0", kogitoBuild.Spec.RuntimeImage.String())
	assert.Equal(t, "http://172.18.0.1:8080/repository/local/", kogitoBuild.Spec.MavenMirrorURL)
	assert.Equal(t, true, kogitoBuild.Spec.EnableMavenDownloadOutput)
}

func Test_InstallBuildService_Failure_K8Cluster(t *testing.T) {
	resource := "https://github.com/kiegroup/kogito-examples/"
	buildFlag := flag.BuildFlags{}
	client := test.CreateFakeClient(nil, nil, nil)

	buildService := buildServiceImpl{}

	err := buildService.InstallBuildService(client, &buildFlag, resource)
	assert.Error(t, err)
}

func Test_DeleteBuildService_Success_WhenBuildExists(t *testing.T) {
	ns := t.Name()
	name := "example-quarkus"
	objects := []runtime.Object{
		&v1alpha1.KogitoBuild{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}},
	}
	client := test.CreateFakeClientOnOpenShift(objects, nil, nil)
	resourceCheckService := new(shared.ResourceCheckServiceMock)
	resourceCheckService.On("CheckKogitoBuildExists", client, name, ns).Return(nil)
	buildService := buildServiceImpl{
		resourceCheckService: resourceCheckService,
	}
	err := buildService.DeleteBuildService(client, name, ns)
	assert.NoError(t, err)
}

func Test_DeleteBuildService_Failure_WhenBuildNotExists(t *testing.T) {
	ns := t.Name()
	name := "example-quarkus"
	client := test.CreateFakeClientOnOpenShift(nil, nil, nil)
	resourceCheckService := new(shared.ResourceCheckServiceMock)
	resourceCheckService.On("CheckKogitoBuildExists", client, name, ns).Return(fmt.Errorf(""))
	buildService := buildServiceImpl{
		resourceCheckService: resourceCheckService,
	}
	err := buildService.DeleteBuildService(client, name, ns)
	assert.Error(t, err)
}

func Test_DeleteBuildService_Failure_K8Cluster(t *testing.T) {
	ns := t.Name()
	name := "example-quarkus"
	client := test.CreateFakeClient(nil, nil, nil)
	resourceCheckService := new(shared.ResourceCheckServiceMock)
	buildService := buildServiceImpl{
		resourceCheckService: resourceCheckService,
	}
	err := buildService.DeleteBuildService(client, name, ns)
	assert.NoError(t, err)
	resourceCheckService.AssertNotCalled(t, "CheckKogitoBuildExists", mock.Anything, mock.Anything, mock.Anything)
}
