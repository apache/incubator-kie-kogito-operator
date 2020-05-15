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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"k8s.io/apimachinery/pkg/types"
)

type builderChain struct {
	Client    *client.Client
	KogitoApp *v1alpha1.KogitoApp
	Resources *KogitoAppResources
	Error     error
}

// TODO: get rid of this once we have https://issues.redhat.com/browse/KOGITO-952

func (c *builderChain) andBuild(f func(*builderChain) *builderChain) *builderChain {
	if c.Error == nil {
		return f(c)
	}
	// break the chain
	return c
}

// GetRequestedResources will fetch for all the kubernetes resources requested by the KogitoApp
func GetRequestedResources(kogitoApp *v1alpha1.KogitoApp, client *client.Client) (*KogitoAppResources, error) {
	chain := &builderChain{
		Client:    client,
		KogitoApp: kogitoApp,
		Resources: &KogitoAppResources{},
	}
	chain.
		andBuild(buildConfigS2IBuilder).
		andBuild(buildConfigRuntimeBuilder).
		andBuild(buildConfigRuntimeBinaryBuilder).
		andBuild(protoBufConfigMap).
		andBuild(appPropConfigMap).
		andBuild(imageStreamBuilder).
		andBuild(runtimeImageGetter).
		andBuild(deploymentConfigBuilder).
		andBuild(serviceBuilder).
		andBuild(routeBuilder).
		andBuild(serviceMonitorBuilder)
	return chain.Resources, chain.Error
}

func buildConfigS2IBuilder(chain *builderChain) *builderChain {
	// if there is gitURI create a s2i build config.
	if !chain.KogitoApp.Spec.IsGitURIEmpty() {
		bc, err := newBuildConfigS2I(chain.KogitoApp)
		if err != nil {
			chain.Error = err
			return chain
		}
		chain.Resources.BuildConfigS2I = &bc

		// otherwise create a s2i build from file
	} else {
		bc, err := newBuildConfigS2IFromFile(chain.KogitoApp)
		if err != nil {
			chain.Error = err
			return chain
		}
		chain.Resources.BuildConfigS2I = &bc
	}
	return chain
}

func buildConfigRuntimeBuilder(chain *builderChain) *builderChain {
	bc, err := newBuildConfigRuntime(
		chain.KogitoApp,
		chain.Resources.BuildConfigS2I,
	)
	if err != nil {
		chain.Error = err
		return chain
	}

	// if gitURI is empty, instantiate a s2i binary build
	if chain.KogitoApp.Spec.IsGitURIEmpty() {
		chain.Resources.BuildConfigBinary = &bc
	}
	chain.Resources.BuildConfigRuntime = &bc

	return chain
}

func buildConfigRuntimeBinaryBuilder(chain *builderChain) *builderChain {
	bc := newBuildConfigRuntimeBinary(chain.KogitoApp)
	chain.Resources.BuildConfigBinary = &bc
	return chain
}

func protoBufConfigMap(chain *builderChain) *builderChain {
	chain.Error = nil
	chain.Resources.ProtoBufCM = newProtoBufConfigMap(chain.KogitoApp)
	return chain
}

func imageStreamBuilder(chain *builderChain) *builderChain {
	if chain.Resources.BuildConfigS2I != nil {
		isS2I := newImageStream(chain.KogitoApp, chain.Resources.BuildConfigS2I.Name)
		chain.Resources.ImageStreamS2I = isS2I
	}

	isRuntime := newImageStream(chain.KogitoApp, chain.KogitoApp.Name)
	chain.Resources.ImageStreamRuntime = isRuntime

	return chain
}

func deploymentConfigBuilder(chain *builderChain) *builderChain {
	chain.Resources.DeploymentConfig = nil
	if chain.Resources.RuntimeImage != nil {
		dc, err := newDeploymentConfig(
			chain.KogitoApp,
			chain.Resources.BuildConfigBinary,
			chain.Resources.RuntimeImage,
			chain.Resources.AppPropContentHash,
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
		svc := newService(chain.KogitoApp, chain.Resources.DeploymentConfig)
		if svc != nil {
			chain.Resources.Service = svc
		}
	}

	return chain
}

func routeBuilder(chain *builderChain) *builderChain {
	// we only create a router if we already have a service
	if chain.Resources.Service != nil {
		route, err := newRoute(chain.KogitoApp, chain.Resources.Service)
		if err != nil {
			chain.Error = err
			return chain
		}

		chain.Resources.Route = route
	}

	return chain
}

func serviceMonitorBuilder(chain *builderChain) *builderChain {
	if chain.Resources.RuntimeImage != nil && chain.Resources.Service != nil {
		sm, err := newServiceMonitor(chain.KogitoApp, chain.Resources.RuntimeImage, chain.Resources.Service, chain.Client)
		if err != nil {
			chain.Error = err
			return chain
		}
		chain.Resources.ServiceMonitor = sm
	}

	return chain
}

func runtimeImageGetter(chain *builderChain) *builderChain {
	imageStreamName := types.NamespacedName{
		Namespace: chain.KogitoApp.Namespace,
		Name:      chain.KogitoApp.Name,
	}
	chain.Resources.DeploymentConfig = nil

	if dockerImage, err := openshift.ImageStreamC(chain.Client).FetchDockerImage(imageStreamName); err != nil {
		chain.Error = err
	} else if dockerImage != nil {
		chain.Resources.RuntimeImage = dockerImage
	} else {
		log.Warnf("Couldn't find an image with name '%s' in the namespace '%s'. The DeploymentConfig will be created once the build is done.", imageStreamName.Name, imageStreamName.Namespace)
	}

	return chain
}

func appPropConfigMap(chain *builderChain) *builderChain {
	chain.Error = nil

	if contentHash, configMap, err := services.GetAppPropConfigMapContentHash(chain.KogitoApp.Name, chain.KogitoApp.Namespace, chain.Client); err != nil {
		chain.Error = err
	} else {
		chain.Resources.AppPropContentHash = contentHash
		chain.Resources.AppPropCM = configMap
	}

	return chain
}
