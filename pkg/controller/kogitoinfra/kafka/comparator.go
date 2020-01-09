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

package kafka

import (
	"github.com/RHsyseng/operator-utils/pkg/resource"
	kafkabetav1 "github.com/kiegroup/kogito-cloud-operator/pkg/apis/kafka/v1beta1"
	"github.com/kiegroup/kogito-cloud-operator/pkg/framework"
	"reflect"
)

// GetComparators gets the comparator for Kafka resources
func GetComparators() []framework.Comparator {
	return []framework.Comparator{createKafkaComparator()}
}

func createKafkaComparator() framework.Comparator {
	return framework.Comparator{
		ResourceType: reflect.TypeOf(kafkabetav1.Kafka{}),
		CompFunc: func(deployed resource.KubernetesResource, requested resource.KubernetesResource) bool {
			kafkaDep := deployed.(*kafkabetav1.Kafka)
			kafkaReq := requested.(*kafkabetav1.Kafka).DeepCopy()
			// we just care for the instance name, other attributes can be changed at will by the user
			return reflect.DeepEqual(kafkaDep.Name, kafkaReq.Name)
		},
	}
}
