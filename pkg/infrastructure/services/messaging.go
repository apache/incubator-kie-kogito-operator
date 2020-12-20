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
	"net/http"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"k8s.io/apimachinery/pkg/runtime"
)

type messagingTopicType string
type messagingEventKind string

const (
	// incoming is the topics type that the application will subscribe
	incoming messagingTopicType = "INCOMING"
	// consumed is the CloudEvents that the application will can consume
	consumed messagingEventKind = "CONSUMED"
	// produced is the CloudEvents that the application will can produce
	produced messagingEventKind = "PRODUCED"
	// topicInfoPath which the topics are fetched
	topicInfoPath = "/messaging/topics"
)

type messagingTopic struct {
	Name       string               `json:"name"`
	Kind       messagingTopicType   `json:"type"`
	EventsMeta []messagingEventMeta `json:"eventsMeta"`
}

type messagingEventMeta struct {
	Type   string             `json:"type"`
	Source string             `json:"source"`
	Kind   messagingEventKind `json:"kind"`
}

type messagingDeployer struct {
	scheme     *runtime.Scheme
	cli        *client.Client
	definition ServiceDefinition
}

// handleMessagingResources handles messaging resources creation.
// These resources can be required by the deployed service through a bound KogitoInfra.
func handleMessagingResources(cli *client.Client, scheme *runtime.Scheme, definition ServiceDefinition, service v1beta1.KogitoService) error {
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

func (m *messagingDeployer) fetchTopicsAndSetCloudEventsStatus(instance v1beta1.KogitoService) ([]messagingTopic, error) {
	topics, err := m.fetchRequiredTopicsForURL(instance, infrastructure.GetKogitoServiceEndpoint(instance))
	if err != nil {
		return nil, err
	}
	m.setCloudEventsStatus(instance, topics)
	return topics, nil
}

func (m *messagingDeployer) setCloudEventsStatus(instance v1beta1.KogitoService, topics []messagingTopic) {
	var eventsConsumed []v1beta1.KogitoCloudEventInfo
	var eventsProduced []v1beta1.KogitoCloudEventInfo
	for _, topic := range topics {
		for _, event := range topic.EventsMeta {
			switch event.Kind {
			case consumed:
				eventsConsumed = append(eventsConsumed, v1beta1.KogitoCloudEventInfo{Type: event.Type, Source: event.Source})
			case produced:
				eventsProduced = append(eventsProduced, v1beta1.KogitoCloudEventInfo{Type: event.Type, Source: event.Source})
			}
		}
	}
	instance.GetStatus().SetCloudEvents(v1beta1.KogitoCloudEventsStatus{
		Consumes: eventsConsumed,
		Produces: eventsProduced,
	})
}

func (m *messagingDeployer) fetchRequiredTopicsForURL(instance v1beta1.KogitoService, serverURL string) ([]messagingTopic, error) {
	available, err := IsDeploymentAvailable(m.cli, instance)
	if err != nil {
		return nil, err
	}
	if !available {
		log.Debug("Deployment not available yet for KogitoService", "KogitoService", instance.GetName())
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
		return nil, errorForServiceNotReachable(resp.StatusCode, topicsURL, "GET")
	}
	var topics []messagingTopic
	if err := json.NewDecoder(resp.Body).Decode(&topics); err != nil {
		return nil, fmt.Errorf("Failed to decode response from %s into topics ", topicsURL)
	}
	return topics, nil
}

func (m *messagingDeployer) fetchInfraDependency(service v1beta1.KogitoService, checker func(*v1beta1.KogitoInfra) bool) (*v1beta1.KogitoInfra, error) {
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
