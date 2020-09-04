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
	"reflect"
	"time"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	enablePersistenceEnvKey = "ENABLE_PERSISTENCE"
	enableEventsEnvKey      = "ENABLE_EVENTS"
)

// TODO: review the way we create those resources on KOGITO-1998: reorganize all those functions on other files within the package

// createRequiredResources creates the required resources given the KogitoService instance
func (s *serviceDeployer) createRequiredResources() (resources map[reflect.Type][]resource.KubernetesResource, reconcileAfter time.Duration, err error) {
	reconcileAfter = 0
	resources = make(map[reflect.Type][]resource.KubernetesResource)
	imageHandler, err := newImageHandler(s.instance, s.definition, s.client)
	if err != nil {
		return
	}
	if imageHandler.HasImageStream() {
		resources[reflect.TypeOf(imgv1.ImageStream{})] = []resource.KubernetesResource{imageHandler.imageStream}
	}

	// we only create the rest of the resources once we have a resolvable image
	// or if the deployment is already there, we don't want to delete it :)
	if image, err := s.getKogitoServiceImage(imageHandler, s.instance); err != nil {
		return resources, reconcileAfter, err
	} else if len(image) > 0 {
		deployment := createRequiredDeployment(s.instance, image, s.definition)
		if err = s.applyDeploymentCustomizations(deployment, imageHandler); err != nil {
			return resources, reconcileAfter, err
		}
		if err = s.applyDataIndexRoute(deployment, s.instance); isRequiresReconciliationError(err) {
			log.Warn(err)
			reconcileAfter = err.(requiresReconciliationError).GetReconcileAfter()
		} else if err != nil {
			return resources, reconcileAfter, err
		}
		if err = s.applyTrustyRoute(deployment, s.instance); isRequiresReconciliationError(err) {
			log.Warn(err)
			reconcileAfter = err.(requiresReconciliationError).GetReconcileAfter()
		} else if err != nil {
			return resources, reconcileAfter, err
		}

		service := createRequiredService(s.instance, deployment)

		appProps := map[string]string{}

		if s.definition.infinispanAware {
			if err = s.applyInfinispanConfigurations(deployment, appProps, s.instance); err != nil {
				return resources, reconcileAfter, err
			}
		}
		if s.definition.kafkaAware {
			if err = s.applyKafkaConfigurations(deployment, appProps, s.instance); err != nil {
				return resources, reconcileAfter, err
			}
		}

		// TODO: refactor GetAppPropConfigMapContentHash to createConfigMap on KOGITO-1998
		contentHash, configMap, err := GetAppPropConfigMapContentHash(s.instance.GetName(), s.instance.GetNamespace(), appProps, s.client)
		if err != nil {
			return resources, reconcileAfter, err
		}
		s.applyApplicationPropertiesConfigurations(contentHash, deployment, s.instance)
		if configMap != nil {
			resources[reflect.TypeOf(corev1.ConfigMap{})] = []resource.KubernetesResource{configMap}
		}

		resources[reflect.TypeOf(appsv1.Deployment{})] = []resource.KubernetesResource{deployment}
		resources[reflect.TypeOf(corev1.Service{})] = []resource.KubernetesResource{service}
		if s.client.IsOpenshift() {
			resources[reflect.TypeOf(routev1.Route{})] = []resource.KubernetesResource{createRequiredRoute(s.instance, service)}
		}
		if err := s.createAdditionalObjects(resources, s.client); err != nil {
			return resources, reconcileAfter, err
		}
	}

	if err := s.setOwner(resources); err != nil {
		return resources, reconcileAfter, err
	}
	return
}

func (s *serviceDeployer) applyDataIndexRoute(deployment *appsv1.Deployment, instance v1alpha1.KogitoService) error {
	if s.definition.RequiresDataIndex {
		dataIndexEndpoints, err := infrastructure.GetDataIndexEndpoints(s.client, s.definition.Request.Namespace)
		if err != nil {
			return err
		}
		if len(dataIndexEndpoints.HTTPRouteURI) == 0 {
			// fallback to env vars directly set on service CR: KOGITO-2827
			if len(framework.GetEnvVarFromContainer(dataIndexEndpoints.HTTPRouteEnv, &deployment.Spec.Template.Spec.Containers[0])) == 0 {
				s.recorder.Eventf(s.client, instance,
					corev1.EventTypeWarning,
					"Failure",
					"Not found Data Index external URL set on %s environment variable. Try setting the env var in '%s' manually using the Kogito service Custom Resource (CR)",
					dataIndexEndpoints.HTTPRouteEnv,
					instance.GetName())
				zeroReplicas := int32(0)
				deployment.Spec.Replicas = &zeroReplicas
				return newKogitoServiceNotReadyError(instance.GetNamespace(), instance.GetName(), "Data Index")
			}
		} else {
			framework.SetEnvVar(dataIndexEndpoints.HTTPRouteEnv, dataIndexEndpoints.HTTPRouteURI, &deployment.Spec.Template.Spec.Containers[0])
			framework.SetEnvVar(dataIndexEndpoints.WSRouteEnv, dataIndexEndpoints.WSRouteURI, &deployment.Spec.Template.Spec.Containers[0])
		}
		deployment.Spec.Replicas = instance.GetSpec().GetReplicas()
	}
	return nil
}

