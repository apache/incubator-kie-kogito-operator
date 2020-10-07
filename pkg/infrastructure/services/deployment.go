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
	"strconv"
)

const (
	portName      = "http"
	singleReplica = int32(1)
	// HTTPPortEnvKey Env variable to define port on which service will listen internally
	HTTPPortEnvKey = "HTTP_PORT"
)

func createRequiredDeployment(service v1alpha1.KogitoService, resolvedImage string, definition ServiceDefinition) *appsv1.Deployment {
	if definition.SingleReplica && *service.GetSpec().GetReplicas() > singleReplica {
		service.GetSpec().SetReplicas(singleReplica)
		log.Warnf("%s can't scale vertically, only one replica is allowed.", service.GetName())
	}
	replicas := service.GetSpec().GetReplicas()
	httpPort := getServiceHTTPPort(service)
	setHTTPPortInEnvVar(httpPort, service)
	probes := getProbeForKogitoService(definition, httpPort)
	labels := service.GetSpec().GetDeploymentLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[framework.LabelAppKey] = service.GetName()

	// clone env var slice so that any changes in deployment env var should not reflect in kogitoInstance env var
	env := make([]corev1.EnvVar, len(service.GetSpec().GetEnvs()))
	copy(env, service.GetSpec().GetEnvs())

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: service.GetName(), Namespace: service.GetNamespace(), Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{framework.LabelAppKey: service.GetName()}},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: service.GetName(),
							Ports: []corev1.ContainerPort{
								{
									Name:          portName,
									ContainerPort: httpPort,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env:             env,
							Resources:       service.GetSpec().GetResources(),
							LivenessProbe:   probes.liveness,
							ReadinessProbe:  probes.readiness,
							ImagePullPolicy: corev1.PullAlways,
							Image:           resolvedImage,
						},
					},
				},
			},
			Strategy: appsv1.DeploymentStrategy{Type: appsv1.RollingUpdateDeploymentStrategyType},
		},
	}

	return deployment
}

// getServiceHTTPPort gets the service port for the given KogitoService based on httpPort CR parameter.
// defaults to 8080
func getServiceHTTPPort(kogitoService v1alpha1.KogitoService) int32 {
	// port should be greater than 0
	httpPort := kogitoService.GetSpec().GetHTTPPort()
	if httpPort < 1 {
		log.Debugf("HTTPPort not set, returning default http port.")
		return framework.DefaultExposedPort
	}
	log.Debugf("HTTPPort is set, returning port number %i", httpPort)
	return httpPort
}

// setHTTPPortInEnvVar will update or add the environment variable into the given kogito service
func setHTTPPortInEnvVar(httpPort int32, kogitoService v1alpha1.KogitoService) {
	httpPortEnvVar := corev1.EnvVar{
		Name:  HTTPPortEnvKey,
		Value: strconv.FormatInt(int64(httpPort), 10),
	}
	envs := kogitoService.GetSpec().GetEnvs()
	modifiedEnv := framework.EnvOverride(envs, httpPortEnvVar)
	kogitoService.GetSpec().SetEnvs(modifiedEnv)
}
