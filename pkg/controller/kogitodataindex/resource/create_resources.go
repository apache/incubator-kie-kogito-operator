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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	routev1 "github.com/openshift/api/route/v1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

var log = logger.GetLogger("resources_kogitodataindex")

// KogitoDataIndexResources is the data structure for all resources managed by KogitoDataIndex CR
type KogitoDataIndexResources struct {
	// StatefulSet is the resource responsible for the Data Index Service image deployment in the cluster
	StatefulSet       *appsv1.StatefulSet
	StatefulSetStatus KogitoDataIndexResourcesStatus
	// Service to expose the data index service internally
	Service       *corev1.Service
	ServiceStatus KogitoDataIndexResourcesStatus
	// Route to expose the service in the Ingress. Supported only on OpenShift for now
	Route       *routev1.Route
	RouteStatus KogitoDataIndexResourcesStatus
	// KafkaTopics are the Kafka Topics required by the Data Index Service
	KafkaTopics      []kafkabetav1.KafkaTopic
	KafkaTopicStatus KogitoDataIndexResourcesStatus
}

// KogitoDataIndexResourcesStatus identifies the status of the resource
type KogitoDataIndexResourcesStatus struct {
	New bool
}

type kogitoDataIndexResourcesFactory struct {
	framework.Factory
	Resources       *KogitoDataIndexResources
	KogitoDataIndex *v1alpha1.KogitoDataIndex
}

// Build will call a builder function if no errors were found
func (f *kogitoDataIndexResourcesFactory) build(fn func(*kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	if f.Error == nil {
		return fn(f)
	}
	// break the chain
	return f
}

func (f *kogitoDataIndexResourcesFactory) buildOnOpenshift(fn func(*kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	if f.Error == nil && f.Context.Client.IsOpenshift() {
		return fn(f)
	}
	// break the chain
	return f
}

// CreateOrFetchResources will create the needed resources in the cluster if they not exists, fetch otherwise
func CreateOrFetchResources(instance *v1alpha1.KogitoDataIndex, context framework.FactoryContext) (KogitoDataIndexResources, error) {
	factory := kogitoDataIndexResourcesFactory{
		Factory:         framework.Factory{Context: &context},
		Resources:       &KogitoDataIndexResources{},
		KogitoDataIndex: instance,
	}

	factory.
		build(createStatefulSet).
		build(createService).
		buildOnOpenshift(createRoute).
		build(createKafkaTopic)

	return *factory.Resources, factory.Error
}

func createStatefulSet(f *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	secret, err := infrastructure.FetchInfinispanCredentials(&f.KogitoDataIndex.Spec, f.KogitoDataIndex.Namespace, f.Context.Client)
	if err != nil {
		f.Error = err
		return f
	}
	externalURI, err := getKafkaServerURI(f.KogitoDataIndex.Spec.Kafka, f.KogitoDataIndex.Namespace, f.Context.Client)
	if err != nil {
		f.Error = err
		return f
	}
	statefulset, err := newStatefulset(f.KogitoDataIndex, secret, externalURI, f.Context.Client)
	if err != nil {
		f.Error = err
		return f
	}
	if err := f.CallPreCreate(statefulset); err != nil {
		f.Error = err
		return f
	}

	if f.Resources.StatefulSetStatus.New, f.Error =
		kubernetes.ResourceC(f.Context.Client).CreateIfNotExists(statefulset); f.Error != nil {
		return f
	}

	if f.CallPostCreate(f.Resources.StatefulSetStatus.New, statefulset); f.Error != nil {
		return f
	}

	f.Resources.StatefulSet = statefulset

	return f
}

func createService(f *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	svc := newService(f.KogitoDataIndex, f.Resources.StatefulSet)
	if f.Error = f.CallPreCreate(svc); f.Error != nil {
		return f
	}

	if f.Resources.ServiceStatus.New, f.Error =
		kubernetes.ResourceC(f.Context.Client).CreateIfNotExists(svc); f.Error != nil {
		return f
	}

	if f.CallPostCreate(f.Resources.ServiceStatus.New, svc); f.Error != nil {
		return f
	}

	f.Resources.Service = svc

	return f
}

func createRoute(f *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	route, err := newRoute(f.KogitoDataIndex, f.Resources.Service)
	if err != nil {
		f.Error = err
		return f
	}

	if f.Error = f.CallPreCreate(route); f.Error != nil {
		return f
	}

	if f.Resources.RouteStatus.New, f.Error = kubernetes.ResourceC(f.Context.Client).CreateIfNotExists(route); f.Error != nil {
		return f
	}

	if f.CallPostCreate(f.Resources.RouteStatus.New, route); f.Error != nil {
		return f
	}

	f.Resources.Route = route

	return f
}

func createKafkaTopic(f *kogitoDataIndexResourcesFactory) *kogitoDataIndexResourcesFactory {
	kafkaName, kafkaReplicas, err := getKafkaServerReplicas(f.KogitoDataIndex.Spec.Kafka, f.KogitoDataIndex.Namespace, f.Context.Client)
	if err != nil {
		f.Error = err
		return f
	} else if len(kafkaName) <= 0 || kafkaReplicas <= 0 {
		return f
	}

	for _, kafkaTopicName := range kafkaTopicNames {
		kafkaTopic := newKafkaTopic(kafkaTopicName, kafkaName, kafkaReplicas, f.KogitoDataIndex.Namespace)

		if f.Error = f.CallPreCreate(kafkaTopic); f.Error != nil {
			return f
		}

		newTopic, err := kubernetes.ResourceC(f.Context.Client).CreateIfNotExists(kafkaTopic)
		if err != nil {
			f.Error = err
			return f
		}

		if f.CallPostCreate(newTopic, kafkaTopic); f.Error != nil {
			return f
		}

		if newTopic {
			f.Resources.KafkaTopicStatus.New = newTopic
		}

		f.Resources.KafkaTopics = append(f.Resources.KafkaTopics, *kafkaTopic)
	}

	return f
}
