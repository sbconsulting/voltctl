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
	DEFAULT_LOGICAL_DEVICE_FORMAT      = "table{{ .Id }}\t{{.DatapathId}}\t{{.RootDeviceId}}\t{{.SerialNumber}}\t{{.Features.NBuffers}}\t{{.Features.NTables}}\t{{.Features.Capabilities}}"
	DEFAULT_LOGICAL_DEVICE_PORT_FORMAT = "table{{.Id}}\t{{.DeviceId}}\t{{.DevicePortNo}}\t{{.RootPort}}\t{{.Openflow.PortNo}}\t{{.Openflow.HwAddr}}\t{{.Openflow.Name}}\t{{.Openflow.State}}\t{{.Openflow.Features.Current}}\t{{.Openflow.Bitrate.Current}}"
)

type LogicalDeviceList struct {
	OutputOptions
}

type LogicalDeviceListOutput struct {
	Id           string `json:"id"`
	DatapathId   string `json:"datapathid"`
	RootDeviceId string `json:"rootdeviceid"`
	SerialNumber string `json:"serialnumber"`
	Features     struct {
		NBuffers     uint32 `json:"nbuffers"`
		NTables      uint32 `json:"ntables"`
		Capabilities string `json:"capabilities"`
	} `json:"features"`
}

type LogicalDevicePortList struct {
	OutputOptions
	Args struct {
		Id string `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type LogicalDevicePortOutput struct {
	Id           string `json:"id"`
	DeviceId     string `json:"deviceid"`
	DevicePortNo uint32 `json:"deviceportno"`
	RootPort     bool   `json:"rootport"`
	Openflow     struct {
		PortNo   uint32 `json:"portno"`
		HwAddr   string `json:"hwaddr"`
		Name     string `json:"name"`
		Config   string `json:"config"`
		State    string `json:"state"`
		Features struct {
			Advertised string `json:"advertised"`
			Current    string `json:"current"`
			Supported  string `json:"supported"`
			Peer       string `json:"peer"`
		} `json:"features"`
		Bitrate struct {
			Current uint32 `json:"current"`
			Max     uint32 `json:"max"`
		}
	} `json:"openflow"`
}

type LogicalDeviceOpts struct {
	List  LogicalDeviceList     `command:"list"`
	Ports LogicalDevicePortList `command:"ports"`
}

var logicalDeviceOpts = LogicalDeviceOpts{}

func RegisterLogicalDeviceCommands(parser *flags.Parser) {
	parser.AddCommand("logicaldevice", "logical device commands", "Commands to query and manipulate VOLTHA logical devices", &logicalDeviceOpts)
}

func (options *LogicalDeviceList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	descriptor, method, err := GetMethod("logical-device-list")
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
		outputFormat = DEFAULT_LOGICAL_DEVICE_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	data := make([]LogicalDeviceListOutput, len(items.([]interface{})))

	for i, item := range items.([]interface{}) {
		val := item.(*dynamic.Message)
		data[i].Id = val.GetFieldByName("id").(string)
		data[i].DatapathId = fmt.Sprintf("%016x", val.GetFieldByName("datapath_id").(uint64))
		data[i].RootDeviceId = val.GetFieldByName("root_device_id").(string)
		desc := val.GetFieldByName("desc").(*dynamic.Message)
		data[i].SerialNumber = desc.GetFieldByName("serial_num").(string)
		features := val.GetFieldByName("switch_features").(*dynamic.Message)
		data[i].Features.NBuffers = features.GetFieldByName("n_buffers").(uint32)
		data[i].Features.NTables = features.GetFieldByName("n_tables").(uint32)
		data[i].Features.Capabilities = fmt.Sprintf("0x%08x", features.GetFieldByName("capabilities").(uint32))
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
	}

	GenerateOutput(&result)
	return nil
}

func (options *LogicalDevicePortList) Execute(args []string) error {

	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	descriptor, method, err := GetMethod("logical-device-ports")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), GlobalConfig.Grpc.Timeout)
	defer cancel()

	h := &RpcEventHandler{
		Fields: map[string]map[string]interface{}{ParamNames[GlobalConfig.ApiVersion]["ID"]: {"id": options.Args.Id}},
	}
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
		outputFormat = DEFAULT_LOGICAL_DEVICE_PORT_FORMAT
	}
	if options.Quiet {
		outputFormat = "{{.Id}}"
	}

	var hw strings.Builder
	var first bool
	data := make([]LogicalDevicePortOutput, len(items.([]interface{})))
	for i, item := range items.([]interface{}) {
		val := item.(*dynamic.Message)
		data[i].Id = val.GetFieldByName("id").(string)
		data[i].DeviceId = val.GetFieldByName("device_id").(string)
		data[i].DevicePortNo = val.GetFieldByName("device_port_no").(uint32)
		data[i].RootPort = val.GetFieldByName("root_port").(bool)
		ofp := val.GetFieldByName("ofp_port").(*dynamic.Message)
		hw.Reset()
		first = true
		for _, b := range ofp.GetFieldByName("hw_addr").([]interface{}) {
			if !first {
				hw.WriteString(":")
			}
			first = false
			hw.WriteString(fmt.Sprintf("%02x", b))
		}
		data[i].Openflow.HwAddr = hw.String()
		data[i].Openflow.PortNo = ofp.GetFieldByName("port_no").(uint32)
		data[i].Openflow.Name = ofp.GetFieldByName("name").(string)
		data[i].Openflow.Config = fmt.Sprintf("0x%08x", ofp.GetFieldByName("config").(uint32))
		data[i].Openflow.State = fmt.Sprintf("0x%08x", ofp.GetFieldByName("state").(uint32))
		data[i].Openflow.Features.Current = fmt.Sprintf("0x%08x", ofp.GetFieldByName("curr").(uint32))
		data[i].Openflow.Features.Advertised = fmt.Sprintf("0x%08x", ofp.GetFieldByName("advertised").(uint32))
		data[i].Openflow.Features.Supported = fmt.Sprintf("0x%08x", ofp.GetFieldByName("supported").(uint32))
		data[i].Openflow.Features.Peer = fmt.Sprintf("0x%08x", ofp.GetFieldByName("peer").(uint32))
		data[i].Openflow.Bitrate.Current = ofp.GetFieldByName("curr_speed").(uint32)
		data[i].Openflow.Bitrate.Max = ofp.GetFieldByName("max_speed").(uint32)
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
	}

	GenerateOutput(&result)
	return nil
}
