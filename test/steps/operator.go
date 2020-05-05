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

package steps

import (
	"fmt"
	"strings"

	"github.com/cucumber/godog"
	"github.com/kiegroup/kogito-cloud-operator/test/framework"
)

func registerOperatorSteps(s *godog.Suite, data *Data) {
	s.Step(`^Kogito Operator should be installed with dependencies$`, data.kogitoOperatorShouldBeInstalledWithDependencies)
	s.Step(`^Kogito Operator should be installed from subscription with dependencies$`, data.kogitoOperatorShouldBeInstalledFromSubscriptionWithDependencies)
	s.Step(`^Kogito Operator is deployed$`, data.kogitoOperatorIsDeployed)
	s.Step(`^Kogito Operator is deployed from subscription using channel "([^"]*)"$`, data.kogitoOperatorIsDeployedFromSubscriptionAndChannel)
	s.Step(`^Kogito Operator is deployed with ((?:Infinispan|Kafka|Keycloak|, | and )+) (?:operator|operators)$`, data.kogitoOperatorIsDeployedWithDependencies)

	s.Step(`^CLI install Kogito operator$`, data.cliInstallKogitoOperator)
}

func (data *Data) kogitoOperatorShouldBeInstalledWithDependencies() error {
	return framework.WaitForKogitoOperatorRunningWithDependencies(data.Namespace)
}

func (data *Data) kogitoOperatorShouldBeInstalledFromSubscriptionWithDependencies() error {
	return framework.WaitForKogitoOperatorRunningFromSubscriptionWithDependencies(data.Namespace)
}

func (data *Data) kogitoOperatorIsDeployed() error {
	// if operator not available, then install via yaml files
	if exists, err := framework.IsKogitoOperatorRunning(data.Namespace); err != nil {
		return fmt.Errorf("Error while trying to retrieve the operator: %v ", err)
	} else if !exists {
		if err := framework.DeployKogitoOperatorFromYaml(data.Namespace); err != nil {
			return fmt.Errorf("Error while deploying operator: %v", err)
		}

		if err := framework.WaitForKogitoOperatorRunning(data.Namespace); err != nil {
			return fmt.Errorf("Error while checking operator running: %v", err)
		}
	}

	return nil
}

func (data *Data) kogitoOperatorIsDeployedWithDependencies(dependencies string) error {
	// Install kogito operator
	if exists, err := framework.IsKogitoOperatorRunning(data.Namespace); err != nil {
		return fmt.Errorf("Error while trying to retrieve the operator: %v ", err)
	} else if !exists {
		if err := framework.DeployKogitoOperatorFromYaml(data.Namespace); err != nil {
			return fmt.Errorf("Error while deploying operator: %v", err)
		}
	}

	// Install operator dependencies
	var installedOperators []string
	for dependentOperator := range framework.KogitoOperatorCommunityDependencies {
		if strings.Contains(dependencies, dependentOperator) {
			if err := framework.InstallCommunityKogitoOperatorDependency(data.Namespace, dependentOperator); err != nil {
				return err
			}
			installedOperators = append(installedOperators, dependentOperator)
		}
	}

	// Wait for kogito operator to be running
	if err := framework.WaitForKogitoOperatorRunning(data.Namespace); err != nil {
		return fmt.Errorf("Error while checking operator running: %v", err)
	}
	// Wait for dependent operators to be running
	for _, installedOperator := range installedOperators {
		if err := framework.WaitForKogitoOperatorDependencyRunning(data.Namespace, installedOperator); err != nil {
			return err
		}
	}

	return nil
}

func (data *Data) kogitoOperatorIsDeployedFromSubscriptionAndChannel(channel string) error {
	return framework.DeployKogitoOperatorFromSubscription(data.Namespace, channel)
}

func (data *Data) cliInstallKogitoOperator() error {
	_, err := framework.ExecuteCliCommandInNamespace(data.Namespace, "install", "operator")
	return err
}
