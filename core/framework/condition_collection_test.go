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

package framework

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/api"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestConditionCollection(t *testing.T) {
	conditions := make([]api.ConditionInterface, 5)
	conditionCollection := NewConditionCollection(conditions)
	newCondition := v1beta1.Condition{
		Type:               api.DeployedConditionType,
		LastTransitionTime: metav1.Now(),
	}
	conditions = conditionCollection.AddCondition(newCondition)
	fmt.Println(conditions)

	newCondition2 := v1beta1.Condition{
		Type:               api.FailedConditionType,
		LastTransitionTime: metav1.Now(),
	}
	conditions = conditionCollection.AddCondition(newCondition2)
	fmt.Println(conditions)

	newCondition3 := v1beta1.Condition{
		Type:               api.ProvisioningConditionType,
		LastTransitionTime: metav1.Now(),
	}
	conditions = conditionCollection.AddCondition(newCondition3)
	fmt.Println(conditions)

	newCondition4 := v1beta1.Condition{
		Type:               api.DeployedConditionType,
		LastTransitionTime: metav1.Now(),
	}
	conditions = conditionCollection.AddCondition(newCondition4)
	fmt.Println(conditions)

	newCondition5 := v1beta1.Condition{
		Type:               api.ProvisioningConditionType,
		LastTransitionTime: metav1.Now(),
	}
	conditions = conditionCollection.AddCondition(newCondition5)
	fmt.Println(conditions)
}
