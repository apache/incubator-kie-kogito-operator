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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/infrastructure"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// DefaultImage is the default image name for the Kogito Data Index Service
	DefaultImage = "quay.io/kiegroup/kogito-data-index:latest"
	// defaultLabelKey is the default label key that should be added to all resources
	defaultLabelKey = "app"
	defaultReplicas = 1
	// defaultExposedPort is the default port exposed by the service.
	// this port can also be found into the docker label openshift.exposed-svc.
	// since we're aiming for cluster agnostic, the image API is out of question.
	// TODO: found an agnostic API to fetch the ImageRaw from the docker image and read this value from there.
	defaultExposedPort       = 8080
	defaultProtobufMountPath = "/home/kogito/data/protobufs"
)

// Collection of Infinispan/Kafka Environment Variables that need to be set in the Data Index image
const (
	kafkaEnvKeyProcessInstancesServer string = "MP_MESSAGING_INCOMING_KOGITO_PROCESSINSTANCES_EVENTS_BOOTSTRAP_SERVERS"
	kafkaEnvKeyUserTaskInstanceServer string = "MP_MESSAGING_INCOMING_KOGITO_USERTASKINSTANCES_EVENTS_BOOTSTRAP_SERVERS"
	kafkaEnvKeyProcessDomainServer    string = "MP_MESSAGING_INCOMING_KOGITO_PROCESSDOMAIN_EVENTS_BOOTSTRAP_SERVERS"
	kafkaEnvKeyUserTaskDomainServer   string = "MP_MESSAGING_INCOMING_KOGITO_USERTASKDOMAIN_EVENTS_BOOTSTRAP_SERVERS"

	kafkaTopicNameProcessInstances  string = "kogito-processinstances-events"
	kafkaTopicNameUserTaskInstances string = "kogito-usertaskinstances-events"
	kafkaTopicNameProcessDomain     string = "kogito-processdomain-events"
	kafkaTopicNameUserTaskDomain    string = "kogito-usertaskdomain-events"

	protoBufKeyFolder string = "KOGITO_PROTOBUF_FOLDER"
	protoBufKeyWatch  string = "KOGITO_PROTOBUF_WATCH"
)

var protoBufEnvs = map[string]string{
	protoBufKeyFolder: defaultProtobufMountPath,
	protoBufKeyWatch:  "true",
}

// managedEnvKeys are a collection of reserved keys
var managedEnvKeys []string

var protoBufKeys = []string{
	protoBufKeyFolder,
	protoBufKeyWatch,
}

var managedKafkaKeys = []string{
	kafkaEnvKeyProcessInstancesServer,
	kafkaEnvKeyUserTaskInstanceServer,
	kafkaEnvKeyProcessDomainServer,
	kafkaEnvKeyUserTaskDomainServer,
}

var kafkaTopicNames = []string{
	kafkaTopicNameProcessInstances,
	kafkaTopicNameUserTaskInstances,
	kafkaTopicNameProcessDomain,
	kafkaTopicNameUserTaskDomain,
}

var defaultAnnotations = map[string]string{
	"org.kie.kogito/managed-by":   "Kogito Operator",
	"org.kie.kogito/operator-crd": "KogitoDataIndex",
}

var defaultProbe = &corev1.Probe{
	TimeoutSeconds:   int32(1),
	PeriodSeconds:    int32(10),
	SuccessThreshold: int32(1),
	FailureThreshold: int32(3),
}

func addDefaultMetadata(objectMeta *metav1.ObjectMeta, instance *v1alpha1.KogitoDataIndex) {
	if objectMeta != nil {
		if objectMeta.Annotations == nil {
			objectMeta.Annotations = map[string]string{}
		}
		if objectMeta.Labels == nil {
			objectMeta.Labels = map[string]string{}
		}
		for key, value := range defaultAnnotations {
			objectMeta.Annotations[key] = value
		}
		objectMeta.Labels[defaultLabelKey] = instance.Name
	}
}

func init() {
	managedEnvKeys = append(managedEnvKeys, protoBufKeys...)
	managedEnvKeys = append(managedEnvKeys, infrastructure.GetInfinispanEnvVarsKeys()...)
	managedEnvKeys = append(managedEnvKeys, managedKafkaKeys...)
}
