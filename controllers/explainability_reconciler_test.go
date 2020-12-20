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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReconcileKogitoSupportingServiceExplainability_Reconcile(t *testing.T) {
	ns := t.Name()
	kogitoKafka := test.CreateFakeKogitoKafka(t.Name())
	instance := &v1beta1.KogitoSupportingService{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "expl",
			Namespace: ns,
		},
		Spec: v1beta1.KogitoSupportingServiceSpec{
			ServiceType: v1beta1.Explainability,
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Infra: []string{
					kogitoKafka.Name,
				},
			},
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(instance, kogitoKafka).OnOpenShift().Build()
	r := &explainabilitySupportingServiceResource{
		log: logger.GetLogger("explainability reconciler"),
	}

	// basic checks
	requeueAfter, err := r.Reconcile(cli, instance, meta.GetRegisteredSchema())
	assert.NoError(t, err)
	assert.True(t, requeueAfter == 0)
}
