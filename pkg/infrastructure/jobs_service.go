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
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	// DefaultJobsServiceImageName is the default image name for the Jobs Service image
	DefaultJobsServiceImageName = "kogito-jobs-service"
	// DefaultJobsServiceName is the default name for the Jobs Services instance service
	DefaultJobsServiceName = "jobs-service"

	// kogito.jobs-service.url
	jobsServicesHTTPURIEnv = "KOGITO_JOBS_SERVICE_URL"
)

// InjectJobsServicesURLIntoKogitoRuntimeServices will query for every KogitoRuntime in the given namespace to inject the Jobs Services route to each one
// Won't trigger an update if the KogitoRuntime already has the route set to avoid unnecessary reconciliation triggers
func InjectJobsServicesURLIntoKogitoRuntimeServices(cli *client.Client, namespace string) error {
	log.Debugf("Querying KogitoRuntime services in the namespace '%s' to inject Jobs Service Route ", namespace)
	deployments, err := getKogitoRuntimeDeployments(namespace, cli)
	if err != nil {
		return err
	}
	var jobServiceEndpoint ServiceEndpoints
	if len(deployments) > 0 {
		log.Debug("Querying Jobs Service route to inject into KogitoRuntime ")
		var err error
		jobServiceEndpoint, err = GetJobsServiceEndpoints(cli, namespace)
		if err != nil {
			return err
		}
		log.Debugf("Jobs Services URI is '%s'", jobServiceEndpoint.HTTPRouteURI)
	}

	for _, dc := range deployments {
		updateHTTP := updateJobsServiceURLIntoKogitoRuntimeEnv(&dc, jobServiceEndpoint)
		if updateHTTP {
			if err := kubernetes.ResourceC(cli).Update(&dc); err != nil {
				return err
			}
		}
	}
	return nil
}

// InjectJobsServiceURLIntoKogitoRuntimeDeployment will inject jobs-service route URL in to kogito runtime deployment env var
func InjectJobsServiceURLIntoKogitoRuntimeDeployment(client *client.Client, namespace string, runtimeDeployment *appsv1.Deployment) error {
	log.Debug("Querying Jobs Service route to inject into Kogito runtime ")
	jobServiceEndpoint, err := GetJobsServiceEndpoints(client, namespace)
	if err != nil {
		return err
	}
	log.Debugf("Jobs service route is '%s'", jobServiceEndpoint.HTTPRouteURI)
	updateJobsServiceURLIntoKogitoRuntimeEnv(runtimeDeployment, jobServiceEndpoint)
	return nil
}

func updateJobsServiceURLIntoKogitoRuntimeEnv(dc *appsv1.Deployment, jobServiceEndpoint ServiceEndpoints) (updateHTTP bool) {
	if len(dc.Spec.Template.Spec.Containers) > 0 {
		updateHTTP = framework.GetEnvVarFromContainer(jobServiceEndpoint.HTTPRouteEnv, &dc.Spec.Template.Spec.Containers[0]) != jobServiceEndpoint.HTTPRouteURI
		if updateHTTP {
			log.Debugf("Updating KogitoRuntime's DC '%s' to inject route %s ", dc.GetName(), jobServiceEndpoint.HTTPRouteURI)
			framework.SetEnvVar(jobServiceEndpoint.HTTPRouteEnv, jobServiceEndpoint.HTTPRouteURI, &dc.Spec.Template.Spec.Containers[0])
		}
	}
	return
}

// GetJobsServiceEndpoints gets Jobs Services published external endpoints
func GetJobsServiceEndpoints(client *client.Client, namespace string) (ServiceEndpoints, error) {
	endpoints := ServiceEndpoints{HTTPRouteEnv: jobsServicesHTTPURIEnv}
	route, err := getSingletonKogitoServiceRoute(client, namespace, &v1alpha1.KogitoJobsServiceList{})
	if err != nil {
		return endpoints, err
	}
	endpoints.HTTPRouteURI = route
	return endpoints, nil
}
