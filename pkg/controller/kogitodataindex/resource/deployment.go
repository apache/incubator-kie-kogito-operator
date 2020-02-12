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

package resource

import (
	"fmt"
	"path"
	"strconv"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	imgv1 "github.com/openshift/api/image/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func newDeployment(instance *v1alpha1.KogitoDataIndex, infinispanSecret *corev1.Secret, kafkaExternalURI string, cli *client.Client, imageStream *imgv1.ImageStream) (*appsv1.Deployment, error) {
	// define the http port
	httpPort := defineDataIndexHTTPPort(instance)
	log.Debugf("The configured internal Data Index port number is [%i]", httpPort)

	// create a standard probe
	probe := defaultProbe
	probe.Handler.TCPSocket = &corev1.TCPSocketAction{Port: intstr.IntOrString{IntVal: httpPort}}

	if instance.Spec.Replicas == 0 {
		instance.Spec.Replicas = defaultReplicas
	}
	if len(instance.Spec.Image) == 0 {
		instance.Spec.Image = infrastructure.DefaultDataIndexImage
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &instance.Spec.Replicas,
			Selector: &metav1.LabelSelector{},
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.RollingUpdateDeploymentStrategyType,
			},
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            instance.Name,
							Image:           instance.Spec.Image,
							Env:             framework.MapToEnvVar(instance.Spec.Env),
							Resources:       extractResources(instance),
							ImagePullPolicy: corev1.PullAlways,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: httpPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							LivenessProbe:  probe,
							ReadinessProbe: probe,
						},
					},
				},
			},
		},
	}

	// protobuf mounting
	if err := mountProtoBufConfigMaps(deployment, cli); err != nil {
		return nil, err
	}

	// add configurable environment variables to the container
	infrastructure.SetInfinispanVariables(instance.Spec.InfinispanProperties, infinispanSecret, &deployment.Spec.Template.Spec.Containers[0])
	setKafkaVariables(kafkaExternalURI, &deployment.Spec.Template.Spec.Containers[0])
	framework.SetEnvVar(DataIndexEnvKeyHTTPPort, strconv.Itoa(int(httpPort)), &deployment.Spec.Template.Spec.Containers[0])

	// metadata information
	meta.SetGroupVersionKind(&deployment.TypeMeta, meta.KindDeployment)
	addDefaultMetadata(&deployment.ObjectMeta, instance)
	addDefaultMetadata(&deployment.Spec.Template.ObjectMeta, instance)
	deployment.Spec.Selector.MatchLabels = deployment.Labels

	// Image Stream
	if imageStream != nil {
		_, _, name, tag := framework.SplitImageTag(instance.Spec.Image)
		imageStreamTag := fmt.Sprintf("%s:%s", name, tag)
		key, value := framework.ResolveImageStreamTriggerAnnotation(imageStreamTag, instance.Name)
		deployment.Annotations[key] = value
	}

	return deployment, nil
}

// mountProtoBufConfigMaps mounts protobuf configMaps from KogitoApps into the given stateful set
func mountProtoBufConfigMaps(deployment *appsv1.Deployment, cli *client.Client) (err error) {
	var cms *corev1.ConfigMapList
	if cms, err = infrastructure.GetProtoBufConfigMaps(deployment.Namespace, cli); err != nil {
		return err
	}

	for _, cm := range cms.Items {
		deployment.Spec.Template.Spec.Volumes =
			append(deployment.Spec.Template.Spec.Volumes, corev1.Volume{
				Name: cm.Name,
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cm.Name,
						},
					},
				},
			})
		deployment.Spec.Template.Spec.Containers[0].VolumeMounts =
			append(deployment.Spec.Template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{Name: cm.Name, MountPath: path.Join(defaultProtobufMountPath, cm.Labels["app"])})
	}
	protoBufEnvs := protoBufEnvsNoVolume
	if len(deployment.Spec.Template.Spec.Volumes) > 0 {
		protoBufEnvs = protoBufEnvsVolumeMounted
	}
	for k, v := range protoBufEnvs {
		framework.SetEnvVar(k, v, &deployment.Spec.Template.Spec.Containers[0])
	}

	return nil
}

// defineDataIndexHTTPPort will define which port the dataindex should be listening to. To set it use httpPort cr parameter.
// defaults to 8080
func defineDataIndexHTTPPort(instance *v1alpha1.KogitoDataIndex) int32 {
	// port should be greater than 0
	if instance.Spec.HTTPPort < 1 {
		log.Debugf("HTTPPort not set, returning default http port.")
		return framework.DefaultExposedPort
	}
	log.Debugf("HTTPPort is set, returning port number %i", int(instance.Spec.HTTPPort))
	return instance.Spec.HTTPPort
}
