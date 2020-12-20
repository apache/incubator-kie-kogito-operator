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

package services

import (
	"crypto/md5"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
)

const (
	appPropConfigMapSuffix = "-properties"
	defaultAppPropContent  = ""

	// AppPropContentHashKey is the annotation key for the content hash of application.properties
	AppPropContentHashKey = "appPropContentHash"
	// AppPropVolumeName is the name of the volume for application.properties
	AppPropVolumeName = "app-prop-config"

	// ConfigMapApplicationPropertyKey is the file name used as a key for ConfigMaps mounted in Kogito services deployments
	ConfigMapApplicationPropertyKey = "application.properties"
	appPropFilePath                 = infrastructure.KogitoHomeDir + "/config"

	appPropConcatPattern = "%s\n%s=%s"
)

// getAppPropConfigMapContentHash calculates the hash of the application.properties contents in the ConfigMap
// If the ConfigMap doesn't exist, create a new one and return it.
func getAppPropConfigMapContentHash(service v1beta1.KogitoService, appProps map[string]string, cli *client.Client) (string, *corev1.ConfigMap, error) {
	configMapName := getAppPropConfigMapName(service)
	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: configMapName, Namespace: service.GetNamespace()}}

	exist, err := kubernetes.ResourceC(cli).Fetch(configMap)
	if err != nil {
		return "", nil, err
	}

	appPropsToApply := getAppPropsFromConfigMap(configMap, exist)
	if err = mergo.Merge(&appPropsToApply, appProps, mergo.WithOverride); err != nil {
		return "", nil, err
	}

	appPropContent := defaultAppPropContent
	if len(appPropsToApply) > 0 {
		var keys []string
		for key := range appPropsToApply {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			appPropContent = fmt.Sprintf(appPropConcatPattern, appPropContent, key, appPropsToApply[key])
		}
	}
	configMap.Data = map[string]string{
		ConfigMapApplicationPropertyKey: appPropContent,
	}

	contentHash := fmt.Sprintf("%x", md5.Sum([]byte(configMap.Data[ConfigMapApplicationPropertyKey])))

	return contentHash, configMap, nil
}

// getAppPropConfigMapName gets the name of the config map for application.properties
func getAppPropConfigMapName(service v1beta1.KogitoService) string {
	if len(service.GetSpec().GetPropertiesConfigMap()) > 0 {
		return service.GetSpec().GetPropertiesConfigMap()
	}
	return service.GetName() + appPropConfigMapSuffix
}

// createAppPropVolumeMount creates a container volume mount for mounting application.properties
func createAppPropVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      AppPropVolumeName,
		MountPath: appPropFilePath,
		ReadOnly:  true,
	}
}

// createAppPropVolume creates a volume for application.properties
func createAppPropVolume(service v1beta1.KogitoService) corev1.Volume {
	return corev1.Volume{
		Name: AppPropVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: getAppPropConfigMapName(service),
				},
				Items: []corev1.KeyToPath{
					{
						Key:  ConfigMapApplicationPropertyKey,
						Path: ConfigMapApplicationPropertyKey,
					},
				},
				DefaultMode: &framework.ModeForPropertyFiles,
			},
		},
	}
}

// getAppPropsFromConfigMap extracts the application properties from the ConfigMap to a string map
func getAppPropsFromConfigMap(configMap *corev1.ConfigMap, exist bool) map[string]string {
	appProps := map[string]string{}
	if exist {
		if _, ok := configMap.Data[ConfigMapApplicationPropertyKey]; ok {
			props := strings.Split(configMap.Data[ConfigMapApplicationPropertyKey], "\n")
			for _, p := range props {
				ps := strings.Split(p, "=")
				if len(ps) > 1 {
					appProps[strings.TrimSpace(ps[0])] = strings.TrimSpace(ps[1])
				}
			}
		}
	}
	return appProps
}
