// Copyright 2019 Red Hat, Inc. and/or its affiliates
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

package context

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/version"
	"github.com/spf13/cobra"
	"io"
)

var (
	rootCmd = &rootCommand{}
	ctx     = &CommandContext{}
)

type rootCommandFlags struct {
	cfgFile string
}

type rootCommand struct {
	CommandContext
	command *cobra.Command
	flags   rootCommandFlags
}

// NewRootCommand is the constructor for the root command
func NewRootCommand(commandContext *CommandContext, output io.Writer) KogitoCommand {
	commandOutput = output
	ctx = commandContext
	rootCmd = &rootCommand{CommandContext: *ctx}
	rootCmd.RegisterHook()
	rootCmd.InitHook()

	return rootCmd
}

func (i *rootCommand) ConfigFile() string {
	return i.flags.cfgFile
}

func (i *rootCommand) Command() *cobra.Command {
	return i.command
}

func (i *rootCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:   "kogito",
		Short: "Kogito CLI",
		Long:  `Kogito CLI deploys your Kogito Services into an OpenShift cluster`,
	}
}

func (i *rootCommand) InitHook() {
	i.flags = rootCommandFlags{}
	i.command.PersistentFlags().StringVarP(&outputFormat, "output", "o", "", "output format (when defined, 'json' is supported)")
	i.command.PersistentFlags().BoolVarP(&logVerbose, "verbose", "v", false, "verbose output")
	i.command.PersistentFlags().Bool("version", false, "display version")
	i.command.Version = version.Version
	i.command.SetVersionTemplate("{{with .Name}}{{printf \"%s \" .}}{{end}}{{printf \"%s\" .Version}}\n")
}
