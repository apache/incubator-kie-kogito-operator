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
	"encoding/base64"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"
	appsv1 "github.com/openshift/api/apps/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	"github.com/stretchr/testify/assert"
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"
)

func Test_addMetadataFromDockerImage_MultiLabel(t *testing.T) {
	dockerImage := &dockerv10.DockerImage{Config: &dockerv10.DockerConfig{
		Labels: map[string]string{
			LabelKeyOrgKie + "layer1":        "value",
			LabelKeyOrgKie + "layer1/layer2": "value",
		},
	}}
	dcWithOutLabels := &appsv1.DeploymentConfig{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Template: &v12.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{},
				},
			},
			Selector: map[string]string{},
		},
	}

	added := MergeImageMetadataWithDeploymentConfig(dcWithOutLabels, dockerImage)
	assert.True(t, added)
	assert.Contains(t, dcWithOutLabels.Labels, "layer1/layer2")
	assert.Contains(t, dcWithOutLabels.Labels, "layer1")
	assert.Len(t, dcWithOutLabels.Labels, 2)
}

func Test_addMetadataFromDockerImage(t *testing.T) {
	dockerImage := &dockerv10.DockerImage{Config: &dockerv10.DockerConfig{
		Labels: map[string]string{
			LabelKeyOrgKie + "myprocess":               "process",
			LabelKeyOrgKie + "myotherlabel":            "value",
			LabelKeyOrgKie + "persistence/anotherfile": "process.proto",
			LabelKeyPrometheus + "/path":               "/metrics",
		},
	}}
	dcWithLabels := &appsv1.DeploymentConfig{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{"myprocess": "process", "myotherlabel": "value"},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Template: &v12.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels:      map[string]string{"myprocess": "process", "myotherlabel": "value"},
					Annotations: map[string]string{LabelKeyPrometheus + "/path": "/metrics"},
				},
			},
			Selector: map[string]string{"myprocess": "process", "myotherlabel": "value"},
		},
	}
	dcWithPartialLabels := &appsv1.DeploymentConfig{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{"myprocess": "process", "myotherlabel": "value"},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Template: &v12.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels:      map[string]string{},
					Annotations: map[string]string{LabelKeyPrometheus + "/path": "/metrics"},
				},
			},
			Selector: map[string]string{"myprocess": "process", "myotherlabel": "value"},
		},
	}
	dcWithOutLabels := &appsv1.DeploymentConfig{
		ObjectMeta: v1.ObjectMeta{
			Labels: map[string]string{},
		},
		Spec: appsv1.DeploymentConfigSpec{
			Template: &v12.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{},
				},
			},
			Selector: map[string]string{},
		},
	}

	type args struct {
		dc          *appsv1.DeploymentConfig
		dockerImage *dockerv10.DockerImage
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"When the DC already have the given labels",
			args{
				dc:          dcWithLabels,
				dockerImage: dockerImage,
			}, false,
		},
		{
			"When the DC doesn't have the given labels",
			args{
				dc:          dcWithOutLabels,
				dockerImage: dockerImage,
			}, true,
		},
		{
			"When the DC is missing some labels",
			args{
				dc:          dcWithPartialLabels,
				dockerImage: dockerImage,
			}, true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MergeImageMetadataWithDeploymentConfig(tt.args.dc, tt.args.dockerImage); got != tt.want {
				t.Errorf("MergeImageMetadataWithDeploymentConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_discoverPortsAndProbesFromImage(t *testing.T) {
	dockerImage := &dockerv10.DockerImage{Config: &dockerv10.DockerConfig{
		Labels: map[string]string{
			openshift.ImageLabelForExposeServices: "8080:http",
		},
	}}
	dc := &appsv1.DeploymentConfig{
		Spec: appsv1.DeploymentConfigSpec{
			Template: &v12.PodTemplateSpec{
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						{
							Name: "service",
						},
					},
				},
			},
		},
	}

	DiscoverPortsAndProbesFromImage(dc, dockerImage)
	assert.Len(t, dc.Spec.Template.Spec.Containers[0].Ports, 1)
	assert.Equal(t, dc.Spec.Template.Spec.Containers[0].LivenessProbe.TCPSocket.Port.IntVal, int32(8080))
	assert.Equal(t, dc.Spec.Template.Spec.Containers[0].ReadinessProbe.TCPSocket.Port.IntVal, int32(8080))
}

func Test_discoverPortsAndProbesFromImageNoPorts(t *testing.T) {
	dockerImage := &dockerv10.DockerImage{Config: &dockerv10.DockerConfig{
		Labels: map[string]string{},
	}}
	dc := &appsv1.DeploymentConfig{
		Spec: appsv1.DeploymentConfigSpec{
			Template: &v12.PodTemplateSpec{
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						{
							Name: "service",
						},
					},
				},
			},
		},
	}

	DiscoverPortsAndProbesFromImage(dc, dockerImage)
	assert.Len(t, dc.Spec.Template.Spec.Containers[0].Ports, 0)
	assert.Nil(t, dc.Spec.Template.Spec.Containers[0].LivenessProbe)
	assert.Nil(t, dc.Spec.Template.Spec.Containers[0].ReadinessProbe)
}

func TestExtractProtoBufFilesFromDockerImage(t *testing.T) {
	base64ProtoFile := test.HelperLoadString(t, "base64-onboarding-proto")
	protoFile, _ := base64.StdEncoding.DecodeString(base64ProtoFile)
	dockerImage := &dockerv10.DockerImage{Config: &dockerv10.DockerConfig{
		Labels: map[string]string{
			LabelKeyOrgKie + "myprocess":                            "process",
			LabelKeyOrgKieProtoBuf + "/onboarding.onboarding.proto": base64ProtoFile,
		},
	}}

	type args struct {
		prefixKey   string
		dockerImage *dockerv10.DockerImage
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			"With required Label",
			args{"onboarding-service", dockerImage},
			map[string]string{"onboarding-service-onboarding.onboarding.proto": string(protoFile)},
		},
		{
			"Without Label",
			args{"", &dockerv10.DockerImage{}},
			map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractProtoBufFilesFromDockerImage(tt.args.prefixKey, tt.args.dockerImage); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractProtoBufFilesFromDockerImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
