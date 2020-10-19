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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

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

		service := createRequiredService(s.instance, deployment)

		appProps := map[string]string{}
		var envProperties []corev1.EnvVar

		if len(s.instance.GetSpec().GetInfra()) > 0 {
			log.Debugf("Infra references are provided")
			infraAppProps, infraEnvProp, err := s.fetchKogitoInfraProperties()
			if err != nil {
				return resources, reconcileAfter, err
			}
			util.AppendToStringMap(infraAppProps, appProps)
			envProperties = append(envProperties, infraEnvProp...)
		}

		if len(s.instance.GetSpec().GetConfig()) > 0 {
			log.Debugf("custom app properties are provided")
			util.AppendToStringMap(s.instance.GetSpec().GetConfig(), appProps)
		}

		s.applyEnvironmentPropertiesConfiguration(envProperties, deployment)

		contentHash, configMap, err := getAppPropConfigMapContentHash(s.instance, appProps, s.client)
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

func (s *serviceDeployer) applyApplicationPropertiesConfigurations(contentHash string, deployment *appsv1.Deployment, instance v1alpha1.KogitoService) {
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = map[string]string{AppPropContentHashKey: contentHash}
	} else {
		deployment.Spec.Template.Annotations[AppPropContentHashKey] = contentHash
	}
	if deployment.Spec.Template.Spec.Volumes == nil {
		deployment.Spec.Template.Spec.Volumes = []corev1.Volume{createAppPropVolume(instance)}
	} else {
		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, createAppPropVolume(instance))
	}
	if deployment.Spec.Template.Spec.Containers[0].VolumeMounts == nil {
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = []corev1.VolumeMount{createAppPropVolumeMount()}
	} else {
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, createAppPropVolumeMount())
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

	if len(s.definition.extraManagedObjectLists) > 0 {
		objectTypes = append(objectTypes, s.definition.extraManagedObjectLists...)
	}

	resources, err = kubernetes.ResourceC(s.client).ListAll(objectTypes, s.instance.GetNamespace(), s.instance)
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

func (s *serviceDeployer) fetchKogitoInfraProperties() (map[string]string, []corev1.EnvVar, error) {
	kogitoInfraReferences := s.instance.GetSpec().GetInfra()
	log.Debugf("Going to fetch kogito infra properties for given references : %s", kogitoInfraReferences)
	consolidateAppProperties := map[string]string{}
	var consolidateEnvProperties []corev1.EnvVar
	for _, kogitoInfraName := range kogitoInfraReferences {
		// load infra resource
		kogitoInfraInstance, err := infrastructure.FetchKogitoInfraInstance(s.client, kogitoInfraName, s.instance.GetNamespace())
		if err != nil {
			return nil, nil, err
		}

		// fetch app properties from Kogito infra instance
		appProp := kogitoInfraInstance.Status.AppProps
		util.AppendToStringMap(appProp, consolidateAppProperties)

		// fetch env properties from Kogito infra instance
		envProp := kogitoInfraInstance.Status.Env
		consolidateEnvProperties = append(consolidateEnvProperties, envProp...)

		// Special handling for Kafka Infra
		if kogitoinfra.IsKafkaResource(kogitoInfraInstance) {
			kafkaURI := getKafkaServerURIFromAppProps(appProp)
			if err = s.createKafkaTopics(kogitoInfraInstance, kafkaURI); err != nil {
				return nil, nil, err
			}
		}
	}
	return consolidateAppProperties, consolidateEnvProperties, nil
}

func (s *serviceDeployer) applyEnvironmentPropertiesConfiguration(envProps []corev1.EnvVar, deployment *appsv1.Deployment) {
	container := &deployment.Spec.Template.Spec.Containers[0]
	container.Env = append(container.Env, envProps...)
}
