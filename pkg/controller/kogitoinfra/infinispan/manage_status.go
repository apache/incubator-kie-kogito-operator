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

package infinispan

import (
	infinispan "github.com/infinispan/infinispan-operator/pkg/apis/infinispan/v1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"k8s.io/apimachinery/pkg/runtime"
)

func updateAppPropsInStatus(cli *client.Client, infinispanInstance *infinispan.Infinispan, instance *v1alpha1.KogitoInfra) error {
	log.Debugf("going to Update Infinispan app properties in kogito infra instance status")
	appProps, err := getInfinispanAppProps(cli, infinispanInstance.Name, infinispanInstance.Namespace)
	if err != nil {
		return err
	}
	instance.Status.AppProps = appProps
	log.Debugf("Following app properties are set infra status : %s", appProps)
	return nil
}

func updateEnvVarsInStatus(cli *client.Client, infinispanInstance *infinispan.Infinispan, instance *v1alpha1.KogitoInfra, scheme *runtime.Scheme) error {
	log.Debugf("going to Update Infinispan env properties in kogito infra instance status")
	envVars, err := getInfinispanSecretEnvVars(cli, infinispanInstance, instance, scheme)
	if err != nil {
		return err
	}
	instance.Status.Env = envVars
	log.Debugf("Following env properties are set infra status : %s", envVars)
	return nil
}
