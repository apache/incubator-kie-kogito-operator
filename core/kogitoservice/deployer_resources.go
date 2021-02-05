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

package kogitoservice

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/framework"
	"github.com/kiegroup/kogito-cloud-operator/core/framework/util"
	"github.com/kiegroup/kogito-cloud-operator/core/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/core/manager"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
)

// createRequiredResources creates the required resources given the KogitoService instance
func (s *serviceDeployer) createRequiredResources() (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	resources = make(map[reflect.Type][]resource.KubernetesResource)
	imageHandler := s.newImageHandler()
	imageStream, err := imageHandler.CreateImageStreamIfNotExists()
	if err != nil {
		return
	}
	if imageStream != nil {
		resources[reflect.TypeOf(imgv1.ImageStream{})] = []resource.KubernetesResource{imageStream}
	}

	// we only create the rest of the resources once we have a resolvable image
	// or if the deployment is already there, we don't want to delete it :)
	if image, err := s.getKogitoServiceImage(imageHandler, s.instance); err != nil {
		return resources, err
	} else if len(image) > 0 {
		deploymentHandler := NewDeploymentHandler(s.client, s.log)
		deployment := deploymentHandler.CreateRequiredDeployment(s.instance, image, s.definition)
		if err = s.onDeploymentCreate(deployment, imageStream); err != nil {
			return resources, err
		}

		serviceHandler := infrastructure.NewServiceHandler(s.log)
		service := serviceHandler.CreateService(s.instance, deployment)

		appProps := map[string]string{}
		var infraVolumes []api.KogitoInfraVolume

		if len(s.instance.GetSpec().GetInfra()) > 0 {
			s.log.Debug("Infra references are provided")
			var infraAppProps map[string]string
			var infraEnvProp []corev1.EnvVar
			infraAppProps, infraEnvProp, infraVolumes, err = s.fetchKogitoInfraProperties()
			if err != nil {
				return resources, err
			}
			util.AppendToStringMap(infraAppProps, appProps)
			deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, infraEnvProp...)
		}

		deployment.Spec.Template.Spec.Containers[0].Env = append(deployment.Spec.Template.Spec.Containers[0].Env, framework.CreateEnvVar(infrastructure.RuntimeTypeKey, string(s.instance.GetSpec().GetRuntime())))

		if len(s.instance.GetSpec().GetConfig()) > 0 {
			s.log.Debug("custom app properties are provided")
			util.AppendToStringMap(s.instance.GetSpec().GetConfig(), appProps)
		}

		appPropsConfigMapHandler := NewAppPropsConfigMapHandler()
		contentHash, configMap, err := appPropsConfigMapHandler.GetAppPropConfigMapContentHash(s.instance, appProps, s.client)
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
			routeHandler := infrastructure.NewRouteHandler(s.client, s.log)
			resources[reflect.TypeOf(routev1.Route{})] = []resource.KubernetesResource{routeHandler.CreateRoute(service)}
		}
		if err := s.onObjectsCreate(resources); err != nil {
			return resources, err
		}
	}

	if err := s.setOwner(resources); err != nil {
		return resources, err
	}
	return
}

func (s *serviceDeployer) onDeploymentCreate(deployment *appsv1.Deployment, imageStream *imgv1.ImageStream) error {
	if imageStream != nil {
		imageHandler := s.newImageHandler()
		key, value := imageHandler.ResolveImageStreamTriggerAnnotation(s.instance.GetName())
		deployment.Annotations = map[string]string{key: value}
	}
	if s.definition.OnDeploymentCreate != nil {
		if err := s.definition.OnDeploymentCreate(deployment); err != nil {
			return err
		}
	}
	return nil
}

