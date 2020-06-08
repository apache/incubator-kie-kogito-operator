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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
	"strings"
)

/*
TODO: review those functions/vars. Should all be private when fixing https://issues.redhat.com/browse/KOGITO-1998
*/

const (
	appPropConfigMapSuffix = "-properties"
	appPropContentKey      = "application.properties"
	defaultAppPropContent  = ""

	// AppPropContentHashKey is the annotation key for the content hash of application.properties
	AppPropContentHashKey = "appPropContentHash"
	// AppPropVolumeName is the name of the volume for application.properties
	AppPropVolumeName = "app-prop-config"

	appPropDefaultMode = int32(420)
	appPropFileName    = "application.properties"
	appPropFilePath    = "/home/kogito/config"

	appPropConcatPattern = "%s\n%s=%s"
)

// GetAppPropConfigMapContentHash calculates the hash of the application.properties contents in the ConfigMap
// If the ConfigMap doesn't exist, create a new one and return it.
func GetAppPropConfigMapContentHash(name, namespace string, appProps map[string]string, cli *client.Client) (string, *corev1.ConfigMap, error) {
	configMapName := GetAppPropConfigMapName(name)
	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: configMapName, Namespace: namespace}}

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
		appPropContentKey: appPropContent,
	}

	contentHash := fmt.Sprintf("%x", md5.Sum([]byte(configMap.Data[appPropContentKey])))

	return contentHash, configMap, nil
}

// GetAppPropConfigMapName generates the name of the config map for application.properties
func GetAppPropConfigMapName(name string) string {
	return name + appPropConfigMapSuffix
}

// CreateAppPropVolumeMount creates a container volume mount for mounting application.properties
func CreateAppPropVolumeMount() corev1.VolumeMount {
	return corev1.VolumeMount{
		Name:      AppPropVolumeName,
		MountPath: appPropFilePath,
		ReadOnly:  true,
	}
}

// CreateAppPropVolume creates a volume for application.properties
func CreateAppPropVolume(name string) corev1.Volume {
	defaultMode := appPropDefaultMode

	return corev1.Volume{
		Name: AppPropVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: GetAppPropConfigMapName(name),
				},
				Items: []corev1.KeyToPath{
					{
						Key:  appPropContentKey,
						Path: appPropFileName,
					},
				},
				DefaultMode: &defaultMode,
			},
		},
	}
}

// getAppPropsFromConfigMap extracts the application properties from the ConfigMap to a string map
func getAppPropsFromConfigMap(configMap *corev1.ConfigMap, exist bool) map[string]string {
	appProps := map[string]string{}
	if exist {
		if _, ok := configMap.Data[appPropContentKey]; ok {
			props := strings.Split(configMap.Data[appPropContentKey], "\n")
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
