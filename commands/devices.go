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
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
	"strings"
)

const (
	DEFAULT_DEVICE_FORMAT = "table{{ .Id }}\t{{.Type}}\t{{.Root}}\t{{.ParentId}}\t{{.SerialNumber}}\t{{.Vlan}}\t{{.AdminState}}\t{{.OperStatus}}\t{{.ConnectStatus}}"
)

type DeviceList struct {
	OutputOptions
}

type DeviceListOutput struct {
	Id            string
	Type          string
	Root          bool
	ParentId      string
	SerialNumber  string
	Vlan          uint32
	AdminState    string
	OperStatus    string
	ConnectStatus string
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

	descriptor, method, err := GetMethod("device-list")
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

	items, err := d.TryGetFieldByName("items")
	if err != nil {
		return err
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DEFAULT_DEVICE_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	data := make([]DeviceListOutput, len(items.([]interface{})))

	for i, item := range items.([]interface{}) {
		val := item.(*dynamic.Message)
		data[i].Id = val.GetFieldByName("id").(string)
		data[i].Type = val.GetFieldByName("type").(string)
		data[i].Root = val.GetFieldByName("root").(bool)
		data[i].ParentId = val.GetFieldByName("parent_id").(string)
		data[i].SerialNumber = val.GetFieldByName("serial_number").(string)
		data[i].Vlan = val.GetFieldByName("vlan").(uint32)
		data[i].AdminState = GetEnumValue(val, "admin_state")
		data[i].OperStatus = GetEnumValue(val, "oper_status")
		data[i].ConnectStatus = GetEnumValue(val, "connect_status")
	}

	result := CommandResult{
		Format:   format.Format(outputFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     data,
	}

	GenerateOutput(&result)
	return nil
}

func (options *DeviceCreate) Execute(args []string) error {

	dm := make(map[string]interface{})
	if options.HostAndPort != "" {
		dm["host_and_port"] = options.HostAndPort
	} else if options.IPAddress != "" {
		dm["ipv4_address"] = options.IPAddress
	} else if options.MACAddress != "" {
		dm["mac_address"] = strings.ToLower(options.MACAddress)
	}
	if options.DeviceType != "" {
		dm["type"] = options.DeviceType
	}

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	descriptor, method, err := GetMethod("device-create")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	h := &RpcEventHandler{
		Fields: map[string]map[string]interface{}{"voltha.Device": dm},
	}
	err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
	if err != nil {
		return err
	} else if h.Status != nil && h.Status.Err() != nil {
		return h.Status.Err()
	}

	resp, err := dynamic.AsDynamicMessage(h.Response)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", resp.GetFieldByName("id").(string))

	return nil
}

func (options *DeviceDelete) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	descriptor, method, err := GetMethod("device-delete")
	if err != nil {
		return err
	}

	for _, i := range options.Args.Ids {

		h := &RpcEventHandler{
			Fields: map[string]map[string]interface{}{ParamNames[GlobalConfig.ApiVersion]["ID"]: {"id": i}},
		}
		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
		defer cancel()

		err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
		if err != nil {
			fmt.Printf("Error while deleting '%s': %s\n", i, err)
			continue
		} else if h.Status != nil && h.Status.Err() != nil {
			fmt.Printf("Error while deleting '%s': %s\n", i, h.Status.Err())
			continue
		}
		fmt.Printf("%s\n", i)
	}

	return nil
}

func (options *DeviceEnable) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	descriptor, method, err := GetMethod("device-enable")
	if err != nil {
		return err
	}

	for _, i := range options.Args.Ids {
		h := &RpcEventHandler{
			Fields: map[string]map[string]interface{}{ParamNames[GlobalConfig.ApiVersion]["ID"]: {"id": i}},
		}
		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
		defer cancel()

		err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
		if err != nil {
			fmt.Printf("Error while enabling '%s': %s\n", i, err)
			continue
		} else if h.Status != nil && h.Status.Err() != nil {
			fmt.Printf("Error while enabling '%s': %s\n", i, h.Status.Err())
			continue
		}
		fmt.Printf("%s\n", i)
	}

	return nil
}

func (options *DeviceDisable) Execute(args []string) error {
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	descriptor, method, err := GetMethod("device-disable")
	if err != nil {
		return err
	}

	for _, i := range options.Args.Ids {
		h := &RpcEventHandler{
			Fields: map[string]map[string]interface{}{ParamNames[GlobalConfig.ApiVersion]["ID"]: {"id": i}},
		}
		ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
		defer cancel()

		err = grpcurl.InvokeRPC(ctx, descriptor, conn, method, []string{}, h, h.GetParams)
		if err != nil {
			fmt.Printf("Error while disabling '%s': %s\n", i, err)
			continue
		} else if h.Status != nil && h.Status.Err() != nil {
			fmt.Printf("Error while disabling '%s': %s\n", i, h.Status.Err())
			continue
		}
		fmt.Printf("%s\n", i)
	}

	return nil
}
