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
	appsv1 "github.com/openshift/client-go/apps/clientset/versioned/typed/apps/v1"
	buildv1 "github.com/openshift/client-go/build/clientset/versioned/typed/build/v1"
	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
	"k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
	controllercli "sigs.k8s.io/controller-runtime/pkg/client"

	monclientv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
)

// NewClientBuilder creates a builder to setup the client
func NewClientBuilder() Builder {
	return &builderStruct{}
}

// Builder wraps information about what to create for a client before building it
type Builder interface {
	// UseConfig sets the restconfig to use for the different CLIs
	UseConfig(kubeconfig *restclient.Config) Builder
	// UseControllerClient sets a specific controllerclient
	UseControllerClient(controllerClient controllercli.Client) Builder
	// WithDiscoveryClient tells the builder to create the discovery client
	WithDiscoveryClient() Builder
	// WithBuildClient tells the builder to create the build client
	WithBuildClient() Builder
	// WithImageClient tells the builder to create the image client
	WithImageClient() Builder
	// WithPrometheusClient tells the builder to create the prometheus client
	WithPrometheusClient() Builder
	// WithDeploymentClient tells the builder to create the deployment client
	WithDeploymentClient() Builder
	// WithAllClients is a shortcut to tell the builder to create all clients
	WithAllClients() Builder
	// Build build the final client structure
	Build() (*Client, error)
}

// Builder wraps information about what to create for a client before building it
type builderStruct struct {
	config        *restclient.Config
	controllerCli controllercli.Client

	isDiscoveryClient  bool
	isBuildClient      bool
	isImageClient      bool
	isPrometheusClient bool
	isDeploymentClient bool
}

func (builder *builderStruct) UseConfig(kubeconfig *restclient.Config) Builder {
	builder.config = kubeconfig
	return builder
}

func (builder *builderStruct) UseControllerClient(controllerClient controllercli.Client) Builder {
	builder.controllerCli = controllerClient
	return builder
}

func (builder *builderStruct) WithDiscoveryClient() Builder {
	builder.isDiscoveryClient = true
	return builder
}

func (builder *builderStruct) WithBuildClient() Builder {
	builder.isBuildClient = true
	return builder
}

func (builder *builderStruct) WithImageClient() Builder {
	builder.isImageClient = true
	return builder
}

func (builder *builderStruct) WithPrometheusClient() Builder {
	builder.isPrometheusClient = true
	return builder
}

func (builder *builderStruct) WithDeploymentClient() Builder {
	builder.isDeploymentClient = true
	return builder
}

func (builder *builderStruct) WithAllClients() Builder {
	builder.WithBuildClient()
	builder.WithDiscoveryClient()
	builder.WithImageClient()
	builder.WithPrometheusClient()
	builder.WithDeploymentClient()
	return builder
}

func (builder *builderStruct) Build() (*Client, error) {
	client := &Client{}

	var err error

	config := builder.config
	if config == nil {
		config, err = buildKubeConnectionConfig()
		if err != nil {
			return nil, fmt.Errorf("Impossible to get Kubernetes local configuration: %v", err)
		}
	}

	client.ControlCli = builder.controllerCli
	if client.ControlCli == nil {
		client.ControlCli, err = newKubeClient(config)
		if err != nil {
			return nil, fmt.Errorf("Impossible to create new Kubernetes client: %v", err)
		}
	}

	if builder.isDiscoveryClient {
		client.Discovery, err = discovery.NewDiscoveryClientForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("Impossible to create new Discovery client: %v", err)
		}
	}
	if builder.isBuildClient {
		client.BuildCli, err = buildv1.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("Error getting build client: %v", err)
		}
	}
	if builder.isImageClient {
		client.ImageCli, err = imagev1.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("Error getting image client: %v", err)
		}
	}
	if builder.isPrometheusClient {
		client.PrometheusCli, err = monclientv1.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("Error getting prometheus client: %v", err)
		}
	}
	if builder.isDeploymentClient {
		client.DeploymentCli, err = appsv1.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("Error getting deployment client: %v", err)
		}
	}
	return client, nil
}
