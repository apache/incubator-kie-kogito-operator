/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package flag

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
)

// GitSourceFlags is common properties used to configure Git
type GitSourceFlags struct {
	Reference  string
	ContextDir string
	Source     string
}

// AddGitSourceFlags adds the Git source flags to the given command
func AddGitSourceFlags(command *cobra.Command, flags *GitSourceFlags) {
	command.Flags().StringVarP(&flags.Reference, "branch", "b", "", "Git branch to use in the git repository")
	command.Flags().StringVarP(&flags.ContextDir, "context-dir", "c", "", "Context/subdirectory where the code is located, relatively to repository root")
}

// CheckGitSourceArgs validates the GitSourceFlags flags
func CheckGitSourceArgs(flags *GitSourceFlags) error {
	if len(flags.Source) > 0 {
		u, err := url.Parse(flags.Source)
		if err != nil {
			return err
		}
		if u.Scheme == "" || u.Host == "" || u.Path == "" {
			return fmt.Errorf("provided Git Source is not valid: %s", flags.Source)
		}
	}
	return nil

}
