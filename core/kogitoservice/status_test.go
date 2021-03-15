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

package kogitoservice

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/api"
	"github.com/kiegroup/kogito-cloud-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/core/operator"
	"github.com/kiegroup/kogito-cloud-operator/core/test"
	"github.com/kiegroup/kogito-cloud-operator/meta"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestReconciliation_ErrorOccur(t *testing.T) {
	instance := test.CreateFakeDataIndex(t.Name())
	cli := test.NewFakeClientBuilder().AddK8sObjects(instance).Build()
	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	statusHandler := NewStatusHandler(context)
	reconciliationError := fmt.Errorf("test error")
	statusHandler.HandleStatusUpdate(instance, &reconciliationError)
	assert.NotNil(t, instance)

	_, err := kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 3)
	failedCondition := getSpecificCondition(instance.GetStatus(), api.FailedConditionType)
	assert.NotNil(t, failedCondition)

	provisionedCondition := getSpecificCondition(instance.GetStatus(), api.ProvisioningConditionType)
	assert.NotNil(t, provisionedCondition)
	assert.Equal(t, metav1.ConditionFalse, provisionedCondition.Status)

	deployedCondition := getSpecificCondition(instance.GetStatus(), api.DeployedConditionType)
	assert.NotNil(t, deployedCondition)
	assert.Equal(t, metav1.ConditionFalse, deployedCondition.Status)
}

func TestReconciliation_RecoverableErrorOccur(t *testing.T) {
	instance := test.CreateFakeDataIndex(t.Name())
	cli := test.NewFakeClientBuilder().AddK8sObjects(instance).Build()
	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	statusHandler := NewStatusHandler(context)
	var reconciliationError error = errorForMonitoring(fmt.Errorf("test error"))
	statusHandler.HandleStatusUpdate(instance, &reconciliationError)
	assert.NotNil(t, instance)

	_, err := kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 3)
	failedCondition := getSpecificCondition(instance.GetStatus(), api.FailedConditionType)
	assert.NotNil(t, failedCondition)

	provisionedCondition := getSpecificCondition(instance.GetStatus(), api.ProvisioningConditionType)
	assert.NotNil(t, provisionedCondition)
	assert.Equal(t, metav1.ConditionTrue, provisionedCondition.Status)

	deployedCondition := getSpecificCondition(instance.GetStatus(), api.DeployedConditionType)
	assert.NotNil(t, deployedCondition)
	assert.Equal(t, metav1.ConditionFalse, deployedCondition.Status)
}

func TestReconciliation(t *testing.T) {
	instance := test.CreateFakeDataIndex(t.Name())
	cli := test.NewFakeClientBuilder().AddK8sObjects(instance).Build()
	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	statusHandler := NewStatusHandler(context)
	var err error = nil
	statusHandler.HandleStatusUpdate(instance, &err)
	assert.NotNil(t, instance)

	_, err = kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 2)

	provisionedCondition := getSpecificCondition(instance.GetStatus(), api.ProvisioningConditionType)
	assert.NotNil(t, provisionedCondition)
	assert.Equal(t, metav1.ConditionTrue, provisionedCondition.Status)

	deployedCondition := getSpecificCondition(instance.GetStatus(), api.DeployedConditionType)
	assert.NotNil(t, deployedCondition)
	assert.Equal(t, metav1.ConditionFalse, deployedCondition.Status)
}

func TestReconciliation_PodAlreadyRunning(t *testing.T) {
	instance := test.CreateFakeDataIndex(t.Name())
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name,
			Namespace: instance.Namespace,
		},
		Status: appsv1.DeploymentStatus{
			Replicas: 1,
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(instance, deployment).Build()
	context := &operator.Context{
		Client: cli,
		Log:    test.TestLogger,
		Scheme: meta.GetRegisteredSchema(),
	}
	statusHandler := NewStatusHandler(context)
	var err error = nil
	statusHandler.HandleStatusUpdate(instance, &err)
	assert.NotNil(t, instance)

	_, err = kubernetes.ResourceC(cli).Fetch(instance)
	assert.NoError(t, err)
	assert.NotNil(t, instance.Status)
	assert.Len(t, instance.Status.Conditions, 2)

	provisionedCondition := getSpecificCondition(instance.GetStatus(), api.ProvisioningConditionType)
	assert.NotNil(t, provisionedCondition)
	assert.Equal(t, metav1.ConditionFalse, provisionedCondition.Status)

	deployedCondition := getSpecificCondition(instance.GetStatus(), api.DeployedConditionType)
	assert.NotNil(t, deployedCondition)
	assert.Equal(t, metav1.ConditionTrue, deployedCondition.Status)
}

func getSpecificCondition(c api.ConditionMetaInterface, conditionType api.ConditionType) *metav1.Condition {
	for _, condition := range c.GetConditions() {
		if condition.Type == string(conditionType) {
			return &condition
		}
	}
	return nil
}
