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

package v1alpha1

const (

	// MonitoringDefaultPath default path
	MonitoringDefaultPath = "/metrics"

	// MonitoringDefaultScheme default scheme
	MonitoringDefaultScheme = "http"
)

// Monitoring properties to connect with Monitoring service
type Monitoring struct {
	// Flag to allow Monitoring to scraping Kogito service.
	Scrape bool `json:"scrape,omitempty"`

	// HTTP scheme to use for scraping.
	// +optional
	Scheme string `json:"scheme,omitempty"`

	// HTTP path to scrape for metrics.
	// +optional
	Path string `json:"path,omitempty"`
}
