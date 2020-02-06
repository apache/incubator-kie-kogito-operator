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
	"github.com/cucumber/godog"
	"github.com/kiegroup/kogito-cloud-operator/test/smoke/framework"
)

// RegisterCliSteps register all CLI steps existing
func registerPrometheusSteps(s *godog.Suite, data *Data) {
	s.Step(`^Prometheus Operator is deployed$`, data.prometheusOperatorIsDeployed)
	s.Step(`^Prometheus instance is deployed, monitoring services with label name "([^"]*)" and value "([^"]*)"$`, data.prometheusInstanceIsDeployed)
}

func (data *Data) prometheusOperatorIsDeployed() error {
	return framework.InstallCommunityOperator(data.Namespace, "prometheus", "beta")
}

func (data *Data) prometheusInstanceIsDeployed(labelName, labelValue string) error {
	err := framework.DeployPrometheusInstance(data.Namespace, labelName, labelValue)
	if err != nil {
		return err
	}
	return framework.WaitForPods(data.Namespace, "app", "prometheus", 1, 3)
}
