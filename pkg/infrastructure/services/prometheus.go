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
	"net/http"

	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/operator"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

const prometheusServerGroup = "monitoring.coreos.com"

func configurePrometheus(client *client.Client, kogitoService v1alpha1.KogitoService, scheme *runtime.Scheme) (failedVerifyAddon bool, err error) {
	prometheusAvailable := isPrometheusAvailable(client)
	if !prometheusAvailable {
		log.Debugf("prometheus operator not available in namespace")
		return
	}

	deploymentAvailable, err := isDeploymentAvailable(client, kogitoService)
	if err != nil {
		return
	}
	if !deploymentAvailable {
		log.Debugf("Deployment is currently not available, will try in next reconciliation loop")
		return
	}

	prometheusAddOnAvailable, err := isPrometheusAddOnAvailable(kogitoService)
	if err != nil {
		return true, err
	}
	if prometheusAddOnAvailable {
		if err = createPrometheusServiceMonitorIfNotExists(client, kogitoService, scheme); err != nil {
			return
		}
	}
	return
}

// isPrometheusAvailable checks if Prometheus CRD is available in the cluster
func isPrometheusAvailable(client *client.Client) bool {
	return client.HasServerGroup(prometheusServerGroup)
}

func isPrometheusAddOnAvailable(kogitoService v1alpha1.KogitoService) (bool, error) {
	url := infrastructure.GetKogitoServiceURL(kogitoService)
	url = url + getMonitoringPath(kogitoService.GetSpec().GetMonitoring())
	if resp, err := http.Head(url); err != nil {
		return false, err
	} else if resp.StatusCode == http.StatusOK {
		return true, nil
	}
	log.Debugf("Non-OK Http Status received")
	return false, nil
}

func createPrometheusServiceMonitorIfNotExists(client *client.Client, kogitoService v1alpha1.KogitoService, scheme *runtime.Scheme) error {
	serviceMonitor, err := loadDeployedServiceMonitor(client, kogitoService.GetName(), kogitoService.GetNamespace())
	if err != nil {
		return err
	}
	if serviceMonitor == nil {
		_, err := createServiceMonitor(client, kogitoService, scheme)
		if err != nil {
			return err
		}
	}
	return nil
}

func loadDeployedServiceMonitor(client *client.Client, instanceName, namespace string) (*monv1.ServiceMonitor, error) {
	log.Debug("fetching deployed Service monitor instance with name %s in namespace %s", instanceName, namespace)
	serviceMonitor := &monv1.ServiceMonitor{}
	if exits, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: instanceName, Namespace: namespace}, serviceMonitor); err != nil {
		log.Error("Error occurs while fetching Service monitor instance")
		return nil, err
	} else if !exits {
		log.Debug("Service monitor instance is not exists")
		return nil, nil
	} else {
		log.Debug("Service monitor instance found")
		return serviceMonitor, nil
	}
}

// createServiceMonitor create ServiceMonitor used for scraping by prometheus for kogito service
func createServiceMonitor(cli *client.Client, kogitoService v1alpha1.KogitoService, scheme *runtime.Scheme) (*monv1.ServiceMonitor, error) {
	monitoring := kogitoService.GetSpec().GetMonitoring()
	endPoint := monv1.Endpoint{}
	endPoint.Path = getMonitoringPath(monitoring)
	endPoint.Scheme = getMonitoringScheme(monitoring)

	serviceSelectorLabels := make(map[string]string)
	serviceSelectorLabels[framework.LabelAppKey] = kogitoService.GetName()

	serviceMonitorLabels := make(map[string]string)
	serviceMonitorLabels["name"] = operator.Name
	serviceMonitorLabels[framework.LabelAppKey] = kogitoService.GetName()

	sm := &monv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kogitoService.GetName(),
			Namespace: kogitoService.GetNamespace(),
			Labels:    serviceMonitorLabels,
		},
		Spec: monv1.ServiceMonitorSpec{
			NamespaceSelector: monv1.NamespaceSelector{
				MatchNames: []string{
					kogitoService.GetNamespace(),
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

	if err := framework.SetOwner(kogitoService, scheme, sm); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(cli).Create(sm); err != nil {
		log.Error("Error occurs while creating Service Monitor instance")
		return nil, err
	}
	return sm, nil
}

func getMonitoringPath(monitoring v1alpha1.Monitoring) string {
	path := monitoring.Path
	if len(path) == 0 {
		path = v1alpha1.MonitoringDefaultPath
	}
	return path
}

func getMonitoringScheme(monitoring v1alpha1.Monitoring) string {
	scheme := monitoring.Scheme
	if len(scheme) == 0 {
		scheme = v1alpha1.MonitoringDefaultScheme
	}
	return scheme
}