// onObjectsCreate calls the OnObjectsCreate hook for clients to add their custom objects/logic to the service
func (s *serviceDeployer) onObjectsCreate(resources map[reflect.Type][]resource.KubernetesResource) error {
	if s.definition.OnObjectsCreate != nil {
		var additionalRes map[reflect.Type][]resource.KubernetesResource
		var err error
		additionalRes, s.definition.extraManagedObjectLists, err = s.definition.OnObjectsCreate(s.instance)
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

func (s *serviceDeployer) getKogitoServiceImage(imageHandler infrastructure.ImageHandler, instance api.KogitoService) (string, error) {
	image, err := imageHandler.ResolveImage()
	if err != nil {
		return "", err
	}
	if len(image) > 0 {
		return image, nil
	}
	s.log.Warn("Image not found for the service")

	deploymentHandler := infrastructure.NewDeploymentHandler(s.client, s.log)
	deploymentDeployed, err := deploymentHandler.MustFetchDeployment(types.NamespacedName{Name: instance.GetName(), Namespace: instance.GetNamespace()})
	if err != nil {
		return "", err
	}
	if len(deploymentDeployed.Spec.Template.Spec.Containers) > 0 {
		s.log.Info("Returning the image resolved from the Deployment")
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
	if err = s.addSharedImageStreamToResources(resources, s.definition.DefaultImageName, s.getNamespace()); err != nil {
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

func (s *serviceDeployer) fetchKogitoInfraProperties() (map[string]string, []corev1.EnvVar, []api.KogitoInfraVolume, error) {
	kogitoInfraReferences := s.instance.GetSpec().GetInfra()
	s.log.Debug("Going to fetch kogito infra properties", "infra", kogitoInfraReferences)
	consolidateAppProperties := map[string]string{}
	var consolidateEnvProperties []corev1.EnvVar
	var volumes []api.KogitoInfraVolume
	for _, kogitoInfraName := range kogitoInfraReferences {
		// load infra resource
		infraManager := manager.NewKogitoInfraManager(s.client, s.log, s.scheme, s.infraHandler)
		kogitoInfraInstance, err := infraManager.MustFetchKogitoInfraInstance(types.NamespacedName{Name: kogitoInfraName, Namespace: s.instance.GetNamespace()})
		if err != nil {
			return nil, nil, nil, err
		}

		runtime := s.instance.GetSpec().GetRuntime()

		// fetch app properties from Kogito infra instance
		appProp := kogitoInfraInstance.GetStatus().GetRuntimeProperties()[runtime].AppProps
		util.AppendToStringMap(appProp, consolidateAppProperties)

		// fetch env properties from Kogito infra instance
		envProp := kogitoInfraInstance.GetStatus().GetRuntimeProperties()[runtime].Env
		consolidateEnvProperties = append(consolidateEnvProperties, envProp...)

		// fetch volume from Kogito infra instance
		volumes = append(volumes, kogitoInfraInstance.GetStatus().GetVolumes()...)
	}
	return consolidateAppProperties, consolidateEnvProperties, volumes, nil
}

func (s *serviceDeployer) mountVolumes(kogitoInfraVolumes []api.KogitoInfraVolume, deployment *appsv1.Deployment) {
	appPropsVolumeHandler := NewAppPropsVolumeHandler()
	deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, appPropsVolumeHandler.CreateAppPropVolume(s.instance))
	deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, appPropsVolumeHandler.CreateAppPropVolumeMount())
	for _, infraVolume := range kogitoInfraVolumes {
		deployment.Spec.Template.Spec.Volumes = append(deployment.Spec.Template.Spec.Volumes, infraVolume.NamedVolume.ToKubeVolume())
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts = append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, infraVolume.Mount)
	}
}

// AddSharedImageStreamToResources adds the shared ImageStream in the given resource map.
// Normally used during reconciliation phase to bring a not yet owned ImageStream to the deployed list.
func (s *serviceDeployer) addSharedImageStreamToResources(resources map[reflect.Type][]resource.KubernetesResource, name, ns string) error {
	if s.client.IsOpenshift() {
		// is the image already there?
		for _, is := range resources[reflect.TypeOf(imgv1.ImageStream{})] {
			if is.GetName() == name &&
				is.GetNamespace() == ns {
				return nil
			}
		}
		// fetch the shared image
		imageStreamHandler := infrastructure.NewImageStreamHandler(s.client, s.log)
		sharedImageStream, err := imageStreamHandler.FetchImageStream(types.NamespacedName{Name: name, Namespace: ns})
		if err != nil {
			return err
		}
		if sharedImageStream != nil {
			resources[reflect.TypeOf(imgv1.ImageStream{})] = append(resources[reflect.TypeOf(imgv1.ImageStream{})], sharedImageStream)
		}
	}
	return nil
}

func (s *serviceDeployer) newImageHandler() infrastructure.ImageHandler {
	addDockerImageReference := len(s.instance.GetSpec().GetImage()) != 0 || !s.definition.CustomService
	var image api.Image
	if len(s.instance.GetSpec().GetImage()) == 0 {
		image = api.Image{
			Name: s.definition.DefaultImageName,
			Tag:  s.definition.DefaultImageTag,
		}
	} else {
		image = framework.ConvertImageTagToImage(s.instance.GetSpec().GetImage())
	}

	return infrastructure.NewImageHandler(&image, s.definition.DefaultImageName, s.definition.DefaultImageName, s.instance.GetNamespace(), addDockerImageReference, s.instance.GetSpec().IsInsecureImageRegistry(), s.client, s.log)
}
