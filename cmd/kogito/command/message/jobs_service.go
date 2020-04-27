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

package message

import "fmt"

var (
	// JobsServiceErrCreating ...
	JobsServiceErrCreating = fmt.Sprintf(serviceErrCreating, "Jobs", "%s")
	// JobsServiceSuccessfulInstalled ...
	JobsServiceSuccessfulInstalled = fmt.Sprintf(serviceSuccessfulInstalled, "Jobs", "%s")
	// JobsServiceCheckStatus ...
	JobsServiceCheckStatus = fmt.Sprintf(serviceCheckStatus, "kogitojobsservice", "%s", "%s")
	// JobsServiceNotInstalledNoKogitoOperator ...
	JobsServiceNotInstalledNoKogitoOperator = fmt.Sprintf(serviceNotInstalledNoKogitoOperator, "Jobs Service", "jobs-service")
)
