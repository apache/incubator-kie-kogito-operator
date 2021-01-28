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

package kogitosupportingservice

import (
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	test2 "github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReconcileKogitoSupportingTrusty_Reconcile(t *testing.T) {
	ns := t.Name()
	kogitoKafka := test2.CreateFakeKogitoKafka(ns)
	instance := test2.CreateFakeTrustyAIService(ns)
	instance.GetSpec().AddInfra(kogitoKafka.GetName())
	cli := test2.NewFakeClientBuilder().AddK8sObjects(instance, kogitoKafka).OnOpenShift().Build()

	r := &trustyAISupportingServiceResource{
		targetContext: targetContext{
			instance:                 instance,
			client:                   cli,
			log:                      logger.GetLogger("trusty ai reconciler"),
			scheme:                   meta.GetRegisteredSchema(),
			infraHandler:             test2.CreateFakeKogitoInfraHandler(cli),
			supportingServiceHandler: test2.CreateFakeKogitoSupportingServiceHandler(cli),
			runtimeHandler:           test2.CreateFakeKogitoRuntimeHandler(cli),
		},
	}
	// basic checks
	requeueAfter, err := r.Reconcile()
	assert.NoError(t, err)
	assert.True(t, requeueAfter == 0)
}
