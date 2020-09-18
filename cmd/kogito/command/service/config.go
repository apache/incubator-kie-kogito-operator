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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/flag"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	configMapPrefix     = "-custom-properties"
	createdByAnnonKey   = "createdBy"
	createdByAnnonValue = "Kogito CLI"
)

// createConfigMapFromFile creates the custom ConfigMap based in the configuration file given in the flags parameter.
// Does nothing if the config file path is empty
func createConfigMapFromFile(cli *client.Client, flags *flag.RuntimeFlags) (cmName string, err error) {
	if len(flags.ConfigFile) == 0 {
		return "", nil
	}
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-%s", flags.Name, configMapPrefix), Namespace: flags.Project},
	}
	fileContent, err := ioutil.ReadFile(flags.ConfigFile)
	if err != nil {
		return "", err
	}
	exists, err := kubernetes.ResourceC(cli).Fetch(cm)
	if err != nil {
		return "", err
	}
	cm.Data = map[string]string{
		services.ConfigMapApplicationPropertyKey: string(fileContent),
	}
	if cm.Annotations == nil {
		cm.Annotations = map[string]string{}
	}
	cm.Annotations[createdByAnnonKey] = createdByAnnonValue
	if exists {
		if err := kubernetes.ResourceC(cli).Update(cm); err != nil {
			return "", err
		}
	} else {
		if err := kubernetes.ResourceC(cli).Create(cm); err != nil {
			return "", err
		}
	}

	return cm.Name, nil
}
