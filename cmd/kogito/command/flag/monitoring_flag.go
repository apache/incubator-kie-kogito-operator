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
	"github.com/kiegroup/kogito-cloud-operator/pkg/apis/app/v1alpha1"
	"github.com/spf13/cobra"
)

// MonitoringFlags is common properties used to configure Monitoring service
type MonitoringFlags struct {
	Scrape bool
	Scheme string
	Path   string
}

// AddMonitoringFlags adds the monitoring flags to the given command
func AddMonitoringFlags(command *cobra.Command, flags *MonitoringFlags) {
	command.Flags().BoolVar(&flags.Scrape, "scrape", false, "Flag to allow Kogito operator to expose Kogito service for scraping")
	command.Flags().StringVar(&flags.Scheme, "monitoring-scheme", v1alpha1.MonitoringDefaultScheme, "HTTP scheme to use for scraping.")
	command.Flags().StringVar(&flags.Path, "monitoring-path", v1alpha1.MonitoringDefaultPath, "HTTP path to scrape for metrics")
}

// CheckMonitoringArgs validates the MonitoringFlags flags
func CheckMonitoringArgs(flags *MonitoringFlags) error {
	return nil
}
