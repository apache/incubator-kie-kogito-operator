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

package resource

import (
	"strconv"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"

	appsv1 "github.com/openshift/api/apps/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	prometheusLabelKeyPrefix = "prometheus.io"
	labelNamespaceSep        = "/"
	orgKieNamespaceLabelKey  = "org.kie" + labelNamespaceSep
	orgKiePersistenceKey     = "persistence"
	dockerLabelServicesSep   = ","
	portSep                  = ":"
	defaultExportedProtocol  = "http"
	portFormatWrongMessage   = "Service %s on " + openshift.ImageLabelForExposeServices + " label in wrong format. Won't be possible to expose Services for this application. Should be PORT_NUMBER:PROTOCOL. e.g. 8080:http"
)

func dockerImageHasLabels(dockerImage *dockerv10.DockerImage) bool {
	if dockerImage == nil || dockerImage.Config == nil || dockerImage.Config.Labels == nil {
		return false
	}
	return true
}

// mergeImageMetadataWithDeploymentConfig retrieves org.kie and prometheus.io labels from DockerImage and adds them to the DeploymentConfig
// returns true if any changes occurred in the deploymentConfig based on the dockerImage labels
func mergeImageMetadataWithDeploymentConfig(dc *appsv1.DeploymentConfig, dockerImage *dockerv10.DockerImage) bool {
	if !dockerImageHasLabels(dockerImage) {
		return false
	}

	log.Debugf("Preparing to read docker labels and add them to the Deployment: %s", dockerImage.Config.Labels)

	if dc.Spec.Template.Annotations == nil {
		dc.Spec.Template.Annotations = map[string]string{}
	}
	if dc.Labels == nil {
		dc.Labels = map[string]string{}
	}
	if dc.Spec.Selector == nil {
		dc.Spec.Selector = map[string]string{}
	}
	if dc.Spec.Template.Labels == nil {
		dc.Spec.Template.Labels = map[string]string{}
	}

	added := false
	for key, value := range dockerImage.Config.Labels {
		if strings.Contains(key, orgKieNamespaceLabelKey) {
			splitedKey := strings.Split(key, labelNamespaceSep)
			// we're only interested on keys like org.kie/something
			if len(splitedKey) > 1 {
				// persistence labels should be treated somewhere else
				if splitedKey[1] != orgKiePersistenceKey {
					importedKey := strings.Join(splitedKey[1:], labelNamespaceSep)

					if !added {
						_, lblPresent := dc.Labels[importedKey]
						_, selectorPresent := dc.Spec.Selector[importedKey]
						_, podLblPresent := dc.Spec.Template.Labels[importedKey]

						added = !(lblPresent && selectorPresent && podLblPresent)
					}

					dc.Labels[importedKey] = value
					dc.Spec.Selector[importedKey] = value
					dc.Spec.Template.Labels[importedKey] = value
				}
			}
		} else if strings.Contains(key, prometheusLabelKeyPrefix) {
			if !added {
				_, present := dc.Spec.Template.Annotations[key]
				added = !present
			}

			dc.Spec.Template.Annotations[key] = value
		}
	}

	return added
}

// discoverPortsAndProbesFromImage set Ports and Probes based on labels set on the DockerImage of this DeploymentConfig
func discoverPortsAndProbesFromImage(dc *appsv1.DeploymentConfig, dockerImage *dockerv10.DockerImage) {
	if !dockerImageHasLabels(dockerImage) {
		return
	}
	containerPorts := []corev1.ContainerPort{}
	var nonSecureProbe *corev1.Probe
	for key, value := range dockerImage.Config.Labels {
		if key == openshift.ImageLabelForExposeServices {
			services := strings.Split(value, dockerLabelServicesSep)
			for _, service := range services {
				ports := strings.Split(service, portSep)
				if len(ports) == 0 {
					log.Warnf(portFormatWrongMessage, service)
					continue
				}
				portNumber, err := strconv.Atoi(strings.Split(service, portSep)[0])
				if err != nil {
					log.Warnf(portFormatWrongMessage, service)
					continue
				}
				portName := ports[1]
				containerPorts = append(containerPorts, corev1.ContainerPort{Name: portName, ContainerPort: int32(portNumber), Protocol: corev1.ProtocolTCP})
				// we have at least one service exported using default HTTP protocols, let's used as a probe!
				if portName == defaultExportedProtocol {
					nonSecureProbe = defaultProbe
					nonSecureProbe.Handler.TCPSocket = &corev1.TCPSocketAction{Port: intstr.FromInt(portNumber)}
				}
			}
			break
		}
	}
	// set the ports we've found
	if len(containerPorts) != 0 {
		dc.Spec.Template.Spec.Containers[0].Ports = containerPorts
		if nonSecureProbe != nil {
			dc.Spec.Template.Spec.Containers[0].LivenessProbe = nonSecureProbe
			dc.Spec.Template.Spec.Containers[0].ReadinessProbe = nonSecureProbe
		}
	}
}
