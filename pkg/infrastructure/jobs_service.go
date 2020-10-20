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
	appsv1 "k8s.io/api/apps/v1"
)

const (
	// DefaultJobsServiceImageName is the default image name for the Jobs Service image
	DefaultJobsServiceImageName = "kogito-jobs-service"
	// DefaultJobsServiceName is the default name for the Jobs Services instance service
	DefaultJobsServiceName = "jobs-service"
	// kogito.jobs-service.url
	jobsServicesHTTPRouteEnv = "KOGITO_JOBS_SERVICE_URL"
)

// InjectJobsServicesURLIntoKogitoRuntimeServices will query for every KogitoRuntime in the given namespace to inject the Jobs Services route to each one
// Won't trigger an update if the KogitoRuntime already has the route set to avoid unnecessary reconciliation triggers
func InjectJobsServicesURLIntoKogitoRuntimeServices(client *client.Client, namespace string) error {
	log.Debugf("Injecting Jobs Service Route in kogito Runtime instances")
	return injectSupportingServiceURLIntoKogitoRuntime(client, namespace, jobsServicesHTTPRouteEnv, "", v1alpha1.JobsService)
}

// InjectJobsServiceURLIntoKogitoRuntimeDeployment will inject jobs-service route URL in to kogito runtime deployment env var
func InjectJobsServiceURLIntoKogitoRuntimeDeployment(client *client.Client, namespace string, deployment *appsv1.Deployment) error {
	log.Debugf("Injecting Data-Index URL in kogito Runtime deployment")
	return injectSupportingServiceURLInToDeployment(client, namespace, jobsServicesHTTPRouteEnv, "", deployment, v1alpha1.JobsService)
}
