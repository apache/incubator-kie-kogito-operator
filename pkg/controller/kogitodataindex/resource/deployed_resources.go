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
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
)

// GetDeployedResources will get the deployed resources for Kogito Data Index
func GetDeployedResources(instance *v1alpha1.KogitoDataIndex, client *client.Client) (
	map[reflect.Type][]resource.KubernetesResource, error) {

	reader := read.New(client.ControlCli).WithNamespace(instance.Namespace).WithOwnerObject(instance)
	var objectTypes []runtime.Object
	if client.IsOpenshift() {
		objectTypes = []runtime.Object{
			&appsv1.DeploymentList{},
			&corev1.ServiceList{},
			&routev1.RouteList{},
			&imgv1.ImageStreamList{},
		}
	} else {
		objectTypes = []runtime.Object{
			&appsv1.DeploymentList{},
			&corev1.ServiceList{},
		}
	}

	if infrastructure.IsStrimziAvailable(client) {
		objectTypes = append(objectTypes, &kafkabetav1.KafkaTopicList{})
	}

	resourceMap, err := reader.ListAll(objectTypes...)
	if err != nil {
		log.Warn("Failed to list deployed objects. ", err)
		return nil, err
	}
	return resourceMap, nil
}
