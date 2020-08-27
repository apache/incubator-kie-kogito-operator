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

package services

import (
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const prometheusServerGroup = "monitoring.coreos.com"

// IsPrometheusAvailable checks if Prometheus CRD is available in the cluster
func IsPrometheusAvailable(client *client.Client) bool {
	return client.HasServerGroup(prometheusServerGroup)
}

// CreateServiceMonitor create ServiceMonitor used for scraping by prometheus for kogito service
func CreateServiceMonitor(kogitoRuntime *v1alpha1.KogitoRuntime) *monv1.ServiceMonitor {
	prometheus := kogitoRuntime.Spec.Prometheus
	endPoint := monv1.Endpoint{}
	endPoint.TargetPort = &intstr.IntOrString{IntVal: getServiceHTTPPort(kogitoRuntime)}

	if len(prometheus.Path) > 0 {
		endPoint.Path = prometheus.Path
	} else {
		endPoint.Path = v1alpha1.PrometheusDefaultPath
	}

	if len(prometheus.Scheme) > 0 {
		endPoint.Scheme = prometheus.Scheme
	} else {
		endPoint.Scheme = v1alpha1.PrometheusDefaultScheme
	}

	serviceSelectorLabels := make(map[string]string)
	serviceSelectorLabels[framework.LabelAppKey] = kogitoRuntime.GetName()

	serviceMonitorLabels := make(map[string]string)
	serviceMonitorLabels["name"] = operator.Name

	sm := &monv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitoRuntime.Name,
			Namespace: kogitoRuntime.Namespace,
			Labels:    serviceMonitorLabels,
		},
		Spec: monv1.ServiceMonitorSpec{
			NamespaceSelector: monv1.NamespaceSelector{
				MatchNames: []string{
					kogitoRuntime.Namespace,
				},
			},
			Selector: metav1.LabelSelector{
				MatchLabels: serviceSelectorLabels,
			},
			Endpoints: []monv1.Endpoint{
				endPoint,
			},
		},
	}
	return sm
}
