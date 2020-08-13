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

package shared

import (
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
)

type servicesRemoval struct {
	namespace string
	client    *kogitocli.Client
	err       error
}

// ServicesRemoval provides an interface for handling infrastructure services removal
type ServicesRemoval interface {
	// RemoveInfinispan removes the installed infinispan instance.
	RemoveInfinispan() ServicesRemoval
	// RemoveKeycloak removes the installed keycloak instance.
	RemoveKeycloak() ServicesRemoval
	// RemoveKafka removes the installed kafka instance.
	RemoveKafka() ServicesRemoval
	// RemovePrometheus removes the installed kafka instance.
	RemovePrometheus() ServicesRemoval
	// RemoveGrafana removes the installed kafka instance.
	RemoveGrafana() ServicesRemoval
	// GetError return any given error during the installation process
	GetError() error
}

func (s servicesRemoval) RemoveInfinispan() ServicesRemoval {
	if s.err == nil {
		s.err = removeInfinispan(s.client, s.namespace)
	}
	return s
}

func (s servicesRemoval) RemoveKeycloak() ServicesRemoval {
	if s.err == nil {
		s.err = removeKeycloak(s.client, s.namespace)
	}
	return s
}

func (s servicesRemoval) RemoveKafka() ServicesRemoval {
	if s.err == nil {
		s.err = removeKafka(s.client, s.namespace)
	}
	return s
}

func (s servicesRemoval) RemovePrometheus() ServicesRemoval {
	if s.err == nil {
		s.err = removePrometheus(s.client, s.namespace)
	}
	return s
}

func (s servicesRemoval) RemoveGrafana() ServicesRemoval {
	if s.err == nil {
		s.err = removeGrafana(s.client, s.namespace)
	}
	return s
}

func (s servicesRemoval) GetError() error {
	return s.err
}

// ServicesRemovalBuilder creates the basic structure for services removal definition.
func ServicesRemovalBuilder(client *kogitocli.Client, namespace string) ServicesRemoval {
	return servicesRemoval{
		namespace: namespace,
		client:    client,
	}
}
