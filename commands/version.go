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
	"github.com/ciena/voltctl/format"
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
	"strings"
)

type VersionOutput struct {
	Client  string `json:"client"`
	Cluster string `json:"cluster"`
}

type VersionOpts struct {
	OutputAs string `short:"o" long:"outputas" default:"table" choice:"table" choice:"json" choice:"yaml" description:"Type of output to generate"`
}

var versionOpts = VersionOpts{}

func RegisterVersionCommands(parent *flags.Parser) {
	parent.AddCommand("version", "display version", "Display client and server version", &versionOpts)
}

func (options *VersionOpts) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	descriptor, method, err := GetMethod("version")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	h := &RpcEventHandler{}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
	if err != nil {
		return err
	}

	if h.Status != nil && h.Status.Err() != nil {
		return h.Status.Err()
	}

	d, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}
	version, err := d.TryGetFieldByName("version")
	if err != nil {
		return err
	}

	result := CommandResult{
		Format:   format.Format("Client Version: {{.Client}}\nCluster Version: {{.Cluster}}"),
		OutputAs: toOutputType(options.OutputAs),
		Data:     VersionOutput{Client: "beta", Cluster: strings.ReplaceAll(version.(string), "\n", "")},
	}

	GenerateOutput(&result)
	return nil
}
