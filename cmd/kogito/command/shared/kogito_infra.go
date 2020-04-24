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

package shared

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
)

func installInfinispan(cli *client.Client, namespace string) error {
	if _, _, err := infrastructure.EnsureKogitoInfra(namespace, cli).WithInfinispan().Apply(); err != nil {
		return err
	}
	context.GetDefaultLogger().Infof(message.InfinispanSuccessfulInstalled, namespace)
	return nil
}

func installKafka(cli *client.Client, namespace string) error {
	if _, _, err := infrastructure.EnsureKogitoInfra(namespace, cli).WithKafka().Apply(); err != nil {
		return err
	}
	context.GetDefaultLogger().Infof(message.KafkaSuccessfulInstalled, namespace)
	return nil
}

func installKeycloak(cli *client.Client, namespace string) error {
	if _, _, err := infrastructure.EnsureKogitoInfra(namespace, cli).WithKeycloak().Apply(); err != nil {
		return err
	}
	context.GetDefaultLogger().Infof(message.KeycloakSuccessfulInstalled, namespace)
	return nil
}

func removeInfinispan(cli *client.Client, namespace string) error {
	if _, _, err := infrastructure.EnsureKogitoInfra(namespace, cli).WithoutInfinispan().Apply(); err != nil {
		return err
	}
	context.GetDefaultLogger().Infof(message.InfinispanSuccessfulRemoved, namespace)
	return nil
}

func removeKeycloak(cli *client.Client, namespace string) error {
	if _, _, err := infrastructure.EnsureKogitoInfra(namespace, cli).WithoutKeycloak().Apply(); err != nil {
		return err
	}
	context.GetDefaultLogger().Infof(message.KeycloakSuccessfulRemoved, namespace)
	return nil
}
func removeKafka(cli *client.Client, namespace string) error {
	if _, _, err := infrastructure.EnsureKogitoInfra(namespace, cli).WithoutKafka().Apply(); err != nil {
		return err
	}
	context.GetDefaultLogger().Infof(message.KafkaSuccessfulRemoved, namespace)
	return nil
}
