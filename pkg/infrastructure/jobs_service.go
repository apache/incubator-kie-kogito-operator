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
)

const (
	// kogito.jobs-service.url
	jobsServicesHTTPURIEnv = "KOGITO_JOBS_SERVICE_URL"
)

// InjectJobsServicesURLIntoKogitoApps will query for every KogitoApp in the given namespace to inject the Jobs Services route to each one
// Won't trigger an update if the KogitoApp already has the route set to avoid unnecessary reconciliation triggers
func InjectJobsServicesURLIntoKogitoApps(cli *client.Client, namespace string) error {
	log.Debugf("Querying KogitoApps in the namespace '%s' to inject Jobs Service Route ", namespace)
	dcs, err := getKogitoAppsDCs(namespace, cli)
	if err != nil {
		return err
	}
	externalURI := ""
	if len(dcs) > 0 {
		log.Debug("Querying Jobs Service URI to inject into KogitoApps ")
		var err error
		externalURI, err = getJobServiceExternalURI(cli, namespace)
		if err != nil {
			return err
		}
		log.Debugf("Jobs Services URI is '%s'", externalURI)
	}

	for _, dc := range dcs {
		// here we compare the current value to avoid updating the app every time
		if len(dc.Spec.Template.Spec.Containers) > 0 &&
			framework.GetEnvVarFromContainer(jobsServicesHTTPURIEnv, dc.Spec.Template.Spec.Containers[0]) != externalURI {
			log.Debugf("Updating kogitoApp's DC '%s' to inject route %s ", dc.GetName(), externalURI)
			framework.SetEnvVar(jobsServicesHTTPURIEnv, externalURI, &dc.Spec.Template.Spec.Containers[0])
			if err := kubernetes.ResourceC(cli).Update(&dc); err != nil {
				return err
			}
		}
	}
	return nil
}

func getJobServiceExternalURI(client *client.Client, namespace string) (string, error) {
	externalURI := ""
	services := &v1alpha1.KogitoJobsServiceList{}
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, services); err != nil {
		return externalURI, err
	}
	if len(services.Items) > 0 {
		// should be only one data index guaranteed by the controller, but still we are looking for the first one
		externalURI = services.Items[0].Status.ExternalURI
	}
	return externalURI, nil
}
