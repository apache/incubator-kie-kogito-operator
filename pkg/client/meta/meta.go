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
	infinispanv1 "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	grafana "github.com/integr8ly/grafana-operator/v3/pkg/apis/integreatly/v1alpha1"
	keycloakv1alpha1 "github.com/keycloak/keycloak-operator/pkg/apis/keycloak/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/api/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/logger"
	mongodb "github.com/mongodb/mongodb-kubernetes-operator/pkg/apis/mongodb/v1"
	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"
	operatormkt "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	coreappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	eventingv1 "knative.dev/eventing/pkg/apis/eventing/v1"
	sourcesv1alpha1 "knative.dev/eventing/pkg/apis/sources/v1alpha1"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"

	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	olmapiv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	olmapiv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
	olmv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/package-server/apis/operators/v1"
)

var log = logger.GetLogger("meta")

// DefinitionKind is a resource kind representation from a Kubernetes/Openshift cluster
type DefinitionKind struct {
	// Name of the resource
	Name string
	// IsFromOpenShift identifies if this Resource only exists on OpenShift cluster
	IsFromOpenShift bool
	// Identifies the group version for the OpenShift APIs
	GroupVersion schema.GroupVersion
}

var (
	// KindService for service
	KindService = DefinitionKind{"Service", false, corev1.SchemeGroupVersion}
	// KindBuildRequest for a BuildRequest
	KindBuildRequest = DefinitionKind{"BuildRequest", true, buildv1.GroupVersion}
	// KindDeployment for Kubernetes Deployment
	KindDeployment = DefinitionKind{"Deployment", false, coreappsv1.SchemeGroupVersion}
)

// SetGroupVersionKind sets the group, version and kind for the resource
func SetGroupVersionKind(typeMeta *metav1.TypeMeta, kind DefinitionKind) {
	typeMeta.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   kind.GroupVersion.Group,
		Version: kind.GroupVersion.Version,
		Kind:    kind.Name,
	})
}

// GetRegisteredSchemeBuilder gets the SchemeBuilder with all the desired APIs registered
func GetRegisteredSchemeBuilder() runtime.SchemeBuilder {
	return runtime.NewSchemeBuilder(
		v1beta1.SchemeBuilder.AddToScheme,
		corev1.AddToScheme,
		coreappsv1.AddToScheme,
		buildv1.Install,
		rbac.AddToScheme,
		appsv1.Install,
		coreappsv1.AddToScheme,
		routev1.Install,
		imgv1.Install,
		apiextensionsv1beta1.AddToScheme,
		kafkabetav1.SchemeBuilder.AddToScheme,
		mongodb.SchemeBuilder.AddToScheme,
		infinispanv1.AddToScheme,
		keycloakv1alpha1.SchemeBuilder.AddToScheme,
		operatormkt.SchemeBuilder.AddToScheme, olmapiv1.AddToScheme, olmapiv1alpha1.AddToScheme,
		monv1.SchemeBuilder.AddToScheme,
		eventingv1.AddToScheme, sourcesv1alpha1.AddToScheme,
		grafana.AddToScheme,
		olmv1.AddToScheme)
}

// GetRegisteredSchema gets all schema and types registered for use with CLI, unit tests, custom clients and so on
func GetRegisteredSchema() *runtime.Scheme {
	s := scheme.Scheme
	schemes := GetRegisteredSchemeBuilder()
	err := schemes.AddToScheme(s)
	if err != nil {
		log.Error(err, "Failed to register APIs schemes")
		panic(err)
	}

	// After upgrading to Operator SDK 0.11.0 we need to add CreateOptions to our own schema. See: https://issues.jboss.org/browse/KOGITO-493
	metav1.AddToGroupVersion(s, v1beta1.GroupVersion)
	// https://issues.jboss.org/browse/KOGITO-617
	metav1.AddToGroupVersion(s, apiextensionsv1beta1.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, operatormkt.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, appsv1.GroupVersion)
	metav1.AddToGroupVersion(s, olmapiv1.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, olmapiv1alpha1.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, monv1.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, routev1.GroupVersion)
	metav1.AddToGroupVersion(s, infinispanv1.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, mongodb.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, kafkabetav1.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, grafana.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, olmv1.SchemeGroupVersion)

	return s
}