func (s *serviceDeployer) applyTrustyRoute(deployment *appsv1.Deployment, instance v1alpha1.KogitoService) error {
	if s.definition.RequiresTrusty {
		trustyEndpoints, err := infrastructure.GetTrustyEndpoints(s.client, s.definition.Request.Namespace)
		if err != nil {
			return err
		}
		if len(trustyEndpoints.HTTPRouteURI) == 0 {
			// fallback to env vars directly set on service CR: KOGITO-2827
			if len(framework.GetEnvVarFromContainer(trustyEndpoints.HTTPRouteEnv, &deployment.Spec.Template.Spec.Containers[0])) == 0 {
				s.recorder.Eventf(s.client, instance,
					corev1.EventTypeWarning,
					"Failure",
					"Not found Trusty external URL set on %s environment variable. Try setting the env var in '%s' manually using the Kogito service Custom Resource (CR)",
					trustyEndpoints.HTTPRouteEnv,
					instance.GetName())
				zeroReplicas := int32(0)
				deployment.Spec.Replicas = &zeroReplicas
				return newKogitoServiceNotReadyError(instance.GetNamespace(), instance.GetName(), "Trusty")
			}
		} else {
			framework.SetEnvVar(trustyEndpoints.HTTPRouteEnv, trustyEndpoints.HTTPRouteURI, &deployment.Spec.Template.Spec.Containers[0])
			framework.SetEnvVar(trustyEndpoints.WSRouteEnv, trustyEndpoints.WSRouteURI, &deployment.Spec.Template.Spec.Containers[0])
		}
		deployment.Spec.Replicas = instance.GetSpec().GetReplicas()
	}
	return nil
}

func (s *serviceDeployer) applyDeploymentCustomizations(deployment *appsv1.Deployment, imageHandler *imageHandler) error {
	if imageHandler.HasImageStream() {
		key, value := framework.ResolveImageStreamTriggerAnnotation(imageHandler.resolveImageNameTag(), s.instance.GetName())
		deployment.Annotations = map[string]string{key: value}
	}
	if s.definition.OnDeploymentCreate != nil {
		if err := s.definition.OnDeploymentCreate(s.client, deployment, s.instance); err != nil {
			return err
		}
	}
	return nil
}

// createAdditionalObjects calls the OnObjectsCreate hook for clients to add their custom objects/logic to the service
func (s *serviceDeployer) createAdditionalObjects(resources map[reflect.Type][]resource.KubernetesResource, cli *client.Client) error {
	if s.definition.OnObjectsCreate != nil {
		var additionalRes map[reflect.Type][]resource.KubernetesResource
		var err error
		additionalRes, s.definition.extraManagedObjectLists, err = s.definition.OnObjectsCreate(cli, s.instance)
		if err != nil {
			return err
		}
		for resType, res := range additionalRes {
			resources[resType] = append(resources[resType], res...)
		}
	}
	return nil
}

