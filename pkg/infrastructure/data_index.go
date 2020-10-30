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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"path"
)

const (
	// DefaultDataIndexImageName is just the image name for the Data Index Service
	DefaultDataIndexImageName = "kogito-data-index-infinispan"
	// DefaultDataIndexName is the default name for the Data Index instance service
	DefaultDataIndexName = "data-index"
	// Data index HTTP URL env
	dataIndexHTTPRouteEnv = "KOGITO_DATAINDEX_HTTP_URL"
	// Data index WS URL env
	dataIndexWSRouteEnv = "KOGITO_DATAINDEX_WS_URL"
	// Default Proto Buf file path
	defaultProtobufMountPath = "/home/kogito/data/protobufs"
	// Proto Buf folder env
	protoBufKeyFolder string = "KOGITO_PROTOBUF_FOLDER"
	// Proto Buf watch env
	protoBufKeyWatch string = "KOGITO_PROTOBUF_WATCH"
	// Proto Buf volume mode
	protoBufConfigMapVolumeDefaultMode int32 = 420
)

// InjectDataIndexURLIntoKogitoRuntimeServices will query for every KogitoRuntime in the given namespace to inject the Data Index route to each one
// Won't trigger an update if the KogitoRuntime already has the route set to avoid unnecessary reconciliation triggers
func InjectDataIndexURLIntoKogitoRuntimeServices(client *client.Client, namespace string) error {
	log.Debugf("Injecting Data-Index Route in kogito apps")
	return injectSupportingServiceURLIntoKogitoRuntime(client, namespace, dataIndexHTTPRouteEnv, dataIndexWSRouteEnv, v1alpha1.DataIndex)
}

// InjectDataIndexURLIntoDeployment will inject data-index route URL in to kogito runtime deployment env var
func InjectDataIndexURLIntoDeployment(client *client.Client, namespace string, deployment *appsv1.Deployment) error {
	log.Debugf("Injecting Data-Index URL in kogito Runtime deployment")
	return injectSupportingServiceURLIntoDeployment(client, namespace, dataIndexHTTPRouteEnv, dataIndexWSRouteEnv, deployment, v1alpha1.DataIndex)
}

// InjectDataIndexURLIntoMgmtConsole will query for every KogitoRuntime in the given namespace to inject the Data Index route to each one
// Won't trigger an update if the KogitoRuntime already has the route set to avoid unnecessary reconciliation triggers
func InjectDataIndexURLIntoMgmtConsole(client *client.Client, namespace string) error {
	log.Debugf("Injecting Data-Index Route in Mgmt console")
	log.Debugf("Querying MgmtConsole instances in the namespace '%s' to inject a route ", namespace)
	deployment, err := getKogitoSupportingServiceDeploymentInstance(client, namespace, v1alpha1.MgmtConsole)
	if err != nil {
		return err
	}
	if deployment == nil {
		log.Debugf("No deployment found for MgmtConsole, skipping to inject %s URL into MgmtConsole", v1alpha1.DataIndex)
		return nil
	}

	log.Debugf("Querying %s route to inject into MgmtConsole", v1alpha1.DataIndex)
	serviceEndpoints, err := getServiceEndpoints(client, namespace, dataIndexHTTPRouteEnv, dataIndexWSRouteEnv, v1alpha1.DataIndex)
	if err != nil {
		return err
	}
	if serviceEndpoints != nil {
		log.Debugf("The %s route is '%s'", v1alpha1.DataIndex, serviceEndpoints.HTTPRouteURI)

		updateHTTP, updateWS := updateServiceEndpointIntoDeploymentEnv(deployment, serviceEndpoints)
		// update only once
		if updateWS || updateHTTP {
			if err := kubernetes.ResourceC(client).Update(deployment); err != nil {
				return err
			}
		}
	}
	return nil
}

// MountProtoBufConfigMapsOnDeployment mounts protobuf configMaps from KogitoRuntime services into the given deployment
func MountProtoBufConfigMapsOnDeployment(client *client.Client, deployment *appsv1.Deployment) (err error) {
	cms, err := getProtoBufConfigMapsForAllRuntimeServices(deployment.Namespace, client)
	if err != nil || len(cms.Items) == 0 {
		return err
	}
	for _, cm := range cms.Items {
		appendProtoBufVolumeIntoDeployment(deployment, cm)
		appendProtoBufVolumeMountIntoDeployment(deployment, cm)
	}
	updateProtoBufPropInToDeploymentEnv(deployment)
	return nil
}

// MountProtoBufConfigMapOnDataIndex mounts protobuf configMaps from KogitoRuntime services into the given deployment instance of DataIndex
func MountProtoBufConfigMapOnDataIndex(client *client.Client, kogitoService v1alpha1.KogitoService) (err error) {
	deployment, err := getKogitoSupportingServiceDeploymentInstance(client, kogitoService.GetNamespace(), v1alpha1.DataIndex)
	if err != nil || deployment == nil {
		return
	}

	cms, err := getProtoBufConfigMapsForSpecificRuntimeService(client, kogitoService.GetName(), kogitoService.GetNamespace())
	if err != nil || len(cms.Items) == 0 {
		return
	}
	for _, cm := range cms.Items {
		appendProtoBufVolumeIntoDeployment(deployment, cm)
		appendProtoBufVolumeMountIntoDeployment(deployment, cm)
	}
	updateProtoBufPropInToDeploymentEnv(deployment)
	if err := kubernetes.ResourceC(client).Update(deployment); err != nil {
		return err
	}
	return nil
}

func appendProtoBufVolumeIntoDeployment(deployment *appsv1.Deployment, cm corev1.ConfigMap) {
	configMapDefaultMode := protoBufConfigMapVolumeDefaultMode
	volumeExists := false
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		if volume.Name == cm.Name {
			volumeExists = true
			break
		}
	}
	if !volumeExists {
		deployment.Spec.Template.Spec.Volumes =
			append(deployment.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: cm.Name,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						DefaultMode: &configMapDefaultMode,
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cm.Name,
						},
					},
				},
			})
	}
}

func appendProtoBufVolumeMountIntoDeployment(deployment *appsv1.Deployment, cm corev1.ConfigMap) {
	for fileName := range cm.Data {

		volumeMountExists := false
		mountPath := path.Join(defaultProtobufMountPath, cm.Labels["app"], fileName)
		for _, volumeMount := range deployment.Spec.Template.Spec.Containers[0].VolumeMounts {
			if volumeMount.MountPath == mountPath {
				volumeMountExists = true
				break
			}
		}

		if !volumeMountExists {
			deployment.Spec.Template.Spec.Containers[0].VolumeMounts =
				append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
					Name:      cm.Name,
					MountPath: mountPath,
					SubPath:   fileName,
				})
		}
	}
}

func updateProtoBufPropInToDeploymentEnv(deployment *appsv1.Deployment) {
	if len(deployment.Spec.Template.Spec.Volumes) > 0 {
		framework.SetEnvVar(protoBufKeyWatch, "true", &deployment.Spec.Template.Spec.Containers[0])
		framework.SetEnvVar(protoBufKeyFolder, defaultProtobufMountPath, &deployment.Spec.Template.Spec.Containers[0])
	} else {
		framework.SetEnvVar(protoBufKeyWatch, "false", &deployment.Spec.Template.Spec.Containers[0])
		framework.SetEnvVar(protoBufKeyFolder, "", &deployment.Spec.Template.Spec.Containers[0])
	}
}
