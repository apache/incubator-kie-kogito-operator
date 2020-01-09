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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"

	appsv1 "github.com/openshift/api/apps/v1"
	buildv1 "github.com/openshift/api/build/v1"
	imgv1 "github.com/openshift/api/image/v1"
	routev1 "github.com/openshift/api/route/v1"

	operatormkt "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"

	coreappsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"

	monv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"

	olmapiv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	olmapiv1alpha1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
)

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
	// KindServiceAccount for serviceAccount
	KindServiceAccount = DefinitionKind{"ServiceAccount", false, corev1.SchemeGroupVersion}
	// KindBuildConfig for a buildConfig
	KindBuildConfig = DefinitionKind{"BuildConfig", true, buildv1.SchemeGroupVersion}
	// KindDeploymentConfig for a DeploymentConfig
	KindDeploymentConfig = DefinitionKind{"DeploymentConfig", true, appsv1.SchemeGroupVersion}
	// KindRoute for a Route
	KindRoute = DefinitionKind{"Route", true, routev1.SchemeGroupVersion}
	// KindImageStreamTag for a ImageStreamTag
	KindImageStreamTag = DefinitionKind{"ImageStreamTag", true, imgv1.SchemeGroupVersion}
	// KindImageStream for a ImageStream
	KindImageStream = DefinitionKind{"ImageStream", true, imgv1.SchemeGroupVersion}
	// KindBuildRequest for a BuildRequest
	KindBuildRequest = DefinitionKind{"BuildRequest", true, buildv1.SchemeGroupVersion}
	// KindNamespace for a Namespace
	KindNamespace = DefinitionKind{"Namespace", false, corev1.SchemeGroupVersion}
	// KindCRD for a CustomResourceDefinition
	KindCRD = DefinitionKind{"CustomResourceDefinition", false, apiextensionsv1beta1.SchemeGroupVersion}
	// KindKogitoApp for a KogitoApp controller
	KindKogitoApp = DefinitionKind{"KogitoApp", false, v1alpha1.SchemeGroupVersion}
	// KindKogitoDataIndex for a KindKogitoDataIndex controller
	KindKogitoDataIndex = DefinitionKind{"KogitoDataIndex", false, v1alpha1.SchemeGroupVersion}
	// KindKogitoDataIndexList for a KindKogitoDataIndexList controller
	KindKogitoDataIndexList = DefinitionKind{"KogitoDataIndexList", false, v1alpha1.SchemeGroupVersion}
	// KindKogitoJobsService for a KogitoJobsService controller
	KindKogitoJobsService = DefinitionKind{"KogitoJobsService", false, v1alpha1.SchemeGroupVersion}
	// KindConfigMap for a ConfigMap
	KindConfigMap = DefinitionKind{"ConfigMap", false, corev1.SchemeGroupVersion}
	// KindDeployment for a Deployment
	KindDeployment = DefinitionKind{"Deployment", false, coreappsv1.SchemeGroupVersion}
	// KindStatefulSet for a StatefulSet
	KindStatefulSet = DefinitionKind{"StatefulSet", false, coreappsv1.SchemeGroupVersion}
	// KindRole ...
	KindRole = DefinitionKind{"Role", false, rbac.SchemeGroupVersion}
	// KindRoleBinding ...
	KindRoleBinding = DefinitionKind{"RoleBinding", false, rbac.SchemeGroupVersion}
	// KindOperatorSource ...
	KindOperatorSource = DefinitionKind{"OperatorSource", false, operatormkt.SchemeGroupVersion}
	// KindServiceMonitor ...
	KindServiceMonitor = DefinitionKind{"ServiceMonitor", false, monv1.SchemeGroupVersion}
	// KindOperatorGroup ...
	KindOperatorGroup = DefinitionKind{"OperatorGroup", false, olmapiv1.SchemeGroupVersion}
	// KindSubscription ...
	KindSubscription = DefinitionKind{"Subscription", false, olmapiv1alpha1.SchemeGroupVersion}
)

// SetGroupVersionKind sets the group, version and kind for the resource
func SetGroupVersionKind(typeMeta *metav1.TypeMeta, kind DefinitionKind) {
	typeMeta.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   kind.GroupVersion.Group,
		Version: kind.GroupVersion.Version,
		Kind:    kind.Name,
	})
}

// GetRegisteredSchema gets all schema and types registered for use with CLI, unit tests, custom clients and so on
func GetRegisteredSchema() *runtime.Scheme {
	s := scheme.Scheme
	s.AddKnownTypes(corev1.SchemeGroupVersion, &corev1.Namespace{}, &corev1.ServiceAccount{})
	s.AddKnownTypes(coreappsv1.SchemeGroupVersion, &coreappsv1.Deployment{})
	s.AddKnownTypes(rbac.SchemeGroupVersion, &rbac.Role{}, &rbac.RoleBinding{})
	s.AddKnownTypes(apiextensionsv1beta1.SchemeGroupVersion, &apiextensionsv1beta1.CustomResourceDefinition{}, &apiextensionsv1beta1.CustomResourceDefinitionList{})
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion,
		&v1alpha1.KogitoApp{}, &v1alpha1.KogitoAppList{},
		&v1alpha1.KogitoDataIndex{}, &v1alpha1.KogitoDataIndexList{},
		&v1alpha1.KogitoInfra{}, &v1alpha1.KogitoInfraList{},
		&v1alpha1.KogitoJobsService{}, &v1alpha1.KogitoJobsServiceList{})
	s.AddKnownTypes(kafkabetav1.SchemeGroupVersion, &kafkabetav1.Kafka{}, &kafkabetav1.KafkaList{}, &kafkabetav1.KafkaTopic{}, &kafkabetav1.KafkaTopicList{})
	s.AddKnownTypes(infinispanv1.SchemeGroupVersion, &infinispanv1.Infinispan{}, &infinispanv1.InfinispanList{})
	s.AddKnownTypes(appsv1.GroupVersion, &appsv1.DeploymentConfig{}, &appsv1.DeploymentConfigList{})
	s.AddKnownTypes(buildv1.GroupVersion, &buildv1.BuildConfig{}, &buildv1.BuildConfigList{})
	s.AddKnownTypes(routev1.GroupVersion, &routev1.Route{}, &routev1.RouteList{})
	s.AddKnownTypes(imgv1.GroupVersion, &imgv1.ImageStreamTag{}, &imgv1.ImageStream{}, &imgv1.ImageStreamList{})
	s.AddKnownTypes(operatormkt.SchemeGroupVersion, &operatormkt.OperatorSource{}, &operatormkt.OperatorSourceList{})
	s.AddKnownTypes(olmapiv1.SchemeGroupVersion, &olmapiv1.OperatorGroup{}, &olmapiv1.OperatorGroupList{})
	s.AddKnownTypes(olmapiv1alpha1.SchemeGroupVersion, &olmapiv1alpha1.Subscription{}, &olmapiv1alpha1.SubscriptionList{})

	// After upgrading to Operator SDK 0.11.0 we need to add CreateOptions to our own schema. See: https://issues.jboss.org/browse/KOGITO-493
	metav1.AddToGroupVersion(s, v1alpha1.SchemeGroupVersion)
	// https://issues.jboss.org/browse/KOGITO-617
	metav1.AddToGroupVersion(s, apiextensionsv1beta1.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, operatormkt.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, appsv1.GroupVersion)
	metav1.AddToGroupVersion(s, olmapiv1.SchemeGroupVersion)
	metav1.AddToGroupVersion(s, olmapiv1alpha1.SchemeGroupVersion)

	return s
}
