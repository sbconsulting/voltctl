/*
 * Copyright 2019-present Ciena Corporation
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package commands

import (
	"context"
	"fmt"
	"github.com/ciena/voltctl/format"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/opencord/voltha/protos/go/voltha"
	"google.golang.org/grpc"
	"strings"
	"time"
)

type OutputType uint8

const (
	OUTPUT_TABLE OutputType = iota
	OUTPUT_JSON
	OUTPUT_YAML
)

var CharReplacer = strings.NewReplacer("\\t", "\t", "\\n", "\n")

type OutputOptions struct {
	Format   string `long:"format" default:"" description:"Format to use to output structured data"`
	Quiet    bool   `short:"q" long:"quiet" description:"Output only the IDs of the objects"`
	OutputAs string `short:"o" long:"outputas" default:"table" choice:"table" choice:"json" choice:"yaml" description:"Type of output to generate"`
}

type CommandNotFound struct {
	command string
}

func (c *CommandNotFound) Error() string {
	return fmt.Sprintf("Unable to locate command '%s'", c.command)
}

func toOutputType(in string) OutputType {
	switch in {
	case "table":
		fallthrough
	default:
		return OUTPUT_TABLE
	case "json":
		return OUTPUT_JSON
	case "yaml":
		return OUTPUT_YAML
	}
}

type CommandResult struct {
	Format   format.Format
	OutputAs OutputType
	Data     interface{}
}

type CommandContext struct {
	Path    []string
	Args    []string
	Command *Command
}

type Command struct {
	Invoke      func(conn *grpc.ClientConn, context *CommandContext) (*CommandResult, error)
	subcommands map[string]*Command
}

func noCommand(conn *grpc.ClientConn, command *CommandContext) (*CommandResult, error) {
	return nil, fmt.Errorf("Unknown command specified")
}

func showVersion(conn *grpc.ClientConn, command *CommandContext) (*CommandResult, error) {
	client := voltha.NewVolthaGlobalServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	got, err := client.GetVoltha(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	fmt.Println("Client Version: beta")
	fmt.Printf("Cluster Version: %s", got.Version)
	return nil, nil
}

var Commands = map[string]*Command{
	"version": {
		Invoke:      showVersion,
		subcommands: nil,
	},
	"adapter": {
		Invoke: noCommand,
		subcommands: map[string]*Command{
			"list": {
				Invoke:      listAllAdapters,
				subcommands: nil,
			},
		},
	},
	"device": {
		Invoke: noCommand,
		subcommands: map[string]*Command{
			"create": {
				Invoke:      createDevice,
				subcommands: nil,
			},
			"delete": {
				Invoke:      deleteDevice,
				subcommands: nil,
			},
			"list": {
				Invoke:      listAllDevices,
				subcommands: nil,
			},
			"enable": {
				Invoke:      enableDevice,
				subcommands: nil,
			},
			"disable": {
				Invoke:      disableDevice,
				subcommands: nil,
			},
		},
	},
}

func lookupCommand(context *CommandContext) error {
	var cmd *Command
	var ok bool

	if context.Args == nil || len(context.Args) == 0 {
		return nil
	}

	if context.Command == nil {
		cmd, ok = Commands[context.Args[0]]
		if !ok {
			return &CommandNotFound{command: context.Args[0]}
		}
	} else {
		cmd, ok = context.Command.subcommands[context.Args[0]]
		if !ok {
			return nil
		}
	}

	context.Path = append(context.Path, context.Args[0])
	context.Args = context.Args[1:]
	context.Command = cmd

	return lookupCommand(context)
}

func LookupCommand(args []string) (*CommandContext, error) {
	context := CommandContext{
		Path:    make([]string, 0),
		Args:    args,
		Command: nil,
	}
	err := lookupCommand(&context)
	if err != nil {
		return nil, err
	}
	return &context, nil
}
