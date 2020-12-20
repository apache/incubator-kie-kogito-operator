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
	"fmt"

	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
)

type serviceInfoMessages struct {
	errCreating string
	installed   string
	checkStatus string
}

type servicesInstallation struct {
	namespace string
	client    *kogitocli.Client
	err       error
}

// ServicesInstallation provides an interface for handling infrastructure services installation
type ServicesInstallation interface {
	// BuildService build kogito service.
	// Depends on the Operator, install it first.
	InstallBuildService(build *v1beta1.KogitoBuild) ServicesInstallation
	// DeployService deploy Runtime service.
	// Depends on the Operator, install it first.
	InstallRuntimeService(runtime *v1beta1.KogitoRuntime) ServicesInstallation
	// InstallSupportingService installs supporting services. If no reference provided, it will install the default instance.
	// Depends on the Operator, install it first.
	InstallSupportingService(supportingService *v1beta1.KogitoSupportingService) ServicesInstallation
	// InstallInfraService install kogito Infra.
	// Depends on the Operator, install it first.
	InstallInfraService(infra *v1beta1.KogitoInfra) ServicesInstallation
	// CheckOperatorCRDs checks wheather the CRDs are available on the cluster or not
	CheckOperatorCRDs() ServicesInstallation
	// GetError return any given error during the installation process
	GetError() error
}

// ServicesInstallationBuilder creates the basic structure for services installation definition.
func ServicesInstallationBuilder(client *kogitocli.Client, namespace string) ServicesInstallation {
	return &servicesInstallation{
		namespace: namespace,
		client:    client,
	}
}

func (s *servicesInstallation) InstallBuildService(build *v1beta1.KogitoBuild) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(build,
			&serviceInfoMessages{
				errCreating: message.BuildServiceErrCreating,
				installed:   message.BuildServiceSuccessfulInstalled,
				checkStatus: message.BuildServiceCheckStatus,
			})
	}
	return s
}

func (s *servicesInstallation) InstallRuntimeService(runtime *v1beta1.KogitoRuntime) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(runtime,
			&serviceInfoMessages{
				errCreating: message.RuntimeServiceErrCreating,
				installed:   message.RuntimeServiceSuccessfulInstalled,
				checkStatus: message.RuntimeServiceCheckStatus,
			})
	}
	return s
}

func (s *servicesInstallation) InstallSupportingService(supportingService *v1beta1.KogitoSupportingService) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(supportingService, getSupportingServiceInfoMessages(supportingService.Spec.ServiceType))
	}
	return s
}

func getSupportingServiceInfoMessages(serviceType v1beta1.ServiceType) *serviceInfoMessages {
	switch serviceType {
	case v1beta1.DataIndex:
		return &serviceInfoMessages{
			errCreating: message.DataIndexErrCreating,
			installed:   message.DataIndexSuccessfulInstalled,
			checkStatus: message.SupportingServiceCheckStatus,
		}
	case v1beta1.JobsService:
		return &serviceInfoMessages{
			errCreating: message.JobsServiceErrCreating,
			installed:   message.JobsServiceSuccessfulInstalled,
			checkStatus: message.SupportingServiceCheckStatus,
		}
	case v1beta1.MgmtConsole:
		return &serviceInfoMessages{
			errCreating: message.MgmtConsoleErrCreating,
			installed:   message.MgmtConsoleSuccessfulInstalled,
			checkStatus: message.SupportingServiceCheckStatus,
		}
	case v1beta1.Explainability:
		return &serviceInfoMessages{
			errCreating: message.ExplainabilityErrCreating,
			installed:   message.ExplainabilitySuccessfulInstalled,
			checkStatus: message.SupportingServiceCheckStatus,
		}
	case v1beta1.TrustyAI:
		return &serviceInfoMessages{
			errCreating: message.TrustyErrCreating,
			installed:   message.TrustySuccessfulInstalled,
			checkStatus: message.SupportingServiceCheckStatus,
		}
	case v1beta1.TrustyUI:
		return &serviceInfoMessages{
			errCreating: message.TrustyUIErrCreating,
			installed:   message.TrustyUISuccessfulInstalled,
			checkStatus: message.SupportingServiceCheckStatus,
		}
	case v1beta1.TaskConsole:
		return &serviceInfoMessages{
			errCreating: message.TaskConsoleErrCreating,
			installed:   message.TaskConsoleSuccessfulInstalled,
			checkStatus: message.SupportingServiceCheckStatus,
		}
	}
	return nil
}

func (s *servicesInstallation) InstallInfraService(infra *v1beta1.KogitoInfra) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(infra,
			&serviceInfoMessages{
				errCreating: message.InfraServiceErrCreating,
				installed:   message.InfraServiceSuccessfulInstalled,
				checkStatus: message.InfraServiceCheckStatus,
			})
	}
	return s
}

func (s *servicesInstallation) GetError() error {
	return s.err
}

func (s *servicesInstallation) installKogitoService(resource meta.ResourceObject, messages *serviceInfoMessages) error {
	if s.err == nil {
		log := context.GetDefaultLogger()
		if err := kubernetes.ResourceC(s.client).Create(resource); err != nil {
			return fmt.Errorf(messages.errCreating, err)
		}
		log.Infof(messages.installed, s.namespace)
		log.Infof(messages.checkStatus, resource.GetName(), s.namespace)
	}
	return nil
}

func (s *servicesInstallation) CheckOperatorCRDs() ServicesInstallation {
	if !s.client.IsKogitoCRDsAvailable() {
		s.err = fmt.Errorf("kogito Operator CRDs not Found in the cluster. Please install operator before using")
	}
	return s
}
