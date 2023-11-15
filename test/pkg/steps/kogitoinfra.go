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
	"github.com/cucumber/messages-go/v16"
	"github.com/kiegroup/kogito-operator/test/pkg/framework"
	"github.com/kiegroup/kogito-operator/test/pkg/steps/mappers"
)

/*
	DataTable for Kogito Infra:
	| config          | <key>       | <value>      |
*/

func registerKogitoInfraSteps(ctx *godog.ScenarioContext, data *Data) {
	ctx.Step(`^Install (Infinispan|MongoDB|Kafka|Keycloak|Broker) Kogito Infra "([^"]*)" targeting service "([^"]*)" within (\d+) (?:minute|minutes)$`, data.installKogitoInfraTargetingServiceWithinMinutes)
	ctx.Step(`^Install (Infinispan|MongoDB|Kafka|Keycloak|Broker) Kogito Infra "([^"]*)" targeting service "([^"]*)" within (\d+) (?:minute|minutes) with configuration:$`, data.installKogitoInfraTargetingServiceWithinMinutesWithConfiguration)
}

func (data *Data) installKogitoInfraTargetingServiceWithinMinutes(targetResourceType, name, targetResourceName string, timeoutInMin int) error {
	return data.installKogitoInfraTargetingServiceWithinMinutesWithConfiguration(targetResourceType, name, targetResourceName, timeoutInMin, &messages.PickleTable{})
}

func (data *Data) installKogitoInfraTargetingServiceWithinMinutesWithConfiguration(targetResourceType, name, targetResourceName string, timeoutInMin int, table *godog.Table) error {
	infraResource, err := framework.GetKogitoInfraResourceStub(data.Namespace, name, targetResourceType, targetResourceName)
	if err != nil {
		return err
	}

	if err := mappers.MapKogitoInfraTable(table, infraResource); err != nil {
		return err
	}

	framework.GetLogger(data.Namespace).Debug("Got kogitoInfra config", "config", infraResource.GetSpec().GetInfraProperties())
	err = framework.InstallKogitoInfraComponent(data.Namespace, framework.GetDefaultInstallerType(), infraResource)
	if err != nil {
		return err
	}

	return framework.WaitForKogitoInfraResource(data.Namespace, name, timeoutInMin, framework.GetKogitoInfraResource)
}
