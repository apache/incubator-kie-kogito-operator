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

package service

import (
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/converter"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/shared"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/util"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RuntimeService is interface to perform Kogito Runtime
type RuntimeService interface {
	InstallRuntimeService(cli *client.Client, flags *flag.RuntimeFlags) (err error)
	DeleteRuntimeService(cli *client.Client, name, project string) (err error)
}

type runtimeService struct {
	resourceCheckService shared.ResourceCheckService
}

// NewRuntimeService create and return runtimeService value
func NewRuntimeService() RuntimeService {
	return runtimeService{
		resourceCheckService: shared.NewResourceCheckService(),
	}
}

// InstallRuntimeService install Kogito runtime service
func (i runtimeService) InstallRuntimeService(cli *client.Client, flags *flag.RuntimeFlags) (err error) {
	log := context.GetDefaultLogger()
	log.Debugf("Installing Kogito Runtime : %s", flags.Name)
	if err := i.resourceCheckService.CheckKogitoRuntimeNotExists(cli, flags.Name, flags.Project); err != nil {
		return err
	}
	configMap, err := converter.CreateConfigMapFromFile(cli, flags.Name, flags.Project, &flags.ConfigFlags)
	if err != nil {
		return err
	}
	kogitoRuntime := v1beta1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{
			Name:      flags.Name,
			Namespace: flags.Project,
		},
		Spec: v1beta1.KogitoRuntimeSpec{
			EnableIstio: flags.EnableIstio,
			Runtime:     converter.FromRuntimeFlagsToRuntimeType(&flags.RuntimeTypeFlags),
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Replicas:              &flags.Replicas,
				Env:                   converter.FromStringArrayToEnvs(flags.Env, flags.SecretEnv),
				Image:                 flags.ImageFlags.Image,
				Resources:             converter.FromPodResourceFlagsToResourceRequirement(&flags.PodResourceFlags),
				ServiceLabels:         util.FromStringsKeyPairToMap(flags.ServiceLabels),
				InsecureImageRegistry: flags.ImageFlags.InsecureImageRegistry,
				PropertiesConfigMap:   configMap,
				Infra:                 flags.Infra,
				Monitoring:            converter.FromMonitoringFlagToMonitoring(&flags.MonitoringFlags),
				Config:                converter.FromConfigFlagsToMap(&flags.ConfigFlags),
				Probes:                converter.FromProbeFlagToKogitoProbe(&flags.ProbeFlags),
			},
		},
		Status: v1beta1.KogitoRuntimeStatus{
			KogitoServiceStatus: v1beta1.KogitoServiceStatus{
				ConditionsMeta: v1beta1.ConditionsMeta{Conditions: []v1beta1.Condition{}},
			},
		},
	}

	log.Debugf("Trying to deploy Kogito Service '%s'", kogitoRuntime.Name)
	// Create the Kogito application
	err = shared.
		ServicesInstallationBuilder(cli, flags.Project).
		CheckOperatorCRDs().
		InstallRuntimeService(&kogitoRuntime).
		GetError()
	if err != nil {
		return err
	}
	if err = printMgmtConsoleInfo(cli, flags.Project); err != nil {
		return err
	}
	return nil
}

func printMgmtConsoleInfo(client *client.Client, project string) error {
	log := context.GetDefaultLogger()
	endpoint, err := infrastructure.GetManagementConsoleEndpoint(client, project)
	if err != nil {
		return err
	}
	if endpoint == nil {
		log.Info(message.RuntimeServiceMgmtConsole)
	} else {
		log.Infof(message.RuntimeServiceMgmtConsoleEndpoint, endpoint.HTTPRouteURI)
	}
	return nil
}

// DeleteRuntimeService delete Kogito runtime service
func (i runtimeService) DeleteRuntimeService(cli *client.Client, name, project string) (err error) {
	log := context.GetDefaultLogger()
	if err := i.resourceCheckService.CheckKogitoRuntimeExists(cli, name, project); err != nil {
		return err
	}
	log.Debugf("About to delete service %s in namespace %s", name, project)
	if err := kubernetes.ResourceC(cli).Delete(&v1beta1.KogitoRuntime{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: project,
		},
	}); err != nil {
		return err
	}
	log.Infof("Successfully deleted Kogito Service %s in the Project %s", name, project)
	return nil
}
