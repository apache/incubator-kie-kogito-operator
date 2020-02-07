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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

const (
	portName                          = "http"
	enablePersistenceEnvKey           = "ENABLE_PERSISTENCE"
	backOffRetryEnvKey                = "BACKOFF_RETRY"
	maxIntervalLimitRetryEnvKey       = "MAX_INTERVAL_LIMIT_RETRY"
	backOffRetryDefaultValue          = 1000
	maxIntervalLimitRetryDefaultValue = 60000
)

var defaultProbe = &corev1.Probe{
	Handler: corev1.Handler{
		TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{IntVal: framework.DefaultExposedPort}},
	},
	TimeoutSeconds:   int32(1),
	PeriodSeconds:    int32(10),
	SuccessThreshold: int32(1),
	FailureThreshold: int32(3),
}

func createRequiredDeployment(jobService *v1alpha1.KogitoJobsService, image *imageHandler, infinispanSecret *corev1.Secret) *appsv1.Deployment {
	if &jobService.Spec.Replicas == nil {
		jobService.Spec.Replicas = 1
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: jobService.Name, Namespace: jobService.Namespace, Labels: map[string]string{labelAppKey: jobService.Name}},
		Spec: appsv1.DeploymentSpec{
			Replicas: &jobService.Spec.Replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{labelAppKey: jobService.Name}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{labelAppKey: jobService.Name}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  jobService.Name,
							Image: image.ResolveImage(),
							Ports: []corev1.ContainerPort{
								{
									Name:          portName,
									ContainerPort: framework.DefaultExposedPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env:             jobService.Spec.Envs,
							Resources:       jobService.Spec.Resources,
							LivenessProbe:   defaultProbe,
							ReadinessProbe:  defaultProbe,
							ImagePullPolicy: corev1.PullAlways,
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RollingUpdateDeploymentStrategyType},
		},
	}

	if image.hasImageStream() {
		key, value := framework.ResolveImageStreamTriggerAnnotation(image.resolveImageNameTag(), jobService.Name)
		deployment.Annotations = map[string]string{key: value}
	}

	infrastructure.SetInfinispanVariables(jobService.Spec.InfinispanProperties, infinispanSecret, &deployment.Spec.Template.Spec.Containers[0])

	if &jobService.Spec.InfinispanProperties != nil &&
		(jobService.Spec.InfinispanProperties.UseKogitoInfra || len(jobService.Spec.InfinispanProperties.URI) > 0) {
		framework.SetEnvVar(enablePersistenceEnvKey, "true", &deployment.Spec.Template.Spec.Containers[0])
	}

	if &jobService.Spec.BackOffRetryMillis != nil {
		if jobService.Spec.BackOffRetryMillis <= 0 {
			jobService.Spec.BackOffRetryMillis = backOffRetryDefaultValue
		}
		framework.SetEnvVar(backOffRetryEnvKey, strconv.FormatInt(jobService.Spec.BackOffRetryMillis, 10), &deployment.Spec.Template.Spec.Containers[0])
	}

	if &jobService.Spec.MaxIntervalLimitToRetryMillis != nil {
		if jobService.Spec.MaxIntervalLimitToRetryMillis <= 0 {
			jobService.Spec.MaxIntervalLimitToRetryMillis = maxIntervalLimitRetryDefaultValue
		}
		framework.SetEnvVar(maxIntervalLimitRetryEnvKey, strconv.FormatInt(jobService.Spec.MaxIntervalLimitToRetryMillis, 10), &deployment.Spec.Template.Spec.Containers[0])
	}

	return deployment
}
