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

package controllers

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReconcileKogitoJobsService_Reconcile(t *testing.T) {
	replicas := int32(1)
	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: v1.ObjectMeta{Name: infrastructure.DefaultJobsServiceName, Namespace: t.Name()},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType:       v1beta1.JobsService,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(instance).OnOpenShift().Build()

	r := &jobsServiceSupportingServiceResource{
		log: logger.GetLogger("jobsService reconciler"),
	}
	// first reconciliation

	requeueAfter, err := r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.NoError(t, err)
	assert.True(t, requeueAfter == 0)
	// second time
	requeueAfter, err = r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.NoError(t, err)
	assert.True(t, requeueAfter == 0)

	_, err = kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 1)
}
