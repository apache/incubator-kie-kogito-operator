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

package infrastructure

import (
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client/kubernetes"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"strconv"
)

const (
	HTTPPortEnvVar = "HTTP_PORT"
)

// getSingletonKogitoServiceRoute gets the route from a kogito service that's unique in the given namespace
func getSingletonKogitoServiceRoute(client *client.Client, namespace string, serviceListRef v1alpha1.KogitoServiceList) (string, error) {
	if err := kubernetes.ResourceC(client).ListWithNamespace(namespace, serviceListRef); err != nil {
		return "", err
	}
	if serviceListRef.GetItemsCount() > 0 {
		return serviceListRef.GetItemAt(0).GetStatus().GetExternalURI(), nil
	}
	return "", nil
}

func SetHttpPortEnvVar(container *v1.Container, kogitoService v1alpha1.KogitoService) {
	httpPort := defineHTTPPort(kogitoService)
	framework.SetEnvVar(HTTPPortEnvVar, strconv.Itoa(int(httpPort)), container)
	container.Ports[0].ContainerPort = httpPort
	if container.ReadinessProbe != nil &&
		container.ReadinessProbe.TCPSocket != nil {
		container.ReadinessProbe.TCPSocket.Port = intstr.IntOrString{IntVal: httpPort}
		container.LivenessProbe.TCPSocket.Port = intstr.IntOrString{IntVal: httpPort}
	} else if container.ReadinessProbe != nil &&
		container.ReadinessProbe.HTTPGet != nil {
		container.ReadinessProbe.HTTPGet.Port = intstr.IntOrString{IntVal: httpPort}
		container.LivenessProbe.HTTPGet.Port = intstr.IntOrString{IntVal: httpPort}
	}
}

// defineHTTPPort will define on which port the service should be listening to. To set it use httpPort cr parameter.
// defaults to 8080
func defineHTTPPort(kogitoService v1alpha1.KogitoService) int32 {
	// port should be greater than 0
	httpPort := kogitoService.GetSpec().GetHTTPPort()
	if httpPort < 1 {
		log.Debugf("HTTPPort not set, returning default http port.")
		return framework.DefaultExposedPort
	} else {
		log.Debugf("HTTPPort is set, returning port number %i", httpPort)
		return httpPort
	}
}
