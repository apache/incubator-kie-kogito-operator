// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package converter

import (
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/flag"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// FromPodResourceFlagsToResourceRequirement converts given PodResourceFlags into ResourceRequirements
func FromPodResourceFlagsToResourceRequirement(flags *flag.PodResourceFlags) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Limits:   FromStringArrayToResources(flags.Limits),
		Requests: FromStringArrayToResources(flags.Requests),
	}
}

// FromStringArrayToResources ...
func FromStringArrayToResources(strings []string) corev1.ResourceList {
	if strings == nil {
		return nil
	}
	res := corev1.ResourceList{}
	mapStr := util.FromStringsKeyPairToMap(strings)
	for k, v := range mapStr {
		res[corev1.ResourceName(k)] = resource.MustParse(v)
	}
	return res
}
