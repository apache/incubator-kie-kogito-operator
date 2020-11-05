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

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/meta"
)

type serviceInfoMessages struct {
	errCreating                  string
	installed                    string
	checkStatus                  string
	notInstalledNoKogitoOperator string
}

type servicesInstallation struct {
	namespace         string
	client            *kogitocli.Client
	operatorInstalled bool
	err               error
}

// ServicesInstallation provides an interface for handling infrastructure services installation
type ServicesInstallation interface {
	// BuildService build kogito service.
	// Depends on the Operator, install it first.
	InstallBuildService(build *v1alpha1.KogitoBuild) ServicesInstallation
	// DeployService deploy Runtime service.
	// Depends on the Operator, install it first.
	InstallRuntimeService(runtime *v1alpha1.KogitoRuntime) ServicesInstallation
	// InstallSupportingService installs supporting services. If no reference provided, it will install the default instance.
	// Depends on the Operator, install it first.
	InstallSupportingService(supportingService *v1alpha1.KogitoSupportingService) ServicesInstallation
	// InstallInfraService install kogito Infra.
	// Depends on the Operator, install it first.
	InstallInfraService(infra *v1alpha1.KogitoInfra) ServicesInstallation
	// InstallOperator installs the Operator.
	InstallOperator(warnIfInstalled bool, operatorImage string, force bool, ch KogitoChannelType) ServicesInstallation
	// SilentlyInstallOperatorIfNotExists installs the operator without a warn if already deployed with the default image
	SilentlyInstallOperatorIfNotExists(ch KogitoChannelType) ServicesInstallation
	// GetError return any given error during the installation process
	GetError() error
}

// ServicesInstallationBuilder creates the basic structure for services installation definition.
func ServicesInstallationBuilder(client *kogitocli.Client, namespace string) ServicesInstallation {
	return &servicesInstallation{
		namespace:         namespace,
		client:            client,
		operatorInstalled: false,
	}
}

func (s *servicesInstallation) InstallBuildService(build *v1alpha1.KogitoBuild) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(build,
			&serviceInfoMessages{
				errCreating:                  message.BuildServiceErrCreating,
				installed:                    message.BuildServiceSuccessfulInstalled,
				checkStatus:                  message.BuildServiceCheckStatus,
				notInstalledNoKogitoOperator: message.BuildServiceNotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallRuntimeService(runtime *v1alpha1.KogitoRuntime) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(runtime,
			&serviceInfoMessages{
				errCreating:                  message.RuntimeServiceErrCreating,
				installed:                    message.RuntimeServiceSuccessfulInstalled,
				checkStatus:                  message.RuntimeServiceCheckStatus,
				notInstalledNoKogitoOperator: message.RuntimeServiceNotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallSupportingService(supportingService *v1alpha1.KogitoSupportingService) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(supportingService, getSupportingServiceInfoMessages(supportingService.Spec.ServiceType))
	}
	return s
}

func getSupportingServiceInfoMessages(serviceType v1alpha1.ServiceType) *serviceInfoMessages {
	switch serviceType {
	case v1alpha1.DataIndex:
		return &serviceInfoMessages{
			errCreating:                  message.DataIndexErrCreating,
			installed:                    message.DataIndexSuccessfulInstalled,
			checkStatus:                  message.SupportingServiceCheckStatus,
			notInstalledNoKogitoOperator: message.DataIndexNotInstalledNoKogitoOperator,
		}
	case v1alpha1.JobsService:
		return &serviceInfoMessages{
			errCreating:                  message.JobsServiceErrCreating,
			installed:                    message.JobsServiceSuccessfulInstalled,
			checkStatus:                  message.SupportingServiceCheckStatus,
			notInstalledNoKogitoOperator: message.JobsServiceNotInstalledNoKogitoOperator,
		}
	case v1alpha1.MgmtConsole:
		return &serviceInfoMessages{
			errCreating:                  message.MgmtConsoleErrCreating,
			installed:                    message.MgmtConsoleSuccessfulInstalled,
			checkStatus:                  message.SupportingServiceCheckStatus,
			notInstalledNoKogitoOperator: message.MgmtConsoleNotInstalledNoKogitoOperator,
		}
	case v1alpha1.Explainability:
		return &serviceInfoMessages{
			errCreating:                  message.ExplainabilityErrCreating,
			installed:                    message.ExplainabilitySuccessfulInstalled,
			checkStatus:                  message.SupportingServiceCheckStatus,
			notInstalledNoKogitoOperator: message.ExplainabilityNotInstalledNoKogitoOperator,
		}
	case v1alpha1.TrustyAI:
		return &serviceInfoMessages{
			errCreating:                  message.TrustyErrCreating,
			installed:                    message.TrustySuccessfulInstalled,
			checkStatus:                  message.SupportingServiceCheckStatus,
			notInstalledNoKogitoOperator: message.TrustyNotInstalledNoKogitoOperator,
		}
	case v1alpha1.TrustyUI:
		return &serviceInfoMessages{
			errCreating:                  message.TrustyUIErrCreating,
			installed:                    message.TrustyUISuccessfulInstalled,
			checkStatus:                  message.SupportingServiceCheckStatus,
			notInstalledNoKogitoOperator: message.TrustyUINotInstalledNoKogitoOperator,
		}
	case v1alpha1.TaskConsole:
		return &serviceInfoMessages{
			errCreating:                  message.TaskConsoleErrCreating,
			installed:                    message.TaskConsoleSuccessfulInstalled,
			checkStatus:                  message.SupportingServiceCheckStatus,
			notInstalledNoKogitoOperator: message.TaskConsoleNotInstalledNoKogitoOperator,
		}
	}
	return nil
}

func (s *servicesInstallation) InstallInfraService(infra *v1alpha1.KogitoInfra) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(infra,
			&serviceInfoMessages{
				errCreating:                  message.InfraServiceErrCreating,
				installed:                    message.InfraServiceSuccessfulInstalled,
				checkStatus:                  message.InfraServiceCheckStatus,
				notInstalledNoKogitoOperator: message.InfraServiceNotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallOperator(warnIfInstalled bool, operatorImage string, force bool, ch KogitoChannelType) ServicesInstallation {
	if s.err == nil && !s.operatorInstalled {
		s.operatorInstalled, s.err = InstallOperatorIfNotExists(s.namespace, operatorImage, s.client, warnIfInstalled, force, ch)
	}
	return s
}

func (s servicesInstallation) SilentlyInstallOperatorIfNotExists(ch KogitoChannelType) ServicesInstallation {
	return s.InstallOperator(false, "", false, ch)
}

func (s *servicesInstallation) GetError() error {
	return s.err
}

func (s *servicesInstallation) installKogitoService(resource meta.ResourceObject, messages *serviceInfoMessages) error {
	if s.err == nil {
		log := context.GetDefaultLogger()
		if !s.operatorInstalled { // depends on operator
			log.Info(messages.notInstalledNoKogitoOperator)
			return nil
		}
		if err := kubernetes.ResourceC(s.client).Create(resource); err != nil {
			return fmt.Errorf(messages.errCreating, err)
		}
		log.Infof(messages.installed, s.namespace)
		log.Infof(messages.checkStatus, resource.GetName(), s.namespace)
	}
	return nil
}
