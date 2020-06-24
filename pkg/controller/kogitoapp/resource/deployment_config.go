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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strconv"
)

const (
	kogitoHome = "/home/kogito"

	defaultReplicas = int32(1)
	// ServiceAccountName is the name of service account used by Kogito Services Runtimes
	ServiceAccountName = "kogito-service-viewer"

	envVarExternalURL        = "KOGITO_SERVICE_URL"
	downwardAPIVolumeName    = "podinfo"
	downwardAPIVolumeMount   = kogitoHome + "/" + downwardAPIVolumeName
	downwardAPIProtoBufCMKey = "protobufcm"

	postHookPersistenceScript = kogitoHome + "/launch/post-hook-persistence.sh"

	envVarNamespace = "NAMESPACE"
)

var (
	downwardAPIDefaultMode = int32(420)

	podStartExecCommand = []string{"/bin/bash", "-c", "if [ -x " + postHookPersistenceScript + " ]; then " + postHookPersistenceScript + "; fi"}
)

// newDeploymentConfig creates a new DeploymentConfig resource for the KogitoApp based on the BuildConfig runner image
func newDeploymentConfig(kogitoApp *v1alpha1.KogitoApp, runnerBC *buildv1.BuildConfig, dockerImage *dockerv10.DockerImage, appPropContentHash string) (dc *appsv1.DeploymentConfig, err error) {
	if runnerBC == nil {
		return nil, fmt.Errorf("Impossible to create the DeploymentConfig without a reference to a the service BuildConfig")
	}

	dc = &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitoApp.Name,
			Namespace: kogitoApp.Namespace,
		},
		Spec: appsv1.DeploymentConfigSpec{
			Strategy: appsv1.DeploymentStrategy{
				Type: appsv1.DeploymentStrategyTypeRolling,
			},
			Template: &corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						downwardAPIProtoBufCMKey: GenerateProtoBufConfigMapName(kogitoApp),
					},
					Annotations: map[string]string{
						services.AppPropContentHashKey: appPropContentHash,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            kogitoApp.Name,
							Env:             kogitoApp.Spec.KogitoServiceSpec.Envs,
							Resources:       kogitoApp.Spec.KogitoServiceSpec.Resources,
							Image:           runnerBC.Spec.Output.To.Name,
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      downwardAPIVolumeName,
									MountPath: downwardAPIVolumeMount,
								},
								services.CreateAppPropVolumeMount(),
							},
							Lifecycle: &corev1.Lifecycle{
								PostStart: &corev1.Handler{
									Exec: &corev1.ExecAction{Command: podStartExecCommand},
								},
							},
						},
					},
					ServiceAccountName: ServiceAccountName,
					Volumes: []corev1.Volume{
						{
							Name: downwardAPIVolumeName,
							VolumeSource: corev1.VolumeSource{
								DownwardAPI: &corev1.DownwardAPIVolumeSource{
									Items: []corev1.DownwardAPIVolumeFile{
										{Path: "name", FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name", APIVersion: "v1"}},
										{Path: downwardAPIProtoBufCMKey, FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.labels['" + downwardAPIProtoBufCMKey + "']", APIVersion: "v1"}},
									},
									DefaultMode: &downwardAPIDefaultMode,
								},
							},
						},
						services.CreateAppPropVolume(kogitoApp.Name),
					},
				},
			},
			Triggers: appsv1.DeploymentTriggerPolicies{
				{Type: appsv1.DeploymentTriggerOnConfigChange},
				{
					Type: appsv1.DeploymentTriggerOnImageChange,
					ImageChangeParams: &appsv1.DeploymentTriggerImageChangeParams{
						Automatic:      true,
						ContainerNames: []string{kogitoApp.Name},
						From:           *runnerBC.Spec.Output.To,
					},
				},
			},
		},
	}

	meta.SetGroupVersionKind(&dc.TypeMeta, meta.KindDeploymentConfig)
	addDefaultMeta(&dc.ObjectMeta, kogitoApp)
	addDefaultMeta(&dc.Spec.Template.ObjectMeta, kogitoApp)
	addDefaultLabels(&dc.Spec.Selector, kogitoApp)
	framework.MergeImageMetadataWithDeploymentConfig(dc, dockerImage)
	framework.DiscoverPortsAndProbesFromImage(dc, dockerImage)

	if kogitoApp.Spec.EnableIstio {
		framework.AddIstioInjectSidecarAnnotation(&dc.Spec.Template.ObjectMeta)
	}

	setReplicas(kogitoApp, dc)
	setNamespaceEnvVars(kogitoApp, dc)
	setHTTPPortEnvVar(kogitoApp, dc)

	return dc, nil
}

func setNamespaceEnvVars(kogitoApp *v1alpha1.KogitoApp, dc *appsv1.DeploymentConfig) {
	framework.SetEnvVar(envVarNamespace, kogitoApp.Namespace, &dc.Spec.Template.Spec.Containers[0])
}

// setReplicas defines the number of container replicas that this DeploymentConfig will have
func setReplicas(kogitoApp *v1alpha1.KogitoApp, dc *appsv1.DeploymentConfig) {
	replicas := defaultReplicas
	if kogitoApp.Spec.KogitoServiceSpec.Replicas != nil {
		replicas = *kogitoApp.Spec.KogitoServiceSpec.Replicas
	}
	dc.Spec.Replicas = replicas
}

// SetExternalRouteEnvVar sets the external URL to the given requested deploymentConfig
func SetExternalRouteEnvVar(cli *client.Client, kogitoApp *v1alpha1.KogitoApp, dc *appsv1.DeploymentConfig) error {
	if dc == nil || kogitoApp == nil {
		return nil
	}

	if exists, route, err := openshift.RouteC(cli).GetHostFromRoute(types.NamespacedName{Namespace: kogitoApp.Namespace, Name: kogitoApp.Name}); err != nil {
		return err
	} else if exists {
		framework.SetEnvVar(envVarExternalURL, fmt.Sprintf("http://%s", route), &dc.Spec.Template.Spec.Containers[0])
	}

	return nil
}

func setHTTPPortEnvVar(kogitoApp *v1alpha1.KogitoApp, dc *appsv1.DeploymentConfig) {
	container := &dc.Spec.Template.Spec.Containers[0]
	// port should be greater than 0
	httpPort := kogitoApp.Spec.HTTPPort
	if httpPort < 1 {
		log.Debugf("HTTPPort not set, returning default http port.")
		httpPort = framework.DefaultExposedPort
	}
	framework.SetEnvVar(services.HTTPPortEnvKey, strconv.Itoa(int(httpPort)), container)
}
