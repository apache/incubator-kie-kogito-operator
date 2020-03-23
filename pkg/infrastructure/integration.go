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
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	v1 "k8s.io/api/core/v1"
)

// ServiceEndpoints represents the endpoints for a deployed Kogito Data Index service
type ServiceEndpoints struct {
	// HTTPRouteURI ...
	HTTPRouteURI string
	// HTTPRouteEnv name of the environment variable that will hold the HTTP URI
	HTTPRouteEnv string
	// WSRouteURI ...
	WSRouteURI string
	// WSRouteEnv name of the environment variable that will hold the HTTP URI
	WSRouteEnv string
}

func (s ServiceEndpoints) String() string {
	return s.HTTPRouteURI
}

// InjectEnvVarsFromExternalServices inject environment variables from external services that the KogitoApp runtime might need
func InjectEnvVarsFromExternalServices(kogitoApp *v1alpha1.KogitoApp, container *v1.Container, client *client.Client) error {
	log.Debugf("Querying external routes to inject into KogitoApp: %s", kogitoApp.GetName())
	dataIndexEndpoints, err := GetKogitoDataIndexEndpoints(client, kogitoApp.GetNamespace())
	if err != nil {
		return err
	}
	log.Debugf("Data Index route is '%s'", dataIndexEndpoints.HTTPRouteURI)
	jobsServiceEndpoints, err := GetJobsServiceEndpoints(client, kogitoApp.GetNamespace())
	if err != nil {
		return err
	}
	log.Debugf("Jobs Service route is '%s'", jobsServiceEndpoints)
	framework.SetEnvVar(dataIndexEndpoints.HTTPRouteEnv, dataIndexEndpoints.HTTPRouteURI, container)
	framework.SetEnvVar(dataIndexEndpoints.WSRouteEnv, dataIndexEndpoints.WSRouteURI, container)
	framework.SetEnvVar(jobsServiceEndpoints.HTTPRouteEnv, jobsServiceEndpoints.HTTPRouteURI, container)
	return nil
}
