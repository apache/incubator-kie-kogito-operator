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

package builder

import (
	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
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
	Inventory *KogitoAppInventory
	Error     error
}

func (c *builderChain) AndBuild(f func(*builderChain) *builderChain) *builderChain {
	if c.Error == nil {
		return f(c)
	}
	// break the chain
	return c
}

// BuildOrFetchObjects will fetch for every resource in KogitoAppInventory and will create it in the cluster if not exists
func BuildOrFetchObjects(context *Context) (inv *KogitoAppInventory, err error) {
	chain := &builderChain{
		Inventory: &KogitoAppInventory{},
		Context:   context,
	}
	chain.AndBuild(buildConfigS2IBuilder).
		AndBuild(buildConfigServiceBuilder).
		AndBuild(imageStreamBuilder).
		AndBuild(deploymentConfigBuilder).
		AndBuild(serviceBuilder).
		AndBuild(routeBuilder)
	return chain.Inventory, chain.Error
}

func callPostCreate(isNew bool, object meta.ResourceObject, chain *builderChain) *builderChain {
	if isNew && chain.Context.PostCreate != nil {
		if chain.Error == nil {
			chain.Error = chain.Context.PostCreate(object)
		}
	}
	return chain
}

func callPreCreate(object meta.ResourceObject, chain *builderChain) error {
	if chain.Error == nil && chain.Context.PreCreate != nil {
		return chain.Context.PreCreate(object)
	}
	return nil
}

func buildConfigS2IBuilder(chain *builderChain) *builderChain {
	bc, err := NewBuildConfigS2I(chain.Context.KogitoApp)
	if err != nil {
		chain.Error = err
		return chain
	}
	if err := callPreCreate(&bc, chain); err != nil {
		chain.Error = err
		return chain
	}
	chain.Inventory.BuildConfigS2IStatus.IsNew, chain.Error =
		kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(&bc)
	chain.Inventory.BuildConfigS2I = &bc
	return callPostCreate(chain.Inventory.BuildConfigS2IStatus.IsNew, &bc, chain)
}

func buildConfigServiceBuilder(chain *builderChain) *builderChain {
	bc, err := NewBuildConfigService(
		chain.Context.KogitoApp,
		chain.Inventory.BuildConfigS2I,
	)
	if err != nil {
		chain.Error = err
		return chain
	}
	if err := callPreCreate(&bc, chain); err != nil {
		chain.Error = err
		return chain
	}
	chain.Inventory.BuildConfigServiceStatus.IsNew, chain.Error =
		kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(&bc)
	chain.Inventory.BuildConfigService = &bc
	return callPostCreate(chain.Inventory.BuildConfigServiceStatus.IsNew, &bc, chain)
}

func imageStreamBuilder(chain *builderChain) *builderChain {
	created := false

	is := NewImageStreamTag(chain.Context.KogitoApp, chain.Inventory.BuildConfigS2I.Name)
	if err := callPreCreate(is, chain); err != nil {
		chain.Error = err
		return chain
	}
	created, chain.Error = kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(is) //openshift.ImageStreamC(chain.Context.Client).CreateTagIfNotExists(is)
	if chain.Error != nil {
		return chain
	}
	chain = callPostCreate(created, is, chain)
	if chain.Error != nil {
		return chain
	}

	is = NewImageStreamTag(chain.Context.KogitoApp, chain.Inventory.BuildConfigService.Name)
	if err := callPreCreate(is, chain); err != nil {
		chain.Error = err
		return chain
	}
	created, chain.Error = kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(is)
	if chain.Error != nil {
		return chain
	}

	return callPostCreate(created, is, chain)
}

func deploymentConfigBuilder(chain *builderChain) *builderChain {
	bcNamespacedName := types.NamespacedName{
		Namespace: chain.Inventory.BuildConfigService.Namespace,
		Name:      chain.Inventory.BuildConfigService.Name,
	}
	chain.Inventory.DeploymentConfig = nil

	if dockerImage, err := openshift.ImageStreamC(chain.Context.Client).FetchDockerImage(bcNamespacedName); err != nil {
		chain.Error = err
	} else if dockerImage != nil {
		dc, err := NewDeploymentConfig(
			chain.Context.KogitoApp,
			chain.Inventory.BuildConfigService,
			dockerImage,
		)
		if err != nil {
			chain.Error = err
			return chain
		}
		if err := callPreCreate(dc, chain); err != nil {
			chain.Error = err
			return chain
		}
		chain.Inventory.DeploymentConfigStatus.IsNew, chain.Error =
			kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(dc)
		chain.Inventory.DeploymentConfig = dc
		chain = callPostCreate(chain.Inventory.DeploymentConfigStatus.IsNew, dc, chain)
	} else {
		log.Warnf("Couldn't find an image with name '%s' in the namespace '%s'. The DeploymentConfig will be created once the build is done.", bcNamespacedName.Name, bcNamespacedName.Namespace)
	}
	return chain
}

func serviceBuilder(chain *builderChain) *builderChain {
	// Service depends on the DC
	if chain.Inventory.DeploymentConfig != nil {
		svc := NewService(chain.Context.KogitoApp, chain.Inventory.DeploymentConfig)
		if svc != nil {
			if err := callPreCreate(svc, chain); err != nil {
				chain.Error = err
				return chain
			}
			chain.Inventory.ServiceStatus.IsNew, chain.Error =
				kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(svc)
			chain.Inventory.Service = svc
			chain = callPostCreate(chain.Inventory.ServiceStatus.IsNew, svc, chain)
		}
	}
	return chain
}

func routeBuilder(chain *builderChain) *builderChain {
	// we only create a router if we already have a service
	if chain.Inventory.Service != nil {
		route, err := NewRoute(chain.Context.KogitoApp, chain.Inventory.Service)
		if err != nil {
			chain.Error = err
			return chain
		}
		if err := callPreCreate(route, chain); err != nil {
			chain.Error = err
			return chain
		}
		chain.Inventory.RouteStatus.IsNew, chain.Error =
			kubernetes.ResourceC(chain.Context.Client).CreateIfNotExists(route)
		chain.Inventory.Route = route
		chain = callPostCreate(chain.Inventory.RouteStatus.IsNew, route, chain)
	}
	return chain
}
