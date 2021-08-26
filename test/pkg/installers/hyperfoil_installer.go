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

package installers

import (
	"github.com/kiegroup/kogito-operator/core/client/kubernetes"
	"github.com/kiegroup/kogito-operator/test/pkg/framework"

	hyperfoilv1alpha2 "github.com/Hyperfoil/hyperfoil-operator/pkg/apis/hyperfoil/v1alpha2"
)

var (
	// hyperfoilOlmNamespacedInstaller installs Hyperfoil into the namespace using OLM
	hyperfoilOlmNamespacedInstaller = OlmNamespacedServiceInstaller{
		SubscriptionName:                  hyperfoilOperatorSubscriptionName,
		Channel:                           hyperfoilOperatorSubscriptionChannel,
		Catalog:                           framework.GetCommunityCatalog,
		InstallationTimeoutInMinutes:      hyperfoilOperatorTimeoutInMin,
		GetAllNamespacedOlmCrsInNamespace: getHyperfoilCrsInNamespace,
	}

	hyperfoilOperatorSubscriptionName    = "hyperfoil-bundle"
	hyperfoilOperatorSubscriptionChannel = "alpha"
	hyperfoilOperatorTimeoutInMin        = 10
)

// GetHyperfoilInstaller returns Hyperfoil installer
func GetHyperfoilInstaller() ServiceInstaller {
	return &hyperfoilOlmNamespacedInstaller
}

func getHyperfoilCrsInNamespace(namespace string) ([]kubernetes.ResourceObject, error) {
	crs := []kubernetes.ResourceObject{}

	hyperfoils := &hyperfoilv1alpha2.HyperfoilList{}
	if err := framework.GetObjectsInNamespace(namespace, hyperfoils); err != nil {
		return nil, err
	}
	for i := range hyperfoils.Items {
		crs = append(crs, &hyperfoils.Items[i])
	}

	return crs, nil
}
