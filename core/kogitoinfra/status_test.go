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

package kogitoinfra

import (
	"errors"
	"github.com/kiegroup/kogito-cloud-operator/api"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	"github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/kiegroup/kogito-cloud-operator/meta"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestUpdateBaseStatus(t *testing.T) {
	instance := &v1beta1.KogitoInfra{
		ObjectMeta: v1.ObjectMeta{
			Name:      "kogito-kafka",
			Namespace: t.Name(),
		},
		Spec: v1beta1.KogitoInfraSpec{
			Resource: v1beta1.Resource{
				Kind:       "Kafka",
				APIVersion: "kafka.strimzi.io/v1beta1",
			},
		},
	}

	cli := test.NewFakeClientBuilder().AddK8sObjects(instance).Build()
	context := &operator.Context{
		Log:    test.TestLogger,
		Client: cli,
		Scheme: meta.GetRegisteredSchema(),
	}
	statusHandler := NewStatusHandler(context)
	err := errors.New("error")
	statusHandler.UpdateBaseStatus(instance, &err)
	test.AssertFetchMustExist(t, cli, instance)
	conditions := *instance.Status.Conditions
	assert.Equal(t, 2, len(conditions))
	assert.Equal(t, string(api.KogitoInfraFailure), conditions[0].Type)
	assert.Equal(t, v1.ConditionTrue, conditions[0].Status)
	assert.Equal(t, string(api.KogitoInfraSuccess), conditions[1].Type)
	assert.Equal(t, v1.ConditionFalse, conditions[1].Status)
}
