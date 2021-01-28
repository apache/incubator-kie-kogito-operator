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
	"github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReconcileKogitoSupportingServiceExplainability_Reconcile(t *testing.T) {
	ns := t.Name()
	kogitoKafka := test.CreateFakeKogitoKafka(ns)
	explainabilityService := test.CreateFakeExplainabilityService(ns)
	explainabilityService.GetSpec().AddInfra(kogitoKafka.GetName())

	cli := test.NewFakeClientBuilder().AddK8sObjects(explainabilityService, kogitoKafka).OnOpenShift().Build()
	r := &explainabilitySupportingServiceResource{
		targetContext: targetContext{
			instance:                 explainabilityService,
			client:                   cli,
			log:                      logger.GetLogger("explainability reconciler"),
			scheme:                   meta.GetRegisteredSchema(),
			infraHandler:             test.CreateFakeKogitoInfraHandler(cli),
			supportingServiceHandler: test.CreateFakeKogitoSupportingServiceHandler(cli),
			runtimeHandler:           test.CreateFakeKogitoRuntimeHandler(cli),
		},
	}

	// basic checks
	requeueAfter, err := r.Reconcile()
	assert.NoError(t, err)
	assert.True(t, requeueAfter == 0)
}
