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

type DeviceList struct {
	OutputOptions
}

type DeviceCreate struct {
	DeviceType  string `short:"t" long:"devicetype" default:"simulated_olt" description:"Device type"`
	MACAddress  string `short:"m" long:"macaddress" default:"00:0c:e2:31:40:00" description:"MAC Address"`
	IPAddress   string `short:"i" long:"ipaddress" default:"" description:"IP Address"`
	HostAndPort string `short:"H" long:"hostandport" default:"" description:"Host and port"`
}

type DeviceDelete struct {
	Args struct {
		Ids []string `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceEnable struct {
	Args struct {
		Ids []string `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceDisable struct {
	Args struct {
		Ids []string `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type DeviceOpts struct {
	List    DeviceList    `command:"list"`
	Create  DeviceCreate  `command:"create"`
	Delete  DeviceDelete  `command:"delete"`
	Enable  DeviceEnable  `command:"enable"`
	Disable DeviceDisable `command:"disable"`
}

var deviceOpts = DeviceOpts{}

func RegisterDeviceCommands(parser *flags.Parser) {
	parser.AddCommand("device", "device commands", "Commands to query and manipulate VOLTHA devices", &deviceOpts)
}

func (options *DeviceList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaGlobalServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	adapters, err := client.ListDevices(ctx, &empty.Empty{})
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
		outputFormat = DEFAULT_DEVICE_FORMAT
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

func (options *DeviceCreate) Execute(args []string) error {

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

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	client := voltha.NewVolthaGlobalServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	got, err := client.CreateDevice(ctx, &device)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	fmt.Println(got.Id)

	return nil
}

func (options *DeviceDelete) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := voltha.NewVolthaGlobalServiceClient(conn)

	for _, i := range options.Args.Ids {
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

	return nil
}

func (options *DeviceEnable) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := voltha.NewVolthaGlobalServiceClient(conn)

	for _, i := range options.Args.Ids {
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

	return nil
}

func (options *DeviceDisable) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()
	client := voltha.NewVolthaGlobalServiceClient(conn)

	for _, i := range options.Args.Ids {
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

	return nil
}
