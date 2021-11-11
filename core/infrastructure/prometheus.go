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

package infrastructure

import (
	api "github.com/kiegroup/kogito-operator/apis"
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/operator"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
)

const (
	prometheusServerGroup = "monitoring.coreos.com"
	// MonitoringPathAnnotation ...
	MonitoringPathAnnotation = "kogito.kie.org/app.monitoring.path"
	// MonitoringSchemeAnnotation ...
	MonitoringSchemeAnnotation = "kogito.kie.org/app.monitoring.scheme"
)

// PrometheusManager ...
type PrometheusManager interface {
	ConfigurePrometheus(deployment *v1.Deployment) error
}

type prometheusManager struct {
	operator.Context
}

// NewPrometheusManager ...
func NewPrometheusManager(context operator.Context) PrometheusManager {
	context.Log = context.Log.WithValues("monitoring", "prometheus")
	return &prometheusManager{
		Context: context,
	}
}

func (m *prometheusManager) ConfigurePrometheus(deployment *v1.Deployment) error {
	m.Log.Debug("Going to configuring prometheus")
	prometheusAvailable := m.isPrometheusAvailable()
	if !prometheusAvailable {
		m.Log.Debug("prometheus operator not available in namespace")
		return nil
	}

	deploymentHandler := NewDeploymentHandler(m.Context)
	deploymentAvailable, err := deploymentHandler.IsDeploymentAvailable(types.NamespacedName{Name: deployment.GetName(), Namespace: deployment.GetNamespace()})
	if err != nil {
		return err
	}
	if !deploymentAvailable {
		m.Log.Debug("Deployment is currently not available, will try in next reconciliation loop")
		return framework.ErrorForDeploymentNotReachable(deployment.GetName())
	}

	prometheusAddOnAvailable := m.isPrometheusAddOnAvailable(deployment)

	if prometheusAddOnAvailable {
		if err := m.createPrometheusServiceMonitorIfNotExists(deployment); err != nil {
			return err
		}
	}
	return nil
}

// isPrometheusAvailable checks if Prometheus CRD is available in the cluster
func (m *prometheusManager) isPrometheusAvailable() bool {
	return m.Client.HasServerGroup(prometheusServerGroup)
}

func (m *prometheusManager) isPrometheusAddOnAvailable(deployment *v1.Deployment) bool {
	kogitoServiceHandler := framework.NewKogitoServiceHandler(m.Context)
	url := kogitoServiceHandler.GetKogitoServiceEndpoint(types.NamespacedName{Name: deployment.GetName(), Namespace: deployment.GetNamespace()})
	url = url + getMonitoringPath(deployment)
	if resp, err := http.Head(url); err != nil {
		m.Recorder.Eventf(deployment, "Normal", "Configuring Prometheus", "Error occurs while checking Prometheus URL. Error : %s", err.Error())
		return false
	} else if resp.StatusCode == http.StatusOK {
		return true
	}
	m.Log.Debug("Prometheus addon not available")
	return false
}

func (m *prometheusManager) createPrometheusServiceMonitorIfNotExists(deployment *v1.Deployment) error {
	serviceMonitor, err := m.loadDeployedServiceMonitor(deployment)
	if err != nil {
		return err
	}
	if serviceMonitor == nil {
		_, err := m.createServiceMonitor(deployment)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *prometheusManager) loadDeployedServiceMonitor(deployment *v1.Deployment) (*monv1.ServiceMonitor, error) {
	m.Log.Debug("fetching deployed Service monitor instance", "instanceName", deployment.GetName(), "namespace", deployment.GetNamespace())
	serviceMonitor := &monv1.ServiceMonitor{}
	if exits, err := kubernetes.ResourceC(m.Client).FetchWithKey(types.NamespacedName{Name: deployment.GetName(), Namespace: deployment.GetNamespace()}, serviceMonitor); err != nil {
		m.Log.Error(err, "Error occurs while fetching Service monitor instance")
		return nil, err
	} else if !exits {
		m.Log.Debug("Service monitor instance is not exists")
		return nil, nil
	} else {
		m.Log.Debug("Service monitor instance found")
		return serviceMonitor, nil
	}
}

// createServiceMonitor create ServiceMonitor used for scraping by prometheus for kogito service
func (m *prometheusManager) createServiceMonitor(deployment *v1.Deployment) (*monv1.ServiceMonitor, error) {
	endPoint := monv1.Endpoint{}
	endPoint.Path = getMonitoringPath(deployment)
	endPoint.Scheme = getMonitoringScheme(deployment)

	serviceSelectorLabels := make(map[string]string)
	serviceSelectorLabels[framework.LabelAppKey] = deployment.GetName()

	serviceMonitorLabels := make(map[string]string)
	serviceMonitorLabels["name"] = operator.Name
	serviceMonitorLabels[framework.LabelAppKey] = deployment.GetName()

	sm := &monv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployment.GetName(),
			Namespace: deployment.GetNamespace(),
			Labels:    serviceMonitorLabels,
		},
		Spec: monv1.ServiceMonitorSpec{
			NamespaceSelector: monv1.NamespaceSelector{
				MatchNames: []string{
					deployment.GetNamespace(),
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

	if err := framework.SetOwner(deployment, m.Scheme, sm); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(m.Client).Create(sm); err != nil {
		m.Log.Error(err, "Error occurs while creating Service Monitor instance")
		return nil, err
	}
	return sm, nil
}

func getMonitoringPath(deployment *v1.Deployment) string {
	path := deployment.GetAnnotations()[MonitoringPathAnnotation]
	if len(path) == 0 {
		path = api.MonitoringDefaultPath
	}
	return path
}

func getMonitoringScheme(deployment *v1.Deployment) string {
	scheme := deployment.GetAnnotations()[MonitoringSchemeAnnotation]
	if len(scheme) == 0 {
		scheme = api.MonitoringDefaultScheme
	}
	return scheme
}
