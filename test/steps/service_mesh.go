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
)

func registerServiceMeshSteps(ctx *godog.ScenarioContext, data *Data) {
	// Deploy
	ctx.Step(`^Elasticsearch Operator is deployed$`, data.elasticsearchOperatorIsDeployed)
	ctx.Step(`^Jaeger Operator is deployed$`, data.jaegerOperatorIsDeployed)
	ctx.Step(`^Kiali Operator is deployed$`, data.kialiOperatorIsDeployed)
	ctx.Step(`^Service Mesh Operator is deployed$`, data.serviceMeshOperatorIsDeployed)
	ctx.Step(`^Service Mesh instance is deployed$`, data.serviceMeshInstanceIsDeployed)
}

func (data *Data) jaegerOperatorIsDeployed() error {
	return framework.InstallOperator(data.Namespace, "jaeger-product", "stable", framework.CommunityCatalog)
}

func (data *Data) elasticsearchOperatorIsDeployed() error {
	return framework.InstallOperator(data.Namespace, "elasticsearch-operator", "4.5", framework.CommunityCatalog)
}

func (data *Data) kialiOperatorIsDeployed() error {
	return framework.InstallOperator(data.Namespace, "kiali-ossm", "stable", framework.CommunityCatalog)
}

func (data *Data) serviceMeshOperatorIsDeployed() error {
	return framework.InstallOperator(data.Namespace, "servicemeshoperator", "stable", framework.CommunityCatalog)
}

func (data *Data) serviceMeshInstanceIsDeployed() error {
	err := framework.DeployServiceMeshInstance(data.Namespace)
	if err != nil {
		return err
	}
	return framework.WaitForPodsWithLabel(data.Namespace, "istio", "citadel", 1, 3)
}
