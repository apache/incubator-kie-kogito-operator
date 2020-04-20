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

package shared

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/message"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"k8s.io/client-go/tools/clientcmd"
	"strings"
)

const (
	kubeConfigContextSep = "/"
)

// GetCurrentNamespaceFromKubeConfig gets the current namespace from the .kubeconfig file registered in the local machine
func GetCurrentNamespaceFromKubeConfig() string {
	file := client.GetKubeConfigFile()
	config := clientcmd.GetConfigFromFileOrDie(file)
	if len(config.CurrentContext) == 0 || config.Contexts[config.CurrentContext] == nil {
		context.GetDefaultLogger().Warnf(message.KubeConfigNoContext, file)
		return ""
	}
	return config.Contexts[config.CurrentContext].Namespace
}

// SetCurrentNamespaceToKubeConfig sets the current namespace to the .kubeconfig file
func SetCurrentNamespaceToKubeConfig(namespace string) error {
	file := client.GetKubeConfigFile()
	config := clientcmd.GetConfigFromFileOrDie(file)

	if len(config.CurrentContext) == 0 {
		return fmt.Errorf(message.KubeConfigNoContext, file)
	}

	currentContext := strings.Split(config.CurrentContext, kubeConfigContextSep)
	currentContext[0] = namespace
	newContext := strings.Join(currentContext, kubeConfigContextSep)
	if _, exists := config.Contexts[newContext]; !exists {
		newContextRef := config.Contexts[config.CurrentContext].DeepCopy()
		newContextRef.Namespace = namespace
		config.Contexts[newContext] = newContextRef
	}
	config.CurrentContext = newContext
	if err := clientcmd.WriteToFile(*config, file); err != nil {
		return fmt.Errorf(message.KubeConfigErrorWriteFile, file, err)
	}
	context.GetDefaultLogger().Debugf("Successfully set current namespace to %s", namespace)
	return nil
}
