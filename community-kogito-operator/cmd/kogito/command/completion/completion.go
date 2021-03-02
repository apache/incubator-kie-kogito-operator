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
	"github.com/kiegroup/kogito-cloud-operator/community-kogito-operator/cmd/kogito/command/context"
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
		Use:     "completion (bash | zsh | fish)",
		Short:   "Generates a completion script for the given shell (bash, zsh or fish)",
		Aliases: []string{"comp"},
		Long: `Description:
  Generates a completion script for the given shell (bash, zsh or fish)

  Bash:
  
    $ source <(kogito completion bash)
    
    # To load completions for each session, execute once:
    Linux:
      $ kogito completion bash > /etc/bash_completion.d/kogito
    MacOS:
      $ kogito completion bash > /usr/local/etc/bash_completion.d/kogito
  
  Zsh:
  
    # If shell completion is not already enabled in your environment you will need
    # to enable it.  You can execute the following once:
    
    $ echo "autoload -U compinit; compinit" >> ~/.zshrc
    
    # To load completions for each session, execute once:
    $ kogito completion zsh > "${fpath[1]}/_kogito"
    
    # You will need to start a new shell for this setup to take effect.
  
  Fish:
  
    $ kogito completion fish | source
    
    # To load completions for each session, execute once:
    $ kogito completion fish > ~/.config/fish/completions/kogito.fish
        `,
		RunE: i.Exec,
		// Args validation
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("requires 1 arg, received %v", len(args))
			}
			if args[0] != "bash" && args[0] != "zsh" && args[0] != "fish" {
				return fmt.Errorf("argument must be 'bash', 'zsh' or 'fish', received %s", args[0])
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
	var err error

	switch shell {
	case "bash":
		err = cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		err = cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		err = cmd.Root().GenFishCompletion(os.Stdout, true)
	}
	if err != nil {
		return fmt.Errorf("Error in creating %s completion file: %v", shell, err)
	}

	return nil
}
