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
	"strings"

	"context"
	"flag"
	"fmt"
	"github.com/opencord/voltha/protos/go/common"
	"github.com/opencord/voltha/protos/go/voltha"
	"time"
)

func listAllDevices(conn *grpc.ClientConn, args []string) (*CommandResult, error) {
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

	result := CommandResult{
		Format: "table{{ .Id }}\t{{.Type}}\t{{.Root}}\t{{.ParentId}}\t{{.SerialNumber}}\t{{.Vlan}}\t{{.AdminState}}\t{{.OperStatus}}\t{{.ConnectStatus}}",
		Data:   adapters.Items,
	}

	return &result, nil
}

func createDevice(conn *grpc.ClientConn, args []string) (*CommandResult, error) {
	var deviceType, macAddress, ipAddress, hostAndPort string
	flags := flag.NewFlagSet("device create", flag.ExitOnError)
	flags.StringVar(&deviceType, "t", "simulated_olt", "Device type")
	flags.StringVar(&macAddress, "m", "00:0c:e2:31:40:00", "MAC Address")
	flags.StringVar(&ipAddress, "i", "", "IP Address")
	flags.StringVar(&hostAndPort, "H", "", "Host and port")

	err := flags.Parse(args)
	if err == flag.ErrHelp {
		return nil, err
	} else if err != nil {
		return nil, err
	}

	device := voltha.Device{}

	if hostAndPort != "" {
		device.Address = &voltha.Device_HostAndPort{
			HostAndPort: hostAndPort,
		}
	} else if ipAddress != "" {
		device.Address = &voltha.Device_Ipv4Address{
			Ipv4Address: ipAddress,
		}
	} else if macAddress != "" {
		device.Address = &voltha.Device_MacAddress{
			MacAddress: strings.ToLower(macAddress),
		}
	}
	if deviceType != "" {
		device.Type = deviceType
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

func deleteDevice(conn *grpc.ClientConn, args []string) (*CommandResult, error) {

	if len(args) == 0 {
		return nil, fmt.Errorf("Must specifyc device(s) to delete")
	}
	client := voltha.NewVolthaGlobalServiceClient(conn)

	for _, i := range args {
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

func enableDevice(conn *grpc.ClientConn, args []string) (*CommandResult, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("Must specifyc device(s) to enable")
	}
	client := voltha.NewVolthaGlobalServiceClient(conn)

	for _, i := range args {
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

func disableDevice(conn *grpc.ClientConn, args []string) (*CommandResult, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("Must specifyc device(s) to disable")
	}
	client := voltha.NewVolthaGlobalServiceClient(conn)

	for _, i := range args {
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
