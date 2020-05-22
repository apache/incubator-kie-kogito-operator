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

package framework

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	imgv1 "github.com/openshift/api/image/v1"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	routev1 "github.com/openshift/api/route/v1"

	v1 "k8s.io/api/core/v1"

	"reflect"
)

// Comparator is a simple struct to encapsulate the complex elements from Operator Utils
type Comparator struct {
	ResourceType reflect.Type
	CompFunc     func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool
}

// ComparatorBuilder creates Comparators to be used during reconciliation phases
type ComparatorBuilder interface {
	// WithCustomComparator it's the custom comparator function that will get called by the Operator Utils
	WithCustomComparator(customComparator func(deployed resource.KubernetesResource, requested resource.KubernetesResource) (equal bool)) ComparatorBuilder
	// WithType defines the comparator resource type
	WithType(resourceType reflect.Type) ComparatorBuilder
	// UseDefaultComparator defines if the comparator will delegate the comparision to inner comparators from Operator Utils
	UseDefaultComparator() ComparatorBuilder
	// Build creates the Comparator in the form of Operator Utils interface
	Build() (reflect.Type, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool)
}

// NewComparatorBuilder creates a new comparator builder for comparision usages
func NewComparatorBuilder() ComparatorBuilder {
	return &comparatorBuilder{
		comparator: &Comparator{},
	}
}

type comparatorBuilder struct {
	comparator        *Comparator
	customComparator  func(deployed resource.KubernetesResource, requested resource.KubernetesResource) (changed bool)
	defaultComparator func(deployed resource.KubernetesResource, requested resource.KubernetesResource) (changed bool)
	client            *client.Client
}

func (c *comparatorBuilder) WithClient(cli *client.Client) ComparatorBuilder {
	c.client = cli
	return c
}

func (c *comparatorBuilder) WithType(resourceType reflect.Type) ComparatorBuilder {
	c.comparator.ResourceType = resourceType
	return c
}

func (c *comparatorBuilder) WithCustomComparator(customComparator func(deployed resource.KubernetesResource, requested resource.KubernetesResource) (changed bool)) ComparatorBuilder {
	c.customComparator = customComparator
	return c
}

func (c *comparatorBuilder) UseDefaultComparator() ComparatorBuilder {
	c.defaultComparator = compare.DefaultComparator().GetComparator(c.comparator.ResourceType)
	// we don't have a default comparator for the given type, call the generic one
	if c.defaultComparator == nil {
		c.defaultComparator = compare.DefaultComparator().GetDefaultComparator()
	}
	return c
}

func (c *comparatorBuilder) Build() (reflect.Type, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	c.comparator.CompFunc = func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		equal := true
		// calls the first comparator defined by the caller
		if c.customComparator != nil {
			equal = c.customComparator(deployed, requested) && equal
		}
		if equal && c.defaultComparator != nil {
			// calls the default comparator from Operator Utils
			equal = c.defaultComparator(deployed, requested) && equal
		}
		return equal
	}
	return c.comparator.ResourceType, c.comparator.CompFunc
}

func containAllLabels(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	deployedLabels := deployed.GetLabels()
	requestedLabels := requested.GetLabels()

	for key, value := range requestedLabels {
		if deployedLabels[key] != value {
			return false
		}
	}

	return true
}

// CreateDeploymentConfigComparator creates a new comparator for DeploymentConfig using Trigger and RollingParams
func CreateDeploymentConfigComparator() func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		dcDeployed := deployed.(*appsv1.DeploymentConfig)
		dcRequested := requested.(*appsv1.DeploymentConfig).DeepCopy()

		for i := range dcDeployed.Spec.Triggers {
			if len(dcRequested.Spec.Triggers) <= i {
				return false
			}
			triggerDeployed := dcDeployed.Spec.Triggers[i]
			triggerRequested := dcRequested.Spec.Triggers[i]
			if triggerDeployed.ImageChangeParams != nil && triggerRequested.ImageChangeParams != nil && triggerRequested.ImageChangeParams.From.Namespace == "" {
				//This value is generated based on image stream being found in current or openshift project:
				triggerDeployed.ImageChangeParams.From.Namespace = ""
			}
		}

		if dcRequested.Spec.Strategy.RollingParams == nil && dcDeployed.Spec.Strategy.Type == dcRequested.Spec.Strategy.Type {
			dcDeployed.Spec.Strategy.RollingParams = dcRequested.Spec.Strategy.RollingParams
		}
		return true
	}
}

// CreateBuildConfigComparator creates a new comparator for BuildConfig using Label, Trigger and SourceStrategy
func CreateBuildConfigComparator() func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		bcDeployed := deployed.(*buildv1.BuildConfig)
		bcRequested := requested.(*buildv1.BuildConfig).DeepCopy()

		if !containAllLabels(bcDeployed, bcRequested) {
			return false
		}
		if len(bcDeployed.Spec.Triggers) > 0 && len(bcRequested.Spec.Triggers) == 0 {
			//Triggers are generated based on provided github repo
			bcDeployed.Spec.Triggers = bcRequested.Spec.Triggers
		}
		return true
	}
}

// CreateServiceComparator creates a new comparator for Service using Label
func CreateServiceComparator() func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		svcDeployed := deployed.(*v1.Service)
		svcRequested := requested.(*v1.Service).DeepCopy()

		if !containAllLabels(svcDeployed, svcRequested) {
			return false
		}
		return true
	}
}

// CreateRouteComparator creates a new comparator for Route using Label
func CreateRouteComparator() func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		rtDeployed := deployed.(*routev1.Route)
		rtRequested := requested.(*routev1.Route).DeepCopy()

		if !containAllLabels(rtDeployed, rtRequested) {
			return false
		}
		return true
	}
}

// CreateConfigMapComparator creates a new comparator for ConfigMap using Label
func CreateConfigMapComparator() func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		cmDeployed := deployed.(*v1.ConfigMap)
		cmRequested := requested.(*v1.ConfigMap).DeepCopy()

		if !containAllLabels(cmDeployed, cmRequested) {
			return false
		}

		return reflect.DeepEqual(cmDeployed.Data, cmRequested.Data)
	}
}

// CreateImageStreamComparator creates a new ImageStream comparator
func CreateImageStreamComparator() func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	return func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		img1 := deployed.(*imgv1.ImageStream)
		img2 := requested.(*imgv1.ImageStream)

		// lets check if the tag is presented in the deployed stream
		for i := range img1.Spec.Tags {
			img1.Spec.Tags[i].Generation = nil
		}
		// there's no tag!
		return reflect.DeepEqual(img1.Spec.Tags, img2.Spec.Tags)
	}
}
