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

package dataindex

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"path"
)

const (
	defaultProtobufMountPath                  = "/home/kogito/data/protobufs"
	protoBufKeyFolder                  string = "KOGITO_PROTOBUF_FOLDER"
	protoBufKeyWatch                   string = "KOGITO_PROTOBUF_WATCH"
	protoBufConfigMapVolumeDefaultMode int32  = 420
)

func onDeploymentCreate(cli *client.Client, deployment *appsv1.Deployment, kogitoService v1alpha1.KogitoService) error {
	if len(deployment.Spec.Template.Spec.Containers) > 0 {
		if err := mountProtoBufConfigMaps(deployment, cli); err != nil {
			return err
		}
	} else {
		log.Warnf("No container definition for service %s. Skipping applying custom Data Index deployment configuration", kogitoService.GetName())
	}

	return nil
}

// mountProtoBufConfigMaps mounts protobuf configMaps from KogitoRuntime services into the given deployment
func mountProtoBufConfigMaps(deployment *appsv1.Deployment, client *client.Client) (err error) {
	var cms *corev1.ConfigMapList
	configMapDefaultMode := protoBufConfigMapVolumeDefaultMode
	if cms, err = infrastructure.GetProtoBufConfigMaps(deployment.Namespace, client); err != nil {
		return err
	}
	for _, cm := range cms.Items {
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
		for fileName := range cm.Data {
			deployment.Spec.Template.Spec.Containers[0].VolumeMounts =
				append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts,
					corev1.VolumeMount{Name: cm.Name, MountPath: path.Join(defaultProtobufMountPath, cm.Labels["app"], fileName), SubPath: fileName})
		}
	}

	if len(deployment.Spec.Template.Spec.Volumes) > 0 {
		framework.SetEnvVar(protoBufKeyWatch, "true", &deployment.Spec.Template.Spec.Containers[0])
		framework.SetEnvVar(protoBufKeyFolder, defaultProtobufMountPath, &deployment.Spec.Template.Spec.Containers[0])
	} else {
		framework.SetEnvVar(protoBufKeyWatch, "false", &deployment.Spec.Template.Spec.Containers[0])
		framework.SetEnvVar(protoBufKeyFolder, "", &deployment.Spec.Template.Spec.Containers[0])
	}

	return nil
}
