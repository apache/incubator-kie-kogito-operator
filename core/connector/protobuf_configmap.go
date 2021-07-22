// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package connector

import (
	"encoding/json"
	"fmt"
	"github.com/kiegroup/kogito-operator/api"
	"github.com/kiegroup/kogito-operator/core/framework"
	"github.com/kiegroup/kogito-operator/core/infrastructure"
	"github.com/kiegroup/kogito-operator/core/kogitoservice"
	"github.com/kiegroup/kogito-operator/core/operator"
	"io/ioutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
	"path"
	"strings"
)

const (
	// DefaultProtobufMountPath Default Proto Buf file path
	DefaultProtobufMountPath = operator.KogitoHomeDir + "/data/protobufs"
	// ConfigMapProtoBufEnabledLabelKey label key used by configMaps that are meant to hold protobuf files
	ConfigMapProtoBufEnabledLabelKey = "kogito-protobuf"
	// protobufConfigMapSuffix Suffix that is appended to Protobuf ConfigMap name
	protobufConfigMapSuffix = "protobuf-files"
	protobufSubdir          = "/persistence/protobuf/"
	protobufListFileName    = "list.json"
)

// ProtoBufConfigMapHandler ...
type ProtoBufConfigMapHandler interface {
	GetProtoBufConfigMapName(runtimeInstance api.KogitoRuntimeInterface) string
	CreateProtoBufConfigMap(runtimeInstance api.KogitoRuntimeInterface) (*corev1.ConfigMap, error)
	FetchProtoBufConfigMap(runtimeInstance api.KogitoRuntimeInterface) (*corev1.ConfigMap, error)
}

type protobufConfigMapHandler struct {
	operator.Context
	deploymentHandler    kogitoservice.DeploymentHandler
	kogitoServiceHandler kogitoservice.ServiceHandler
	configMapHandler     infrastructure.ConfigMapHandler
}

// NewProtoBufConfigMapHandler ...
func NewProtoBufConfigMapHandler(context operator.Context) ProtoBufConfigMapHandler {
	return &protobufConfigMapHandler{
		Context:              context,
		deploymentHandler:    kogitoservice.NewDeploymentHandler(context),
		kogitoServiceHandler: kogitoservice.NewKogitoServiceHandler(context),
		configMapHandler:     infrastructure.NewConfigMapHandler(context),
	}
}

func (p *protobufConfigMapHandler) FetchProtoBufConfigMap(runtimeInstance api.KogitoRuntimeInterface) (*corev1.ConfigMap, error) {
	labelSelector := map[string]string{
		framework.LabelAppKey:            runtimeInstance.GetName(),
		ConfigMapProtoBufEnabledLabelKey: "true",
	}
	configMapList, err := p.configMapHandler.FetchConfigMapsForLabel(runtimeInstance.GetNamespace(), labelSelector)
	if err != nil {
		return nil, err
	}
	if len(configMapList.Items) > 0 {
		return &configMapList.Items[0], nil
	}
	return nil, nil
}

func (p *protobufConfigMapHandler) CreateProtoBufConfigMap(runtimeInstance api.KogitoRuntimeInterface) (*corev1.ConfigMap, error) {
	protoBufData, err := p.getProtobufData(runtimeInstance)
	if err != nil {
		return nil, err
	}
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: runtimeInstance.GetNamespace(),
			Name:      p.GetProtoBufConfigMapName(runtimeInstance),
			Labels: map[string]string{
				ConfigMapProtoBufEnabledLabelKey: "true",
				framework.LabelAppKey:            runtimeInstance.GetName(),
			},
			Annotations: map[string]string{
				infrastructure.FromFileKey:  "true",
				infrastructure.MountPathKey: path.Join(DefaultProtobufMountPath, runtimeInstance.GetName()),
				infrastructure.FileModeKey:  fmt.Sprint(framework.ModeForProtoBufConfigMapVolume),
			},
		},
		Data: protoBufData,
	}
	return configMap, nil
}

func (p *protobufConfigMapHandler) getProtobufData(runtimeInstance api.KogitoRuntimeInterface) (map[string]string, error) {
	available, err := p.deploymentHandler.IsDeploymentAvailable(runtimeInstance)
	if err != nil {
		p.Log.Error(err, "failed to check deployment status")
		return nil, err
	}
	if !available {
		p.Log.Debug("deployment not available")
		return nil, infrastructure.ErrorForDeploymentNotReachable(runtimeInstance.GetName())
	}

	protobufEndpoint := p.kogitoServiceHandler.GetKogitoServiceEndpoint(runtimeInstance) + protobufSubdir
	protobufListURL := protobufEndpoint + protobufListFileName
	protobufListBytes, err := getHTTPFileBytes(protobufListURL)
	if err != nil {
		p.Log.Error(err, "failed to get protobuf file list", "protobufListURL", protobufListURL)
		return nil, err
	}
	if protobufListBytes == nil {
		p.Log.Debug("no protobuf list found", "protobuf file", protobufListURL)
		return nil, nil
	}
	var protobufList []string
	err = json.Unmarshal(protobufListBytes, &protobufList)
	if err != nil {
		p.Log.Error(err, "failed to parse protobuf file list")
		return nil, err
	}
	p.Log.Debug("Protobuf List", "files", strings.Join(protobufList, ","))

	var protobufFileBytes []byte
	data := map[string]string{}
	for _, fileName := range protobufList {
		protobufFileURL := protobufEndpoint + fileName
		protobufFileBytes, err = getHTTPFileBytes(protobufFileURL)
		if err != nil {
			p.Log.Error(err, "failed to fetch protobuf", "Protobuf Url", protobufFileURL)
			continue
		}
		if protobufFileBytes == nil {
			p.Log.Error(fmt.Errorf("protobuf Files not found"), "Protobuf Files not found", "Protobuf URL", protobufFileURL)
			continue
		}
		data[fileName] = string(protobufFileBytes)
	}
	return data, nil
}

// GetProtoBufConfigMapName gets the name of the protobuf configMap based the given KogitoRuntime instance
func (p *protobufConfigMapHandler) GetProtoBufConfigMapName(runtimeInstance api.KogitoRuntimeInterface) string {
	return fmt.Sprintf("%s-%s", runtimeInstance.GetName(), protobufConfigMapSuffix)
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
