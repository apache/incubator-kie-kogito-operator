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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	maistrav1 "github.com/maistra/istio-operator/pkg/apis/maistra/v1"
)

// DeployServiceMeshInstance deploys an instance of Service Mesh
func DeployServiceMeshInstance(namespace string) error {
	GetLogger(namespace).Info("Creating Service Mesh CR to spin up instance.")

	serviceMeshControlPlaneCR := &maistrav1.ServiceMeshControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "control-pane",
			Namespace: namespace,
		},
		Spec: maistrav1.ControlPlaneSpec{},
	}
	if err := kubernetes.ResourceC(kubeClient).Create(serviceMeshControlPlaneCR); err != nil {
		return fmt.Errorf("Error while creating Service Mesh Control Plane CR: %v ", err)
	}

	return nil
}
