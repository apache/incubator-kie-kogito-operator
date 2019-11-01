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
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/kiegroup/kogito-cloud-operator/pkg/client/openshift"

	appsv1 "github.com/openshift/api/apps/v1"
	dockerv10 "github.com/openshift/api/image/docker10"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

const (
	// DefaultExportedPort is the default protocol exposed by inner services specified in image metadata
	DefaultExportedPort = "http"
	// LabelKeyPrometheus is the label key for Prometheus metadata
	LabelKeyPrometheus = "prometheus.io"
	// LabelKeyOrgKie is the label key for KIE metadata
	LabelKeyOrgKie = "org.kie" + labelNamespaceSep
	// LabelKeyOrgKiePersistence is the label key for Persistence metadata
	LabelKeyOrgKiePersistence = "org.kie" + labelNamespaceSep + "persistence"
	// LabelKeyOrgKieProtoBuf is the label key for ProtoBuf metadata
	LabelKeyOrgKieProtoBuf = "org.kie" + labelNamespaceSep + "persistence" + labelNamespaceSep + "proto"

	labelNamespaceSep               = "/"
	dockerLabelServicesSep, portSep = ",", ":"
	portFormatWrongMessage          = "Service %s on " + openshift.ImageLabelForExposeServices + " label in wrong format. Won't be possible to expose Services for this application. Should be PORT_NUMBER:PROTOCOL. e.g. 8080:http"
)

var defaultProbe = &corev1.Probe{
	TimeoutSeconds:   int32(1),
	PeriodSeconds:    int32(10),
	SuccessThreshold: int32(1),
	FailureThreshold: int32(3),
}

func dockerImageHasLabels(dockerImage *dockerv10.DockerImage) bool {
	if dockerImage == nil || dockerImage.Config == nil || dockerImage.Config.Labels == nil {
		return false
	}
	return true
}

// ExtractProtoBufFilesFromDockerImage will extract the protobuf files from the DockerImage labels prefixed with
func ExtractProtoBufFilesFromDockerImage(prefixKey string, dockerImage *dockerv10.DockerImage) map[string]string {
	files := map[string]string{}
	if !dockerImageHasLabels(dockerImage) {
		return files
	}
	for key, value := range dockerImage.Config.Labels {
		if strings.Contains(key, LabelKeyOrgKieProtoBuf) {
			splitKey := strings.Split(key, labelNamespaceSep)
			fileName := fmt.Sprintf("%s-%s", prefixKey, splitKey[len(splitKey)-1])
			if fileContent, err := decompressBase64GZip(value); err != nil && len(fileContent) == 0 {
				log.Errorf("Error while trying to read file %s from image label: %s", fileName, err)
			} else {
				files[fileName] = fileContent
			}
		}
	}

	return files
}

func decompressBase64GZip(contents string) (string, error) {
	var decode []byte
	var err error
	var reader *gzip.Reader
	defer func() {
		if reader != nil {
			reader.Close()
		}
	}()
	if decode, err = base64.StdEncoding.DecodeString(contents); err != nil {
		return "", fmt.Errorf("Error while converting contents from base64: %s", err)
	}
	if reader, err = gzip.NewReader(bytes.NewReader(decode)); err != nil {
		// the file might not being compressed, we should support old versions where the labels are not compressed
		err = fmt.Errorf("Error while decompressing contents: %s", err)
		if strings.Contains(err.Error(), "invalid header") {
			return string(decode), err
		}
		return "", err
	}
	if decode, err = ioutil.ReadAll(reader); err != nil {
		return "", fmt.Errorf("Error while reading contents after decompressing: %s", err)
	}
	return string(decode), nil
}

// MergeImageMetadataWithDeploymentConfig retrieves org.kie and prometheus.io labels from DockerImage and adds them to the DeploymentConfig
// returns true if any changes occurred in the deploymentConfig based on the dockerImage labels
func MergeImageMetadataWithDeploymentConfig(dc *appsv1.DeploymentConfig, dockerImage *dockerv10.DockerImage) bool {
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
		if strings.Contains(key, LabelKeyOrgKie) && !strings.Contains(key, LabelKeyOrgKiePersistence) {
			splitedKey := strings.Split(key, labelNamespaceSep)
			// we're only interested on keys like org.kie/something
			if len(splitedKey) > 1 {
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
		} else if strings.Contains(key, LabelKeyPrometheus) {
			if !added {
				_, present := dc.Spec.Template.Annotations[key]
				added = !present
			}

			dc.Spec.Template.Annotations[key] = value
		}
	}

	return added
}

// DiscoverPortsAndProbesFromImage set Ports and Probes based on labels set on the DockerImage of this DeploymentConfig
func DiscoverPortsAndProbesFromImage(dc *appsv1.DeploymentConfig, dockerImage *dockerv10.DockerImage) {
	if !dockerImageHasLabels(dockerImage) {
		return
	}
	var containerPorts []corev1.ContainerPort
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
				if portName == DefaultExportedPort {
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
