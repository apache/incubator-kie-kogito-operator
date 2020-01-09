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

package shared

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/test"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/controller/kogitoapp/resource"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"
	"github.com/kiegroup/kogito-cloud-operator/pkg/operator"
	olmapiv1 "github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"

	"github.com/stretchr/testify/assert"

	operatormkt "github.com/operator-framework/operator-marketplace/pkg/apis/operators/v1"
	apps "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"testing"
)

func Test_InstallOperatorWithYaml(t *testing.T) {
	ns := t.Name()
	client := test.SetupFakeKubeCli(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}})
	image := "docker.io/myrepo/custom-operator:1.0"

	err := installOperatorWithYamlFiles(image, ns, client)
	assert.NoError(t, err)

	serviceAccount := v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resource.ServiceAccountName,
			Namespace: ns,
		},
	}

	_, err = kubernetes.ResourceC(client).Fetch(&serviceAccount)
	assert.NoError(t, err)
	assert.Equal(t, resource.ServiceAccountName, serviceAccount.Name)

	serviceAccount = v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.Name,
			Namespace: ns,
		},
	}

	_, err = kubernetes.ResourceC(client).Fetch(&serviceAccount)
	assert.NoError(t, err)
	assert.Equal(t, operator.Name, serviceAccount.Name)

	deployment := &apps.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      operator.Name,
			Namespace: ns,
		},
	}
	_, err = kubernetes.ResourceC(client).Fetch(deployment)
	assert.NoError(t, err)
	assert.Equal(t, operator.Name, deployment.Name)
	assert.Equal(t, image, deployment.Spec.Template.Spec.Containers[0].Image)

	// checks CRD
	crds := &apiextensionsv1beta1.CustomResourceDefinitionList{}
	err = kubernetes.ResourceC(client).ListWithNamespace(ns, crds)
	assert.NoError(t, err)
	assert.Len(t, crds.Items, 4)
	assert.Contains(t, crds.Items[0].Name, "app.kiegroup.org")
	assert.Contains(t, crds.Items[1].Name, "app.kiegroup.org")
	assert.Contains(t, crds.Items[2].Name, "app.kiegroup.org")
}

func TestMustInstallOperatorIfNotExists_WithOperatorHub(t *testing.T) {
	ns := operatorMarketplaceNamespace
	operatorSource := &operatormkt.OperatorSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      communityOperatorSource,
			Namespace: operatorMarketplaceNamespace,
		},
		Status: operatormkt.OperatorSourceStatus{
			Packages: "cert-utils-operator,spark-gcp,metering,spinnaker-operator,apicurito,kubefed,prometheus,hawtio-operator,t8c,hazelcast-enterprise,opsmx-spinnaker-operator,ibmcloud-operator,openebs,iot-simulator,submariner,microcks,enmasse,teiid,federation,aqua,eclipse-che,3scale-community-operator,jaeger,openshift-pipelines-operator,awss3-operator-registry,service-binding-operator,node-network-operator,myvirtualdirectory,triggermesh,namespace-configuration-operator,maistraoperator,camel-k,federatorai,knative-serving-operator,syndesis,knative-kafka-operator,postgresql,event-streams-topic,planetscale,kiali,ripsaw,esindex-operator,halkyon,quay,kogito-operator,seldon-operator,cockroachdb,atlasmap-operator,strimzi-kafka-operator,knative-camel-operator,lightbend-console-operator,descheduler,node-problem-detector,opendatahub-operator,radanalytics-spark,hco-operatorhub,smartgateway-operator,etcd,knative-eventing-operator,postgresql-operator-dev4devs-com,twistlock,microsegmentation-operator,open-liberty,akka-cluster-operator,grafana-operator,kubeturbo,appsody-community-operator,infinispan",
		},
	}
	client := test.SetupFakeKubeCli(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}, operatorSource)
	// Operator is there in the hub and not exists in the given namespace, let's check if there's no error
	installed, err := MustInstallOperatorIfNotExists(ns, defaultOperatorImageName, client, false)
	assert.NoError(t, err)
	assert.True(t, installed)

	groups := &olmapiv1.OperatorGroupList{}
	err = kubernetes.ResourceC(client).ListWithNamespace(ns, groups)
	assert.NoError(t, err)
	assert.Contains(t, groups.Items[0].Name, ns)

	subs := &v1alpha1.Subscription{ObjectMeta: metav1.ObjectMeta{Name: defaultOperatorPackageName, Namespace: ns}}
	exists, err := kubernetes.ResourceC(client).Fetch(subs)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestTryToInstallOperatorIfNotExists_WithOperatorHub(t *testing.T) {
	ns := operatorMarketplaceNamespace
	operatorSource := &operatormkt.OperatorSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      communityOperatorSource,
			Namespace: operatorMarketplaceNamespace,
		},
		Status: operatormkt.OperatorSourceStatus{
			Packages: "cert-utils-operator,spark-gcp,metering,spinnaker-operator,apicurito,kubefed,prometheus,hawtio-operator,t8c,hazelcast-enterprise,opsmx-spinnaker-operator,ibmcloud-operator,openebs,iot-simulator,submariner,microcks,enmasse,teiid,federation,aqua,eclipse-che,3scale-community-operator,jaeger,openshift-pipelines-operator,awss3-operator-registry,service-binding-operator,node-network-operator,myvirtualdirectory,triggermesh,namespace-configuration-operator,maistraoperator,camel-k,federatorai,knative-serving-operator,syndesis,knative-kafka-operator,postgresql,event-streams-topic,planetscale,kiali,ripsaw,esindex-operator,halkyon,quay,kogito-operator,seldon-operator,cockroachdb,atlasmap-operator,strimzi-kafka-operator,knative-camel-operator,lightbend-console-operator,descheduler,node-problem-detector,opendatahub-operator,radanalytics-spark,hco-operatorhub,smartgateway-operator,etcd,knative-eventing-operator,postgresql-operator-dev4devs-com,twistlock,microsegmentation-operator,open-liberty,akka-cluster-operator,grafana-operator,kubeturbo,appsody-community-operator,infinispan",
		},
	}
	client := test.SetupFakeKubeCli(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}, operatorSource)
	// Operator is in the hub but does not exist in the given namespace. Don't raise an error.
	installed, err := SilentlyInstallOperatorIfNotExists(ns, defaultOperatorImageName, client)
	assert.NoError(t, err)
	assert.True(t, installed)
}

func TestMustInstallOperatorIfNotExists_WithoutOperatorHub(t *testing.T) {
	ns := t.Name()

	client := test.SetupFakeKubeCli(
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}},
		&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: operatorMarketplaceNamespace}},
	)
	// Operator is not in the hub. Install with yaml files.
	installed, err := MustInstallOperatorIfNotExists(ns, defaultOperatorImageName, client, false)
	assert.NoError(t, err)
	assert.True(t, installed)
	// Operator is now in the hub, but no pods are running because this is a controlled test environment
	exist, err := infrastructure.CheckKogitoOperatorExists(client, ns)
	assert.Error(t, err)
	assert.True(t, exist)
	assert.Contains(t, err.Error(), "kogito-operator Operator seems to be created in the namespace")
}
