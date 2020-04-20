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

package client

import (
	"fmt"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	"github.com/stretchr/testify/assert"
	restclient "k8s.io/client-go/rest"
)

func TestMain(m *testing.M) {
	// safe backup to not jeopardize user's envs
	oldEnvVar := util.GetOSEnv(clientcmd.RecommendedConfigPathEnvVar, "")
	defer func() {
		os.Setenv(clientcmd.RecommendedConfigPathEnvVar, oldEnvVar)
	}()
	os.Setenv(clientcmd.RecommendedConfigPathEnvVar, "/tmp/config")

	result := m.Run()

	os.Exit(result)
}

func Test_simpleBuild(t *testing.T) {
	client := build(NewClientBuilder())

	assert.NotNil(t, client.ControlCli)
	assert.Nil(t, client.BuildCli)
	assert.Nil(t, client.Discovery)
	assert.Nil(t, client.ImageCli)
	assert.Nil(t, client.PrometheusCli)
	assert.Nil(t, client.DeploymentCli)
	assert.Nil(t, client.KubernetesExtensionCli)

}

func Test_useControllerClient(t *testing.T) {
	config := &restclient.Config{}
	controllerCli, err := newKubeClient(config)
	if err != nil {
		t.Fatal(fmt.Sprintf("Impossible to create new Kubernetes client: %v", err))
	}

	client := build(NewClientBuilder().UseControllerClient(controllerCli))

	assert.NotNil(t, client.ControlCli)
	assert.Equal(t, controllerCli, client.ControlCli)

	assert.Nil(t, client.BuildCli)
	assert.Nil(t, client.Discovery)
	assert.Nil(t, client.ImageCli)
	assert.Nil(t, client.PrometheusCli)
	assert.Nil(t, client.DeploymentCli)
	assert.Nil(t, client.KubernetesExtensionCli)

}

func Test_WithBuildClient(t *testing.T) {
	client := build(NewClientBuilder().WithBuildClient())

	assert.NotNil(t, client.ControlCli)
	assert.NotNil(t, client.BuildCli)
	assert.Nil(t, client.Discovery)
	assert.Nil(t, client.ImageCli)
	assert.Nil(t, client.PrometheusCli)
	assert.Nil(t, client.DeploymentCli)
	assert.Nil(t, client.KubernetesExtensionCli)
}

func Test_WithDiscoveryClient(t *testing.T) {
	client := build(NewClientBuilder().WithDiscoveryClient())

	assert.NotNil(t, client.ControlCli)
	assert.Nil(t, client.BuildCli)
	assert.NotNil(t, client.Discovery)
	assert.Nil(t, client.ImageCli)
	assert.Nil(t, client.PrometheusCli)
	assert.Nil(t, client.DeploymentCli)
	assert.Nil(t, client.KubernetesExtensionCli)
}

func Test_WithImageClient(t *testing.T) {
	client := build(NewClientBuilder().WithImageClient())

	assert.NotNil(t, client.ControlCli)
	assert.Nil(t, client.BuildCli)
	assert.Nil(t, client.Discovery)
	assert.NotNil(t, client.ImageCli)
	assert.Nil(t, client.PrometheusCli)
	assert.Nil(t, client.DeploymentCli)
	assert.Nil(t, client.KubernetesExtensionCli)
}

func Test_WithPrometheusClient(t *testing.T) {
	client := build(NewClientBuilder().WithPrometheusClient())

	assert.NotNil(t, client.ControlCli)
	assert.Nil(t, client.BuildCli)
	assert.Nil(t, client.Discovery)
	assert.Nil(t, client.ImageCli)
	assert.NotNil(t, client.PrometheusCli)
	assert.Nil(t, client.DeploymentCli)
	assert.Nil(t, client.KubernetesExtensionCli)
}

func Test_WithDeploymentClient(t *testing.T) {
	client := build(NewClientBuilder().WithDeploymentClient())

	assert.NotNil(t, client.ControlCli)
	assert.Nil(t, client.BuildCli)
	assert.Nil(t, client.Discovery)
	assert.Nil(t, client.ImageCli)
	assert.Nil(t, client.PrometheusCli)
	assert.NotNil(t, client.DeploymentCli)
	assert.Nil(t, client.KubernetesExtensionCli)
}

func Test_WithKubernetesClient(t *testing.T) {
	client := build(NewClientBuilder().WithKubernetesExtensionClient())

	assert.NotNil(t, client.ControlCli)
	assert.Nil(t, client.BuildCli)
	assert.Nil(t, client.Discovery)
	assert.Nil(t, client.ImageCli)
	assert.Nil(t, client.PrometheusCli)
	assert.Nil(t, client.DeploymentCli)
	assert.NotNil(t, client.KubernetesExtensionCli)
}

func Test_WithAllClients(t *testing.T) {
	client := build(NewClientBuilder().WithAllClients())

	assert.NotNil(t, client.ControlCli)
	assert.NotNil(t, client.BuildCli)
	assert.NotNil(t, client.Discovery)
	assert.NotNil(t, client.ImageCli)
	assert.NotNil(t, client.PrometheusCli)
	assert.NotNil(t, client.DeploymentCli)
	assert.NotNil(t, client.KubernetesExtensionCli)
}

func build(clientBuilder Builder) *Client {
	// Fix to avoid to read kube config from server
	cb := clientBuilder.(*builderStruct)
	if cb.config == nil {
		cb.config = &restclient.Config{}
	}

	// Build the client
	client, err := clientBuilder.Build()
	if err != nil {
		panic(err)
	}
	return client
}
