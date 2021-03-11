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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetDeployed(t *testing.T) {
	now := metav1.Now()
	conditionsMeta := &v1beta1.ConditionsMeta{}
	statusHandler := statusHandler{}
	status := statusHandler.SetDeployed(conditionsMeta)
	assert.True(t, status)

	assert.NotEmpty(t, conditionsMeta)
	assert.Equal(t, api.DeployedConditionType, conditionsMeta.Conditions[0].Type)
	assert.Equal(t, corev1.ConditionTrue, conditionsMeta.Conditions[0].Status)
	assert.True(t, now.Before(&conditionsMeta.Conditions[0].LastTransitionTime))
}

func TestSetDeployedSkipUpdate(t *testing.T) {
	conditionsMeta := &v1beta1.ConditionsMeta{}
	statusHandler := statusHandler{}
	status := statusHandler.SetDeployed(conditionsMeta)
	assert.True(t, status)

	// trying to set same deployed status again
	status = statusHandler.SetDeployed(conditionsMeta)
	assert.False(t, status)

	assert.Equal(t, 1, len(conditionsMeta.Conditions))
	assert.Equal(t, api.DeployedConditionType, conditionsMeta.Conditions[0].Type)
}

func TestSetProvisioning(t *testing.T) {
	now := metav1.Now()
	conditionsMeta := &v1beta1.ConditionsMeta{}
	statusHandler := statusHandler{}
	status := statusHandler.SetProvisioning(conditionsMeta)
	assert.True(t, status)
	assert.NotEmpty(t, conditionsMeta.Conditions)
	assert.Equal(t, api.ProvisioningConditionType, conditionsMeta.Conditions[0].Type)
	assert.Equal(t, corev1.ConditionTrue, conditionsMeta.Conditions[0].Status)
	assert.True(t, now.Before(&conditionsMeta.Conditions[0].LastTransitionTime))
}

func TestSetProvisioningSkipUpdate(t *testing.T) {
	conditionsMeta := &v1beta1.ConditionsMeta{}
	statusHandler := statusHandler{}
	status := statusHandler.SetProvisioning(conditionsMeta)
	assert.True(t, status)

	// trying to set same deployed status again
	status = statusHandler.SetProvisioning(conditionsMeta)
	assert.False(t, status)

	assert.Equal(t, 1, len(conditionsMeta.Conditions))
	assert.Equal(t, api.ProvisioningConditionType, conditionsMeta.Conditions[0].Type)
}

func TestSetProvisioningAndThenDeployed(t *testing.T) {
	now := metav1.Now()
	// we set a sleep to not conflict the time
	time.Sleep(1 * time.Second)
	conditionsMeta := &v1beta1.ConditionsMeta{}
	statusHandler := statusHandler{}
	assert.True(t, statusHandler.SetProvisioning(conditionsMeta))
	assert.True(t, statusHandler.SetDeployed(conditionsMeta))
	assert.NotEmpty(t, conditionsMeta.Conditions)
	assert.Equal(t, 2, len(conditionsMeta.Conditions))
	condition := conditionsMeta.Conditions[0]
	assert.Equal(t, api.ProvisioningConditionType, condition.Type)
	assert.Equal(t, corev1.ConditionTrue, condition.Status)
	assert.True(t, now.Before(&condition.LastTransitionTime))

	assert.Equal(t, api.DeployedConditionType, conditionsMeta.Conditions[1].Type)
	assert.Equal(t, corev1.ConditionTrue, conditionsMeta.Conditions[1].Status)
	assert.True(t, condition.LastTransitionTime.Before(&conditionsMeta.Conditions[1].LastTransitionTime))
}

func TestSetFailed(t *testing.T) {
	failureMessage := "Unknown error occurs"
	conditionsMeta := &v1beta1.ConditionsMeta{}
	statusHandler := statusHandler{}
	statusHandler.SetFailed(conditionsMeta, api.ServiceReconciliationFailure, fmt.Errorf(failureMessage))
	assert.NotEmpty(t, conditionsMeta.Conditions)
	assert.Equal(t, 1, len(conditionsMeta.Conditions))
	condition := conditionsMeta.Conditions[0]
	assert.Equal(t, api.FailedConditionType, condition.Type)
	assert.Equal(t, corev1.ConditionFalse, condition.Status)
	assert.Equal(t, api.ServiceReconciliationFailure, condition.Reason)
	assert.Equal(t, failureMessage, condition.Message)
}
