// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package command

import (
	"io"
	"os"

	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/completion"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/context"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/deploy"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/install"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/project"
	"github.com/apache/incubator-kie-kogito-operator/cmd/kogito/command/remove"
	"github.com/apache/incubator-kie-kogito-operator/core/client"
	"github.com/apache/incubator-kie-kogito-operator/meta"
	"github.com/spf13/cobra"
)

// DefaultBuildCommands creates a new start command for the Kogito CLI
func DefaultBuildCommands() *cobra.Command {
	return BuildCommands(client.NewForConsole(meta.GetRegisteredSchema()), os.Stdout)
}

// BuildCommands creates a customized start command for the Kogito CLI
func BuildCommands(kubeClient *client.Client, output io.Writer) *cobra.Command {
	ctx := &context.CommandContext{Client: kubeClient}

	rootCommand := context.NewRootCommand(ctx, output)
	completion.BuildCommands(ctx, rootCommand.Command())
	deploy.BuildCommands(ctx, rootCommand.Command())
	install.BuildCommands(ctx, rootCommand.Command())
	remove.BuildCommands(ctx, rootCommand.Command())
	project.BuildCommands(ctx, rootCommand.Command())

	return rootCommand.Command()
}
