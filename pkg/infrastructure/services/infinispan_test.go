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

package services

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func Test_SetInfinispanVariables(t *testing.T) {
	type args struct {
		runtime              v1alpha1.RuntimeType
		connectionProperties v1alpha1.InfinispanConnectionProperties
		secret               *corev1.Secret
	}
	tests := []struct {
		name            string
		args            args
		expectedEnvVars []corev1.EnvVar
	}{
		{
			"EmptyProperties",
			args{
				connectionProperties: v1alpha1.InfinispanConnectionProperties{},
				secret:               nil,
			},
			[]corev1.EnvVar{
				{Name: infinispanEnvKeyUseAuth, Value: "false"},
				{Name: envVarInfinispanQuarkus[envVarInfinispanUseAuth], Value: "false"},
			},
		},
		{
			"Uri",
			args{
				connectionProperties: v1alpha1.InfinispanConnectionProperties{
					URI: "custom-uri:123",
				},
				secret: nil,
			},
			[]corev1.EnvVar{
				{Name: infinispanEnvKeyUseAuth, Value: "false"},
				{Name: envVarInfinispanQuarkus[envVarInfinispanUseAuth], Value: "false"},
				{Name: envVarInfinispanQuarkus[envVarInfinispanServerList], Value: "custom-uri:123"},
				{Name: infinispanEnvKeyServerList, Value: "custom-uri:123"},
			},
		},
		{
			"AuthRealm",
			args{
				connectionProperties: v1alpha1.InfinispanConnectionProperties{
					AuthRealm: "custom-realm",
				},
				secret: nil,
			},
			[]corev1.EnvVar{
				{Name: infinispanEnvKeyUseAuth, Value: "false"},
				{Name: envVarInfinispanQuarkus[envVarInfinispanUseAuth], Value: "false"},
				{Name: infinispanEnvKeyAuthRealm, Value: "custom-realm"},
				{Name: envVarInfinispanQuarkus[envVarInfinispanAuthRealm], Value: "custom-realm"},
			},
		},
		{
			"SaslMechanism",
			args{
				connectionProperties: v1alpha1.InfinispanConnectionProperties{
					SaslMechanism: "DIGEST-MD5",
				},
				secret: nil,
			},
			[]corev1.EnvVar{
				{Name: infinispanEnvKeyUseAuth, Value: "false"},
				{Name: envVarInfinispanQuarkus[envVarInfinispanUseAuth], Value: "false"},
				{Name: infinispanEnvKeySasl, Value: "DIGEST-MD5"},
				{Name: envVarInfinispanQuarkus[envVarInfinispanSaslMechanism], Value: "DIGEST-MD5"},
			},
		},
		{
			"CustomSecretDefaultKeys",
			args{
				connectionProperties: v1alpha1.InfinispanConnectionProperties{
					Credentials: v1alpha1.SecretCredentialsType{
						SecretName: "custom-secret",
					},
				},
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "custom-secret",
						Namespace: t.Name(),
					},
					StringData: map[string]string{
						infrastructure.InfinispanSecretUsernameKey: "user1",
						infrastructure.InfinispanSecretPasswordKey: "pass1",
					},
				},
			},
			[]corev1.EnvVar{
				{
					Name: infinispanEnvKeyUsername,
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  infrastructure.InfinispanSecretUsernameKey,
						},
					},
				},
				{
					Name: envVarInfinispanQuarkus[envVarInfinispanUser],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  infrastructure.InfinispanSecretUsernameKey,
						},
					},
				},
				{
					Name: infinispanEnvKeyPassword,
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  infrastructure.InfinispanSecretPasswordKey,
						},
					},
				},
				{
					Name: envVarInfinispanQuarkus[envVarInfinispanPassword],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  infrastructure.InfinispanSecretPasswordKey,
						},
					},
				},
				{Name: infinispanEnvKeyCredSecret, Value: "custom-secret"},
				{Name: infinispanEnvKeyUseAuth, Value: "true"},
				{Name: envVarInfinispanQuarkus[envVarInfinispanUseAuth], Value: "true"},
				{Name: infinispanEnvKeySasl, Value: string(v1alpha1.SASLPlain)},
				{Name: envVarInfinispanQuarkus[envVarInfinispanSaslMechanism], Value: string(v1alpha1.SASLPlain)},
			},
		},
		{
			"CustomSecretCustomKeys",
			args{
				connectionProperties: v1alpha1.InfinispanConnectionProperties{
					Credentials: v1alpha1.SecretCredentialsType{
						SecretName:  "custom-secret",
						UsernameKey: "custom-username",
						PasswordKey: "custom-password",
					},
				},
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "custom-secret",
						Namespace: t.Name(),
					},
					StringData: map[string]string{
						"custom-username": "user1",
						"custom-password": "pass1",
					},
				},
			},
			[]corev1.EnvVar{
				{
					Name: infinispanEnvKeyUsername,
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  "custom-username",
						},
					},
				},
				{
					Name: envVarInfinispanQuarkus[envVarInfinispanUser],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  "custom-username",
						},
					},
				},
				{
					Name: infinispanEnvKeyPassword,
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  "custom-password",
						},
					},
				},
				{
					Name: envVarInfinispanQuarkus[envVarInfinispanPassword],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  "custom-password",
						},
					},
				},
				{Name: infinispanEnvKeyCredSecret, Value: "custom-secret"},
				{Name: infinispanEnvKeyUseAuth, Value: "true"},
				{Name: envVarInfinispanQuarkus[envVarInfinispanUseAuth], Value: "true"},
				{Name: infinispanEnvKeySasl, Value: string(v1alpha1.SASLPlain)},
				{Name: envVarInfinispanQuarkus[envVarInfinispanSaslMechanism], Value: string(v1alpha1.SASLPlain)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := &corev1.Container{}

			setInfinispanVariables(tt.args.runtime, tt.args.connectionProperties, tt.args.secret, container)

			assert.Equal(t, len(tt.expectedEnvVars), len(container.Env))
			for _, expectedEnvVar := range tt.expectedEnvVars {
				assert.Contains(t, container.Env, expectedEnvVar)
			}
		})
	}
}
