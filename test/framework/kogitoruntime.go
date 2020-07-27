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

package framework

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	bddtypes "github.com/kiegroup/kogito-cloud-operator/test/types"
)

// DeployRuntimeService deploy a Kogito service
func DeployRuntimeService(namespace string, installerType InstallerType, serviceHolder *bddtypes.KogitoServiceHolder) error {
	return DeployService(serviceHolder, installerType)
}

// SetKogitoRuntimeReplicas sets the number of replicas for a Kogito application
func SetKogitoRuntimeReplicas(namespace, name string, nbPods int) error {
	GetLogger(namespace).Infof("Set Kogito application %s replica number to %d", name, nbPods)
	kogitoRuntime, err := getKogitoRuntime(namespace, name)
	if err != nil {
		return err
	} else if kogitoRuntime == nil {
		return fmt.Errorf("No KogitoRuntime found with name %s in namespace %s", name, namespace)
	}
	replicas := int32(nbPods)
	kogitoRuntime.Spec.KogitoServiceSpec.Replicas = &replicas
	return kubernetes.ResourceC(kubeClient).Update(kogitoRuntime)
}

// GetKogitoRuntimeStub Get basic KogitoRuntime stub with all needed fields initialized
func GetKogitoRuntimeStub(namespace, runtimeType, name, imageTag string) *v1alpha1.KogitoRuntime {
	image := framework.ConvertImageTagToImage(imageTag)
	replicas := int32(1)
	kogitoRuntime := &v1alpha1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Status: v1alpha1.KogitoRuntimeStatus{
			KogitoServiceStatus: v1alpha1.KogitoServiceStatus{
				ConditionsMeta: v1alpha1.ConditionsMeta{Conditions: []v1alpha1.Condition{}},
			},
		},
		Spec: v1alpha1.KogitoRuntimeSpec{
			Runtime: v1alpha1.RuntimeType(runtimeType),
			KogitoServiceSpec: v1alpha1.KogitoServiceSpec{
				Image: image,
				// Use insecure registry flag in tests
				InsecureImageRegistry: true,
				Replicas:              &replicas,
			},
		},
	}

	return kogitoRuntime
}

func getKogitoRuntime(namespace, name string) (*v1alpha1.KogitoRuntime, error) {
	kogitoRuntime := &v1alpha1.KogitoRuntime{}
	if exists, err := kubernetes.ResourceC(kubeClient).FetchWithKey(types.NamespacedName{Name: name, Namespace: namespace}, kogitoRuntime); err != nil && !errors.IsNotFound(err) {
		return nil, fmt.Errorf("Error while trying to look for KogitoRuntime %s: %v ", name, err)
	} else if errors.IsNotFound(err) || !exists {
		return nil, nil
	}
	return kogitoRuntime, nil
}
