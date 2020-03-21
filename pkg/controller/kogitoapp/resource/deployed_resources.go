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
	"context"
	"fmt"
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/prometheus"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
)

// GetDeployedResources will get the deployed resources for KogitoApp
func GetDeployedResources(instance *v1alpha1.KogitoApp, client *client.Client) (
	map[reflect.Type][]resource.KubernetesResource, error) {

	reader := read.New(&reader{client}).WithNamespace(instance.Namespace).WithOwnerObject(instance)

	objectLists := []runtime.Object{
		&buildv1.BuildConfigList{},
		&imgv1.ImageStreamList{},
		&appsv1.DeploymentConfigList{},
		&corev1.ServiceList{},
		&routev1.RouteList{},
		&corev1.ConfigMapList{},
	}

	if isPrometheusOperatorReady(client) {
		objectLists = append(objectLists, &monv1.ServiceMonitorList{})
	}

	resourceMap, err := reader.ListAll(objectLists...)
	if err != nil {
		log.Warn("Failed to list deployed objects. ", err)
		return nil, err
	}

	services.ExcludeAppPropConfigMapFromResource(instance.GetName(), resourceMap)
	return resourceMap, nil
}

type reader struct {
	client *client.Client
}

func (r *reader) List(ctx context.Context, list runtime.Object, opts ...clientv1.ListOption) error {
	switch l := list.(type) {
	case *monv1.ServiceMonitorList:
		for _, opt := range opts {
			if namespace, ok := opt.(clientv1.InNamespace); ok {
				sList, err := prometheus.ServiceMonitorC(r.client).List(string(namespace))
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
		return r.client.ControlCli.List(ctx, l, opts...)
	}
}

func (r *reader) Get(ctx context.Context, key clientv1.ObjectKey, obj runtime.Object) error {
	return r.client.ControlCli.Get(ctx, key, obj)
}
