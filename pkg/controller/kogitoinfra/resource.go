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

package kogitoinfra

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/infinispan"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/kafka"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"

	"reflect"
)

// getDeployedResources fetches for deployed resources managed by KogitoInfra controller
func (r *ReconcileKogitoInfra) getDeployedResources(instance *v1alpha1.KogitoInfra) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	resourcesInfinispan, err := infinispan.GetDeployedResources(instance, r.client)
	if err != nil {
		return
	}
	resourcesKafka, err := kafka.GetDeployedResources(instance, r.client)
	if err != nil {
		return
	}
	resources = make(map[reflect.Type][]resource.KubernetesResource, len(resourcesInfinispan)+len(resourcesKafka))
	mergeResourceMaps(resources, resourcesKafka, resourcesInfinispan)
	return
}

// createRequiredResources creates the structure of the required resources needed to have
func (r *ReconcileKogitoInfra) createRequiredResources(instance *v1alpha1.KogitoInfra) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	resourcesInfinispan, err := infinispan.CreateRequiredResources(instance, r.client)
	if err != nil {
		return
	}
	resourcesKafka, err := kafka.CreateRequiredResources(instance, r.client)
	if err != nil {
		return
	}
	resources = make(map[reflect.Type][]resource.KubernetesResource, len(resourcesInfinispan)+len(resourcesKafka))
	mergeResourceMaps(resources, resourcesKafka, resourcesInfinispan)
	return
}

// getComparator gets the comparator map from the resources managed by the KogitoInfra controller
func (r *ReconcileKogitoInfra) getComparator() compare.MapComparator {
	var comparators []framework.Comparator
	comparators = append(comparators, infinispan.GetComparators()...)
	comparators = append(comparators, kafka.GetComparators()...)

	resourceComparator := compare.DefaultComparator()
	for _, comparator := range comparators {
		resourceComparator.SetComparator(comparator.ResourceType, comparator.CompFunc)
	}

	return compare.MapComparator{Comparator: resourceComparator}
}

func mergeResourceMaps(target map[reflect.Type][]resource.KubernetesResource, source ...map[reflect.Type][]resource.KubernetesResource) {
	for _, s := range source {
		for k, v := range s {
			target[k] = v
		}
	}
}
