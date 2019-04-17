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
	"os"
	"strings"

	"context"
	"fmt"
	"github.com/ciena/voltctl/format"
	flags "github.com/jessevdk/go-flags"
	"github.com/opencord/voltha/protos/go/common"
	"github.com/opencord/voltha/protos/go/voltha"
	"time"
)

const (
	DEFAULT_DEVICE_FORMAT = "table{{ .Id }}\t{{.Type}}\t{{.Root}}\t{{.ParentId}}\t{{.SerialNumber}}\t{{.Vlan}}\t{{.AdminState}}\t{{.OperStatus}}\t{{.ConnectStatus}}"
)

func listAllDevices(conn *grpc.ClientConn, command *CommandContext) (*CommandResult, error) {
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

	adapters, err := client.ListDevices(ctx, &empty.Empty{})
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
		outputFormat = DEFAULT_DEVICE_FORMAT
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

type CreateDeviceOptions struct {
	DeviceType  string `short:"t" long:"devicetype" default:"simulated_olt" description:"Device type"`
	MACAddress  string `short:"m" long:"macaddress" default:"00:0c:e2:31:40:00" description:"MAC Address"`
	IPAddress   string `short:"i" long:"ipaddress" default:"" description:"IP Address"`
	HostAndPort string `short:"H" long:"hostandport" default:"" description:"Host and port"`
}

func createDevice(conn *grpc.ClientConn, command *CommandContext) (*CommandResult, error) {

	options := CreateDeviceOptions{}

	parser := flags.NewNamedParser(strings.Join(command.Path, " "), flags.Default)
	_, err := parser.AddGroup("Command Options", "", &options)
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
	device := voltha.Device{}

	if options.HostAndPort != "" {
		device.Address = &voltha.Device_HostAndPort{
			HostAndPort: options.HostAndPort,
		}
	} else if options.IPAddress != "" {
		device.Address = &voltha.Device_Ipv4Address{
			Ipv4Address: options.IPAddress,
		}
	} else if options.MACAddress != "" {
		device.Address = &voltha.Device_MacAddress{
			MacAddress: strings.ToLower(options.MACAddress),
		}
	}
	if options.DeviceType != "" {
		device.Type = options.DeviceType
	}

	client := voltha.NewVolthaGlobalServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	got, err := client.CreateDevice(ctx, &device)
	if err != nil {
		return nil, err
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	result := CommandResult{
		Format: "{{.Id}}",
		Data:   []*voltha.Device{got},
	}

	return &result, nil
}

func deleteDevice(conn *grpc.ClientConn, command *CommandContext) (*CommandResult, error) {

	if len(command.Args) == 0 {
		return nil, fmt.Errorf("Must specifyc device(s) to delete")
	}
	client := voltha.NewVolthaGlobalServiceClient(conn)

	for _, i := range command.Args {
		id := common.ID{
			Id: i,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		_, err := client.DeleteDevice(ctx, &id)
		if err != nil {
			fmt.Printf("Error while deleting '%s': %s\n", i, err)
			continue
		}

		select {
		case <-ctx.Done():
			fmt.Printf("Error while deleting '%s': %s\n", i, ctx.Err())
		default:
			fmt.Println(i)
		}
	}

	return nil, nil
}

func enableDevice(conn *grpc.ClientConn, command *CommandContext) (*CommandResult, error) {
	if len(command.Args) == 0 {
		return nil, fmt.Errorf("Must specifyc device(s) to enable")
	}
	client := voltha.NewVolthaGlobalServiceClient(conn)

	for _, i := range command.Args {
		id := common.ID{
			Id: i,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		_, err := client.EnableDevice(ctx, &id)
		if err != nil {
			fmt.Printf("Error while enabling '%s': %s\n", i, err)
			continue
		}

		select {
		case <-ctx.Done():
			fmt.Printf("Error while enabling '%s': %s\n", i, ctx.Err())
		default:
			fmt.Println(i)
		}
	}

	return nil, nil
}

func disableDevice(conn *grpc.ClientConn, command *CommandContext) (*CommandResult, error) {
	if len(command.Args) == 0 {
		return nil, fmt.Errorf("Must specifyc device(s) to disable")
	}
	client := voltha.NewVolthaGlobalServiceClient(conn)

	for _, i := range command.Args {
		id := common.ID{
			Id: i,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		_, err := client.DisableDevice(ctx, &id)
		if err != nil {
			fmt.Printf("Error while disabling '%s': %s\n", i, err)
			continue
		}

		select {
		case <-ctx.Done():
			fmt.Printf("Error while disabling '%s': %s\n", i, ctx.Err())
		default:
			fmt.Println(i)
		}
	}

	return nil, nil
}
