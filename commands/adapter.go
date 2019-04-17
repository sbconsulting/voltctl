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
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"

	"context"
	"github.com/ciena/voltctl/format"
	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltha/protos/go/voltha"
	"os"
	"strings"
	"time"
)

const (
	DEFAULT_OUTPUT_FORMAT = "table{{ .Id }}\t{{.Vendor}}\t{{.Version}}"
)

func listAllAdapters(conn *grpc.ClientConn, command *CommandContext) (*CommandResult, error) {
	outputOpts := OutputOptions{}

	parser := flags.NewNamedParser(strings.Join(command.Path, " "), flags.Default)
	_, err := parser.AddGroup("Command Options", "", &outputOpts)
	if err != nil {
		panic(err)
	}

	_, err = parser.ParseArgs(command.Args)
	if err != nil {
		real := err.(*flags.Error)
		if real.Type == flags.ErrHelp {
			return nil, nil
		}
		os.Exit(1)
	}

	client := voltha.NewVolthaGlobalServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	adapters, err := client.ListAdapters(ctx, &empty.Empty{})
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	outputFormat := CharReplacer.Replace(outputOpts.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_OUTPUT_FORMAT
	}
	if outputOpts.Quiet {
		outputFormat = "{{.Id}}"
	}

	result := CommandResult{
		Format:   format.Format(outputFormat),
		OutputAs: toOutputType(outputOpts.OutputAs),
		Data:     adapters.Items,
	}

	return &result, nil
}
