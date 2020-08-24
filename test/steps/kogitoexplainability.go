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
	"github.com/kiegroup/kogito-cloud-operator/test/framework"
	"github.com/kiegroup/kogito-cloud-operator/test/steps/mappers"
	bddtypes "github.com/kiegroup/kogito-cloud-operator/test/types"
)

/*
	DataTable for Explainability:
	| runtime-request | cpu/memory  | value                     |
	| runtime-limit   | cpu/memory  | value                     |
	| runtime-env     | varName     | varValue                  |
	| kafka           | externalURI | kafka-bootstrap:9092      |
	| kafka           | instance    | external-kafka            |
*/

func registerKogitoExplainabilityServiceSteps(ctx *godog.ScenarioContext, data *Data) {
	ctx.Step(`^Install Kogito Explainability with (\d+) replicas$`, data.installKogitoExplainabilityServiceWithReplicas)
	ctx.Step(`^Install Kogito Explainability with (\d+) replicas with configuration:$`, data.installKogitoExplainabilityServiceWithReplicasWithConfiguration)
	ctx.Step(`^Kogito Explainability has (\d+) pods running within (\d+) minutes$`, data.kogitoExplainabilityHasPodsRunningWithinMinutes)
}

func (data *Data) installKogitoExplainabilityServiceWithReplicas(replicas int) error {
	explainability := framework.GetKogitoExplainabilityResourceStub(data.Namespace, replicas)
	return framework.InstallKogitoExplainabilityService(data.Namespace, framework.GetDefaultInstallerType(), &bddtypes.KogitoServiceHolder{KogitoService: explainability})
}

func (data *Data) installKogitoExplainabilityServiceWithReplicasWithConfiguration(replicas int, table *godog.Table) error {
	explainability := &bddtypes.KogitoServiceHolder{
		KogitoService: framework.GetKogitoExplainabilityResourceStub(data.Namespace, replicas),
	}

	if err := mappers.MapKogitoServiceTable(table, explainability); err != nil {
		return err
	}

	return framework.InstallKogitoExplainabilityService(data.Namespace, framework.GetDefaultInstallerType(), explainability)
}

func (data *Data) kogitoExplainabilityHasPodsRunningWithinMinutes(podNb, timeoutInMin int) error {
	return framework.WaitForKogitoExplainabilityService(data.Namespace, podNb, timeoutInMin)
}
