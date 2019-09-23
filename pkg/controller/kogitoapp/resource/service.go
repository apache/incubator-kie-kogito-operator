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
	v1alpha1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	appsv1 "github.com/openshift/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// NewService creates a Service resource based on the DC Containers ports exposed. Returns nil if no ports is found on Deployment Config
func NewService(kogitoApp *v1alpha1.KogitoApp, deploymentConfig *appsv1.DeploymentConfig) (service *corev1.Service) {
	if deploymentConfig == nil {
		// we can't create a service without a DC
		return nil
	}

	ports := buildServicePorts(deploymentConfig)
	if len(ports) == 0 {
		return nil
	}

	service = &corev1.Service{
		ObjectMeta: deploymentConfig.ObjectMeta,
		Spec: corev1.ServiceSpec{
			Selector: deploymentConfig.Spec.Selector,
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    ports,
		},
	}

	meta.SetGroupVersionKind(&service.TypeMeta, meta.KindService)
	addDefaultMeta(&service.ObjectMeta, kogitoApp)
	addServiceLabels(&service.ObjectMeta, kogitoApp)
	service.ResourceVersion = ""
	return service
}

func buildServicePorts(deploymentConfig *appsv1.DeploymentConfig) (ports []corev1.ServicePort) {
	ports = []corev1.ServicePort{}

	// for now, we should have at least 1 container defined.
	if len(deploymentConfig.Spec.Template.Spec.Containers) == 0 ||
		len(deploymentConfig.Spec.Template.Spec.Containers[0].Ports) == 0 {
		log.Warnf("The deploymentConfig spec '%s' doesn't have any ports exposed", deploymentConfig.Name)
		return ports
	}

	for _, port := range deploymentConfig.Spec.Template.Spec.Containers[0].Ports {
		ports = append(ports, corev1.ServicePort{
			Name:       port.Name,
			Protocol:   port.Protocol,
			Port:       port.ContainerPort,
			TargetPort: intstr.FromInt(int(port.ContainerPort)),
		},
		)
	}
	return ports
}
