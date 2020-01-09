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
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	dockerv10 "github.com/openshift/api/image/docker10"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// newServiceMonitor creates a new ServiceMonitor resource for the KogitoApp based on the Prometheus labels of the runtime image
func newServiceMonitor(kogitoApp *v1alpha1.KogitoApp, dockerImage *dockerv10.DockerImage, service *corev1.Service, client *client.Client) (*monv1.ServiceMonitor, error) {
	if !isPrometheusOperatorReady(client) {
		return nil, nil
	}

	scrape, scheme, path, port, err := framework.ExtractPrometheusConfigurationFromImage(dockerImage)

	if err != nil {
		return nil, err
	}

	if scrape {
		endPoint := monv1.Endpoint{}
		if port != nil {
			endPoint.TargetPort = port
		} else {
			endPoint.Port = "http"
		}

		if len(path) > 0 {
			endPoint.Path = path
		} else {
			endPoint.Path = "/metrics"
		}

		if len(scheme) > 0 {
			endPoint.Scheme = scheme
		} else {
			endPoint.Scheme = "http"
		}

		sm := &monv1.ServiceMonitor{
			ObjectMeta: metav1.ObjectMeta{
				Name:      kogitoApp.Name,
				Namespace: kogitoApp.Namespace,
			},
			Spec: monv1.ServiceMonitorSpec{
				NamespaceSelector: monv1.NamespaceSelector{
					MatchNames: []string{
						kogitoApp.Namespace,
					},
				},
				Selector: metav1.LabelSelector{
					MatchLabels: service.ObjectMeta.Labels,
				},
				Endpoints: []monv1.Endpoint{
					endPoint,
				},
			},
		}

		meta.SetGroupVersionKind(&sm.TypeMeta, meta.KindServiceMonitor)
		addDefaultMeta(&sm.ObjectMeta, kogitoApp)

		return sm, nil
	}

	return nil, nil
}

func isPrometheusOperatorReady(client *client.Client) bool {
	return client.HasServerGroup("monitoring.coreos.com")
}
