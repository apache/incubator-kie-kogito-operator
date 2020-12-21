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
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_getProtoBufConfigMapsForAllRuntimeServices(t *testing.T) {
	cm1 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs1",
			Labels: map[string]string{
				ConfigMapProtoBufEnabledLabelKey: "true",
				framework.LabelAppKey:            "my-domain-protobufs1",
			},
		},
		Data: map[string]string{"mydomain.proto": "This is a protobuf file"},
	}
	cm2 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs2",
			Labels: map[string]string{
				ConfigMapProtoBufEnabledLabelKey: "true",
				framework.LabelAppKey:            "my-domain-protobufs2",
			},
		},
		Data: map[string]string{"mydomain2.proto": "This is a protobuf file"},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(cm1, cm2).Build()
	cms, err := getProtoBufConfigMapsForAllRuntimeServices(t.Name(), cli)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(cms.Items))
}

func Test_getProtoBufConfigMapsForSpecificRuntimeService(t *testing.T) {
	cm1 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs1",
			Labels: map[string]string{
				ConfigMapProtoBufEnabledLabelKey: "true",
				framework.LabelAppKey:            "my-domain-protobufs1",
			},
		},
		Data: map[string]string{"mydomain.proto": "This is a protobuf file"},
	}
	cm2 := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: t.Name(),
			Name:      "my-domain-protobufs2",
			Labels: map[string]string{
				ConfigMapProtoBufEnabledLabelKey: "true",
				framework.LabelAppKey:            "my-domain-protobufs2",
			},
		},
		Data: map[string]string{"mydomain2.proto": "This is a protobuf file"},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(cm1, cm2).Build()
	cms, err := getProtoBufConfigMapsForSpecificRuntimeService(cli, "my-domain-protobufs1", t.Name())
	assert.NoError(t, err)
	assert.Equal(t, 1, len(cms.Items))
	assert.Equal(t, "my-domain-protobufs1", cms.Items[0].Name)
}

func TestFetchKogitoRuntimeService_InstanceFound(t *testing.T) {
	ns := t.Name()
	name := "kogito-runtime"
	kogitoRuntime := &v1beta1.KogitoRuntime{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}
	cli := test.NewFakeClientBuilder().AddK8sObjects(kogitoRuntime).Build()
	instance, err := FetchKogitoRuntimeService(cli, name, ns)
	assert.NoError(t, err)
	assert.NotNil(t, instance)
}

func TestFetchKogitoRuntimeService_InstanceNotFound(t *testing.T) {
	ns := t.Name()
	name := "kogito-runtime"
	cli := test.NewFakeClientBuilder().Build()
	instance, err := FetchKogitoRuntimeService(cli, name, ns)
	assert.NoError(t, err)
	assert.Nil(t, instance)
}
