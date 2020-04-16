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
	"strconv"
	"strings"
	"testing"

	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure/services"
	"github.com/kiegroup/kogito-cloud-operator/pkg/test"

	"github.com/stretchr/testify/assert"

	appsv1 "github.com/openshift/api/apps/v1"
	dockerv10 "github.com/openshift/api/image/docker10"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func contains(env []v1.EnvVar, key string) bool {
	for i := range env {
		envVar := env[i]
		if envVar.Name == key {
			return true
		}
	}
	return false
}

func createTestKogitoApp(runtime v1alpha1.RuntimeType) *v1alpha1.KogitoApp {
	return &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Runtime: runtime,
			Build:   &v1alpha1.KogitoAppBuildObject{},
		},
	}
}

func createTestKogitoInfra() *v1alpha1.KogitoInfra {
	return &v1alpha1.KogitoInfra{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoInfraSpec{
			InstallInfinispan: true,
		},
		Status: v1alpha1.KogitoInfraStatus{
			Infinispan: v1alpha1.InfinispanInstallStatus{
				InfraComponentInstallStatusType: v1alpha1.InfraComponentInstallStatusType{
					Service: "test",
				},
				CredentialSecret: "test",
			},
		},
	}
}

func createTestDeploymentConfig() *appsv1.DeploymentConfig {
	return &appsv1.DeploymentConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: appsv1.DeploymentConfigSpec{
			Template: &v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Env: []v1.EnvVar{},
						},
					},
				},
			},
		},
	}
}

func createTestService(kogitoInfra *v1alpha1.KogitoInfra, port int) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: kogitoInfra.Status.Infinispan.Service, Namespace: kogitoInfra.Namespace},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					TargetPort: intstr.FromInt(port),
				},
			},
		},
	}
}

func createTestSecret(kogitoInfra *v1alpha1.KogitoInfra, username, password string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: kogitoInfra.Status.Infinispan.CredentialSecret, Namespace: kogitoInfra.Namespace},
		Data: map[string][]byte{
			infrastructure.InfinispanSecretUsernameKey: []byte(username),
			infrastructure.InfinispanSecretPasswordKey: []byte(password),
		},
	}
}

func Test_deploymentConfigResource_NewWithValidDocker(t *testing.T) {
	uri := "https://github.com/kiegroup/kogito-examples"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Build: &v1alpha1.KogitoAppBuildObject{
				GitSource: v1alpha1.GitSource{
					URI:        uri,
					ContextDir: "process-quarkus-example",
				},
			},
		},
	}
	dockerImage := &dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				// notice the semicolon
				openshift.ImageLabelForExposeServices: "8080:http,8181;https",
				framework.LabelKeyOrgKie + "operator": "kogito",
				framework.LabelPrometheusPath:         "/metrics",
				framework.LabelPrometheusPort:         "8080",
				framework.LabelPrometheusScheme:       "http",
				framework.LabelPrometheusScrape:       "true",
			},
		},
	}
	bcS2I, _ := newBuildConfigS2I(kogitoApp)
	bcRuntime, _ := newBuildConfigRuntime(kogitoApp, &bcS2I)
	dc, err := newDeploymentConfig(kogitoApp, &bcRuntime, dockerImage, "abc")
	assert.Nil(t, err)
	assert.NotNil(t, dc)
	// we should have only one port. the 8181 is invalid.
	assert.Len(t, dc.Spec.Template.Spec.Containers[0].Ports, 1)
	assert.Equal(t, int32(8080), dc.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort)
	// this one where added by the docker image :)
	assert.Equal(t, "kogito", dc.Labels["operator"])
	// prometheus labels
	assert.Equal(t, "/metrics", dc.Spec.Template.Annotations[framework.LabelPrometheusPath])
	assert.Equal(t, "8080", dc.Spec.Template.Annotations[framework.LabelPrometheusPort])
	assert.Equal(t, "http", dc.Spec.Template.Annotations[framework.LabelPrometheusScheme])
	assert.Equal(t, "true", dc.Spec.Template.Annotations[framework.LabelPrometheusScrape])
}

