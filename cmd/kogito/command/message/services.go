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

const (
	serviceErrCreating                  = "Error while trying to create a new Kogito %s Service: %s "
	serviceSuccessfulInstalled          = "Kogito %s Service successfully installed in the Project %s."
	serviceCheckStatus                  = "Check the Service status by running 'oc describe %s/%s -n %s'"
	serviceNotInstalledNoKogitoOperator = "Skipping %s install since Kogito Operator is not available. Use 'kogito install %s' after installing the operator"
	// ServiceInfinispanNotAvailable ...
	ServiceInfinispanNotAvailable = "Infinispan is not available in the Project: %s. Please make sure to install the Infinispan Operator before deploying the service without infinispan-url provided "
	// ServiceInfinispanOperatorNotAvailable ...
	ServiceInfinispanOperatorNotAvailable = "Infinispan CRDs are available in the cluster, but looks like that the Infinispan Operator has not been installed yet. The service could take more time to deploy until the Infinispan Operator is ready"
	// ServiceKafkaNotAvailable ...
	ServiceKafkaNotAvailable = "Kafka is not available in the Project: %s. Please make sure to install the Strimzi Operator before deploying the service without Kafka provided "
)
