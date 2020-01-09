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

package kogitojobsservice

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestReconcileKogitoJobsService_Reconcile(t *testing.T) {
	instance := &v1alpha1.KogitoJobsService{
		ObjectMeta: v1.ObjectMeta{Name: "jobs-service", Namespace: t.Name()},
		Spec:       v1alpha1.KogitoJobsServiceSpec{Replicas: 1},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)

	r := ReconcileKogitoJobsService{client: cli, scheme: meta.GetRegisteredSchema()}

	// first reconciliation
	result, err := r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Requeue)

	// second time
	result, err = r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Requeue)

	_, err = kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 1)
}

func TestReconcileKogitoJobsService_Reconcile_WithInfinispan(t *testing.T) {
	instance := &v1alpha1.KogitoJobsService{
		ObjectMeta: v1.ObjectMeta{Name: "jobs-service", Namespace: t.Name()},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			InfinispanMeta: v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
				UseKogitoInfra: true,
			}},
			Replicas: 1,
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)

	r := ReconcileKogitoJobsService{client: cli, scheme: meta.GetRegisteredSchema()}

	result, err := r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// there's no infra controller to deploy infinispan, keep reconciling
	assert.True(t, result.Requeue)

	result, err = r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}})
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Requeue)

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultKogitoInfraName, Namespace: instance.Namespace},
	}
	_, err = kubernetes.ResourceC(cli).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, kogitoInfra.Spec.InstallInfinispan)
}
