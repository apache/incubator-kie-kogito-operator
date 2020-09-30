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

package infrastructure

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/apimachinery/pkg/types"
)

// FetchKogitoInfraInstance load infra instance
func FetchKogitoInfraInstance(client *client.Client, name string, namespace string) (*v1alpha1.KogitoInfra, error) {
	log.Debugf("going to fetch deployed kogito infra instance %s", name)
	instance := &v1alpha1.KogitoInfra{}
	if exists, resultErr := kubernetes.ResourceC(client).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, instance); resultErr != nil {
		log.Errorf("Error occurs while fetching deployed kogito infra instance %s", name)
		return nil, resultErr
	} else if !exists {
		return nil, fmt.Errorf("kogito Infra resource with name %s not found in namespace %s", name, namespace)
	} else {
		log.Debugf("Successfully fetch deployed kogito infra reference %s", name)
		return instance, nil
	}
}
