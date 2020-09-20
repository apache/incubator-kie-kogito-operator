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

var defaultReplicas = int32(1)
var defaultServiceStatus = v1alpha1.KogitoServiceStatus{ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}}}
var defaultServiceSpec = v1alpha1.KogitoServiceSpec{Replicas: &defaultReplicas}

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
	// InstallDataIndex installs Data Index. If no reference provided, it will install the default instance.
	// Depends on the Operator, install it first.
	InstallDataIndex(dataIndex *v1alpha1.KogitoDataIndex) ServicesInstallation
	// InstallExplainability installs Explainability. If no reference provided, it will install the default instance.
	// Depends on the Operator, install it first.
	InstallExplainability(explainability *v1alpha1.KogitoExplainability) ServicesInstallation
	// InstallTrusty installs Trusty. If no reference provided, it will install the default instance.
	// Depends on the Operator, install it first.
	InstallTrusty(trusty *v1alpha1.KogitoTrusty) ServicesInstallation
	// InstallJobsService installs Jobs Service. If no reference provided, it will install the default instance.
	// Depends on the Operator, install it first.
	InstallJobsService(jobsService *v1alpha1.KogitoJobsService) ServicesInstallation
	// InstallMgmtConsole installs Management Console. If no reference provided, it will install the default instance.
	// Depends on the Operator, install it first.
	InstallMgmtConsole(mgmtConsole *v1alpha1.KogitoMgmtConsole) ServicesInstallation
	// InstallTrustyUI installs Trusty UI. If no reference provided, it will install the default instance.
	// Depends on the Operator, install it first.
	InstallTrustyUI(trustyUI *v1alpha1.KogitoTrustyUI) ServicesInstallation
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
			serviceInfoMessages{
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
			serviceInfoMessages{
				errCreating:                  message.RuntimeServiceErrCreating,
				installed:                    message.RuntimeServiceSuccessfulInstalled,
				checkStatus:                  message.RuntimeServiceCheckStatus,
				notInstalledNoKogitoOperator: message.RuntimeServiceNotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallDataIndex(dataIndex *v1alpha1.KogitoDataIndex) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(dataIndex,
			serviceInfoMessages{
				errCreating:                  message.DataIndexErrCreating,
				installed:                    message.DataIndexSuccessfulInstalled,
				checkStatus:                  message.DataIndexCheckStatus,
				notInstalledNoKogitoOperator: message.DataIndexNotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallExplainability(explainability *v1alpha1.KogitoExplainability) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(explainability,
			serviceInfoMessages{
				errCreating:                  message.ExplainabilityErrCreating,
				installed:                    message.ExplainabilitySuccessfulInstalled,
				checkStatus:                  message.ExplainabilityCheckStatus,
				notInstalledNoKogitoOperator: message.ExplainabilityNotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallTrusty(trusty *v1alpha1.KogitoTrusty) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(trusty,
			serviceInfoMessages{
				errCreating:                  message.TrustyErrCreating,
				installed:                    message.TrustySuccessfulInstalled,
				checkStatus:                  message.TrustyCheckStatus,
				notInstalledNoKogitoOperator: message.TrustyNotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallJobsService(jobsService *v1alpha1.KogitoJobsService) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(jobsService,
			serviceInfoMessages{
				errCreating:                  message.JobsServiceErrCreating,
				installed:                    message.JobsServiceSuccessfulInstalled,
				checkStatus:                  message.JobsServiceCheckStatus,
				notInstalledNoKogitoOperator: message.JobsServiceNotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallMgmtConsole(mgmtConsole *v1alpha1.KogitoMgmtConsole) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(mgmtConsole,
			serviceInfoMessages{
				errCreating:                  message.MgmtConsoleErrCreating,
				installed:                    message.MgmtConsoleSuccessfulInstalled,
				checkStatus:                  message.MgmtConsoleCheckStatus,
				notInstalledNoKogitoOperator: message.MgmtConsoleNotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallTrustyUI(trustyUI *v1alpha1.KogitoTrustyUI) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(trustyUI,
			serviceInfoMessages{
				errCreating:                  message.TrustyUIErrCreating,
				installed:                    message.TrustyUISuccessfulInstalled,
				checkStatus:                  message.TrustyUICheckStatus,
				notInstalledNoKogitoOperator: message.TrustyUINotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallInfraService(infra *v1alpha1.KogitoInfra) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(infra,
			serviceInfoMessages{
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

func (s *servicesInstallation) installKogitoService(resource meta.ResourceObject, messages serviceInfoMessages) error {
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
