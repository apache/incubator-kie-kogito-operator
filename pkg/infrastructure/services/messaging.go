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

package services

import (
	"encoding/json"
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
)

type messageTopicKind string

const (
	// consumed is the topics type that the application can consume
	consumed messageTopicKind = "CONSUMED"
	// topicInfoPath which the topics are fetched
	topicInfoPath = "/messaging/topics"
)

type messageTopic struct {
	Name string           `json:"name"`
	Kind messageTopicKind `json:"type"`
}

type messagingDeployer struct {
	scheme     *runtime.Scheme
	cli        *client.Client
	definition ServiceDefinition
}

// handleMessagingResources handles messaging resources creation.
// These resources can be required by the deployed service through a bound KogitoInfra.
func handleMessagingResources(cli *client.Client, scheme *runtime.Scheme, definition ServiceDefinition, service v1alpha1.KogitoService) error {
	m := messagingDeployer{
		scheme:     scheme,
		cli:        cli,
		definition: definition,
	}
	knativeHandler := knativeMessagingDeployer{messagingDeployer: m}
	kafkaHandler := kafkaMessagingDeployer{messagingDeployer: m}
	if err := knativeHandler.createRequiredResources(service); err != nil {
		return err
	}
	if err := kafkaHandler.createRequiredResources(service); err != nil {
		return err
	}
	return nil
}

func (m *messagingDeployer) fetchRequiredTopics(instance v1alpha1.KogitoService) ([]messageTopic, error) {
	svcURL := infrastructure.CreateKogitoServiceURI(instance)
	return m.fetchRequiredTopicsForURL(instance, svcURL)
}

func (m *messagingDeployer) fetchRequiredTopicsForURL(instance v1alpha1.KogitoService, serverURL string) ([]messageTopic, error) {
	available, err := isDeploymentAvailable(m.cli, instance)
	if err != nil {
		return nil, err
	}
	if !available {
		log.Debugf("Deployment not available yet for KogitoService %s ", instance.GetName())
		return nil, nil
	}
	topicsURL := fmt.Sprintf("%s%s", serverURL, topicInfoPath)
	resp, err := http.Get(topicsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Received NOT expected status code %d while making GET request to %s ", resp.StatusCode, topicsURL)
	}
	var topics []messageTopic
	if err := json.NewDecoder(resp.Body).Decode(&topics); err != nil {
		return nil, fmt.Errorf("Failed to decode response from %s into topics ", topicsURL)
	}
	return topics, nil
}

func (m *messagingDeployer) fetchInfraDependency(service v1alpha1.KogitoService, checker func(*v1alpha1.KogitoInfra) bool) (*v1alpha1.KogitoInfra, error) {
	for _, infraName := range service.GetSpec().GetInfra() {
		infra, err := infrastructure.MustFetchKogitoInfraInstance(m.cli, infraName, service.GetNamespace())
		if err != nil {
			return nil, err
		}
		if checker(infra) {
			return infra, nil
		}
	}
	return nil, nil
}
