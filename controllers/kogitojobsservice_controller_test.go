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

package controllers

import (
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/api/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestReconcileKogitoJobsService_Reconcile(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoJobsService{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultJobsServiceName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)

	r := KogitoJobsServiceReconciler{Client: cli, Scheme: meta.GetRegisteredSchema(), Log: test_log}

	// first reconciliation
	test.AssertReconcileMustNotRequeue(t, &r, instance)
	// second time
	test.AssertReconcileMustNotRequeue(t, &r, instance)

	_, err := kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 1)
}

func TestReconcileKogitoJobsService_Reconcile_WithInfinispan(t *testing.T) {
	replicas := int32(1)
	instance := &v1alpha1.KogitoJobsService{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultJobsServiceName, Namespace: t.Name()},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			InfinispanMeta: v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
				UseKogitoInfra: true,
			}},
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.CreateFakeClientOnOpenShift([]runtime.Object{instance}, nil, nil)

	r := KogitoJobsServiceReconciler{Client: cli, Scheme: meta.GetRegisteredSchema(), Log: test_log}

	test.AssertReconcileMustRequeue(t, &r, instance)
	test.AssertReconcileMustRequeue(t, &r, instance)

	kogitoInfra := &v1alpha1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultKogitoInfraName, Namespace: instance.Namespace},
	}
	_, err := kubernetes.ResourceC(cli).Fetch(kogitoInfra)
	assert.NoError(t, err)
	assert.True(t, kogitoInfra.Spec.InstallInfinispan)
}
