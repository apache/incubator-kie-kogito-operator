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

package completion

import (
	"fmt"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/spf13/cobra"
	"os"
)

type completionCommand struct {
	context.CommandContext
	command *cobra.Command
	Parent  *cobra.Command
}

// initCompletionCommand is the constructor for the completion command
func initCompletionCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := &completionCommand{CommandContext: *ctx, Parent: parent}
	cmd.RegisterHook()
	cmd.InitHook()
	return cmd
}

func (i *completionCommand) Command() *cobra.Command {
	return i.command
}

func (i *completionCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:     "completion (bash | zsh)",
		Short:   "Generates a completion script for the given shell (bash or zsh)",
		Aliases: []string{"comp"},
		Long: `Description:
  Generates a completion script for the given shell (bash or zsh)

Bash:
  To load in the current session:
  . <(kogito completion bash)

  To load in all new sessions:
  echo ". <(kogito completion bash)" >> ~/.bashrc

  To load in all new sessions for all users:
  kogito completion bash | sudo tee /etc/bash_completion.d/kogito

Zsh:
  To load in the current session:
  . <(kogito completion zsh); compdef _kogito kogito

  To load in all new sessions:
  echo ". <(kogito completion zsh); compdef _kogito kogito" >> ~/.zshrc

  To load in all new sessions for all users:
  kogito completion zsh | sudo tee /usr/share/zsh/site-functions/_kogito
        `,
		RunE: i.Exec,
		// Args validation
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg, received %v", len(args))
			}
			if args[0] != "bash" && args[0] != "zsh" {
				return fmt.Errorf("argument must be 'bash' or 'zsh', received %s", args[0])
			}
			return nil
		},
	}
}

func (i *completionCommand) InitHook() {
	i.Parent.AddCommand(i.command)
}

func (i *completionCommand) Exec(cmd *cobra.Command, args []string) error {
	shell := args[0]

	if shell == "bash" {
		cmd.Root().GenBashCompletion(os.Stdout)
	} else if shell == "zsh" {
		cmd.Root().GenZshCompletion(os.Stdout)
	}

	return nil
}
