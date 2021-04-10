// Copyright 2020 Red Hat, Inc. and/or its affiliates
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

var (
	getProcessInstancesNameQuery = `
	{
	  ProcessInstances{
		processName
	  }
	}
  `
	getProcessInstancesIDByNameQuery = `
	{
	  ProcessInstances(where: {processName: {equal: "$name"}}, pagination: {offset: $offset, limit: $limit}) {
		id
	  }
	}
  `
	getJobsIDQuery = `
	{
	  Jobs{
		id
	  }
	}
  `
)

// GraphqlDataIndexProcessInstancesQueryResponse Query response type of Data Index GraphQL endpoint containing process instances
type GraphqlDataIndexProcessInstancesQueryResponse struct {
	ProcessInstances []struct {
		ProcessName string `json:"processName,omitempty"`
	}
}

// GraphqlDataIndexProcessInstancesIDQueryResponse Query response type of Data Index GraphQL endpoint containing process instance IDs
type GraphqlDataIndexProcessInstancesIDQueryResponse struct {
	ProcessInstances []struct {
		ID string `json:"id,omitempty"`
	}
}

// GraphqlDataIndexJobsQueryResponse Query response type of Data Index GraphQL endpoint containing jobs
type GraphqlDataIndexJobsQueryResponse struct {
	Jobs []struct {
		ID string `json:"id,omitempty"`
	}
}
