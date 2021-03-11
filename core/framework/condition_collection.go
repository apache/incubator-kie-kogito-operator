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

import "github.com/kiegroup/kogito-cloud-operator/api"

// ConditionCollection ...
type ConditionCollection interface {
	AddCondition(condition api.ConditionInterface) []api.ConditionInterface
}

type conditionCollection struct {
	conditions []api.ConditionInterface
}

// NewConditionCollection ...
func NewConditionCollection(conditions []api.ConditionInterface) ConditionCollection {
	return &conditionCollection{
		conditions: conditions,
	}
}

func (c *conditionCollection) AddCondition(newCondition api.ConditionInterface) []api.ConditionInterface {
	for i, condition := range c.conditions {
		if condition != nil && condition.GetType() == newCondition.GetType() {
			c.removeOldCondition(i)
		}
	}
	c.pushCondition(newCondition)
	return c.conditions
}

func (c *conditionCollection) removeOldCondition(index int) {
	c.conditions = append(c.conditions[:index], c.conditions[index+1:]...)
}

func (c *conditionCollection) pushCondition(newCondition api.ConditionInterface) {
	c.conditions = append(c.conditions, newCondition)
}
