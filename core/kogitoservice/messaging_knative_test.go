// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package kogitoservice

import (
	"github.com/apache/incubator-kie-kogito-operator/core/client/kubernetes"
	"github.com/apache/incubator-kie-kogito-operator/core/framework"
	"github.com/apache/incubator-kie-kogito-operator/core/operator"
	"github.com/apache/incubator-kie-kogito-operator/core/test"
	"github.com/apache/incubator-kie-kogito-operator/internal/app"
	"github.com/apache/incubator-kie-kogito-operator/meta"
	"github.com/stretchr/testify/assert"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	"testing"
)

func Test_knativeMessagingDeployer_CreateRequiredResources(t *testing.T) {
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
	server := test.MockKogitoSvcReplies(t, test.ServerHandler{Path: topicInfoPath, JSONResponse: responseWithTopics})
	defer server.Close()
	deferFn := test.SetSharedEnv(EnvVarKogitoServiceURL, server.URL)
	defer deferFn()

	kogitoSvc := createServiceInstance(t)
	request := newReconcileRequest(kogitoSvc.GetNamespace())
	request.Name = kogitoSvc.GetName()
	knativeInfra := test.CreateFakeKogitoKnative(t.Name())
	kogitoSvc.GetSpec().AddInfra(knativeInfra.GetName())

	client := test.NewFakeClientBuilder().AddK8sObjects(kogitoSvc, knativeInfra, createAvailableDeployment(kogitoSvc)).Build()
	context := operator.Context{
		Client: client,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	infraHandler := app.NewKogitoInfraHandler(context)
	knativeDeployer := NewKnativeMessagingDeployer(context, ServiceDefinition{Request: request}, infraHandler)

	err := knativeDeployer.CreateRequiredResources(kogitoSvc)
	assert.NoError(t, err)
	// check trigger
	triggers := &eventingv1.TriggerList{}
	labels := map[string]string{
		framework.LabelAppKey: kogitoSvc.GetName(),
		topicIdentifier:       "travellers",
	}
	err = kubernetes.ResourceC(client).ListWithNamespaceAndLabel(kogitoSvc.GetNamespace(), triggers, labels)
	assert.NoError(t, err)
	assert.Len(t, triggers.Items, 1)
	assert.Equal(t, "travellers", triggers.Items[0].Spec.Filter.Attributes["type"])
}
