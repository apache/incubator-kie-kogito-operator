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

package v1alpha1

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestSetDeployed(t *testing.T) {
	now := metav1.Now()
	conditions := ConditionsMeta{}
	assert.True(t, conditions.SetDeployed())

	assert.NotEmpty(t, conditions)
	assert.Equal(t, DeployedConditionType, conditions.Conditions[0].Type)
	assert.Equal(t, corev1.ConditionTrue, conditions.Conditions[0].Status)
	assert.True(t, now.Before(&conditions.Conditions[0].LastTransitionTime))
}

func TestSetDeployedSkipUpdate(t *testing.T) {
	conditionsMeta := ConditionsMeta{}
	conditionsMeta.SetDeployed()

	assert.NotEmpty(t, conditionsMeta)
	condition := conditionsMeta.Conditions[0]

	assert.False(t, conditionsMeta.SetDeployed())
	assert.Equal(t, 1, len(conditionsMeta.Conditions))
	assert.Equal(t, condition, conditionsMeta.Conditions[0])
}

func TestSetProvisioning(t *testing.T) {
	now := metav1.Now()
	conditionsMeta := ConditionsMeta{}
	assert.True(t, conditionsMeta.SetProvisioning())

	assert.NotEmpty(t, conditionsMeta.Conditions)
	assert.Equal(t, ProvisioningConditionType, conditionsMeta.Conditions[0].Type)
	assert.Equal(t, corev1.ConditionTrue, conditionsMeta.Conditions[0].Status)
	assert.True(t, now.Before(&conditionsMeta.Conditions[0].LastTransitionTime))
}

func TestSetProvisioningSkipUpdate(t *testing.T) {
	conditionsMeta := ConditionsMeta{}
	assert.True(t, conditionsMeta.SetProvisioning())

	assert.NotEmpty(t, conditionsMeta.Conditions)
	condition := conditionsMeta.Conditions[0]

	assert.False(t, conditionsMeta.SetProvisioning())
	assert.Equal(t, 1, len(conditionsMeta.Conditions))
	assert.Equal(t, condition, conditionsMeta.Conditions[0])
}

func TestSetProvisioningAndThenDeployed(t *testing.T) {
	now := metav1.Now()
	conditionsMeta := ConditionsMeta{}

	assert.True(t, conditionsMeta.SetProvisioning())
	assert.True(t, conditionsMeta.SetDeployed())

	assert.NotEmpty(t, conditionsMeta.Conditions)
	condition := conditionsMeta.Conditions[0]
	assert.Equal(t, 2, len(conditionsMeta.Conditions))
	assert.Equal(t, ProvisioningConditionType, condition.Type)
	assert.Equal(t, corev1.ConditionTrue, condition.Status)
	assert.True(t, now.Before(&condition.LastTransitionTime))

	assert.Equal(t, DeployedConditionType, conditionsMeta.Conditions[1].Type)
	assert.Equal(t, corev1.ConditionTrue, conditionsMeta.Conditions[1].Status)
	assert.True(t, condition.LastTransitionTime.Before(&conditionsMeta.Conditions[1].LastTransitionTime))
}

func TestBuffer(t *testing.T) {
	conditionsMeta := ConditionsMeta{}
	for i := 0; i < maxBufferCondition+2; i++ {
		conditionsMeta.SetFailed(UnknownReason, fmt.Errorf("error %d", i))
	}
	size := len(conditionsMeta.Conditions)
	assert.Equal(t, maxBufferCondition, size)
	assert.Equal(t, "error 6", conditionsMeta.Conditions[size-1].Message)
}
