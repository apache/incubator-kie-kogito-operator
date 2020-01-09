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
	// Build creates the Comparator
	Build() *Comparator
	// BuildAsFunc creates the Comparator in the form of Operator Utils interface
	BuildAsFunc() (reflect.Type, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool)
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

func (c *comparatorBuilder) Build() *Comparator {
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
	return c.comparator
}

func (c *comparatorBuilder) BuildAsFunc() (reflect.Type, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	c.Build()
	return c.comparator.ResourceType, c.comparator.CompFunc
}
