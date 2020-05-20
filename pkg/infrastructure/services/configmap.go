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
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
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
)

// GetAppPropConfigMapContentHash calculates the hash of the application.properties contents in the ConfigMap
// If the ConfigMap doesn't exist, create a new one and return it.
func GetAppPropConfigMapContentHash(name, namespace string, cli *client.Client) (string, *corev1.ConfigMap, error) {
	configMapName := GetAppPropConfigMapName(name)
	configMap := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: configMapName, Namespace: namespace}}

	exist, err := kubernetes.ResourceC(cli).Fetch(configMap)
	if err != nil {
		return "", nil, err
	}

	if exist {
		if _, ok := configMap.Data[appPropContentKey]; !ok {
			exist = false
		}
	}

	if !exist {
		configMap.Data = map[string]string{
			appPropContentKey: defaultAppPropContent,
		}
	}

	contentHash := fmt.Sprintf("%x", md5.Sum([]byte(configMap.Data[appPropContentKey])))

	if exist {
		return contentHash, nil, nil
	}

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

// ExcludeAppPropConfigMapFromResource excludes the application properties config map from the resources to be compared for changes.
// The config map is supposed to be modified by users directly, the operator should not overwrite the changes made by users.
func ExcludeAppPropConfigMapFromResource(name string, resourceMap map[reflect.Type][]resource.KubernetesResource) {
	tp := reflect.TypeOf(corev1.ConfigMap{})
	if configmaps, ok := resourceMap[tp]; ok {
		name := GetAppPropConfigMapName(name)
		for i, cm := range configmaps {
			if cm.GetName() == name {
				if i == len(configmaps)-1 {
					resourceMap[tp] = configmaps[:i]
				} else {
					resourceMap[tp] = append(configmaps[:i], configmaps[i+1:]...)
				}
				return
			}
		}
	}
}
