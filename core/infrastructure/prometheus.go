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
	"github.com/kiegroup/kogito-operator/core/framework"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/core/operator"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
	ConfigurePrometheus() error
}

type prometheusManager struct {
	operator.Context
	instance client.Object
}

// NewPrometheusManager ...
func NewPrometheusManager(context operator.Context, instance client.Object) PrometheusManager {
	context.Log = context.Log.WithValues("monitoring", "prometheus")
	return &prometheusManager{
		Context:  context,
		instance: instance,
	}
}

func (m *prometheusManager) ConfigurePrometheus() error {
	m.Log.Debug("Going to configuring prometheus")
	prometheusAvailable := m.isPrometheusAvailable()
	if !prometheusAvailable {
		m.Log.Debug("prometheus operator not available in namespace")
		return nil
	}

	deploymentHandler := NewDeploymentHandler(m.Context)
	deploymentAvailable, err := deploymentHandler.IsDeploymentAvailable(types.NamespacedName{Name: m.instance.GetName(), Namespace: m.instance.GetNamespace()})
	if err != nil {
		return err
	}
	if !deploymentAvailable {
		m.Log.Debug("Deployment is currently not available, will try in next reconciliation loop")
		return framework.ErrorForDeploymentNotReachable(m.instance.GetName())
	}

	prometheusAddOnAvailable := m.isPrometheusAddOnAvailable()

	if prometheusAddOnAvailable {
		if err := m.createPrometheusServiceMonitorIfNotExists(); err != nil {
			return err
		}
	}
	return nil
}

// isPrometheusAvailable checks if Prometheus CRD is available in the cluster
func (m *prometheusManager) isPrometheusAvailable() bool {
	return m.Client.HasServerGroup(prometheusServerGroup)
}

func (m *prometheusManager) isPrometheusAddOnAvailable() bool {
	kogitoServiceHandler := framework.NewKogitoServiceHandler(m.Context)
	url := kogitoServiceHandler.GetKogitoServiceEndpoint(types.NamespacedName{Name: m.instance.GetName(), Namespace: m.instance.GetNamespace()})
	url = url + getMonitoringPath(m.instance)
	if resp, err := http.Head(url); err != nil {
		m.Recorder.Eventf(m.instance, "Normal", "Configuring Prometheus", "Error occurs while checking Prometheus URL. Error : %s", err.Error())
		return false
	} else if resp.StatusCode == http.StatusOK {
		return true
	}
	m.Log.Debug("Prometheus addon not available")
	return false
}

func (m *prometheusManager) createPrometheusServiceMonitorIfNotExists() error {
	serviceMonitor, err := m.loadDeployedServiceMonitor()
	if err != nil {
		return err
	}
	if serviceMonitor == nil {
		_, err := m.createServiceMonitor()
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *prometheusManager) loadDeployedServiceMonitor() (*monv1.ServiceMonitor, error) {
	m.Log.Debug("fetching deployed Service monitor instance", "instanceName", m.instance.GetName(), "namespace", m.instance.GetNamespace())
	serviceMonitor := &monv1.ServiceMonitor{}
	if exits, err := kubernetes.ResourceC(m.Client).FetchWithKey(types.NamespacedName{Name: m.instance.GetName(), Namespace: m.instance.GetNamespace()}, serviceMonitor); err != nil {
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
func (m *prometheusManager) createServiceMonitor() (*monv1.ServiceMonitor, error) {
	endPoint := monv1.Endpoint{}
	endPoint.Path = getMonitoringPath(m.instance)
	endPoint.Scheme = getMonitoringScheme(m.instance)

	serviceSelectorLabels := make(map[string]string)
	serviceSelectorLabels[framework.LabelAppKey] = m.instance.GetName()

	serviceMonitorLabels := make(map[string]string)
	serviceMonitorLabels["name"] = operator.Name
	serviceMonitorLabels[framework.LabelAppKey] = m.instance.GetName()

	sm := &monv1.ServiceMonitor{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.instance.GetName(),
			Namespace: m.instance.GetNamespace(),
			Labels:    serviceMonitorLabels,
		},
		Spec: monv1.ServiceMonitorSpec{
			NamespaceSelector: monv1.NamespaceSelector{
				MatchNames: []string{
					m.instance.GetNamespace(),
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

	if err := framework.SetOwner(m.instance, m.Scheme, sm); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(m.Client).Create(sm); err != nil {
		m.Log.Error(err, "Error occurs while creating Service Monitor instance")
		return nil, err
	}
	return sm, nil
}

func getMonitoringPath(instance client.Object) string {
	path := instance.GetAnnotations()[MonitoringPathAnnotation]
	if len(path) == 0 {
		path = api.MonitoringDefaultPath
	}
	return path
}

func getMonitoringScheme(instance client.Object) string {
	scheme := instance.GetAnnotations()[MonitoringSchemeAnnotation]
	if len(scheme) == 0 {
		scheme = api.MonitoringDefaultScheme
	}
	return scheme
}
