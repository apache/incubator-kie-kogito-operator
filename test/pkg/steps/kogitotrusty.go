/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package steps

import (
	"github.com/cucumber/godog"
	"github.com/kiegroup/kogito-operator/test/pkg/framework"
	"github.com/kiegroup/kogito-operator/test/pkg/steps/mappers"
	bddtypes "github.com/kiegroup/kogito-operator/test/pkg/types"
)

/*
	DataTable for Trusty:
	| config          | infra       | <KogitoInfra name>        |
	| runtime-request | cpu/memory  | value                     |
	| runtime-limit   | cpu/memory  | value                     |
	| runtime-env     | varName     | varValue                  |
*/

func registerKogitoTrustyServiceSteps(ctx *godog.ScenarioContext, data *Data) {
	ctx.Step(`^Install Kogito Trusty with (\d+) replicas$`, data.installKogitoTrustyServiceWithReplicas)
	ctx.Step(`^Install Kogito Trusty with (\d+) replicas with configuration:$`, data.installKogitoTrustyServiceWithReplicasWithConfiguration)
	ctx.Step(`^Kogito Trusty has (\d+) pods running within (\d+) minutes$`, data.kogitoTrustyHasPodsRunningWithinMinutes)
}

func (data *Data) installKogitoTrustyServiceWithReplicas(replicas int) error {
	trusty := framework.GetKogitoTrustyResourceStub(data.Namespace, replicas)
	return framework.InstallKogitoTrustyService(data.Namespace, framework.GetDefaultInstallerType(), &bddtypes.KogitoServiceHolder{KogitoService: trusty})
}

func (data *Data) installKogitoTrustyServiceWithReplicasWithConfiguration(replicas int, table *godog.Table) error {
	trusty := &bddtypes.KogitoServiceHolder{
		KogitoService: framework.GetKogitoTrustyResourceStub(data.Namespace, replicas),
	}

	if err := mappers.MapKogitoServiceTable(table, trusty); err != nil {
		return err
	}

	return framework.InstallKogitoTrustyService(data.Namespace, framework.GetDefaultInstallerType(), trusty)
}

func (data *Data) kogitoTrustyHasPodsRunningWithinMinutes(podNb, timeoutInMin int) error {
	return framework.WaitForKogitoTrustyService(data.Namespace, podNb, timeoutInMin)
}
