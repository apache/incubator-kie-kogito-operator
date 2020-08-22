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
	"fmt"
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/prometheus"
	"k8s.io/apimachinery/pkg/runtime"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
)

// CustomReader provides capability to identify client at runtime
type CustomReader struct {
	Client *Client
}

// CustomReaderC provide reader reference for given client
func CustomReaderC(cli *Client) clientv1.Reader {
	return &CustomReader{
		Client: cli,
	}
}

// List retrieves list of objects for a given namespace and list options.
func (r *CustomReader) List(ctx context.Context, list runtime.Object, opts ...clientv1.ListOption) error {
	switch l := list.(type) {
	case *monv1.ServiceMonitorList:
		for _, opt := range opts {
			if namespace, ok := opt.(clientv1.InNamespace); ok {
				sList, err := prometheus.ServiceMonitorC(r.Client.PrometheusCli).List(string(namespace))
				if err != nil {
					return err
				}
				serviceMonitorList := list.(*monv1.ServiceMonitorList)
				*serviceMonitorList = *sList
				return nil
			}
		}
		return fmt.Errorf("namespace is not specified, cannot list prometheuses")
	default:
		return r.Client.ControlCli.List(ctx, l, opts...)
	}
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
func (r *CustomReader) Get(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
	switch o := obj.(type) {
	case *monv1.ServiceMonitor:
		sObj, err := prometheus.ServiceMonitorC(r.Client.PrometheusCli).Get(ctx, key)
		if err != nil {
			return err
		}
		serviceMonitor := obj.(*monv1.ServiceMonitor)
		*serviceMonitor = *sObj
		return nil
	default:
		return r.Client.ControlCli.Get(ctx, key, o)
	}
}
