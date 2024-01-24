// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package app

import (
	"github.com/apache/incubator-kie-kogito-operator/apis"
	"github.com/apache/incubator-kie-kogito-operator/apis/app/v1beta1"
	"github.com/apache/incubator-kie-kogito-operator/core/client/kubernetes"
	"github.com/apache/incubator-kie-kogito-operator/core/manager"
	"github.com/apache/incubator-kie-kogito-operator/core/operator"
	"k8s.io/apimachinery/pkg/types"
)

type kogitoBuildHandler struct {
	operator.Context
}

// NewKogitoBuildHandler ...
func NewKogitoBuildHandler(context operator.Context) manager.KogitoBuildHandler {
	return &kogitoBuildHandler{
		context,
	}
}

func (k *kogitoBuildHandler) FetchKogitoBuildInstance(key types.NamespacedName) (api.KogitoBuildInterface, error) {
	instance := &v1beta1.KogitoBuild{}
	if exists, err := kubernetes.ResourceC(k.Client).FetchWithKey(key, instance); err != nil {
		return nil, err
	} else if !exists {
		return nil, nil
	}
	return instance, nil
}

func (k *kogitoBuildHandler) CreateBuild() api.BuildsInterface {
	return &v1beta1.Builds{}
}
