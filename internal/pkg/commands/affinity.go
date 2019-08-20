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
	"github.com/ciena/voltctl/pkg/format"
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
)

type GoroutinecountOutput struct {
	count uint32
}

type GoroutinecountOpts struct {
	OutputOptions
}

type AffinityOpts struct {
	Goroutinecount GoroutinecountOpts `command:"goroutinecount"`
}

var affinityOpts = AffinityOpts{}

var goroutinecountInfo = GoroutinecountOutput{}

func RegisterAffinityCommands(parent *flags.Parser) {
	_, err := parent.AddCommand("affinity", "affinity router stuff", "Affinity Router Stuff", &affinityOpts)
	if err != nil {
		panic(err)
	}
}

const DefaultGoroutinecountFormat = `{{.count}}`

func (options *GoroutinecountOpts) Execute(args []string) error {
	conn, err := NewAffinityConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	descriptor, method, err := GetMethod("get-goroutine-count")
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
	count, err := d.TryGetFieldByName("count")
	if err != nil {
		return err
	}

	goroutinecountInfo.count = count.(uint32)

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DefaultGoroutinecountFormat
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      goroutinecountInfo,
	}

	GenerateOutput(&result)
	return nil
}
