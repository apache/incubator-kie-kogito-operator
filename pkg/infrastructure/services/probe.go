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

package services

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// HealthCheckProbeType defines the supported probes for the ServiceDefinition
type HealthCheckProbeType string

const (
	// QuarkusHealthCheckProbe probe implemented with Quarkus Microprofile Health. See: https://quarkus.io/guides/microprofile-health.
	// the operator will set the probe to the default path /health/live and /health/ready for liveness and readiness probes, respectively.
	QuarkusHealthCheckProbe HealthCheckProbeType = "quarkus"
	// TCPHealthCheckProbe default health check probe that binds to port 8080
	TCPHealthCheckProbe HealthCheckProbeType = "TCP"

	quarkusProbeLivenessPath  = "/health/live"
	quarkusProbeReadinessPath = "/health/ready"
)

type healthCheckProbe struct {
	readiness *corev1.Probe
	liveness  *corev1.Probe
}

// getProbeForKogitoService gets the appropriate liveness (index 0) and readiness (index 1) probes based on the given service definition
func getProbeForKogitoService(serviceDefinition ServiceDefinition, httpPort int32) healthCheckProbe {
	switch serviceDefinition.HealthCheckProbe {
	case QuarkusHealthCheckProbe:
		return healthCheckProbe{
			readiness: getQuarkusHealthCheckReadiness(httpPort),
			liveness:  getQuarkusHealthCheckLiveness(httpPort),
		}
	case TCPHealthCheckProbe:
		return healthCheckProbe{
			readiness: getTCPHealthCheckProbe(httpPort),
			liveness:  getTCPHealthCheckProbe(httpPort),
		}
	default:
		return healthCheckProbe{
			readiness: getTCPHealthCheckProbe(httpPort),
			liveness:  getTCPHealthCheckProbe(httpPort),
		}
	}
}

func getTCPHealthCheckProbe(httpPort int32) *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			TCPSocket: &corev1.TCPSocketAction{Port: intstr.IntOrString{IntVal: httpPort}},
		},
		TimeoutSeconds:   int32(1),
		PeriodSeconds:    int32(10),
		SuccessThreshold: int32(1),
		FailureThreshold: int32(3),
	}
}

func getQuarkusHealthCheckLiveness(httpPort int32) *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   quarkusProbeLivenessPath,
				Port:   intstr.IntOrString{IntVal: httpPort},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		TimeoutSeconds:   int32(1),
		PeriodSeconds:    int32(10),
		SuccessThreshold: int32(1),
		FailureThreshold: int32(3),
	}
}

func getQuarkusHealthCheckReadiness(httpPort int32) *corev1.Probe {
	return &corev1.Probe{
		Handler: corev1.Handler{
			HTTPGet: &corev1.HTTPGetAction{
				Path:   quarkusProbeReadinessPath,
				Port:   intstr.IntOrString{IntVal: httpPort},
				Scheme: corev1.URISchemeHTTP,
			},
		},
		TimeoutSeconds:   int32(1),
		PeriodSeconds:    int32(10),
		SuccessThreshold: int32(1),
		FailureThreshold: int32(3),
	}
}
