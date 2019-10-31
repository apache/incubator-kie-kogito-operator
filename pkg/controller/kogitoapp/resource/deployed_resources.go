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
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	corev1 "k8s.io/api/core/v1"
	"reflect"
	clientv1 "sigs.k8s.io/controller-runtime/pkg/client"
)

// GetDeployedResources will get the deployed resources for KogitoApp
func GetDeployedResources(instance *v1alpha1.KogitoApp, client clientv1.Reader) (
	map[reflect.Type][]resource.KubernetesResource, error) {

	reader := read.New(client).WithNamespace(instance.Namespace).WithOwnerObject(instance)
	resourceMap, err := reader.ListAll(
		&buildv1.BuildConfigList{},
		&imgv1.ImageStreamList{},
		&appsv1.DeploymentConfigList{},
		&corev1.ServiceList{},
		&routev1.RouteList{},
	)
	if err != nil {
		log.Warn("Failed to list deployed objects. ", err)
		return nil, err
	}
	return resourceMap, nil
}
