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

package meta

import (
	"github.com/kiegroup/kogito-operator/apis/app/v1beta1"
	v1 "github.com/kiegroup/kogito-operator/apis/rhpam/v1"
	"github.com/kiegroup/kogito-operator/core/framework/util"
	grafana "github.com/kiegroup/kogito-operator/core/infrastructure/grafana/v1alpha1"
	infinispan "github.com/kiegroup/kogito-operator/core/infrastructure/infinispan/v1"
	"github.com/kiegroup/kogito-operator/core/infrastructure/kafka/v1beta2"
	keycloakv1alpha1 "github.com/kiegroup/kogito-operator/core/infrastructure/keycloak/v1alpha1"
	mongodb "github.com/kiegroup/kogito-operator/core/infrastructure/mongodb/v1"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	coreappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	sourcesv1 "knative.dev/eventing/pkg/apis/sources/v1"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	monv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
)

// GetRegisteredSchema gets all schema and types registered for use with CLI, unit tests, custom clients and so on
func GetRegisteredSchema() *runtime.Scheme {
	s := runtime.NewScheme()
	schemes := getRegisteredSchemeBuilder()
	err := schemes.AddToScheme(s)
	if err != nil {
		panic(err)
	}

	// After upgrading to Operator SDK 0.11.0 we need to add CreateOptions to our own schema. See: https://issues.jboss.org/browse/KOGITO-493
	if util.IsProductMode() {
		metav1.AddToGroupVersion(s, v1.GroupVersion)
	} else {
		metav1.AddToGroupVersion(s, v1beta1.GroupVersion)
	}
	// https://issues.jboss.org/browse/KOGITO-617
	metav1.AddToGroupVersion(s, apiextensionsv1.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, appsv1.GroupVersion)
	metav1.AddToGroupVersion(s, monv1.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, routev1.GroupVersion)
	metav1.AddToGroupVersion(s, infinispan.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, mongodb.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, v1beta2.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, grafana.GroupVersion)
	metav1.AddToGroupVersion(s, eventingv1.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, sourcesv1.SchemeGroupVersion)
	return s
}

// getRegisteredSchemeBuilder gets the SchemeBuilder with all the desired APIs registered
func getRegisteredSchemeBuilder() runtime.SchemeBuilder {
	return runtime.NewSchemeBuilder(
		v1beta1.SchemeBuilder.AddToScheme,
		v1.SchemeBuilder.AddToScheme,
		clientgoscheme.AddToScheme,
		corev1.AddToScheme,
		coreappsv1.AddToScheme,
		buildv1.Install,
		rbac.AddToScheme,
		appsv1.Install,
		coreappsv1.AddToScheme,
		routev1.Install,
		imgv1.Install,
		apiextensionsv1.AddToScheme,
		v1beta2.SchemeBuilder.AddToScheme,
		mongodb.SchemeBuilder.AddToScheme,
		infinispan.AddToScheme,
		keycloakv1alpha1.SchemeBuilder.AddToScheme,
		monv1.SchemeBuilder.AddToScheme,
		eventingv1.AddToScheme, sourcesv1.AddToScheme,
		grafana.AddToScheme)
}
