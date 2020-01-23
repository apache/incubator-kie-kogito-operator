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

// Messages for Infrastructure components
const (
	KogitoInfraErrCreating = "Error while trying to create a new Kogito Infra: %s "

	InfinispanErrInstalling       = "Error while trying to install Infinispan: %s "
	InfinispanSuccessfulInstalled = "Infinispan has been successfully installed in the Project %s."
	InfinispanErrRemoving         = "Error while trying to remove Infinispan: %s "
	InfinispanSuccessfulRemoved   = "Infinispan has been successfully removed from the Project %s."

	KafkaErrInstalling       = "Error while trying to install Kafka: %s "
	KafkaSuccessfulInstalled = "Kafka has been successfully installed in the Project %s."

	KeycloakErrInstalling       = "Error while trying to install Keycloak: %s "
	KeycloakSuccessfulInstalled = "Keycloak has been successfully installed in the Project %s."
	KeycloakErrRemoving         = "Error while trying to remove Keycloak: %s "
	KeycloakSuccessfulRemoved   = "Keycloak has been successfully removed from the Project %s."
)
