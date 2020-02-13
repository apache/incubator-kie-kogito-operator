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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	istioSidecarInjectAnnotation = "sidecar.istio.io/inject"
)

// AddIstioInjectSidecarAnnotation adds the annotation to be read by the Istio operator to setup sidecars in the given Pod
func AddIstioInjectSidecarAnnotation(objectMeta *metav1.ObjectMeta) {
	if objectMeta == nil {
		return
	}
	if objectMeta.Annotations == nil {
		objectMeta.Annotations = map[string]string{}
	}
	objectMeta.Annotations[istioSidecarInjectAnnotation] = "true"
}
