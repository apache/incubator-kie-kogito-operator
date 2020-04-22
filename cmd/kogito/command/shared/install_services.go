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
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
)

type servicesInstallation struct {
	namespace         string
	client            *kogitocli.Client
	operatorInstalled bool
	err               error
}

// ServicesInstallation provides an interface for handling infrastructure services installation
type ServicesInstallation interface {
	// InstallDataIndex installs Data Index. If no reference provided, it will install the default instance.
	// Depends on the Operator, install it first.
	InstallDataIndex(dataIndex *v1alpha1.KogitoDataIndex) ServicesInstallation
	// InstallOperator installs the Operator.
	InstallOperator(warnIfInstalled bool, operatorImage string, force bool) ServicesInstallation
	// InstallInfinispan install an infinispan instance.
	InstallInfinispan() ServicesInstallation
	// InstallKeycloak install a keycloak instance.
	InstallKeycloak() ServicesInstallation
	// InstallKafka install a kafka instance.
	InstallKafka() ServicesInstallation
	// SilentlyInstallOperator installs the operator without a warn if already deployed with the default image
	SilentlyInstallOperator() ServicesInstallation
	// OperatorInstalled assumes operator is already installed
	OperatorInstalled() ServicesInstallation
	// GetError return any given error during the installation process
	GetError() error
}

func (s servicesInstallation) OperatorInstalled() ServicesInstallation {
	s.operatorInstalled = true
	return s
}

func (s servicesInstallation) InstallDataIndex(dataIndex *v1alpha1.KogitoDataIndex) ServicesInstallation {
	if s.err == nil {
		if !s.operatorInstalled { // depends on operator
			log.Info(message.DataIndexNotInstalledNoOperator)
			return s
		}
		if dataIndex == nil {
			s.err = installDefaultDataIndex(s.client, s.namespace)
		} else {
			s.err = installCustomizedDataIndex(s.client, s.namespace, dataIndex)
		}
	}
	return s
}

func (s servicesInstallation) InstallOperator(warnIfInstalled bool, operatorImage string, force bool) ServicesInstallation {
	if s.err == nil && !s.operatorInstalled {
		s.operatorInstalled, s.err = InstallOperatorIfNotExists(s.namespace, operatorImage, s.client, warnIfInstalled, force)
	}
	return s
}

func (s servicesInstallation) SilentlyInstallOperator() ServicesInstallation {
	return s.InstallOperator(false, "", false)
}

func (s servicesInstallation) InstallInfinispan() ServicesInstallation {
	if s.err == nil {
		s.err = installInfinispan(s.client, s.namespace)
	}
	return s
}

func (s servicesInstallation) InstallKeycloak() ServicesInstallation {
	if s.err == nil {
		s.err = installKeycloak(s.client, s.namespace)
	}
	return s
}

func (s servicesInstallation) InstallKafka() ServicesInstallation {
	if s.err == nil {
		s.err = installKafka(s.client, s.namespace)
	}
	return s
}

func (s servicesInstallation) GetError() error {
	return s.err
}

// ServicesInstallationBuilder creates the basic structure for services installation definition.
func ServicesInstallationBuilder(client *kogitocli.Client, namespace string) ServicesInstallation {
	return servicesInstallation{
		namespace:         namespace,
		client:            client,
		operatorInstalled: false,
	}
}
