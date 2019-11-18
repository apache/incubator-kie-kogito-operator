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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
)

func TestReconcileKogitoDataIndex_Reconcile(t *testing.T) {
	ns := t.Name()
	instance := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-data-index",
			Namespace: ns,
		},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Infinispan: v1alpha1.InfinispanConnectionProperties{UseKogitoInfra: true},
		},
	}
	client := test.CreateFakeClient([]runtime.Object{instance}, nil, nil)
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
	infra, created, err := infrastructure.CreateOrFetchInfra(ns, client)
	assert.NoError(t, err)
	assert.False(t, created)
	assert.NotNil(t, infra)
	assert.Equal(t, infrastructure.DefaultKogitoInfraName, infra.GetName())
}
