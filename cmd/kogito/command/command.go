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

package command

import (
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/context"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/deploy"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/install"
	"github.com/kiegroup/kogito-cloud-operator/cmd/kogito/command/project"
	"github.com/kiegroup/kogito-cloud-operator/pkg/client"
	"github.com/spf13/cobra"
	"io"
)

// DefaultBuildCommands will create a new start command for the Kogito CLI
func DefaultBuildCommands() *cobra.Command {
	return BuildCommands(&client.Client{}, nil)
}

// BuildCommands will create a customized start command for the Kogito CLI
func BuildCommands(kubeClient *client.Client, output io.Writer) *cobra.Command {
	ctx := &context.CommandContext{Client: kubeClient}

	rootCommand := context.NewRootCommand(ctx, output)
	deploy.BuildCommands(ctx, rootCommand.Command())
	install.BuildCommands(ctx, rootCommand.Command())
	project.BuildCommands(ctx, rootCommand.Command())

	return rootCommand.Command()
}
