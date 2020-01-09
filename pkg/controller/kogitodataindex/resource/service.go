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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newService creates a Service resource based on deployment spec. It's expected that the container exposes at least one port to be able to create a valid service
func newService(instance *v1alpha1.KogitoDataIndex, statefulset *appsv1.StatefulSet) *corev1.Service {
	ports := framework.ExtractPortsFromContainer(&statefulset.Spec.Template.Spec.Containers[0])
	if len(ports) == 0 {
		// a service without port to expose doesn't exist
		log.Warnf("The deployment spec '%s' doesn't have any ports exposed. Won't be possible to create a new service.", statefulset.Name)
		return nil
	}
	svc := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports:    ports,
			Selector: statefulset.Spec.Selector.MatchLabels,
			Type:     corev1.ServiceTypeClusterIP,
		},
	}

	addDefaultMetadata(&svc.ObjectMeta, instance)
	meta.SetGroupVersionKind(&svc.TypeMeta, meta.KindService)

	statefulset.Spec.ServiceName = svc.Name

	return &svc
}
