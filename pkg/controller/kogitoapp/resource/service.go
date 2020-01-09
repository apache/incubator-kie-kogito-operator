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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"strings"

	appsv1 "github.com/openshift/api/apps/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// newService creates a Service resource based on the DC Containers ports exposed. Returns nil if no ports is found on Deployment Config
func newService(kogitoApp *v1alpha1.KogitoApp, deploymentConfig *appsv1.DeploymentConfig) (service *corev1.Service) {
	if deploymentConfig == nil {
		// we can't create a service without a DC
		return nil
	}

	ports := buildServicePorts(deploymentConfig)
	if len(ports) == 0 {
		return nil
	}

	service = &corev1.Service{
		ObjectMeta: *deploymentConfig.ObjectMeta.DeepCopy(),
		Spec: corev1.ServiceSpec{
			Selector: deploymentConfig.Spec.Selector,
			Type:     corev1.ServiceTypeClusterIP,
			Ports:    ports,
		},
	}

	meta.SetGroupVersionKind(&service.TypeMeta, meta.KindService)
	addDefaultMeta(&service.ObjectMeta, kogitoApp)
	addServiceLabels(&service.ObjectMeta, kogitoApp)
	importPrometheusAnnotations(deploymentConfig, service)
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

// importPrometheusAnnotations will import any annotations on the deploymentConfig template spec to the given service
// returns true if any modification has been made to the service
func importPrometheusAnnotations(deploymentConfig *appsv1.DeploymentConfig, service *corev1.Service) bool {
	if &deploymentConfig.Spec == nil || deploymentConfig.Spec.Template == nil || deploymentConfig.Spec.Template.Annotations == nil {
		return false
	}
	if service.Annotations == nil {
		service.Annotations = map[string]string{}
	}

	present := true
	for key, value := range deploymentConfig.Spec.Template.Annotations {
		if strings.Contains(key, framework.LabelKeyPrometheus) {
			if present {
				_, present = service.Annotations[key]
			}
			service.Annotations[key] = value
		}
	}
	return !present
}

// addServiceLabels adds the service labels
func addServiceLabels(objectMeta *metav1.ObjectMeta, kogitoApp *v1alpha1.KogitoApp) {
	if objectMeta != nil {
		if objectMeta.Labels == nil {
			objectMeta.Labels = map[string]string{}
		}

		addServiceLabelsToMap(objectMeta.Labels, kogitoApp)
	}
}

func addServiceLabelsToMap(labelsMap map[string]string, kogitoApp *v1alpha1.KogitoApp) {
	if kogitoApp.Spec.Service.Labels == nil {
		labelsMap[LabelKeyServiceName] = kogitoApp.Name
	} else {
		for key, value := range kogitoApp.Spec.Service.Labels {
			labelsMap[key] = value
		}
	}
}
