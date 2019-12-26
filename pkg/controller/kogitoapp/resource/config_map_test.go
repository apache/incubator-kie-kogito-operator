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

package resource

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_newProtoBufConfigMap(t *testing.T) {
	kogitoApp := &v1alpha1.KogitoApp{ObjectMeta: metav1.ObjectMeta{Name: "myapp", Namespace: t.Name()}}
	cm := newProtoBufConfigMap(kogitoApp)
	assert.NotNil(t, cm)
	assert.Equal(t, GenerateProtoBufConfigMapName(kogitoApp), cm.Name)
}

func TestCheckProtoBufConfigMapIntegrity(t *testing.T) {
	type args struct {
		configMap *v1.ConfigMap
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"HasIntegrity",
			args{configMap: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s", t.Name(), protobufConfigMapSuffix),
					Annotations: map[string]string{
						"org.kie.kogito.protobuf.hash/file1": "098f6bcd4621d373cade4e832627b4f6",
					},
				},
				Data: map[string]string{"file1.proto": "test"},
			}},
			true,
		},
		{
			"DoesntHaveIntegrity",
			args{configMap: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name: fmt.Sprintf("%s-%s", t.Name(), protobufConfigMapSuffix),
					Annotations: map[string]string{
						"org.kie.kogito.protobuf.hash/file1": "098f6bcd4621d373cade4e832627b4f6",
						"org.kie.kogito.protobuf.hash/file2": "098f6bcd4621d373cade4e832627b4f6",
					},
				},
				// missing data for file2
				Data: map[string]string{"file1.proto": "test"},
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckProtoBufConfigMapIntegrity(tt.args.configMap); got != tt.want {
				t.Errorf("CheckProtoBufConfigMapIntegrity() = %v, want %v", got, tt.want)
			}
		})
	}
}
