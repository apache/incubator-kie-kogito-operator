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

package client

import (
	"context"
	"github.com/apache/incubator-kie-kogito-operator/apis/app/v1beta1"
	"github.com/apache/incubator-kie-kogito-operator/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

// TestBasicCheckClientSetGenerated simple verification for the generated code
func TestBasicCheckClientSetGenerated(t *testing.T) {
	kogitoClient := fake.NewSimpleClientset(&v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mykogitosvc",
			Namespace: "kogito",
		},
		Spec: v1beta1.KogitoRuntimeSpec{
			KogitoServiceSpec: v1beta1.KogitoServiceSpec{
				Image: "quay.io/kiegroup/process-example:latest",
			},
		},
	})
	kogitoSvc, err :=
		kogitoClient.AppV1beta1().KogitoRuntimes("kogito").Get(context.TODO(), "mykogitosvc", metav1.GetOptions{})
	assert.NoError(t, err)
	assert.NotNil(t, kogitoSvc)
	assert.Equal(t, "mykogitosvc", kogitoSvc.GetName())
}
