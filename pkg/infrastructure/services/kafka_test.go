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

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoinfra/kafka"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Test_createKafkaTopics(t *testing.T) {

	appProps := map[string]string{}
	appProps[kafka.QuarkusKafkaBootstrapAppProp] = "kogito-kafka:9092"

	kogitoInfraInstance := &v1alpha1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kogito-kafka",
			Namespace: "mynamespace",
		},
		Spec: v1alpha1.KogitoInfraSpec{
			Resource: v1alpha1.Resource{
				APIVersion: infrastructure.KafkaAPIVersion,
				Kind:       infrastructure.KafkaKind,
			},
		},
	}

	client := test.CreateFakeClient(nil, nil, nil)

	serviceDeployer := serviceDeployer{
		client: client,
		definition: ServiceDefinition{
			KafkaTopics: []string{
				"kogito-processinstances-events",
			},
		},
	}

	err := serviceDeployer.createKafkaTopics(kogitoInfraInstance, "kogito-kafka:9092")
	assert.NoError(t, err)

	kafkaTopic := &v1beta1.KafkaTopic{}
	exists, err := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Namespace: "mynamespace", Name: "kogito-processinstances-events"}, kafkaTopic)
	assert.NoError(t, err)
	assert.True(t, exists)
}
