// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package manager

import (
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/framework"
	test2 "github.com/kiegroup/kogito-cloud-operator/core/test"
	api2 "github.com/kiegroup/kogito-cloud-operator/core/test/api"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestMustFetchKogitoInfraInstance_InstanceNotFound(t *testing.T) {
	ns := t.Name()
	name := "InfinispanInfra"
	cli := test2.NewFakeClientBuilder().Build()
	infraHandler := test2.CreateFakeKogitoInfraHandler(cli)
	infraManager := NewKogitoInfraManager(cli, test2.TestLogger, test2.GetRegisteredSchema(), infraHandler)
	_, err := infraManager.MustFetchKogitoInfraInstance(types.NamespacedName{Name: name, Namespace: ns})
	assert.Error(t, err)
}

func TestRemoveKogitoInfraOwnership(t *testing.T) {
	ns := t.Name()
	scheme := meta.GetRegisteredSchema()
	travels := &api2.KogitoRuntimeTest{
		ObjectMeta: v1.ObjectMeta{
			Name:      "travels",
			Namespace: ns,
			UID:       test2.GenerateUID(),
		},
		Spec: api2.KogitoRuntimeSpecTest{
			KogitoServiceSpec: api.KogitoServiceSpec{
				Infra: []string{
					"infinispan_infra",
					"kafka_infra",
				},
			},
		},
	}

	kafkaInfra := &api2.KogitoInfraTest{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kafka_infra",
			Namespace: ns,
		},
	}
	err := framework.AddOwnerReference(travels, scheme, kafkaInfra)
	assert.NoError(t, err)

	cli := test2.NewFakeClientBuilder().AddK8sObjects(kafkaInfra).Build()
	infraHandler := test2.CreateFakeKogitoInfraHandler(cli)
	infraManager := NewKogitoInfraManager(cli, test2.TestLogger, test2.GetRegisteredSchema(), infraHandler)
	err = infraManager.RemoveKogitoInfraOwnership(types.NamespacedName{Name: kafkaInfra.Name, Namespace: ns}, travels)
	assert.NoError(t, err)
	actualKafkaInfra := &api2.KogitoInfraTest{}
	exists, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: "kafka_infra", Namespace: ns}, actualKafkaInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, 0, len(actualKafkaInfra.GetOwnerReferences()))
}
