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

package services

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	portName      = "http"
	labelAppKey   = "app"
	singleReplica = 1
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

func createRequiredDeployment(service v1alpha1.KogitoService, image *imageHandler, definition ServiceDefinition) *appsv1.Deployment {
	if definition.SingleReplica && service.GetSpec().GetReplicas() > singleReplica {
		service.GetSpec().SetReplicas(singleReplica)
		log.Warnf("%s can't scale vertically, only one replica is allowed.", service.GetName())
	}
	replicas := service.GetSpec().GetReplicas()

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: service.GetName(), Namespace: service.GetNamespace(), Labels: map[string]string{labelAppKey: service.GetName()}},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{labelAppKey: service.GetName()}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{labelAppKey: service.GetName()}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: service.GetName(),
							Ports: []corev1.ContainerPort{
								{
									Name:          portName,
									ContainerPort: framework.DefaultExposedPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env:             service.GetSpec().GetEnvs(),
							Resources:       service.GetSpec().GetResources(),
							LivenessProbe:   defaultProbe,
							ReadinessProbe:  defaultProbe,
							ImagePullPolicy: corev1.PullAlways,
							Image:           image.resolveImage(),
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RollingUpdateDeploymentStrategyType},
		},
	}

	if image.hasImageStream() {
		key, value := framework.ResolveImageStreamTriggerAnnotation(image.resolveImageNameTag(), service.GetName())
		deployment.Annotations = map[string]string{key: value}
	}

	return deployment
}
