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
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/api/v1beta1"
	util2 "github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/util"
	"github.com/kiegroup/kogito-cloud-operator/pkg/util"
	"github.com/spf13/cobra"
)

var (
	validWebHookTypes = []string{string(v1beta1.GitHubWebHook), string(v1beta1.GenericWebHook)}
)

// WebHookFlags is common properties used to configure Git
type WebHookFlags struct {
	WebHook []string
}

// AddWebHookFlags adds the WebHook flags to the given command
func AddWebHookFlags(command *cobra.Command, flags *WebHookFlags) {
	command.Flags().StringArrayVar(&flags.WebHook, "web-hook", nil, "Secrets for source to image builds based on Git repositories (Remote Sources). For example 'WEB_HOOK_TYPE=WEB_HOOT_SECRET'. Can be set more than once.")
}

// CheckWebHookArgs validates the WebHookFlags flags
func CheckWebHookArgs(flags *WebHookFlags) error {
	webHookMap := util2.FromStringsKeyPairToMap(flags.WebHook)
	for webHookType := range webHookMap {
		if !util.Contains(webHookType, validWebHookTypes) {
			return fmt.Errorf("WebHook type not valid. Valid types are %s. Received %s", validWebHookTypes, webHookType)
		}
	}
	return nil
}
