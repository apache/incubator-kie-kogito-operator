// Copyright 2021 Red Hat, Inc. and/or its affiliates
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

package installers

import (
	"fmt"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"

	operatorframework "github.com/kiegroup/kogito-operator/core/framework"
	mongodbv1 "github.com/kiegroup/kogito-operator/core/infrastructure/mongodb/v1"
	"github.com/kiegroup/kogito-operator/test/pkg/framework"
	coreapps "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

var (
	// mongoDbYamlNamespacedInstaller installs MongoDB namespaced using YAMLs
	mongoDbYamlNamespacedInstaller = YamlNamespacedServiceInstaller{
		InstallNamespacedYaml:           installMongoDbUsingYaml,
		WaitForNamespacedServiceRunning: waitForMongoDbUsingYamlRunning,
		GetAllNamespaceYamlCrs:          getMongoDbCrsInNamespace,
		UninstallNamespaceYaml:          uninstallMongoDbUsingYaml,
		NamespacedYamlServiceName:       mongoDBOperatorServiceName,
	}

	mongoDBOperatorServiceName    = "Mongo DB"
	mongoDBOperatorVersion        = "v0.7.0"
	mongoDBOperatorDeployFilesURI = "https://raw.githubusercontent.com/mongodb/mongodb-kubernetes-operator/" + mongoDBOperatorVersion + "/"

	// Used for CRD creation in case of parallel execution of scenarios
	mongoDBCrdMux = &sync.Mutex{}
)

// GetMongoDbInstaller returns MongoDB installer
func GetMongoDbInstaller() ServiceInstaller {
	return &mongoDbYamlNamespacedInstaller
}

func installMongoDbUsingYaml(namespace string) error {
	framework.GetLogger(namespace).Info("Deploy MongoDB from yaml files", "file uri", mongoDBOperatorDeployFilesURI)

	// Lock to avoid parallel creation of crds
	var err error
	mongoDBCrdMux.Lock()
	if !framework.IsMongoDBAvailable(namespace) {
		err = framework.LoadResource(namespace, mongoDBOperatorDeployFilesURI+"config/crd/bases/mongodbcommunity.mongodb.com_mongodbcommunity.yaml", &apiextensionsv1.CustomResourceDefinition{}, nil)
	}
	mongoDBCrdMux.Unlock()
	if err != nil {
		return err
	}

	// rbac
	if err = framework.LoadResource(namespace, mongoDBOperatorDeployFilesURI+"config/rbac/service_account.yaml", &corev1.ServiceAccount{}, nil); err != nil {
		return err
	}
	if err = framework.LoadResource(namespace, mongoDBOperatorDeployFilesURI+"config/rbac/role.yaml", &rbac.Role{}, nil); err != nil {
		return err
	}
	if err = framework.LoadResource(namespace, mongoDBOperatorDeployFilesURI+"config/rbac/role_binding.yaml", &rbac.RoleBinding{}, nil); err != nil {
		return err
	}

	// Then operator
	if err = framework.LoadResource(namespace, mongoDBOperatorDeployFilesURI+"config/manager/manager.yaml", &coreapps.Deployment{}, func(object interface{}) {
		// Override with v0.7.0 values as the branch does not have the correct ones ...
		object.(*coreapps.Deployment).Spec.Template.Spec.Containers[0].Image = "quay.io/mongodb/mongodb-kubernetes-operator:0.7.0"
		object.(*coreapps.Deployment).Spec.Template.Spec.Containers[0].Env = operatorframework.EnvOverride(object.(*coreapps.Deployment).Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{Name: "AGENT_IMAGE", Value: "quay.io/mongodb/mongodb-agent:11.0.5.6963-1"},
			corev1.EnvVar{Name: "VERSION_UPGRADE_HOOK_IMAGE", Value: "quay.io/mongodb/mongodb-kubernetes-operator-version-upgrade-post-start-hook:1.0.2"},
			corev1.EnvVar{Name: "READINESS_PROBE_IMAGE", Value: "quay.io/mongodb/mongodb-kubernetes-readinessprobe:1.0.4"},
		)
		if framework.IsOpenshift() {
			framework.GetLogger(namespace).Debug("Setup MANAGED_SECURITY_CONTEXT env in MongoDB operator for Openshift")
			object.(*coreapps.Deployment).Spec.Template.Spec.Containers[0].Env = operatorframework.EnvOverride(object.(*coreapps.Deployment).Spec.Template.Spec.Containers[0].Env,
				corev1.EnvVar{Name: "MANAGED_SECURITY_CONTEXT", Value: "true"},
			)
		}
	}); err != nil {
		return err
	}

	// Set correct file to be deployed
	if framework.IsOpenshift() {
		// Used to give correct access to pvc/secret
		// https://github.com/mongodb/mongodb-kubernetes-operator/issues/212#issuecomment-704744307
		output, err := framework.CreateCommand("oc", "adm", "policy", "add-scc-to-user", "anyuid", "system:serviceaccount:"+namespace+":mongodb-kubernetes-operator").WithLoggerContext(namespace).Sync("add-scc-to-user").Execute()
		if err != nil {
			framework.GetLogger(namespace).Error(err, "Error while trying to set specific rights for MongoDB deployments")
			return err
		}
		framework.GetLogger(namespace).Info(output)
	}

	return nil
}

func waitForMongoDbUsingYamlRunning(namespace string) error {
	return framework.WaitForMongoDBOperatorRunning(namespace)
}

func uninstallMongoDbUsingYaml(namespace string) error {
	framework.GetMainLogger().Info("Uninstalling Mongo DB")

	var originalError error

	output, err := framework.CreateCommand("oc", "adm", "policy", "remove-scc-from-user", "anyuid", "system:serviceaccount:"+namespace+":mongodb-kubernetes-operator").WithLoggerContext(namespace).Execute()
	if err != nil {
		framework.GetLogger(namespace).Error(err, fmt.Sprintf("Deleting Mongo DB operator failed, output: %s", output))
		originalError = err
	}

	output, err = framework.CreateCommand("oc", "delete", "-f", mongoDBOperatorDeployFilesURI+"manager/manager.yaml").WithLoggerContext(namespace).Execute()
	if err != nil {
		framework.GetLogger(namespace).Error(err, fmt.Sprintf("Deleting Mongo DB operator failed, output: %s", output))
		if originalError == nil {
			originalError = err
		}
	}

	output, err = framework.CreateCommand("oc", "delete", "-f", mongoDBOperatorDeployFilesURI+"rbac/role_binding.yaml").WithLoggerContext(namespace).Execute()
	if err != nil {
		framework.GetLogger(namespace).Error(err, fmt.Sprintf("Deleting Mongo DB role binding failed, output: %s", output))
		if originalError == nil {
			originalError = err
		}
	}

	output, err = framework.CreateCommand("oc", "delete", "-f", mongoDBOperatorDeployFilesURI+"rbac/role.yaml").WithLoggerContext(namespace).Execute()
	if err != nil {
		framework.GetLogger(namespace).Error(err, fmt.Sprintf("Deleting Mongo DB role failed, output: %s", output))
		if originalError == nil {
			originalError = err
		}
	}

	output, err = framework.CreateCommand("oc", "delete", "-f", mongoDBOperatorDeployFilesURI+"rbac/service_account.yaml").WithLoggerContext(namespace).Execute()
	if err != nil {
		framework.GetLogger(namespace).Error(err, fmt.Sprintf("Deleting Mongo DB service account failed, output: %s", output))
		if originalError == nil {
			originalError = err
		}
	}

	return originalError
}

func getMongoDbCrsInNamespace(namespace string) ([]client.Object, error) {
	var crs []client.Object

	mongoDbs := &mongodbv1.MongoDBCommunityList{}
	if err := framework.GetObjectsInNamespace(namespace, mongoDbs); err != nil {
		return nil, err
	}
	for i := range mongoDbs.Items {
		crs = append(crs, &mongoDbs.Items[i])
	}

	return crs, nil
}
