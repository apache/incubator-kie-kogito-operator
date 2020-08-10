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
	"context"
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
)

type writer struct {
	client *client.Client
}

// Create saves the object obj in the Kubernetes cluster.
func (w *writer) Create(ctx context.Context, obj runtime.Object, opts ...clientv1.CreateOption) error {
	switch obj.(type) {
	case *monv1.ServiceMonitor:
		serviceMonitor := obj.(*monv1.ServiceMonitor)
		return prometheus.ServiceMonitorC(w.client).Create(ctx, serviceMonitor, opts...)
	default:
		return w.client.ControlCli.Create(ctx, obj, opts...)
	}
}

// Delete deletes the given obj from Kubernetes cluster.
func (w *writer) Delete(ctx context.Context, obj runtime.Object, opts ...clientv1.DeleteOption) error {
	switch obj.(type) {
	case *monv1.ServiceMonitor:
		serviceMonitor := obj.(*monv1.ServiceMonitor)
		return prometheus.ServiceMonitorC(w.client).Delete(ctx, serviceMonitor, opts...)
	default:
		return w.client.ControlCli.Delete(ctx, obj, opts...)
	}
}

// Update updates the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (w *writer) Update(ctx context.Context, obj runtime.Object, opts ...clientv1.UpdateOption) error {
	switch obj.(type) {
	case *monv1.ServiceMonitor:
		serviceMonitor := obj.(*monv1.ServiceMonitor)
		return prometheus.ServiceMonitorC(w.client).Update(ctx, serviceMonitor, opts...)
	default:
		return w.client.ControlCli.Update(ctx, obj, opts...)
	}
}

// Patch patches the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (w *writer) Patch(ctx context.Context, obj runtime.Object, patch clientv1.Patch, opts ...clientv1.PatchOption) error {
	switch obj.(type) {
	case *monv1.ServiceMonitor:
		serviceMonitor := obj.(*monv1.ServiceMonitor)
		return prometheus.ServiceMonitorC(w.client).Patch(ctx, serviceMonitor, patch, opts...)
	default:
		return w.client.ControlCli.Patch(ctx, obj, patch, opts...)
	}
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (w *writer) DeleteAllOf(ctx context.Context, obj runtime.Object, opts ...clientv1.DeleteAllOfOption) error {
	switch obj.(type) {
	case *monv1.ServiceMonitor:
		serviceMonitor := obj.(*monv1.ServiceMonitor)
		return prometheus.ServiceMonitorC(w.client).DeleteAllOf(ctx, serviceMonitor, opts...)
	default:
		return w.client.ControlCli.DeleteAllOf(ctx, obj, opts...)
	}
}
