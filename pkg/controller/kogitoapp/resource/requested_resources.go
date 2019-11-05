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

package resource

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/resource"
	"k8s.io/apimachinery/pkg/types"
)

// Context is the context for building KogitoApp resources
type Context struct {
	resource.FactoryContext
	//KogitoApp is the cached instance of the created CR
	KogitoApp *v1alpha1.KogitoApp
}

type builderChain struct {
	Context   *Context
	Resources *KogitoAppResources
	Error     error
}

func (c *builderChain) AndBuild(f func(*builderChain) *builderChain) *builderChain {
	if c.Error == nil {
		return f(c)
	}
	// break the chain
	return c
}

// GetRequestedResources will fetch for all the kubernetes resources requested by the KogitoApp
func GetRequestedResources(context *Context) (*KogitoAppResources, error) {
	chain := &builderChain{
		Resources: &KogitoAppResources{},
		Context:   context,
	}
	chain.
		AndBuild(buildConfigS2IBuilder).
		AndBuild(buildConfigRuntimeBuilder).
		AndBuild(imageStreamBuilder).
		AndBuild(runtimeImageGetter).
		AndBuild(deploymentConfigBuilder).
		AndBuild(serviceBuilder).
		AndBuild(routeBuilder).
		AndBuild(servicemonitorBuilder)
	return chain.Resources, chain.Error
}

func buildConfigS2IBuilder(chain *builderChain) *builderChain {
	bc, err := NewBuildConfigS2I(chain.Context.KogitoApp)
	if err != nil {
		chain.Error = err
		return chain
	}

	chain.Resources.BuildConfigS2I = &bc

	return chain
}

func buildConfigRuntimeBuilder(chain *builderChain) *builderChain {
	bc, err := NewBuildConfigRuntime(
		chain.Context.KogitoApp,
		chain.Resources.BuildConfigS2I,
	)
	if err != nil {
		chain.Error = err
		return chain
	}

	chain.Resources.BuildConfigRuntime = &bc

	return chain
}

func imageStreamBuilder(chain *builderChain) *builderChain {
	isS2I := NewImageStreamTag(chain.Context.KogitoApp, chain.Resources.BuildConfigS2I.Name)
	chain.Resources.ImageStreamS2I = isS2I

	isRuntime := NewImageStreamTag(chain.Context.KogitoApp, chain.Resources.BuildConfigRuntime.Name)
	chain.Resources.ImageStreamRuntime = isRuntime

	return chain
}

func deploymentConfigBuilder(chain *builderChain) *builderChain {
	chain.Resources.DeploymentConfig = nil

	if chain.Resources.RuntimeImage != nil {
		dc, err := NewDeploymentConfig(
			chain.Context.KogitoApp,
			chain.Resources.BuildConfigRuntime,
			chain.Resources.RuntimeImage,
		)
		if err != nil {
			chain.Error = err
			return chain
		}

		chain.Resources.DeploymentConfig = dc
	}

	return chain
}

func serviceBuilder(chain *builderChain) *builderChain {
	// Service depends on the DC
	if chain.Resources.DeploymentConfig != nil {
		svc := NewService(chain.Context.KogitoApp, chain.Resources.DeploymentConfig)
		if svc != nil {
			chain.Resources.Service = svc
		}
	}
	return chain
}

func routeBuilder(chain *builderChain) *builderChain {
	// we only create a router if we already have a service
	if chain.Resources.Service != nil {
		route, err := NewRoute(chain.Context.KogitoApp, chain.Resources.Service)
		if err != nil {
			chain.Error = err
			return chain
		}

		chain.Resources.Route = route
	}
	return chain
}

func servicemonitorBuilder(chain *builderChain) *builderChain {
	if chain.Resources.RuntimeImage != nil && chain.Resources.Service != nil {
		sm, err := NewServiceMonitor(chain.Context.KogitoApp, chain.Resources.RuntimeImage, chain.Resources.Service, chain.Context.Client)
		if err != nil {
			chain.Error = err
			return chain
		}
		chain.Resources.ServiceMonitor = sm
	}

	return chain
}

func runtimeImageGetter(chain *builderChain) *builderChain {
	bcNamespacedName := types.NamespacedName{
		Namespace: chain.Resources.BuildConfigRuntime.Namespace,
		Name:      chain.Resources.BuildConfigRuntime.Name,
	}
	chain.Resources.DeploymentConfig = nil

	if dockerImage, err := openshift.ImageStreamC(chain.Context.Client).FetchDockerImage(bcNamespacedName); err != nil {
		chain.Error = err
	} else if dockerImage != nil {
		chain.Resources.RuntimeImage = dockerImage
	} else {
		log.Warnf("Couldn't find an image with name '%s' in the namespace '%s'. The DeploymentConfig will be created once the build is done.", bcNamespacedName.Name, bcNamespacedName.Namespace)
	}

	return chain
}
