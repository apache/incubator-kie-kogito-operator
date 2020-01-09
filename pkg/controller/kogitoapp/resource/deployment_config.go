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
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/shared"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

const (
	kogitoHome = "/home/kogito"

	defaultReplicas = int32(1)
	// ServiceAccountName is the name of service account used by Kogito Services Runtimes
	ServiceAccountName = "kogito-service-viewer"

	envVarInfinispanServerList     = "SERVER_LIST"
	envVarInfinispanUser           = "USERNAME"
	envVarInfinispanPassword       = "PASSWORD"
	envVarInfinispanSaslMechanism  = "SASL_MECHANISM"
	defaultInfinispanSaslMechanism = v1alpha1.SASLPlain

	envVarKafkaBootstrapURI    = "KAFKA_BOOTSTRAP_SERVERS"
	envVarKafkaBootstrapSuffix = "_BOOTSTRAP_SERVERS"

	envVarExternalURL        = "KOGITO_SERVICE_URL"
	downwardAPIVolumeName    = "podinfo"
	downwardAPIVolumeMount   = kogitoHome + "/" + downwardAPIVolumeName
	downwardAPIProtoBufCMKey = "protobufcm"

	postHookPersistenceScript = kogitoHome + "/launch/post-hook-persistence.sh"
)

var (
	/*
		Infinispan variables for the KogitoInfra deployed infrastructure.
		For Quarkus: https://quarkus.io/guides/infinispan-client#quarkus-infinispan-client_configuration
		For Spring: https://github.com/infinispan/infinispan-spring-boot/blob/master/infinispan-spring-boot-starter-remote/src/test/resources/test-application.properties
	*/

	envVarInfinispanQuarkus = map[string]string{
		envVarInfinispanServerList:    "QUARKUS_INFINISPAN_CLIENT_SERVER_LIST",
		envVarInfinispanUser:          "QUARKUS_INFINISPAN_CLIENT_AUTH_USERNAME",
		envVarInfinispanPassword:      "QUARKUS_INFINISPAN_CLIENT_AUTH_PASSWORD",
		envVarInfinispanSaslMechanism: "QUARKUS_INFINISPAN_CLIENT_SASL_MECHANISM",
	}
	envVarInfinispanSpring = map[string]string{
		envVarInfinispanServerList:    "INFINISPAN_REMOTE_SERVER_LIST",
		envVarInfinispanUser:          "INFINISPAN_REMOTE_AUTH_USER_NAME",
		envVarInfinispanPassword:      "INFINISPAN_REMOTE_AUTH_PASSWORD",
		envVarInfinispanSaslMechanism: "INFINISPAN_REMOTE_SASL_MECHANISM",
	}

	downwardAPIDefaultMode = int32(420)

	podStartExecCommand = []string{"/bin/bash", "-c", "if [ -x " + postHookPersistenceScript + " ]; then " + postHookPersistenceScript + "; fi"}
)

// newDeploymentConfig creates a new DeploymentConfig resource for the KogitoApp based on the BuildConfig runner image
func newDeploymentConfig(kogitoApp *v1alpha1.KogitoApp, runnerBC *buildv1.BuildConfig, dockerImage *dockerv10.DockerImage) (dc *appsv1.DeploymentConfig, err error) {
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
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{
					downwardAPIProtoBufCMKey: GenerateProtoBufConfigMapName(kogitoApp),
				}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: kogitoApp.Name,
							// this conversion will be removed in future versions
							Env: shared.FromEnvToEnvVar(kogitoApp.Spec.Env),
							// this conversion will be removed in future versions
							Resources:       shared.FromResourcesToResourcesRequirements(kogitoApp.Spec.Resources),
							Image:           runnerBC.Spec.Output.To.Name,
							ImagePullPolicy: corev1.PullAlways,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      downwardAPIVolumeName,
									MountPath: downwardAPIVolumeMount,
								},
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
	setReplicas(kogitoApp, dc)

	return dc, nil
}

// setReplicas defines the number of container replicas that this DeploymentConfig will have
func setReplicas(kogitoApp *v1alpha1.KogitoApp, dc *appsv1.DeploymentConfig) {
	replicas := defaultReplicas
	if kogitoApp.Spec.Replicas != nil {
		replicas = *kogitoApp.Spec.Replicas
	}
	dc.Spec.Replicas = replicas
}

// SetInfinispanEnvVars sets Infinispan variables to the given KogitoApp instance DeploymentConfig by reading information from the KogitoInfra
func SetInfinispanEnvVars(cli *client.Client, kogitoInfra *v1alpha1.KogitoInfra, kogitoApp *v1alpha1.KogitoApp, dc *appsv1.DeploymentConfig) error {
	if dc != nil && kogitoApp != nil &&
		(kogitoInfra != nil && &kogitoInfra.Status != nil && &kogitoInfra.Status.Infinispan != nil) {
		uri, err := infrastructure.GetInfinispanServiceURI(cli, kogitoInfra)
		if err != nil {
			return err
		}
		secret, err := infrastructure.GetInfinispanCredentialsSecret(cli, kogitoInfra)
		if err != nil {
			return err
		}

		// inject credentials to deploymentConfig container
		if len(dc.Spec.Template.Spec.Containers) > 0 {
			vars := envVarInfinispanQuarkus
			if kogitoApp.Spec.Runtime == v1alpha1.SpringbootRuntimeType {
				vars = envVarInfinispanSpring
			}
			framework.SetEnvVar(vars[envVarInfinispanServerList], uri, &dc.Spec.Template.Spec.Containers[0])
			framework.SetEnvVar(vars[envVarInfinispanSaslMechanism], string(defaultInfinispanSaslMechanism), &dc.Spec.Template.Spec.Containers[0])
			framework.SetEnvVarFromSecret(vars[envVarInfinispanUser], infrastructure.InfinispanSecretUsernameKey, secret, &dc.Spec.Template.Spec.Containers[0])
			framework.SetEnvVarFromSecret(vars[envVarInfinispanPassword], infrastructure.InfinispanSecretPasswordKey, secret, &dc.Spec.Template.Spec.Containers[0])
		}
	}
	return nil
}

// SetKafkaEnvVars sets Kafka variables to the given KogitoApp instance DeploymentConfig by reading information from the KogitoInfra
func SetKafkaEnvVars(cli *client.Client, kogitoInfra *v1alpha1.KogitoInfra, kogitoApp *v1alpha1.KogitoApp, dc *appsv1.DeploymentConfig) error {
	if dc != nil && kogitoApp != nil &&
		(kogitoInfra != nil && &kogitoInfra.Status != nil && &kogitoInfra.Status.Kafka != nil) {
		uri, err := infrastructure.GetKafkaServiceURI(cli, kogitoInfra)
		if err != nil {
			return err
		}
		if len(dc.Spec.Template.Spec.Containers) > 0 {
			framework.SetEnvVar(envVarKafkaBootstrapURI, uri, &dc.Spec.Template.Spec.Containers[0])
			// let's also add a secret feature that injects all _BOOTSTRAP_SERVERS env vars with the correct uri :p
			for _, env := range dc.Spec.Template.Spec.Containers[0].Env {
				if strings.HasSuffix(env.Name, envVarKafkaBootstrapSuffix) {
					framework.SetEnvVar(env.Name, uri, &dc.Spec.Template.Spec.Containers[0])
				}
			}
		}
	}
	return nil
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
