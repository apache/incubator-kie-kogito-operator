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
		appProps             map[string]string
	}
	tests := []struct {
		name             string
		args             args
		expectedEnvVars  []corev1.EnvVar
		expectedAppProps map[string]string
	}{
		{
			"EmptyProperties, Quarkus",
			args{
				connectionProperties: v1alpha1.InfinispanConnectionProperties{},
				secret:               nil,
				appProps:             map[string]string{},
			},
			[]corev1.EnvVar{},
			map[string]string{
				PropertiesInfinispanQuarkus[AppPropInfinispanUseAuth]: "false",
			},
		},
		{
			"Uri, Quarkus",
			args{
				connectionProperties: v1alpha1.InfinispanConnectionProperties{
					URI: "custom-uri:123",
				},
				secret:   nil,
				appProps: map[string]string{},
			},
			[]corev1.EnvVar{
				{Name: infinispanEnvKeyServerList, Value: "custom-uri:123"},
			},
			map[string]string{
				PropertiesInfinispanQuarkus[AppPropInfinispanUseAuth]:    "false",
				PropertiesInfinispanQuarkus[AppPropInfinispanServerList]: "custom-uri:123",
			},
		},
		{
			"AuthRealm, Quarkus",
			args{
				connectionProperties: v1alpha1.InfinispanConnectionProperties{
					AuthRealm: "custom-realm",
				},
				secret:   nil,
				appProps: map[string]string{},
			},
			[]corev1.EnvVar{},
			map[string]string{
				PropertiesInfinispanQuarkus[AppPropInfinispanUseAuth]:   "false",
				PropertiesInfinispanQuarkus[AppPropInfinispanAuthRealm]: "custom-realm",
			},
		},
		{
			"SaslMechanism, Quarkus",
			args{
				connectionProperties: v1alpha1.InfinispanConnectionProperties{
					SaslMechanism: "DIGEST-MD5",
				},
				secret:   nil,
				appProps: map[string]string{},
			},
			[]corev1.EnvVar{},
			map[string]string{
				PropertiesInfinispanQuarkus[AppPropInfinispanUseAuth]:       "false",
				PropertiesInfinispanQuarkus[AppPropInfinispanSaslMechanism]: "DIGEST-MD5",
			},
		},
		{
			"CustomSecretDefaultKeys, Quarkus",
			args{
				runtime: v1alpha1.QuarkusRuntimeType,
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
				appProps: map[string]string{},
			},
			[]corev1.EnvVar{
				{
					Name: PropertiesInfinispanQuarkus[EnvVarInfinispanUser],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  infrastructure.InfinispanSecretUsernameKey,
						},
					},
				},
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
					Name: PropertiesInfinispanQuarkus[EnvVarInfinispanPassword],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  infrastructure.InfinispanSecretPasswordKey,
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
				{Name: infinispanEnvKeyCredSecret, Value: "custom-secret"},
			},
			map[string]string{
				PropertiesInfinispanQuarkus[AppPropInfinispanUseAuth]:       "true",
				PropertiesInfinispanQuarkus[AppPropInfinispanSaslMechanism]: string(v1alpha1.SASLPlain),
			},
		},
		{
			"CustomSecretCustomKeys, Quarkus",
			args{
				runtime: v1alpha1.QuarkusRuntimeType,
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
				appProps: map[string]string{},
			},
			[]corev1.EnvVar{
				{
					Name: PropertiesInfinispanQuarkus[EnvVarInfinispanUser],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  "custom-username",
						},
					},
				},
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
					Name: PropertiesInfinispanQuarkus[EnvVarInfinispanPassword],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  "custom-password",
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
				{Name: infinispanEnvKeyCredSecret, Value: "custom-secret"},
			},
			map[string]string{
				PropertiesInfinispanQuarkus[AppPropInfinispanUseAuth]:       "true",
				PropertiesInfinispanQuarkus[AppPropInfinispanSaslMechanism]: string(v1alpha1.SASLPlain),
			},
		},
		{
			"EmptyProperties, Spring",
			args{
				runtime:              v1alpha1.SpringbootRuntimeType,
				connectionProperties: v1alpha1.InfinispanConnectionProperties{},
				secret:               nil,
				appProps:             map[string]string{},
			},
			[]corev1.EnvVar{},
			map[string]string{
				PropertiesInfinispanSpring[AppPropInfinispanUseAuth]: "false",
			},
		},
		{
			"Uri, Spring",
			args{
				runtime: v1alpha1.SpringbootRuntimeType,
				connectionProperties: v1alpha1.InfinispanConnectionProperties{
					URI: "custom-uri:123",
				},
				secret:   nil,
				appProps: map[string]string{},
			},
			[]corev1.EnvVar{
				{Name: infinispanEnvKeyServerList, Value: "custom-uri:123"},
			},
			map[string]string{
				PropertiesInfinispanSpring[AppPropInfinispanUseAuth]:    "false",
				PropertiesInfinispanSpring[AppPropInfinispanServerList]: "custom-uri:123",
			},
		},
		{
			"AuthRealm, Spring",
			args{
				runtime: v1alpha1.SpringbootRuntimeType,
				connectionProperties: v1alpha1.InfinispanConnectionProperties{
					AuthRealm: "custom-realm",
				},
				secret:   nil,
				appProps: map[string]string{},
			},
			[]corev1.EnvVar{},
			map[string]string{
				PropertiesInfinispanSpring[AppPropInfinispanUseAuth]:   "false",
				PropertiesInfinispanSpring[AppPropInfinispanAuthRealm]: "custom-realm",
			},
		},
		{
			"SaslMechanism, Spring",
			args{
				runtime: v1alpha1.SpringbootRuntimeType,
				connectionProperties: v1alpha1.InfinispanConnectionProperties{
					SaslMechanism: "DIGEST-MD5",
				},
				secret:   nil,
				appProps: map[string]string{},
			},
			[]corev1.EnvVar{},
			map[string]string{
				PropertiesInfinispanSpring[AppPropInfinispanUseAuth]:       "false",
				PropertiesInfinispanSpring[AppPropInfinispanSaslMechanism]: "DIGEST-MD5",
			},
		},
		{
			"CustomSecretDefaultKeys, Spring",
			args{
				runtime: v1alpha1.SpringbootRuntimeType,
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
				appProps: map[string]string{},
			},
			[]corev1.EnvVar{
				{
					Name: PropertiesInfinispanSpring[EnvVarInfinispanUser],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  infrastructure.InfinispanSecretUsernameKey,
						},
					},
				},
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
					Name: PropertiesInfinispanSpring[EnvVarInfinispanPassword],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  infrastructure.InfinispanSecretPasswordKey,
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
				{Name: infinispanEnvKeyCredSecret, Value: "custom-secret"},
			},
			map[string]string{
				PropertiesInfinispanSpring[AppPropInfinispanUseAuth]:       "true",
				PropertiesInfinispanSpring[AppPropInfinispanSaslMechanism]: string(v1alpha1.SASLPlain),
			},
		},
		{
			"CustomSecretCustomKeys, Spring",
			args{
				runtime: v1alpha1.SpringbootRuntimeType,
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
				appProps: map[string]string{},
			},
			[]corev1.EnvVar{
				{
					Name: PropertiesInfinispanSpring[EnvVarInfinispanUser],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  "custom-username",
						},
					},
				},
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
					Name: PropertiesInfinispanSpring[EnvVarInfinispanPassword],
					ValueFrom: &corev1.EnvVarSource{
						SecretKeyRef: &corev1.SecretKeySelector{
							LocalObjectReference: corev1.LocalObjectReference{Name: "custom-secret"},
							Key:                  "custom-password",
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
				{Name: infinispanEnvKeyCredSecret, Value: "custom-secret"},
			},
			map[string]string{
				PropertiesInfinispanSpring[AppPropInfinispanUseAuth]:       "true",
				PropertiesInfinispanSpring[AppPropInfinispanSaslMechanism]: string(v1alpha1.SASLPlain),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			container := &corev1.Container{}

			setInfinispanVariables(tt.args.runtime, tt.args.connectionProperties, tt.args.secret, container, tt.args.appProps)

			assert.Equal(t, len(tt.expectedEnvVars), len(container.Env))
			for _, expectedEnvVar := range tt.expectedEnvVars {
				assert.Contains(t, container.Env, expectedEnvVar)
			}
			assert.Equal(t, tt.expectedAppProps, tt.args.appProps)
		})
	}
}
