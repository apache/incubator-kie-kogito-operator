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
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

const (
	labelAppKey = "app"
)

var log = logger.GetLogger("resource_jobs_service")

// CreateRequiredResources creates the required resources given the KogitoJobsService instance
func CreateRequiredResources(instance *v1alpha1.KogitoJobsService, cli *client.Client) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	resources = make(map[reflect.Type][]resource.KubernetesResource)
	var infinispanSecret *corev1.Secret
	infinispanSecret, err = infrastructure.FetchInfinispanCredentials(&instance.Spec, instance.Namespace, cli)
	if err != nil {
		return
	}
	image := newImageHandler(instance, cli)
	deployment := createRequiredDeployment(instance, image, infinispanSecret)
	service := createRequiredService(instance, deployment)

	resources[reflect.TypeOf(appsv1.Deployment{})] = []resource.KubernetesResource{deployment}
	resources[reflect.TypeOf(corev1.Service{})] = []resource.KubernetesResource{service}

	if image.hasImageStream() {
		resources[reflect.TypeOf(imgv1.ImageStream{})] = []resource.KubernetesResource{image.imageStream}
	}
	if cli.IsOpenshift() {
		resources[reflect.TypeOf(routev1.Route{})] = []resource.KubernetesResource{createRequiredRoute(instance, service)}
	}

	return
}

// GetDeployedResources gets the deployed resources in the cluster owned by the given instance
func GetDeployedResources(instance *v1alpha1.KogitoJobsService, cli *client.Client) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	reader := read.New(cli.ControlCli).WithNamespace(instance.Namespace).WithOwnerObject(instance)
	if cli.IsOpenshift() {
		resources, err = reader.ListAll(&appsv1.DeploymentList{}, &corev1.ServiceList{}, &routev1.RouteList{}, &imgv1.ImageStreamList{})
	} else {
		resources, err = reader.ListAll(&appsv1.DeploymentList{}, &corev1.ServiceList{})
	}

	return
}

// GetComparator gets the comparator for the owned resources
func GetComparator() compare.MapComparator {
	resourceComparator := compare.DefaultComparator()

	resourceComparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(appsv1.Deployment{})).
			UseDefaultComparator().
			BuildAsFunc())

	resourceComparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(corev1.Service{})).
			UseDefaultComparator().
			BuildAsFunc())

	resourceComparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(routev1.Route{})).
			UseDefaultComparator().
			BuildAsFunc())

	resourceComparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(imgv1.ImageStream{})).
			UseDefaultComparator().
			WithCustomComparator(equalImageStreamTag).
			BuildAsFunc())

	return compare.MapComparator{Comparator: resourceComparator}
}

func equalImageStreamTag(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
	img1 := deployed.(*imgv1.ImageStream)
	img2 := requested.(*imgv1.ImageStream)

	// lets check if the tag is presented in the deployed stream
	for i := range img1.Spec.Tags {
		img1.Spec.Tags[i].Generation = nil
	}
	// there's no tag!
	return reflect.DeepEqual(img1.Spec.Tags, img2.Spec.Tags)
}
