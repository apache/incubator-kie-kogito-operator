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
	grafana "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	// InstanceName is the default name for the Grafana provisioned instance
	instanceName = "kogito-grafana"
)

var log = logger.GetLogger("kogitografana_resource")

func loadDeployedGrafanaInstance(cli *client.Client, instanceName string, namespace string) (*grafana.Grafana, error) {
	log.Debug("fetching deployed kogito grafana instance")
	grafanaInstance := &grafana.Grafana{}
	if exits, err := kubernetes.ResourceC(cli).FetchWithKey(types.NamespacedName{Name: instanceName, Namespace: namespace}, grafanaInstance); err != nil {
		log.Error(err, "Error occurs while fetching kogito grafana instance")
		return nil, err
	} else if !exits {
		log.Debug("Kogito grafana instance is not exists")
		return nil, nil
	} else {
		log.Debug("Kogito grafana instance found")
		return grafanaInstance, nil
	}
}

func createNewGrafanaInstance(cli *client.Client, name string, namespace string, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) (*grafana.Grafana, error) {
	log.Debug("Going to create kogito grafana instance")
	grafanaRes := infrastructure.GetGrafanaDefaultResource(name, namespace)
	if err := controllerutil.SetOwnerReference(instance, grafanaRes, scheme); err != nil {
		return nil, err
	}
	if err := kubernetes.ResourceC(cli).Create(grafanaRes); err != nil {
		log.Error("Error occurs while creating kogito grafana instance")
		return nil, err
	}
	log.Debug("Kogito grafana instance created successfully")
	return grafanaRes, nil
}
