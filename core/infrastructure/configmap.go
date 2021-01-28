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
	"github.com/RHsyseng/operator-utils/pkg/resource"
	"github.com/kiegroup/kogito-cloud-operator/core/framework"
	"github.com/kiegroup/kogito-cloud-operator/core/logger"
	"github.com/kiegroup/kogito-cloud-operator/core/record"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

// ConfigMapHandler ...
type ConfigMapHandler interface {
	TakeConfigMapOwnership(key types.NamespacedName, owner resource.KubernetesResource) (updated bool, err error)
}

type configMapHandler struct {
	client   *client.Client
	log      logger.Logger
	scheme   *runtime.Scheme
	recorder record.EventRecorder
}

// NewConfigMapHandler ...
func NewConfigMapHandler(client *client.Client, log logger.Logger, scheme *runtime.Scheme, recorder record.EventRecorder) ConfigMapHandler {
	return &configMapHandler{
		client:   client,
		log:      log,
		scheme:   scheme,
		recorder: recorder,
	}
}

func (s *configMapHandler) TakeConfigMapOwnership(key types.NamespacedName, owner resource.KubernetesResource) (updated bool, err error) {
	cm := &corev1.ConfigMap{}
	exists, err := kubernetes.ResourceC(s.client).FetchWithKey(key, cm)
	if err != nil {
		return
	}
	if !exists {
		s.recorder.Eventf(s.client, owner, corev1.EventTypeWarning, "NotExists", "ConfigMap %s does not exist", key.Name)
		return
	}
	if framework.IsOwner(cm, owner) {
		return
	}
	if err = framework.AddOwnerReference(owner, s.scheme, cm); err != nil {
		return
	}
	if err = kubernetes.ResourceC(s.client).Update(cm); err != nil {
		return
	}
	return true, nil
}
