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
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/test/framework"
)

func registerGraphQLSteps(ctx *godog.ScenarioContext, data *Data) {
	ctx.Step(`^GraphQL request on service "([^"]*)" is successful within (\d+) minutes with path "([^"]*)" and query:$`, data.graphqlRequestOnServiceIsSuccessfulWithinMinutesWithPathAndQuery)
	ctx.Step(`^GraphQL request on service "([^"]*)" is successful using access token "([^"]*)" within (\d+) minutes with path "([^"]*)" and query:$`, data.graphqlRequestOnServiceIsSuccessfulUsingAccessTokenWithinMinutesWithPathAndQuery)
	ctx.Step(`^GraphQL request on Data Index service returns ProcessInstances processName "([^"]*)" within (\d+) minutes$`, data.graphqlRequestOnDataIndexReturnsProcessInstancesProcessNameWithinMinutes)
	ctx.Step(`^GraphQL request on Data Index service returns Jobs ID "([^"]*)" within (\d+) minutes$`, data.graphqlRequestOnDataIndexReturnsJobsIDWithinMinutes)
}

func (data *Data) graphqlRequestOnServiceIsSuccessfulWithinMinutesWithPathAndQuery(serviceName string, timeoutInMin int, path string, query *godog.DocString) error {
	framework.GetLogger(data.Namespace).Debugf("graphqlRequestOnServiceWithPathAndBodyIsSuccessfulWithinMinutes with service %s, path %s, query %s and timeout %d", serviceName, path, query, timeoutInMin)
	uri, err := framework.WaitAndRetrieveEndpointURI(data.Namespace, serviceName)
	if err != nil {
		return err
	}
	var response interface{}
	return framework.WaitForSuccessfulGraphQLRequest(data.Namespace, uri, path, query.GetContent(), timeoutInMin, response, nil)
}

func (data *Data) graphqlRequestOnServiceIsSuccessfulUsingAccessTokenWithinMinutesWithPathAndQuery(serviceName, accessToken string, timeoutInMin int, path string, query *godog.DocString) error {
	accessToken = data.ResolveWithScenarioContext(accessToken)
	framework.GetLogger(data.Namespace).Debugf("graphqlRequestOnServiceIsSuccessfulUsingAccessTokenWithinMinutesWithPathAndQuery with service %s, path %s, query %s, access token %s and timeout %d", serviceName, path, query, accessToken, timeoutInMin)
	uri, err := framework.WaitAndRetrieveEndpointURI(data.Namespace, serviceName)
	if err != nil {
		return err
	}
	var response interface{}
	return framework.WaitForSuccessfulGraphQLRequestUsingAccessToken(data.Namespace, uri, path, query.GetContent(), accessToken, timeoutInMin, response, nil)
}

func (data *Data) graphqlRequestOnDataIndexReturnsProcessInstancesProcessNameWithinMinutes(processName string, timeoutInMin int) error {
	serviceName := infrastructure.DefaultDataIndexName
	query := getProcessInstancesNameQuery
	path := "graphql"

	framework.GetLogger(data.Namespace).Debugf("graphqlProcessNameRequestOnDataIndexIsSuccessfulWithinMinutes with service %s, path %s, query %s and timeout %d", serviceName, path, query, timeoutInMin)
	uri, err := framework.WaitAndRetrieveEndpointURI(data.Namespace, serviceName)
	if err != nil {
		return err
	}
	response := GraphqlDataIndexProcessInstancesQueryResponse{}
	return framework.WaitForSuccessfulGraphQLRequest(data.Namespace, uri, path, query, timeoutInMin, &response, func(response interface{}) (bool, error) {
		resp := response.(*GraphqlDataIndexProcessInstancesQueryResponse)
		for _, processInstance := range resp.ProcessInstances {
			if processInstance.ProcessName == processName {
				return true, nil
			}
		}
		return false, nil
	})
}

func (data *Data) graphqlRequestOnDataIndexReturnsJobsIDWithinMinutes(id string, timeoutInMin int) error {
	serviceName := infrastructure.DefaultDataIndexName
	query := getJobsIDQuery
	path := "graphql"

	framework.GetLogger(data.Namespace).Debugf("graphqlRequestOnDataIndexReturnsJobsIDWithinMinutes with service %s, path %s, query %s and timeout %d", serviceName, path, query, timeoutInMin)
	uri, err := framework.WaitAndRetrieveEndpointURI(data.Namespace, serviceName)
	if err != nil {
		return err
	}
	response := GraphqlDataIndexJobsQueryResponse{}
	return framework.WaitForSuccessfulGraphQLRequest(data.Namespace, uri, path, query, timeoutInMin, &response, func(response interface{}) (bool, error) {
		resp := response.(*GraphqlDataIndexJobsQueryResponse)
		for _, job := range resp.Jobs {
			if job.ID == id {
				return true, nil
			}
		}
		return false, nil
	})
}
