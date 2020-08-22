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

package prometheus

import (
	"context"
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	monclientv1 "github.com/coreos/prometheus-operator/pkg/client/versioned/typed/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceMonitorInterface has functions that interact with Service Monitor instances in the Kubernetes cluster
type ServiceMonitorInterface interface {
	Get(ctx context.Context, key clientv1.ObjectKey) (*monv1.ServiceMonitor, error)
	List(namespace string) (*monv1.ServiceMonitorList, error)
	Create(ctx context.Context, serviceMonitor *monv1.ServiceMonitor, opts ...clientv1.CreateOption) error
	Update(ctx context.Context, serviceMonitor *monv1.ServiceMonitor, opts ...clientv1.UpdateOption) error
	Delete(ctx context.Context, serviceMonitor *monv1.ServiceMonitor, opts ...clientv1.DeleteOption) error
	DeleteAllOf(ctx context.Context, serviceMonitor *monv1.ServiceMonitor, opts ...clientv1.DeleteAllOfOption) error
	Patch(ctx context.Context, serviceMonitor *monv1.ServiceMonitor, patch clientv1.Patch, opts ...clientv1.PatchOption) error
}
type serviceMonitor struct {
	prometheusCli monclientv1.MonitoringV1Interface
}

func newServiceMonitor(c monclientv1.MonitoringV1Interface) ServiceMonitorInterface {
	return &serviceMonitor{
		prometheusCli: c,
	}
}

func (s *serviceMonitor) Get(ctx context.Context, key clientv1.ObjectKey) (*monv1.ServiceMonitor, error) {
	log.Debugf("Get service monitor instance : %s", key.Name)
	return s.prometheusCli.ServiceMonitors(key.Namespace).Get(ctx, key.Name, metav1.GetOptions{})
}

func (s *serviceMonitor) List(namespace string) (*monv1.ServiceMonitorList, error) {
	log.Debugf("List service monitor instances from namespace %s", namespace)
	return s.prometheusCli.ServiceMonitors(namespace).List(context.TODO(), metav1.ListOptions{})
}

// Create implements client.Client
func (s *serviceMonitor) Create(ctx context.Context, serviceMonitor *monv1.ServiceMonitor, opts ...clientv1.CreateOption) error {
	log.Debugf("Create service monitor instance : %s", serviceMonitor)
	createOpts := &clientv1.CreateOptions{}
	createOpts.ApplyOptions(opts)
	_, err := s.prometheusCli.ServiceMonitors(serviceMonitor.Namespace).Create(ctx, serviceMonitor, *createOpts.AsCreateOptions())
	return err
}

func (s *serviceMonitor) Update(ctx context.Context, serviceMonitor *monv1.ServiceMonitor, opts ...clientv1.UpdateOption) error {
	log.Debugf("Update service monitor instance : %s", serviceMonitor)
	updateOpts := &clientv1.UpdateOptions{}
	updateOpts.ApplyOptions(opts)
	_, err := s.prometheusCli.ServiceMonitors(serviceMonitor.Namespace).Update(ctx, serviceMonitor, *updateOpts.AsUpdateOptions())
	return err
}

func (s *serviceMonitor) Delete(ctx context.Context, serviceMonitor *monv1.ServiceMonitor, opts ...clientv1.DeleteOption) error {
	log.Debugf("Delete service monitor instance : %s", serviceMonitor)
	DeleteOpts := &clientv1.DeleteOptions{}
	DeleteOpts.ApplyOptions(opts)
	err := s.prometheusCli.ServiceMonitors(serviceMonitor.Namespace).Delete(ctx, serviceMonitor.Name, *DeleteOpts.AsDeleteOptions())
	return err
}

func (s *serviceMonitor) DeleteAllOf(ctx context.Context, serviceMonitor *monv1.ServiceMonitor, opts ...clientv1.DeleteAllOfOption) error {
	DeleteAllOpts := &clientv1.DeleteAllOfOptions{}
	DeleteAllOpts.ApplyOptions(opts)
	err := s.prometheusCli.ServiceMonitors(serviceMonitor.Namespace).DeleteCollection(ctx, *DeleteAllOpts.AsDeleteOptions(), *DeleteAllOpts.AsListOptions())
	return err
}

func (s *serviceMonitor) Patch(ctx context.Context, serviceMonitor *monv1.ServiceMonitor, patch clientv1.Patch, opts ...clientv1.PatchOption) error {
	log.Debugf("Patch service monitor instance : %s", serviceMonitor)
	PatchOpts := &clientv1.PatchOptions{}
	PatchOpts.ApplyOptions(opts)
	data, err := patch.Data(serviceMonitor)
	if err != nil {
		return err
	}
	_, err = s.prometheusCli.ServiceMonitors(serviceMonitor.Namespace).Patch(ctx, serviceMonitor.Name, patch.Type(), data, *PatchOpts.AsPatchOptions(), "")
	return err
}