// setOwner sets this service instance as the owner of each resource.
func (s *serviceDeployer) setOwner(resources map[reflect.Type][]resource.KubernetesResource) error {
	for _, resourceArr := range resources {
		for _, res := range resourceArr {
			if err := framework.SetOwner(s.instance, s.scheme, res); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *serviceDeployer) getKogitoServiceImage(imageHandler *imageHandler, instance v1alpha1.KogitoService) (string, error) {
	image, err := imageHandler.resolveImage()
	if err != nil {
		return "", err
	}
	if len(image) > 0 {
		return image, nil
	}
	if !s.definition.customService {
		log.Warnf("Image for the service %s not found yet in the namespace %s. Please make sure that the informed image %s exists in the given registry.",
			instance.GetName(), instance.GetNamespace(), imageHandler.resolveRegistryImage())
	} else {
		log.Warnf("Image for the service %s not found yet in the namespace %s. ",
			instance.GetName(), instance.GetNamespace())
	}

	deploymentDeployed := &appsv1.Deployment{ObjectMeta: v1.ObjectMeta{Name: instance.GetName(), Namespace: instance.GetNamespace()}}
	if exists, err := kubernetes.ResourceC(s.client).Fetch(deploymentDeployed); err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}
	if len(deploymentDeployed.Spec.Template.Spec.Containers) > 0 {
		log.Infof("Returning the image resolved from the Deployment")
		return deploymentDeployed.Spec.Template.Spec.Containers[0].Image, nil
	}
	return "", nil
}

func (s *serviceDeployer) applyInfinispanConfigurations(deployment *appsv1.Deployment, appProps map[string]string, instance v1alpha1.KogitoService) error {
	var infinispanSecret *corev1.Secret
	infinispanAware := instance.GetSpec().(v1alpha1.InfinispanAware)
	infinispanSecret, err := fetchInfinispanCredentials(infinispanAware, instance.GetNamespace(), s.client)
	if err != nil {
		return err
	}
	setInfinispanVariables(
		s.instance.GetSpec().GetRuntime(),
		infinispanAware.GetInfinispanProperties(),
		infinispanSecret,
		&deployment.Spec.Template.Spec.Containers[0],
		appProps)

	if infinispanAware.GetInfinispanProperties().UseKogitoInfra || len(infinispanAware.GetInfinispanProperties().URI) > 0 {
		framework.SetEnvVar(enablePersistenceEnvKey, "true", &deployment.Spec.Template.Spec.Containers[0])
	}
	return nil
}

func (s *serviceDeployer) applyKafkaConfigurations(deployment *appsv1.Deployment, appProps map[string]string, instance v1alpha1.KogitoService) error {
	URI, err := getKafkaServerURI(*instance.GetSpec().(v1alpha1.KafkaAware).GetKafkaProperties(), s.getNamespace(), s.client)
	if err != nil {
		return err
	}

	if len(URI) > 0 {
		framework.SetEnvVar(enableEventsEnvKey, "true", &deployment.Spec.Template.Spec.Containers[0])
		if s.instance.GetSpec().GetRuntime() == v1alpha1.SpringBootRuntimeType {
			appProps[SpringBootstrapAppProp] = URI
		} else {
			for _, kafkaTopic := range s.definition.KafkaTopics {
				appProps[fromKafkaTopicToQuarkusAppProp(kafkaTopic)] = URI
			}
			appProps[QuarkusBootstrapAppProp] = URI
			framework.SetEnvVar(quarkusBootstrapEnvVar, URI, &deployment.Spec.Template.Spec.Containers[0])
		}
	} else {
		framework.SetEnvVar(enableEventsEnvKey, "false", &deployment.Spec.Template.Spec.Containers[0])
	}

	return nil
}

func (s *serviceDeployer) applyApplicationPropertiesConfigurations(contentHash string, deployment *appsv1.Deployment, instance v1alpha1.KogitoService) {
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = map[string]string{AppPropContentHashKey: contentHash}
	} else {
		deployment.Spec.Template.Annotations[AppPropContentHashKey] = contentHash
	}
	if deployment.Spec.Template.Spec.Volumes == nil {
		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{CreateAppPropVolume(instance.GetName())}
	} else {
		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, CreateAppPropVolume(instance.GetName()))
	}
	if deployment.Spec.Template.Spec.Containers[0].VolumeMounts == nil {
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{CreateAppPropVolumeMount()}
	} else {
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, CreateAppPropVolumeMount())
	}
}

// getDeployedResources gets the deployed resources in the cluster owned by the given instance
func (s *serviceDeployer) getDeployedResources() (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	var objectTypes []runtime.Object
	if s.client.IsOpenshift() {
		objectTypes = []runtime.Object{&appsv1.DeploymentList{}, &corev1.ServiceList{}, &corev1.ConfigMapList{}, &routev1.RouteList{}, &imgv1.ImageStreamList{}}
	} else {
		objectTypes = []runtime.Object{&appsv1.DeploymentList{}, &corev1.ServiceList{}, &corev1.ConfigMapList{}}
	}

	if infrastructure.IsStrimziAvailable(s.client) {
		objectTypes = append(objectTypes, &kafkabetav1.KafkaTopicList{})
	}

	if IsPrometheusAvailable(s.client) {
		objectTypes = append(objectTypes, &monv1.ServiceMonitorList{})
	}

	if len(s.definition.extraManagedObjectLists) > 0 {
		objectTypes = append(objectTypes, s.definition.extraManagedObjectLists...)
	}

	resources, err = kubernetes.ResourceC(s.client).ListALL(objectTypes, s.instance.GetNamespace(), s.instance)
	if err != nil {
		return
	}
	if err = AddSharedImageStreamToResources(resources, s.definition.DefaultImageName, s.getNamespace(), s.client); err != nil {
		return
	}

	return
}

// getComparator gets the comparator for the owned resources
func (s *serviceDeployer) getComparator() compare.MapComparator {
	resourceComparator := compare.DefaultComparator()

	resourceComparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(appsv1.Deployment{})).
			UseDefaultComparator().
			WithCustomComparator(framework.CreateDeploymentComparator()).
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
			WithCustomComparator(framework.CreateSharedImageStreamComparator()).
			Build())

	if s.definition.OnGetComparators != nil {
		s.definition.OnGetComparators(resourceComparator)
	}

	return compare.MapComparator{Comparator: resourceComparator}
}
