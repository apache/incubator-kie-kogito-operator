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

package kogitodataindex

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/stretchr/testify/assert"
)

func TestReconcileKogitoDataIndex_Reconcile(t *testing.T) {
	ns := t.Name()
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			InfinispanMeta: v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{UseKogitoInfra: true}},
		},
	}
	kafkaList := &kafkabetav1.KafkaList{
		Items: []kafkabetav1.Kafka{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "kafka",
					Namespace: ns,
				},
				Spec: kafkabetav1.KafkaSpec{
					Kafka: kafkabetav1.KafkaClusterSpec{
						Replicas: 1,
					},
				},
				Status: kafkabetav1.KafkaStatus{
					Listeners: []kafkabetav1.ListenerStatus{
						{
							Type: "plain",
							Addresses: []kafkabetav1.ListenerAddress{
								{
									Host: "kafka",
									Port: 9092,
								},
							},
						},
					},
				},
			},
		},
	}
	client := test.CreateFakeClient([]runtime.Object{instance, kafkaList}, nil, nil)
	r := &ReconcileKogitoDataIndex{
		client: client,
		scheme: meta.GetRegisteredSchema(),
	}
	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
	}

	// basic checks
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	if res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// check infra
	infra, created, ready, err := infrastructure.EnsureKogitoInfra(ns, client).WithInfinispan()
	assert.NoError(t, err)
	assert.False(t, created) // the created = true were returned when the infra was created during the reconcile phase
	assert.False(t, ready)   // we don't have status defined since the KogitoInfra controller is not running
	assert.NotNil(t, infra)  // we have a infra instance created during reconciliation phase
	assert.Equal(t, infrastructure.DefaultKogitoInfraName, infra.GetName())
}
