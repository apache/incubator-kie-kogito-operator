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
	"github.com/kiegroup/kogito-cloud-operator/core/api"
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
	command.Flags().StringVar(&flags.Scheme, "monitoring-scheme", api.MonitoringDefaultScheme, "HTTP scheme to use for scraping.Default is "+api.MonitoringDefaultScheme)
	command.Flags().StringVar(&flags.Path, "monitoring-path", api.MonitoringDefaultPath, "HTTP path to scrape for metrics. Default is "+api.MonitoringDefaultPath)
}

// CheckMonitoringArgs validates the MonitoringFlags flags
func CheckMonitoringArgs(flags *MonitoringFlags) error {
	return nil
}
