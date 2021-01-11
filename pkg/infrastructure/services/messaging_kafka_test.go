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
	"testing"

	kafkav1beta1 "github.com/kiegroup/kogito-cloud-operator/api/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Test_createKafkaTopics(t *testing.T) {

	appProps := map[string]string{}
	appProps[QuarkusKafkaBootstrapAppProp] = "kogito-kafka:9092"

	kogitoInfraInstance := &v1beta1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogito-kafka",
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoInfraSpec{
			Resource: v1beta1.Resource{
				APIVersion: infrastructure.KafkaAPIVersion,
				Kind:       infrastructure.KafkaKind,
			},
		},
		Status: v1beta1.KogitoInfraStatus{
			RuntimeProperties: map[v1beta1.RuntimeType]v1beta1.RuntimeProperties{
				v1beta1.QuarkusRuntimeType: {
					AppProps: appProps,
				},
			},
		},
	}
	service := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "data-index",
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: v1beta1.DataIndex,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Infra: []string{kogitoInfraInstance.Name},
			},
		},
	}

	client := test.NewFakeClientBuilder().AddK8sObjects(kogitoInfraInstance, service).Build()
	k := kafkaMessagingDeployer{
		messagingDeployer{
			scheme: meta.GetRegisteredSchema(),
			cli:    client,
			definition: ServiceDefinition{
				KafkaTopics: []string{
					"kogito-processinstances-events",
				},
			}}}
	err := k.createRequiredResources(service)
	assert.NoError(t, err)

	kafkaTopic := &kafkav1beta1.KafkaTopic{}
	exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Namespace: t.Name(), Name: "kogito-processinstances-events"}, kafkaTopic)
	assert.NoError(t, err)
	assert.True(t, exists)
}
