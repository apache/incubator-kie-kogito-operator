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

package kogitoruntime

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"io/ioutil"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
)

const (
	serviceAccountName = "kogito-service-viewer"

	envVarExternalURL = "KOGITO_SERVICE_URL"

	// protobufConfigMapSuffix Suffix that is appended to Protobuf ConfigMap name
	protobufConfigMapSuffix = "protobuf-files"
	protobufSubdir          = "/persistence/protobuf/"
	protobufListFileName    = "list.json"

	envVarNamespace = "NAMESPACE"
)

func onGetComparators(comparator compare.ResourceComparator) {
	comparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(corev1.ConfigMap{})).
			WithCustomComparator(framework.CreateConfigMapComparator()).
			Build())

	comparator.SetComparator(
		framework.NewComparatorBuilder().
			WithType(reflect.TypeOf(monv1.ServiceMonitor{})).
			WithCustomComparator(framework.CreateServiceMonitorComparator()).
			Build())
}

func onObjectsCreate(cli *client.Client, kogitoService v1beta1.KogitoService) (resources map[reflect.Type][]resource.KubernetesResource, lists []runtime.Object, err error) {
	resources = make(map[reflect.Type][]resource.KubernetesResource)

	resObjectList, resType, res := createProtoBufConfigMap(cli, kogitoService)
	lists = append(lists, resObjectList)
	resources[resType] = []resource.KubernetesResource{res}
	return
}

func getProtobufData(cli *client.Client, kogitoService v1beta1.KogitoService) map[string]string {
	available, err := services.IsDeploymentAvailable(cli, kogitoService)
	if err != nil {
		log.Errorf("failed to check status of %s: %v", kogitoService.GetName(), err)
		return nil
	}
	if !available {
		log.Debugf("deployment not available yet for %s ", kogitoService.GetName())
		return nil
	}

	protobufEndpoint := infrastructure.GetKogitoServiceEndpoint(kogitoService) + protobufSubdir
	protobufListURL := protobufEndpoint + protobufListFileName
	protobufListBytes, err := getHTTPFileBytes(protobufListURL)
	if err != nil {
		log.Errorf("failed to get %s protobuf file list: %v", kogitoService.GetName(), err)
		return nil
	}
	if protobufListBytes == nil {
		log.Debugf("no protobuf list found for %s at %s", kogitoService.GetName(), protobufListURL)
		return nil
	}
	var protobufList []string
	err = json.Unmarshal(protobufListBytes, &protobufList)
	if err != nil {
		log.Errorf("failed to parse %s protobuf file list: %v", kogitoService.GetName(), err)
		return nil
	}
	log.Debugf("%s Protobuf List: %v", kogitoService.GetName(), protobufList)

	var protobufFileBytes []byte
	data := map[string]string{}
	for _, fileName := range protobufList {
		protobufFileURL := protobufEndpoint + fileName
		protobufFileBytes, err = getHTTPFileBytes(protobufFileURL)
		if err != nil {
			log.Errorf("failed to get protobuf file for %s at %s: %v", kogitoService.GetName(), protobufFileURL, err)
			continue
		}
		if protobufFileBytes == nil {
			log.Errorf("%s protobuf list not found at %s", kogitoService.GetName(), protobufFileURL)
			continue
		}
		data[fileName] = string(protobufFileBytes)
	}
	return data
}

func createProtoBufConfigMap(cli *client.Client, kogitoService v1beta1.KogitoService) (runtime.Object, reflect.Type, resource.KubernetesResource) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kogitoService.GetNamespace(),
			Name:      getProtoBufConfigMapName(kogitoService.GetName()),
			Labels: map[string]string{
				infrastructure.ConfigMapProtoBufEnabledLabelKey: "true",
				framework.LabelAppKey:                           kogitoService.GetName(),
			},
		},
		Data: getProtobufData(cli, kogitoService),
	}
	return &corev1.ConfigMapList{}, reflect.TypeOf(corev1.ConfigMap{}), configMap
}

// onDeploymentCreate hooks into the infrastructure package to add additional capabilities/properties to the deployment creation
func onDeploymentCreate(cli *client.Client, deployment *v1.Deployment, kogitoService v1beta1.KogitoService) error {
	kogitoRuntime := kogitoService.(*v1beta1.KogitoRuntime)
	// NAMESPACE service discovery
	framework.SetEnvVar(envVarNamespace, kogitoService.GetNamespace(), &deployment.Spec.Template.Spec.Containers[0])
	// external URL
	if kogitoService.GetStatus().GetExternalURI() != "" {
		framework.SetEnvVar(envVarExternalURL, kogitoService.GetStatus().GetExternalURI(), &deployment.Spec.Template.Spec.Containers[0])
	}
	// sa
	deployment.Spec.Template.Spec.ServiceAccountName = serviceAccountName
	// istio
	if kogitoRuntime.Spec.EnableIstio {
		framework.AddIstioInjectSidecarAnnotation(&deployment.Spec.Template.ObjectMeta)
	}

	if err := infrastructure.InjectDataIndexURLIntoDeployment(cli, kogitoService.GetNamespace(), deployment); err != nil {
		return err
	}

	if err := infrastructure.InjectJobsServiceURLIntoKogitoRuntimeDeployment(cli, kogitoService.GetNamespace(), deployment); err != nil {
		return err
	}

	if err := infrastructure.InjectTrustyURLIntoDeployment(cli, kogitoService.GetNamespace(), deployment); err != nil {
		return err
	}

	return nil
}

// getProtoBufConfigMapName gets the name of the protobuf configMap based the given KogitoRuntime instance
func getProtoBufConfigMapName(serviceName string) string {
	return fmt.Sprintf("%s-%s", serviceName, protobufConfigMapSuffix)
}

func getHTTPFileBytes(fileURL string) ([]byte, error) {
	res, err := http.Get(fileURL)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, nil
	}
	fileBytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	return fileBytes, nil
}
