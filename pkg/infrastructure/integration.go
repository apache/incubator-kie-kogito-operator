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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
)

// InjectEnvVarsFromExternalServices inject environment variables from external services that the KogitoApp runtime might need
func InjectEnvVarsFromExternalServices(kogitoApp *v1alpha1.KogitoApp, client *client.Client) error {
	log.Debugf("Querying Data Index route to inject into KogitoApp: %s", kogitoApp.GetName())
	// We look for a deployed data index to inject into the runtime service
	// later we could also integrate with other external services like Kafka, Infinispan and SSO
	route, err := getKogitoDataIndexRoute(client, kogitoApp.GetNamespace())
	if err != nil {
		return err
	}
	log.Debugf("Data Index route is '%s'", route)
	kogitoApp.Spec.Env = util.AppendOrReplaceEnv(v1alpha1.Env{Name: kogitoDataIndexRouteEnv, Value: route}, kogitoApp.Spec.Env)
	return nil
}
