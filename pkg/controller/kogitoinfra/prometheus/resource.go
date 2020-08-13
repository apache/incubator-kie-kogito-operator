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

package prometheus

import (
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/read"
	prometheusv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// InstanceName is the default name for the Prometheus provisioned instance
	InstanceName = "kogito-prometheus"
	replicasSize = int32(1)
)

var log = logger.GetLogger("kogitoinfra_resource")

// GetDeployedResources gets the resources deployed as is
func GetDeployedResources(kogitoInfra *v1alpha1.KogitoInfra, cli *client.Client) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	if infrastructure.IsPrometheusAvailable(cli) {
		reader := read.New(cli.ControlCli).WithNamespace(kogitoInfra.Namespace).WithOwnerObject(kogitoInfra)
		resources, err = reader.ListAll(&prometheusv1.PrometheusList{})
		if err != nil {
			log.Warn("Failed to list deployed objects. ", err)
			return nil, err
		}
	}

	return
}

// CreateRequiredResources creates the very basic resources to have Prometheus in the namespace
func CreateRequiredResources(kogitoInfra *v1alpha1.KogitoInfra) (resources map[reflect.Type][]resource.KubernetesResource, err error) {
	resources = make(map[reflect.Type][]resource.KubernetesResource, 1)
	if kogitoInfra.Spec.InstallPrometheus {
		log.Debugf("Creating default resources for Prometheus installation for Kogito Infra on %s namespace", kogitoInfra.Namespace)
		replicas := replicasSize
		prometheus := &prometheusv1.Prometheus{
			ObjectMeta: metav1.ObjectMeta{Namespace: kogitoInfra.Namespace, Name: InstanceName},
			Spec: prometheusv1.PrometheusSpec{
				Replicas: &replicas,
			},
		}
		resources[reflect.TypeOf(prometheusv1.Prometheus{})] = []resource.KubernetesResource{prometheus}
		log.Debugf("Requested objects created as %s", resources)
	}
	return
}