func Test_deploymentConfigResource_NewWithInvalidDocker(t *testing.T) {
	uri := "https://github.com/kiegroup/kogito-examples"
	kogitoApp := &v1alpha1.KogitoApp{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Spec: v1alpha1.KogitoAppSpec{
			Build: &v1alpha1.KogitoAppBuildObject{
				GitSource: v1alpha1.GitSource{
					URI:        uri,
					ContextDir: "process-quarkus-example",
				},
			},
		},
	}
	bcS2I, _ := newBuildConfigS2I(kogitoApp)
	bcRuntime, _ := newBuildConfigRuntime(kogitoApp, &bcS2I)
	dc, err := newDeploymentConfig(kogitoApp, &bcRuntime, &dockerv10.DockerImage{}, "abc")
	assert.Nil(t, err)
	assert.NotNil(t, dc)
	assert.Len(t, dc.Spec.Selector, 1)
	assert.Len(t, dc.Spec.Template.Spec.Containers, 1)
	assert.Equal(t, bcRuntime.Spec.Output.To.Name, dc.Spec.Template.Spec.Containers[0].Image)
	assert.Equal(t, "test", dc.Labels[LabelKeyAppName])
	assert.Equal(t, "test", dc.Spec.Selector[LabelKeyAppName])
	assert.Equal(t, "test", dc.Spec.Template.Labels[LabelKeyAppName])
}

func Test_SetInfinispanEnvVars_QuarkusRuntime(t *testing.T) {
	kogitoApp := createTestKogitoApp(v1alpha1.QuarkusRuntimeType)
	kogitoInfra := createTestKogitoInfra()
	dc := createTestDeploymentConfig()
	service := createTestService(kogitoInfra, 11222)
	secret := createTestSecret(kogitoInfra, "test", "test")

	objs := []runtime.Object{service, secret}
	fakeClient := test.CreateFakeClient(objs, nil, []runtime.Object{})
	err := SetInfinispanEnvVars(fakeClient, kogitoInfra, kogitoApp, dc)
	assert.NoError(t, err)

	container := dc.Spec.Template.Spec.Containers[0]
	assert.Equal(t, 5, len(container.Env))
	assert.True(t, contains(container.Env, envVarInfinispanQuarkus[envVarInfinispanServerList]))
	assert.True(t, contains(container.Env, envVarInfinispanQuarkus[envVarInfinispanUseAuth]))
	assert.True(t, contains(container.Env, envVarInfinispanQuarkus[envVarInfinispanUser]))
	assert.True(t, contains(container.Env, envVarInfinispanQuarkus[envVarInfinispanPassword]))
	assert.True(t, contains(container.Env, envVarInfinispanQuarkus[envVarInfinispanSaslMechanism]))
}

func Test_IstioEnabled(t *testing.T) {
	uri := "https://github.com/kiegroup/kogito-examples"
	kogitoApp := createTestKogitoApp(v1alpha1.QuarkusRuntimeType)
	kogitoApp.Spec.EnableIstio = true
	kogitoApp.Spec.Build.GitSource = v1alpha1.GitSource{URI: uri}
	dockerImage := &dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				// notice the semicolon
				openshift.ImageLabelForExposeServices: "8080:http,8181;https",
				framework.LabelKeyOrgKie + "operator": "kogito",
				framework.LabelPrometheusPath:         "/metrics",
				framework.LabelPrometheusPort:         "8080",
				framework.LabelPrometheusScheme:       "http",
				framework.LabelPrometheusScrape:       "true",
			},
		},
	}
	bcS2I, _ := newBuildConfigS2I(kogitoApp)
	bcRuntime, _ := newBuildConfigRuntime(kogitoApp, &bcS2I)
	dc, err := newDeploymentConfig(kogitoApp, &bcRuntime, dockerImage, "abc")
	assert.NoError(t, err)
	assert.NotNil(t, dc)

	template := dc.Spec.Template
	for k, v := range template.Annotations {
		if strings.Contains(k, "istio") {
			annotationValue, err := strconv.ParseBool(v)
			assert.NoError(t, err)
			assert.True(t, annotationValue)
			return
		}
	}
	assert.Fail(t, "Should have istio annotation")
}

func Test_SetInfinispanEnvVars_SpringBootRuntime(t *testing.T) {
	kogitoApp := createTestKogitoApp(v1alpha1.SpringbootRuntimeType)
	kogitoInfra := createTestKogitoInfra()
	dc := createTestDeploymentConfig()
	service := createTestService(kogitoInfra, 11222)
	secret := createTestSecret(kogitoInfra, "test", "test")

	objs := []runtime.Object{service, secret}
	fakeClient := test.CreateFakeClient(objs, nil, []runtime.Object{})
	err := SetInfinispanEnvVars(fakeClient, kogitoInfra, kogitoApp, dc)
	assert.NoError(t, err)

	container := dc.Spec.Template.Spec.Containers[0]
	assert.Equal(t, 5, len(container.Env))
	assert.True(t, contains(container.Env, envVarInfinispanSpring[envVarInfinispanServerList]))
	assert.True(t, contains(container.Env, envVarInfinispanSpring[envVarInfinispanUseAuth]))
	assert.True(t, contains(container.Env, envVarInfinispanSpring[envVarInfinispanUser]))
	assert.True(t, contains(container.Env, envVarInfinispanSpring[envVarInfinispanPassword]))
	assert.True(t, contains(container.Env, envVarInfinispanSpring[envVarInfinispanSaslMechanism]))
}

