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

package resource

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func Test_DeploymentSetShouldHaveProtoBufEnvVars(t *testing.T) {
	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Namespace: t.Name(), Name: "kogito-proto", Labels: map[string]string{infrastructure.ConfigMapProtoBufEnabledLabelKey: "true"}},
		Data:       map[string]string{"myproto.proto": "import whatever; do whatever;"},
	}
	dataIndex := &v1alpha1.KogitoDataIndex{
		ObjectMeta: metav1.ObjectMeta{Namespace: t.Name(), Name: "data-index"},
		Spec: v1alpha1.KogitoDataIndexSpec{
			Replicas: 1,
			InfinispanMeta: v1alpha1.InfinispanMeta{
				InfinispanProperties: v1alpha1.InfinispanConnectionProperties{
					UseKogitoInfra: true,
				},
			},
		},
	}
	cli := test.CreateFakeClient([]runtime.Object{cm, dataIndex}, nil, nil)
	is := newImage(dataIndex)
	deployment, err := newDeployment(dataIndex, nil, "", cli, is)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)

	imgKey, _ := framework.ResolveImageStreamTriggerAnnotation("", "")
	assert.NotNil(t, deployment.Annotations[imgKey])
	assert.Contains(t, deployment.Annotations[imgKey], "data-index")

	assert.Equal(t, "true", framework.GetEnvVarFromContainer(protoBufKeyWatch, &deployment.Spec.Template.Spec.Containers[0]))
	assert.Equal(t, defaultProtobufMountPath, framework.GetEnvVarFromContainer(protoBufKeyFolder, &deployment.Spec.Template.Spec.Containers[0]))

	// since we don't have a CM anymore, the deployment should not mount the folder, thus the watch should be set to false
	err = kubernetes.ResourceC(cli).Delete(cm)
	assert.NoError(t, err)

	deployment, err = newDeployment(dataIndex, nil, "", cli, nil)
	assert.NoError(t, err)
	assert.NotNil(t, deployment)

	assert.Equal(t, "false", framework.GetEnvVarFromContainer(protoBufKeyWatch, &deployment.Spec.Template.Spec.Containers[0]))
	assert.Equal(t, "", framework.GetEnvVarFromContainer(protoBufKeyFolder, &deployment.Spec.Template.Spec.Containers[0]))
}
