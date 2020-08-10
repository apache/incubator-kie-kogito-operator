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

package flag

import (
	"github.com/spf13/cobra"
)

// PrometheusFlags is common properties used to configure Prometheus
type PrometheusFlags struct {
	Scrape bool
	Scheme string
	Path   string
}

// AddPrometheusFlags adds the prometheus flags to the given command
func AddPrometheusFlags(command *cobra.Command, flags *PrometheusFlags) {
	command.Flags().BoolVar(&flags.Scrape, "prome-scrape", false, "Flag to allow Prometheus to scraping Kogito service")
	command.Flags().StringVar(&flags.Scheme, "prome-scheme", "http", "HTTP scheme to use for scraping.")
	command.Flags().StringVar(&flags.Path, "prome-path", "/metrics", "HTTP path to scrape for metrics")
}

// CheckPrometheusArgs validates the PrometheusFlags flags
func CheckPrometheusArgs(flags *PrometheusFlags) error {
	return nil
}