func Test_deploymentConfigReplicas(t *testing.T) {
	uri := "https://github.com/kiegroup/kogito-examples"
	kogitoApp := createTestKogitoApp(v1alpha1.QuarkusRuntimeType)
	kogitoApp.Spec.Build.GitSource = v1alpha1.GitSource{URI: uri}
	dockerImage := &dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				// notice the semicolon
				openshift.ImageLabelForExposeServices: "8080:http,8181;https",
				framework.LabelKeyOrgKie + "operator": "kogito",
				framework.LabelPrometheusPath:         "/metrics",
				framework.LabelPrometheusPort:         "8080",
				framework.LabelPrometheusScheme:       "http",
				framework.LabelPrometheusScrape:       "true",
			},
		},
	}
	bcS2I, _ := newBuildConfigS2I(kogitoApp)
	bcRuntime, _ := newBuildConfigRuntime(kogitoApp, &bcS2I)

	{
		dc, err := newDeploymentConfig(kogitoApp, &bcRuntime, dockerImage, "abc")
		assert.NoError(t, err)
		assert.NotNil(t, dc)
		assert.Equal(t, defaultReplicas, dc.Spec.Replicas)
	}

	{
		zeroReplica := int32(0)
		kogitoApp.Spec.Replicas = &zeroReplica
		dc, err := newDeploymentConfig(kogitoApp, &bcRuntime, dockerImage, "abc")
		assert.NoError(t, err)
		assert.NotNil(t, dc)
		assert.Equal(t, int32(0), dc.Spec.Replicas)
	}
}

func Test_applicationProperties(t *testing.T) {
	uri := "https://github.com/kiegroup/kogito-examples"
	kogitoApp := createTestKogitoApp(v1alpha1.QuarkusRuntimeType)
	kogitoApp.Spec.Build.GitSource = v1alpha1.GitSource{URI: uri}
	dockerImage := &dockerv10.DockerImage{
		Config: &dockerv10.DockerConfig{
			Labels: map[string]string{
				// notice the semicolon
				openshift.ImageLabelForExposeServices: "8080:http,8181;https",
				framework.LabelKeyOrgKie + "operator": "kogito",
				framework.LabelPrometheusPath:         "/metrics",
				framework.LabelPrometheusPort:         "8080",
				framework.LabelPrometheusScheme:       "http",
				framework.LabelPrometheusScrape:       "true",
			},
		},
	}
	bcS2I, _ := newBuildConfigS2I(kogitoApp)
	bcRuntime, _ := newBuildConfigRuntime(kogitoApp, &bcS2I)

	contentHash := "abc24680"

	dc, err := newDeploymentConfig(kogitoApp, &bcRuntime, dockerImage, contentHash)
	assert.NoError(t, err)
	assert.NotNil(t, dc)
	assert.Equal(t, contentHash, dc.Spec.Template.ObjectMeta.Annotations[services.AppPropContentHashKey])

	foundVolume := false
	for _, volume := range dc.Spec.Template.Spec.Volumes {
		if volume.Name == services.AppPropVolumeName {
			foundVolume = true
			break
		}
	}
	assert.True(t, foundVolume)

	foundVolumeMount := false
	for _, volumeMount := range dc.Spec.Template.Spec.Containers[0].VolumeMounts {
		if volumeMount.Name == services.AppPropVolumeName {
			foundVolumeMount = true
			break
		}
	}
	assert.True(t, foundVolumeMount)
}

func Test_namespaceEnvVarCorrectSet(t *testing.T) {
	kogitoApp := createTestKogitoApp(v1alpha1.QuarkusRuntimeType)
	kogitoApp.Spec.Build.GitSource.URI = "http://example.com"
	bcS2I, _ := newBuildConfigS2I(kogitoApp)
	bcRuntime, _ := newBuildConfigRuntime(kogitoApp, &bcS2I)
	dc, err := newDeploymentConfig(kogitoApp, &bcRuntime, nil, "")
	assert.NoError(t, err)
	assert.True(t, contains(dc.Spec.Template.Spec.Containers[0].Env, envVarNamespace))
	assert.Equal(t, kogitoApp.Namespace, framework.GetEnvVarFromContainer(envVarNamespace, dc.Spec.Template.Spec.Containers[0]))
}
