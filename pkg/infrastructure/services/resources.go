// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

package services

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

const (
	enablePersistenceEnvKey = "ENABLE_PERSISTENCE"
)

// createRequiredResources creates the required resources given the KogitoService instance
func (s *serviceDeployer) createRequiredResources(instance v1alpha1.KogitoService) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	resources = make(map[reflect.Type][]resource.KubernetesResource)
	imageHandler := newImageHandler(instance, s.definition.DefaultImageName, s.client)
	deployment := createRequiredDeployment(instance, imageHandler, s.definition)
	if s.definition.OnDeploymentCreate != nil {
		if err = s.definition.OnDeploymentCreate(deployment, instance); err != nil {
			return
		}
	}

	service := createRequiredService(instance, deployment)

	if s.definition.infinispanAware {
		if err = s.applyInfinispanConfigurations(deployment, instance); err != nil {
			return
		}
	}

	if s.definition.kafkaAware {
		if err = s.applyKafkaConfigurations(deployment, instance); err != nil {
			return
		}
	}

	resources[reflect.TypeOf(appsv1.Deployment{})] = []resource.KubernetesResource{deployment}
	resources[reflect.TypeOf(corev1.Service{})] = []resource.KubernetesResource{service}

	if imageHandler.hasImageStream() {
		resources[reflect.TypeOf(imgv1.ImageStream{})] = []resource.KubernetesResource{imageHandler.imageStream}
	}
	if s.client.IsOpenshift() {
		resources[reflect.TypeOf(routev1.Route{})] = []resource.KubernetesResource{createRequiredRoute(instance, service)}
	}

	return
}

func (s *serviceDeployer) applyInfinispanConfigurations(deployment *appsv1.Deployment, instance v1alpha1.KogitoService) error {
	var infinispanSecret *corev1.Secret
	infinispanAware := instance.GetSpec().(v1alpha1.InfinispanAware)
	infinispanSecret, err := infrastructure.FetchInfinispanCredentials(infinispanAware, instance.GetNamespace(), s.client)
	if err != nil {
		return err
	}
	infrastructure.SetInfinispanVariables(
		*infinispanAware.GetInfinispanProperties(),
		infinispanSecret,
		&deployment.Spec.Template.Spec.Containers[0])

	if infinispanAware.GetInfinispanProperties().UseKogitoInfra || len(infinispanAware.GetInfinispanProperties().URI) > 0 {
		framework.SetEnvVar(enablePersistenceEnvKey, "true", &deployment.Spec.Template.Spec.Containers[0])
	}
	return nil
}

func (s *serviceDeployer) applyKafkaConfigurations(deployment *appsv1.Deployment, instance v1alpha1.KogitoService) error {
	URI, err := getKafkaServerURI(*instance.GetSpec().(v1alpha1.KafkaAware).GetKafkaProperties(), s.getNamespace(), s.client)
	if err != nil {
		return err
	}

	if len(URI) > 0 {
		for _, kafkaTopic := range s.definition.KafkaTopics {
			framework.SetEnvVar(fromKafkaTopicToQuarkusEnvVar(kafkaTopic), URI, &deployment.Spec.Template.Spec.Containers[0])
		}
		framework.SetEnvVar(quarkusBootstrapEnvVar, URI, &deployment.Spec.Template.Spec.Containers[0])
	}

	return nil
}

// getDeployedResources gets the deployed resources in the cluster owned by the given instance
func (s *serviceDeployer) getDeployedResources(instance v1alpha1.KogitoService) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	reader := read.New(s.client.ControlCli).WithNamespace(instance.GetNamespace()).WithOwnerObject(instance)
	var objectTypes []runtime.Object
	if s.client.IsOpenshift() {
		objectTypes = []runtime.Object{&appsv1.DeploymentList{}, &corev1.ServiceList{}, &routev1.RouteList{}, &imgv1.ImageStreamList{}}
	} else {
		objectTypes = []runtime.Object{&appsv1.DeploymentList{}, &corev1.ServiceList{}}
	}

	if infrastructure.IsStrimziAvailable(s.client) {
		objectTypes = append(objectTypes, &kafkabetav1.KafkaTopicList{})
	}
	return reader.ListAll(objectTypes...)
}

// getComparator gets the comparator for the owned resources
func (s *serviceDeployer) getComparator() compare.MapComparator {
	resourceComparator := compare.DefaultComparator()

	resourceComparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(appsv1.Deployment{})).
			UseDefaultComparator().
			Build())

	resourceComparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(corev1.Service{})).
			UseDefaultComparator().
			Build())

	resourceComparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(routev1.Route{})).
			UseDefaultComparator().
			WithCustomComparator(framework.CreateRouteComparator()).
			Build())

	resourceComparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(imgv1.ImageStream{})).
			UseDefaultComparator().
			WithCustomComparator(framework.CreateImageStreamComparator()).
			Build())

	return compare.MapComparator{Comparator: resourceComparator}
}
