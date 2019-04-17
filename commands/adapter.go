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

	"context"
	"github.com/ciena/voltctl/format"
	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltha/protos/go/voltha"
)

const (
	DEFAULT_OUTPUT_FORMAT = "table{{ .Id }}\t{{.Vendor}}\t{{.Version}}"
)

type AdapterList struct {
	OutputOptions
}

type AdapterOpts struct {
	List AdapterList `command:"list"`
}

var adapterOpts = AdapterOpts{}

func RegisterAdapterCommands(parent *flags.Parser) {
	parent.AddCommand("adapter", "adapter commands", "Commands to query and manipulate VOLTHA adapters", &adapterOpts)
}

func (options *AdapterList) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaGlobalServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	adapters, err := client.ListAdapters(ctx, &empty.Empty{})
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_OUTPUT_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	result := CommandResult{
		Format:   format.Format(outputFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     adapters.Items,
	}
	GenerateOutput(&result)

	return nil
}
