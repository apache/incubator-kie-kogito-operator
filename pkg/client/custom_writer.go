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

package client

import (
	"context"
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
)

// CustomWriter provides capability to identify client at runtime
type CustomWriter struct {
	Client *Client
}

// Create saves the object obj in the Kubernetes cluster.
func (w *CustomWriter) Create(ctx context.Context, obj runtime.Object, opts ...clientv1.CreateOption) error {
	switch o := obj.(type) {
	case *monv1.ServiceMonitor:
		serviceMonitor := obj.(*monv1.ServiceMonitor)
		return prometheus.ServiceMonitorC(w.Client.PrometheusCli).Create(ctx, serviceMonitor, opts...)
	default:
		return w.Client.ControlCli.Create(ctx, o, opts...)
	}
}

// Delete deletes the given obj from Kubernetes cluster.
func (w *CustomWriter) Delete(ctx context.Context, obj runtime.Object, opts ...clientv1.DeleteOption) error {
	switch o := obj.(type) {
	case *monv1.ServiceMonitor:
		serviceMonitor := obj.(*monv1.ServiceMonitor)
		return prometheus.ServiceMonitorC(w.Client.PrometheusCli).Delete(ctx, serviceMonitor, opts...)
	default:
		return w.Client.ControlCli.Delete(ctx, o, opts...)
	}
}

// Update updates the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (w *CustomWriter) Update(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
	switch o := obj.(type) {
	case *monv1.ServiceMonitor:
		serviceMonitor := obj.(*monv1.ServiceMonitor)
		return prometheus.ServiceMonitorC(w.Client.PrometheusCli).Update(ctx, serviceMonitor, opts...)
	default:
		return w.Client.ControlCli.Update(ctx, o, opts...)
	}
}

// Patch patches the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (w *CustomWriter) Patch(ctx context.Context, obj runtime.Object, patch clientv1.Patch, opts ...clientv1.PatchOption) error {
	switch o := obj.(type) {
	case *monv1.ServiceMonitor:
		serviceMonitor := obj.(*monv1.ServiceMonitor)
		return prometheus.ServiceMonitorC(w.Client.PrometheusCli).Patch(ctx, serviceMonitor, patch, opts...)
	default:
		return w.Client.ControlCli.Patch(ctx, o, patch, opts...)
	}
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (w *CustomWriter) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...clientv1.DeleteAllOfOption) error {
	switch o := obj.(type) {
	case *monv1.ServiceMonitor:
		serviceMonitor := obj.(*monv1.ServiceMonitor)
		return prometheus.ServiceMonitorC(w.Client.PrometheusCli).DeleteAllOf(ctx, serviceMonitor, opts...)
	default:
		return w.Client.ControlCli.DeleteAllOf(ctx, o, opts...)
	}
}
