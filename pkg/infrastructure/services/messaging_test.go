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

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_fetchRequiredTopics(t *testing.T) {
	responseWithTopics := `[
   {
      "name":"kogito_incoming_stream",
      "type":"INCOMING",
      "eventsMeta":[
         {
            "type":"travellers",
            "source":"",
            "kind":"CONSUMED"
         }
      ]
   },
   {
      "name":"processedtravellers",
      "type":"OUTGOING",
      "eventsMeta":[
         {
            "type":"process.travelers.processedtravellers",
            "source":"/process/travelers",
            "kind":"PRODUCED"
         }
      ]
   }
]`
	instance := createServiceInstance(t)

	server := mockKogitoSvcReplies(t, serverHandler{Path: topicInfoPath, JSONResponse: responseWithTopics})
	defer server.Close()

	m := messagingDeployer{cli: test.NewFakeClientBuilder().AddK8sObjects(createAvailableDeployment(instance)).Build()}
	topics, err := m.fetchRequiredTopicsForURL(instance, server.URL)
	assert.NoError(t, err)
	assert.NotEmpty(t, topics)
	m.setCloudEventsStatus(instance, topics)
	assert.NotNil(t, instance.GetStatus().GetCloudEvents())
}

func Test_fetchRequiredTopicsWithEmptyReply(t *testing.T) {
	emptyResponse := "[]"
	instance := createServiceInstance(t)

	server := mockKogitoSvcReplies(t, serverHandler{Path: topicInfoPath, JSONResponse: emptyResponse})
	defer server.Close()

	m := messagingDeployer{cli: test.NewFakeClientBuilder().AddK8sObjects(createAvailableDeployment(instance)).Build()}
	topics, err := m.fetchRequiredTopicsForURL(instance, server.URL)
	assert.NoError(t, err)
	assert.Empty(t, topics)
}

func createAvailableDeployment(instance v1beta1.KogitoService) *v1.Deployment {
	return &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: instance.GetName(), Namespace: instance.GetNamespace()},
		Status: v1.DeploymentStatus{
			AvailableReplicas: 1,
		},
	}

}

func createServiceInstance(t *testing.T) v1beta1.KogitoService {
	return &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{Name: "quarkus-test", Namespace: t.Name()},
		Status: v1beta1.KogitoRuntimeStatus{
			KogitoServiceStatus: v1beta1.KogitoServiceStatus{
				DeploymentConditions: []v1.DeploymentCondition{
					{
						Type:           v1.DeploymentAvailable,
						Status:         corev1.ConditionTrue,
						LastUpdateTime: metav1.Now(),
					},
				},
			},
		},
	}
}
