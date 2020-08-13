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

package grafana

import (
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	grafanav1 "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// InstanceName is the default name for the Grafana provisioned instance
	InstanceName = "kogito-grafana"
)

var log = logger.GetLogger("kogitoinfra_resource")

// GetDeployedResources gets the resources deployed as is
func GetDeployedResources(kogitoInfra *v1alpha1.KogitoInfra, cli *client.Client) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	if infrastructure.IsGrafanaAvailable(cli) {
		reader := read.New(cli.ControlCli).WithNamespace(kogitoInfra.Namespace).WithOwnerObject(kogitoInfra)
		resources, err = reader.ListAll(&grafanav1.GrafanaList{})
		if err != nil {
			log.Warn("Failed to list deployed objects. ", err)
			return nil, err
		}
	}

	return
}

// CreateRequiredResources creates the very basic resources to have Grafana in the namespace
func CreateRequiredResources(kogitoInfra *v1alpha1.KogitoInfra) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	resources = make(map[reflect.Type][]resource.KubernetesResource, 1)
	if kogitoInfra.Spec.InstallGrafana {
		log.Debugf("Creating default resources for Grafana installation for Kogito Infra on %s namespace", kogitoInfra.Namespace)
		Grafana := &grafanav1.Grafana{
			ObjectMeta: metav1.ObjectMeta{Namespace: kogitoInfra.Namespace, Name: InstanceName},
			Spec: grafanav1.GrafanaSpec{
				InitialReplicas: 1,
			},
		}
		resources[reflect.TypeOf(grafanav1.Grafana{})] = []resource.KubernetesResource{Grafana}
		log.Debugf("Requested objects created as %s", resources)
	}
	return
}
