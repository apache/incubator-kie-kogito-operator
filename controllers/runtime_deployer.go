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

package controllers

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/connector"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/kogitoservice"
	"github.com/kiegroup/kogito-operator/core/manager"
	"github.com/kiegroup/kogito-operator/core/operator"

	"io/ioutil"
	"net/http"

	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/RHsyseng/operator-utils/pkg/resource/compare"
	"github.com/kiegroup/kogito-operator/core/framework"
	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	envVarExternalURL = "KOGITO_SERVICE_URL"

	// protobufConfigMapSuffix Suffix that is appended to Protobuf ConfigMap name
	protobufConfigMapSuffix = "protobuf-files"
	protobufSubdir          = "/persistence/protobuf/"
	protobufListFileName    = "list.json"

	envVarNamespace = "NAMESPACE"
)

// RuntimeDeployerHandler ...
type RuntimeDeployerHandler interface {
	OnGetComparators(comparator compare.ResourceComparator)
	OnObjectsCreate(kogitoService api.KogitoService) (resources map[reflect.Type][]resource.KubernetesResource, lists []runtime.Object, err error)
	OnDeploymentCreate(deployment *v1.Deployment) error
}

type runtimeDeployerHandler struct {
	operator.Context
	instance                 api.KogitoRuntimeInterface
	supportingServiceHandler manager.KogitoSupportingServiceHandler
	runtimeHandler           manager.KogitoRuntimeHandler
}

// NewRuntimeDeployerHandler ...
func NewRuntimeDeployerHandler(context operator.Context, instance api.KogitoRuntimeInterface, supportingServiceHandler manager.KogitoSupportingServiceHandler, runtimeHandler manager.KogitoRuntimeHandler) RuntimeDeployerHandler {
	return &runtimeDeployerHandler{
		Context:                  context,
		instance:                 instance,
		supportingServiceHandler: supportingServiceHandler,
		runtimeHandler:           runtimeHandler,
	}
}

func (d *runtimeDeployerHandler) OnGetComparators(comparator compare.ResourceComparator) {
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

func (d *runtimeDeployerHandler) OnObjectsCreate(kogitoService api.KogitoService) (resources map[reflect.Type][]resource.KubernetesResource, lists []runtime.Object, err error) {
	resources = make(map[reflect.Type][]resource.KubernetesResource)

	resObjectList, resType, res := d.createProtoBufConfigMap(kogitoService)
	lists = append(lists, resObjectList)
	resources[resType] = []resource.KubernetesResource{res}
	return
}

// onDeploymentCreate hooks into the infrastructure package to add additional capabilities/properties to the deployment creation
func (d *runtimeDeployerHandler) OnDeploymentCreate(deployment *v1.Deployment) error {
	// NAMESPACE service discovery
	framework.SetEnvVar(envVarNamespace, d.instance.GetNamespace(), &deployment.Spec.Template.Spec.Containers[0])
	// external URL
	if d.instance.GetStatus().GetExternalURI() != "" {
		framework.SetEnvVar(envVarExternalURL, d.instance.GetStatus().GetExternalURI(), &deployment.Spec.Template.Spec.Containers[0])
	}
	// sa
	deployment.Spec.Template.Spec.ServiceAccountName = infrastructure.RuntimeServiceAccountName
	// istio
	if d.instance.GetRuntimeSpec().IsEnableIstio() {
		framework.AddIstioInjectSidecarAnnotation(&deployment.Spec.Template.ObjectMeta)
	}

	urlHandler := connector.NewURLHandler(d.Context, d.runtimeHandler, d.supportingServiceHandler)
	if err := urlHandler.InjectDataIndexURLIntoDeployment(d.instance.GetNamespace(), deployment); err != nil {
		return err
	}

	if err := urlHandler.InjectJobsServiceURLIntoKogitoRuntimeDeployment(d.instance.GetNamespace(), deployment); err != nil {
		return err
	}

	if err := urlHandler.InjectTrustyURLIntoDeployment(d.instance.GetNamespace(), deployment); err != nil {
		return err
	}

	return nil
}

func (d *runtimeDeployerHandler) createProtoBufConfigMap(kogitoService api.KogitoService) (runtime.Object, reflect.Type, resource.KubernetesResource) {
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kogitoService.GetNamespace(),
			Name:      getProtoBufConfigMapName(kogitoService.GetName()),
			Labels: map[string]string{
				connector.ConfigMapProtoBufEnabledLabelKey: "true",
				framework.LabelAppKey:                      kogitoService.GetName(),
			},
		},
		Data: d.getProtobufData(kogitoService),
	}
	return &corev1.ConfigMapList{}, reflect.TypeOf(corev1.ConfigMap{}), configMap
}

func (d *runtimeDeployerHandler) getProtobufData(kogitoService api.KogitoService) map[string]string {
	deployerHandler := kogitoservice.NewDeploymentHandler(d.Context)
	available, err := deployerHandler.IsDeploymentAvailable(kogitoService)
	if err != nil {
		d.Log.Error(err, "failed to check deployment status")
		return nil
	}
	if !available {
		d.Log.Debug("deployment not available")
		return nil
	}

	kogitoServiceHandler := kogitoservice.NewKogitoServiceHandler(d.Context)
	protobufEndpoint := kogitoServiceHandler.GetKogitoServiceEndpoint(kogitoService) + protobufSubdir
	protobufListURL := protobufEndpoint + protobufListFileName
	protobufListBytes, err := getHTTPFileBytes(protobufListURL)
	if err != nil {
		d.Log.Error(err, "failed to get protobuf file list")
		return nil
	}
	if protobufListBytes == nil {
		d.Log.Debug("no protobuf list found", "protobuf file", protobufListURL)
		return nil
	}
	var protobufList []string
	err = json.Unmarshal(protobufListBytes, &protobufList)
	if err != nil {
		d.Log.Error(err, "failed to parse protobuf file list")
		return nil
	}
	d.Log.Debug("Protobuf List", "files", strings.Join(protobufList, ","))

	var protobufFileBytes []byte
	data := map[string]string{}
	for _, fileName := range protobufList {
		protobufFileURL := protobufEndpoint + fileName
		protobufFileBytes, err = getHTTPFileBytes(protobufFileURL)
		if err != nil {
			d.Log.Error(err, "failed to fetch protobuf", "Protobuf Url", protobufFileURL)
			continue
		}
		if protobufFileBytes == nil {
			d.Log.Error(fmt.Errorf("protobuf Files not found"), "Protobuf Files not found", "Protobuf URL", protobufFileURL)
			continue
		}
		data[fileName] = string(protobufFileBytes)
	}
	return data
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
