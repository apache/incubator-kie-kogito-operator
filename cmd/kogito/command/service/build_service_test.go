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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_InstallBuildService(t *testing.T) {
	ns := t.Name()
	resource := "https://github.com/kiegroup/kogito-examples/"
	buildFlag := flag.BuildFlags{
		Name:    "example-quarkus",
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
	client := test.SetupFakeKubeCli(&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})

	err := InstallBuildService(client, &buildFlag, resource)
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

func Test_DeleteBuildService_WhenBuildExists(t *testing.T) {
	ns := t.Name()
	name := "example-quarkus"
	client := test.SetupFakeKubeCli(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&v1alpha1.KogitoBuild{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns}})
	err := DeleteBuildService(client, name, ns)
	assert.NoError(t, err)
}

func Test_DeleteBuildService_WhenBuildNotExists(t *testing.T) {
	ns := t.Name()
	name := "example-quarkus"
	client := test.SetupFakeKubeCli(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	err := DeleteBuildService(client, name, ns)
	assert.Error(t, err)
}
