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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

// createRequiredResources creates the required resources given the KogitoService instance
func (s *serviceDeployer) createRequiredResources() (resources map[reflect.Type][]resource.KubernetesResource, err error) {
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
		return resources, err
	} else if len(image) > 0 {
		deployment := createRequiredDeployment(s.instance, image, s.definition)
		if err = s.onDeploymentCreate(deployment, imageHandler); err != nil {
			return resources, err
		}
		service := createRequiredService(s.instance, deployment)

		appProps := map[string]string{}
		var infraVolumes []v1beta1.KogitoInfraVolume

		if len(s.instance.GetSpec().GetInfra()) > 0 {
			log.Debug("Infra references are provided")
			var infraAppProps map[string]string
			var infraEnvProp []corev1.EnvVar
			infraAppProps, infraEnvProp, infraVolumes, err = s.fetchKogitoInfraProperties()
			if err != nil {
				return resources, err
			}
			util.AppendToStringMap(infraAppProps, appProps)
			deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, infraEnvProp...)
		}

		if len(s.instance.GetSpec().GetConfig()) > 0 {
			log.Debug("custom app properties are provided")
			util.AppendToStringMap(s.instance.GetSpec().GetConfig(), appProps)
		}

		contentHash, configMap, err := getAppPropConfigMapContentHash(s.instance, appProps, s.client)
		if err != nil {
			return resources, err
		}

		s.applyApplicationPropertiesAnnotations(contentHash, deployment)

		s.mountVolumes(infraVolumes, deployment)

		if configMap != nil {
			resources[reflect.TypeOf(corev1.ConfigMap{})] = []resource.KubernetesResource{configMap}
		}
		resources[reflect.TypeOf(appsv1.Deployment{})] = []resource.KubernetesResource{deployment}
		resources[reflect.TypeOf(corev1.Service{})] = []resource.KubernetesResource{service}
		if s.client.IsOpenshift() {
			resources[reflect.TypeOf(routev1.Route{})] = []resource.KubernetesResource{createRequiredRoute(s.instance, service)}
		}
		if err := s.onObjectsCreate(resources, s.client); err != nil {
			return resources, err
		}
	}

	if err := s.setOwner(resources); err != nil {
		return resources, err
	}
	return
}

func (s *serviceDeployer) onDeploymentCreate(deployment *appsv1.Deployment, imageHandler *imageHandler) error {
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

// onObjectsCreate calls the OnObjectsCreate hook for clients to add their custom objects/logic to the service
func (s *serviceDeployer) onObjectsCreate(resources map[reflect.Type][]resource.KubernetesResource, cli *client.Client) error {
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

func (s *serviceDeployer) getKogitoServiceImage(imageHandler *imageHandler, instance v1beta1.KogitoService) (string, error) {
	image, err := imageHandler.resolveImage()
	if err != nil {
		return "", err
	}
	if len(image) > 0 {
		return image, nil
	}
	log.Warn("Image not found for the service", "service",
		instance.GetName(), "namespace", instance.GetNamespace())

	deploymentDeployed := &appsv1.Deployment{ObjectMeta: v1.ObjectMeta{Name: instance.GetName(), Namespace: instance.GetNamespace()}}
	if exists, err := kubernetes.ResourceC(s.client).Fetch(deploymentDeployed); err != nil {
		return "", err
	} else if !exists {
		return "", nil
	}
	if len(deploymentDeployed.Spec.Template.Spec.Containers) > 0 {
		log.Info("Returning the image resolved from the Deployment")
		return deploymentDeployed.Spec.Template.Spec.Containers[0].Image, nil
	}
	return "", nil
}

func (s *serviceDeployer) applyApplicationPropertiesAnnotations(contentHash string, deployment *appsv1.Deployment) {
	if deployment.Spec.Template.Annotations == nil {
		deployment.Spec.Template.Annotations = map[string]string{AppPropContentHashKey: contentHash}
	} else {
		deployment.Spec.Template.Annotations[AppPropContentHashKey] = contentHash
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

func (s *serviceDeployer) fetchKogitoInfraProperties() (map[string]string, []corev1.EnvVar, []v1beta1.KogitoInfraVolume, error) {
	kogitoInfraReferences := s.instance.GetSpec().GetInfra()
	log.Debug("Going to fetch kogito infra properties", "infra", kogitoInfraReferences)
	consolidateAppProperties := map[string]string{}
	var consolidateEnvProperties []corev1.EnvVar
	var volumes []v1beta1.KogitoInfraVolume
	for _, kogitoInfraName := range kogitoInfraReferences {
		// load infra resource
		kogitoInfraInstance, err := infrastructure.MustFetchKogitoInfraInstance(s.client, kogitoInfraName, s.instance.GetNamespace())
		if err != nil {
			return nil, nil, nil, err
		}

		runtime := s.instance.GetSpec().GetRuntime()

		// fetch app properties from Kogito infra instance
		appProp := kogitoInfraInstance.Status.RuntimeProperties[runtime].AppProps
		util.AppendToStringMap(appProp, consolidateAppProperties)

		// fetch env properties from Kogito infra instance
		envProp := kogitoInfraInstance.Status.RuntimeProperties[runtime].Env
		consolidateEnvProperties = append(consolidateEnvProperties, envProp...)
		volumes = append(volumes, kogitoInfraInstance.Status.Volume...)
	}
	return consolidateAppProperties, consolidateEnvProperties, volumes, nil
}

func (s *serviceDeployer) mountVolumes(kogitoInfraVolumes []v1beta1.KogitoInfraVolume, deployment *appsv1.Deployment) {
	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, createAppPropVolume(s.instance))
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, createAppPropVolumeMount())
	for _, infraVolume := range kogitoInfraVolumes {
		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, infraVolume.NamedVolume.ToKubeVolume())
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, infraVolume.Mount)
	}
}
