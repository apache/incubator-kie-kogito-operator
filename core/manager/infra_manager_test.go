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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/core/framework"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	"github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/kiegroup/kogito-cloud-operator/internal"
	"github.com/kiegroup/kogito-cloud-operator/meta"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func TestMustFetchKogitoInfraInstance_InstanceNotFound(t *testing.T) {
	ns := t.Name()
	name := "InfinispanInfra"
	cli := test.NewFakeClientBuilder().Build()
	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	infraHandler := internal.NewKogitoInfraHandler(context)
	infraManager := NewKogitoInfraManager(context, infraHandler)
	_, err := infraManager.MustFetchKogitoInfraInstance(types.NamespacedName{Name: name, Namespace: ns})
	assert.Error(t, err)
}

func TestRemoveKogitoInfraOwnership(t *testing.T) {
	ns := t.Name()
	scheme := meta.GetRegisteredSchema()
	travels := &v1beta1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{
			Name:      "travels",
			Namespace: ns,
			UID:       test.GenerateUID(),
		},
		Spec: v1beta1.KogitoRuntimeSpec{
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Infra: []string{
					"infinispan_infra",
					"kafka_infra",
				},
			},
		},
	}

	kafkaInfra := &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kafka_infra",
			Namespace: ns,
		},
	}
	err := framework.AddOwnerReference(travels, scheme, kafkaInfra)
	assert.NoError(t, err)

	cli := test.NewFakeClientBuilder().AddK8sObjects(kafkaInfra).Build()
	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	infraHandler := internal.NewKogitoInfraHandler(context)
	infraManager := NewKogitoInfraManager(context, infraHandler)
	err = infraManager.RemoveKogitoInfraOwnership(types.NamespacedName{Name: kafkaInfra.Name, Namespace: ns}, travels)
	assert.NoError(t, err)
	actualKafkaInfra := &v1beta1.KogitoInfra{}
	exists, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: "kafka_infra", Namespace: ns}, actualKafkaInfra)
	assert.NoError(t, err)
	assert.True(t, exists)
	assert.Equal(t, 0, len(actualKafkaInfra.GetOwnerReferences()))
}
