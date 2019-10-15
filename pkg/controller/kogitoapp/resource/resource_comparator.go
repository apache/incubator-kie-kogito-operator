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
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	routev1 "github.com/openshift/api/route/v1"
	v1 "k8s.io/api/core/v1"
	"reflect"
)

// GetComparator returns the comparator for the kubernetes resources of the KogitoApp
func GetComparator() compare.MapComparator {
	resourceComparator := compare.DefaultComparator()

	resourceComparator.SetComparator(createDeploymentConfigComparator(resourceComparator))
	resourceComparator.SetComparator(createBuildConfigComparator(resourceComparator))
	resourceComparator.SetComparator(createServiceComparator(resourceComparator))
	resourceComparator.SetComparator(createRouteComparator(resourceComparator))

	return compare.MapComparator{Comparator: resourceComparator}
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

func createDeploymentConfigComparator(resourceComparator compare.ResourceComparator) (
	reflect.Type, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	dcType := reflect.TypeOf(appsv1.DeploymentConfig{})
	defaultDCComparator := resourceComparator.GetComparator(dcType)
	return dcType, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		dc1 := deployed.(*appsv1.DeploymentConfig)
		dc2 := requested.(*appsv1.DeploymentConfig).DeepCopy()

		for i := range dc1.Spec.Triggers {
			if len(dc2.Spec.Triggers) <= i {
				return false
			}
			trigger1 := dc1.Spec.Triggers[i]
			trigger2 := dc2.Spec.Triggers[i]
			if trigger1.ImageChangeParams != nil && trigger2.ImageChangeParams != nil && trigger2.ImageChangeParams.From.Namespace == "" {
				//This value is generated based on image stream being found in current or openshift project:
				trigger1.ImageChangeParams.From.Namespace = ""
			}
		}

		if dc2.Spec.Strategy.RollingParams == nil && dc1.Spec.Strategy.Type == dc2.Spec.Strategy.Type {
			dc1.Spec.Strategy.RollingParams = dc2.Spec.Strategy.RollingParams
		}

		return defaultDCComparator(deployed, requested)
	}
}

func createBuildConfigComparator(resourceComparator compare.ResourceComparator) (
	reflect.Type, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	bcType := reflect.TypeOf(buildv1.BuildConfig{})
	defaultBCComparator := resourceComparator.GetComparator(bcType)
	return bcType, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		bc1 := deployed.(*buildv1.BuildConfig)
		bc2 := requested.(*buildv1.BuildConfig).DeepCopy()

		if !containAllLabels(bc1, bc2) {
			return false
		}
		if bc1.Spec.Strategy.SourceStrategy != nil {
			//This value is generated based on image stream being found in current or openshift project:
			bc1.Spec.Strategy.SourceStrategy.From.Namespace = bc2.Spec.Strategy.SourceStrategy.From.Namespace
		}
		if len(bc1.Spec.Triggers) > 0 && len(bc2.Spec.Triggers) == 0 {
			//Triggers are generated based on provided github repo
			bc1.Spec.Triggers = bc2.Spec.Triggers
		}
		return defaultBCComparator(deployed, requested)
	}
}

func createServiceComparator(resourceComparator compare.ResourceComparator) (
	reflect.Type, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	svcType := reflect.TypeOf(v1.Service{})
	defaultSvcComparator := resourceComparator.GetComparator(svcType)
	return svcType, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		svc1 := deployed.(*v1.Service)
		svc2 := requested.(*v1.Service).DeepCopy()

		if !containAllLabels(svc1, svc2) {
			return false
		}
		return defaultSvcComparator(deployed, requested)
	}
}

func createRouteComparator(resourceComparator compare.ResourceComparator) (
	reflect.Type, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool) {
	rtType := reflect.TypeOf(routev1.Route{})
	defaultRtComparator := resourceComparator.GetComparator(rtType)
	return rtType, func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
		rt1 := deployed.(*routev1.Route)
		rt2 := requested.(*routev1.Route).DeepCopy()

		if !containAllLabels(rt1, rt2) {
			return false
		}
		return defaultRtComparator(deployed, requested)
	}
}
