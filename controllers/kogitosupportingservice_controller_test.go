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
	"github.com/kiegroup/kogito-cloud-operator/core/api"
	"github.com/kiegroup/kogito-cloud-operator/core/kogitosupportingservice"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/kiegroup/kogito-cloud-operator/internal"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReconcileKogitoSupportingService_Reconcile(t *testing.T) {
	replicas := int32(1)
	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: v1.ObjectMeta{Name: kogitosupportingservice.DefaultJobsServiceName, Namespace: t.Name()},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType:       api.JobsService,
			KogitoServiceSpec: api.KogitoServiceSpec{Replicas: &replicas},
		},
	}
	cli := test.NewFakeClientBuilder().UseScheme(internal.GetRegisteredSchema()).AddK8sObjects(instance).OnOpenShift().Build()

	r := &KogitoSupportingServiceReconciler{
		Client: cli,
		Log:    logger.GetLogger("KogitoSupportingService"),
		Scheme: internal.GetRegisteredSchema(),
	}
	test.AssertReconcileMustNotRequeue(t, r, instance)
}

func TestContains(t *testing.T) {
	allServices := []api.ServiceType{
		api.MgmtConsole,
		api.JobsService,
		api.TrustyAI,
	}
	testService := api.DataIndex

	assert.False(t, contains(allServices, testService))
}

// Check is the testService is available in the slice of allServices
func contains(allServices []api.ServiceType, testService api.ServiceType) bool {
	for _, a := range allServices {
		if a == testService {
			return true
		}
	}
	return false
}
