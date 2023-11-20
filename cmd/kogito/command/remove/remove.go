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

package remove

import (
	"github.com/kiegroup/kogito-operator/cmd/kogito/command/context"
	"github.com/spf13/cobra"
)

type removeCommand struct {
	context.CommandContext
	command *cobra.Command
	Parent  *cobra.Command
}

func initRemoveCommand(ctx *context.CommandContext, parent *cobra.Command) context.KogitoCommand {
	cmd := removeCommand{
		CommandContext: *ctx,
		Parent:         parent,
	}
	cmd.RegisterHook()
	cmd.InitHook()
	return &cmd
}

func (i *removeCommand) Command() *cobra.Command {
	return i.command
}

func (i *removeCommand) RegisterHook() {
	i.command = &cobra.Command{
		Use:    "remove",
		Short:  "remove all sorts of infrastructure components from your Kogito project",
		PreRun: i.CommonPreRun,
	}
}

func (i *removeCommand) InitHook() {
	i.Parent.AddCommand(i.command)
}
