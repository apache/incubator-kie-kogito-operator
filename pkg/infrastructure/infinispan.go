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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"k8s.io/api/apps/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	infinispanServerGroup = "infinispan.org"
	operatorName          = "infinispan-operator"
)

// IsInfinispanOperatorAvailable verify if Infinispan Operator is running in the given namespace and the CRD is available
func IsInfinispanOperatorAvailable(cli *client.Client, namespace string) (bool, error) {
	log.Debugf("Checking if Infinispan Operator is available in the namespace %s", namespace)
	// first check for CRD
	if cli.HasServerGroup(infinispanServerGroup) {
		log.Debugf("Infinispan CRDs available. Checking if Infinispan Operator is deployed in the namespace %s", namespace)
		// then check if there's a Infinispan Operator deployed
		deployment := &v1.Deployment{ObjectMeta: v12.ObjectMeta{Namespace: namespace, Name: operatorName}}
		exists := false
		var err error
		if exists, err = kubernetes.ResourceC(cli).Fetch(deployment); err != nil {
			return false, nil
		}
		if exists {
			log.Debugf("Infinispan Operator is available in the namespace %s", namespace)
			return true, nil
		}
	} else {
		log.Debug("Couldn't find Infinispan CRDs")
	}
	log.Debugf("Looks like Infinispan Operator is not available in the namespace %s", namespace)
	return false, nil
}
