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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKogitoSupportingServiceDataIndex_Reconcile(t *testing.T) {
	ns := t.Name()
	kogitoKafka := test.CreateFakeKogitoKafka(t.Name())
	kogitoInfinispan := test.CreateFakeKogitoInfinispan(t.Name())
	dataIndex := test.CreateFakeDataIndex(ns)
	dataIndex.GetSpec().AddInfra(kogitoKafka.GetName())
	dataIndex.GetSpec().AddInfra(kogitoInfinispan.GetName())
	cli := test.NewFakeClientBuilder().AddK8sObjects(dataIndex, kogitoKafka, kogitoInfinispan).OnOpenShift().Build()
	r := &dataIndexSupportingServiceResource{
		targetContext: targetContext{
			instance:                 dataIndex,
			client:                   cli,
			log:                      logger.GetLogger("data index reconciler"),
			scheme:                   test.GetRegisteredSchema(),
			supportingServiceHandler: test.CreateFakeKogitoSupportingServiceHandler(cli),
			infraHandler:             test.CreateFakeKogitoInfraHandler(cli),
			runtimeHandler:           test.CreateFakeKogitoRuntimeHandler(cli),
		},
	}
	requeueAfter, err := r.Reconcile()
	assert.NoError(t, err)
	assert.True(t, requeueAfter == 0)
}
