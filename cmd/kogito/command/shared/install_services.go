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
	"reflect"

	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kogitocli "github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var defaultReplicas = int32(1)
var defaultServiceStatus = v1alpha1.KogitoServiceStatus{ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}}}
var defaultServiceSpec = v1alpha1.KogitoServiceSpec{Replicas: &defaultReplicas}

type serviceInfoMessages struct {
	errCreating                  string
	installed                    string
	checkStatus                  string
	notInstalledNoKogitoOperator string
	serviceExists                string
}

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
	// InstallJobsService installs Jobs Service. If no reference provided, it will install the default instance.
	// Depends on the Operator, install it first.
	InstallJobsService(jobsService *v1alpha1.KogitoJobsService) ServicesInstallation
	// InstallMgmtConsole installs Management Console. If no reference provided, it will install the default instance.
	// Depends on the Operator, install it first.
	InstallMgmtConsole(mgmtConsole *v1alpha1.KogitoMgmtConsole) ServicesInstallation
	// InstallOperator installs the Operator.
	InstallOperator(warnIfInstalled bool, operatorImage string, force bool) ServicesInstallation
	// InstallInfinispan install an infinispan instance.
	InstallInfinispan() ServicesInstallation
	// InstallKeycloak install a keycloak instance.
	InstallKeycloak() ServicesInstallation
	// InstallKafka install a kafka instance.
	InstallKafka() ServicesInstallation
	// SilentlyInstallOperatorIfNotExists installs the operator without a warn if already deployed with the default image
	SilentlyInstallOperatorIfNotExists() ServicesInstallation
	// WarnIfDependenciesNotReady checks if the given dependencies are installed, warn if they are not ready
	WarnIfDependenciesNotReady(infinispan, kafka bool) ServicesInstallation
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

func (s *servicesInstallation) WarnIfDependenciesNotReady(infinispan, kafka bool) ServicesInstallation {
	log := context.GetDefaultLogger()
	if infinispan {
		if infrastructure.IsInfinispanAvailable(s.client) {
			if available, err := infrastructure.IsInfinispanOperatorAvailable(s.client, s.namespace); err != nil {
				s.err = err
			} else if !available {
				log.Info(message.ServiceInfinispanOperatorNotAvailable)
			}
		} else {
			log.Infof(message.ServiceInfinispanNotAvailable, s.namespace)
		}
	}
	if kafka {
		if available := infrastructure.IsStrimziAvailable(s.client); !available {
			log.Infof(message.ServiceKafkaNotAvailable, s.namespace)
		}
	}
	return s
}

func (s *servicesInstallation) InstallDataIndex(dataIndex *v1alpha1.KogitoDataIndex) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(dataIndex, s.getDefaultDataIndex,
			serviceInfoMessages{
				errCreating:                  message.DataIndexErrCreating,
				installed:                    message.DataIndexSuccessfulInstalled,
				checkStatus:                  message.DataIndexCheckStatus,
				notInstalledNoKogitoOperator: message.DataIndexNotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallJobsService(jobsService *v1alpha1.KogitoJobsService) ServicesInstallation {
	if s.err == nil {
		s.err = s.installKogitoService(jobsService, s.getDefaultJobsService,
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
		s.err = s.installKogitoService(mgmtConsole, s.getDefaultMgmtConsole,
			serviceInfoMessages{
				errCreating:                  message.MgmtConsoleErrCreating,
				installed:                    message.MgmtConsoleSuccessfulInstalled,
				checkStatus:                  message.MgmtConsoleCheckStatus,
				notInstalledNoKogitoOperator: message.MgmtConsoleNotInstalledNoKogitoOperator,
			})
	}
	return s
}

func (s *servicesInstallation) InstallOperator(warnIfInstalled bool, operatorImage string, force bool) ServicesInstallation {
	if s.err == nil && !s.operatorInstalled {
		s.operatorInstalled, s.err = InstallOperatorIfNotExists(s.namespace, operatorImage, s.client, warnIfInstalled, force)
	}
	return s
}

func (s servicesInstallation) SilentlyInstallOperatorIfNotExists() ServicesInstallation {
	return s.InstallOperator(false, "", false)
}

func (s *servicesInstallation) InstallInfinispan() ServicesInstallation {
	if s.err == nil {
		s.err = installInfinispan(s.client, s.namespace)
	}
	return s
}

func (s *servicesInstallation) InstallKeycloak() ServicesInstallation {
	if s.err == nil {
		s.err = installKeycloak(s.client, s.namespace)
	}
	return s
}

func (s *servicesInstallation) InstallKafka() ServicesInstallation {
	if s.err == nil {
		s.err = installKafka(s.client, s.namespace)
	}
	return s
}

func (s *servicesInstallation) GetError() error {
	return s.err
}

func (s *servicesInstallation) installKogitoService(service v1alpha1.KogitoService, getService func() v1alpha1.KogitoService, messages serviceInfoMessages) error {
	if s.err == nil {
		log := context.GetDefaultLogger()
		if !s.operatorInstalled { // depends on operator
			log.Info(messages.notInstalledNoKogitoOperator)
			return nil
		}
		if service == nil || reflect.ValueOf(service).IsNil() {
			service = getService()
		}
		if err := kubernetes.ResourceC(s.client).Create(service); err != nil {
			return fmt.Errorf(messages.errCreating, err)
		}
		log.Infof(messages.installed, s.namespace)
		log.Infof(messages.checkStatus, service.GetName(), s.namespace)
	}
	return nil
}

// getDefaultDataIndex gets the default Data Index instance
func (s *servicesInstallation) getDefaultDataIndex() v1alpha1.KogitoService {
	return &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultDataIndexName, Namespace: s.namespace},
		Spec: v1alpha1.KogitoDataIndexSpec{
			KogitoServiceSpec: defaultServiceSpec,
			InfinispanMeta:    v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{UseKogitoInfra: true}},
			KafkaMeta:         v1alpha1.KafkaMeta{KafkaProperties: v1alpha1.KafkaConnectionProperties{UseKogitoInfra: true}},
		},
		Status: v1alpha1.KogitoDataIndexStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}

// getDefaultJobsService gets the default Jobs Service instance
func (s *servicesInstallation) getDefaultJobsService() v1alpha1.KogitoService {
	return &v1alpha1.KogitoJobsService{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultJobsServiceName, Namespace: s.namespace},
		Spec: v1alpha1.KogitoJobsServiceSpec{
			KogitoServiceSpec: defaultServiceSpec,
			InfinispanMeta:    v1alpha1.InfinispanMeta{InfinispanProperties: v1alpha1.InfinispanConnectionProperties{UseKogitoInfra: false}},
			KafkaMeta:         v1alpha1.KafkaMeta{KafkaProperties: v1alpha1.KafkaConnectionProperties{UseKogitoInfra: false}},
		},
		Status: v1alpha1.KogitoJobsServiceStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}

// getDefaultMgmtConsole gets the default Management Console instance
func (s *servicesInstallation) getDefaultMgmtConsole() v1alpha1.KogitoService {
	return &v1alpha1.KogitoMgmtConsole{
		ObjectMeta: metav1.ObjectMeta{Name: infrastructure.DefaultMgmtConsoleName, Namespace: s.namespace},
		Spec:       v1alpha1.KogitoMgmtConsoleSpec{KogitoServiceSpec: defaultServiceSpec},
		Status:     v1alpha1.KogitoMgmtConsoleStatus{KogitoServiceStatus: defaultServiceStatus},
	}
}
