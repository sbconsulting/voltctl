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
	"sort"
	"strings"
)

const (
	DEFAULT_FLOW_OUTPUT_FORMAT = "table{{.Id}}\t{{.TableId}}\t{{.Priority}}\t{{.Cookie}}\t{{.InPort}}\t{{.EthType}}\t{{.IpProto}}\t{{.UdpSrc}}\t{{.UdpDst}}\t{{.Output}}"
)

type FlowList struct {
	OutputOptions
	Args struct {
		Id string `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`
}

type FlowOpts struct {
	List FlowList `command:"list"`

	parser *flags.Parser
}

type FlowOutput struct {
	Id         string
	TableId    uint32
	Priority   uint32
	Cookie     string
	InPort     string
	EthType    string
	VlanId     string
	IpProto    string
	UdpSrc     string
	UdpDst     string
	Metadata   string
	SetVlanId  string
	PopVlan    string
	PushVlanId string
	Output     string
}

var flowOpts = FlowOpts{}

func RegisterFlowCommands(parent *flags.Parser) {
	parent.AddCommand("flow", "flow commands", "Commands to query and manipulate VOLTHA flows", &flowOpts)
	flowOpts.parser = parent
}

var (
	SORT_ORDER = map[string]byte{
		"Id":         0,
		"TableId":    1,
		"Priority":   2,
		"Cookie":     3,
		"InPort":     4,
		"EthType":    5,
		"VlanId":     6,
		"IpProto":    7,
		"UdpSrc":     8,
		"UdpDst":     9,
		"Metadata":   10,
		"SetVlanId":  11,
		"PopVlan":    12,
		"PushVlanId": 13,
		"Output":     100,
	}
)

func buildOutputFormat(fields map[string]byte) string {
	keys := make([]string, len(fields))
	i := 0
	for k := range fields {
		keys[i] = k
		i += 1
	}
	sort.Slice(keys, func(i, j int) bool {
		return SORT_ORDER[keys[i]] < SORT_ORDER[keys[j]]
	})
	var b strings.Builder
	b.WriteString("table")
	first := true
	for _, k := range keys {
		if !first {
			b.WriteString("\t")
		}
		first = false
		b.WriteString("{{.")
		b.WriteString(k)
		b.WriteString("}}")
	}
	return b.String()
}

func (options *FlowList) Execute(args []string) error {

	if len(args) > 0 {
		return fmt.Errorf("only a single argument 'DEVICE_ID' can be provided")
	}
	conn, err := NewConnection()
	if err != nil {
		return err
	}
	defer conn.Close()

	descriptor, method, err := GetMethod("flow-list")
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
	} else if h.Status != nil && h.Status.Err() != nil {
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

	outFields := map[string]byte{
		"Id":       0,
		"TableId":  0,
		"Priority": 0,
		"Cookie":   0,
	}

	data := make([]FlowOutput, len(items.([]interface{})))

	for i, item := range items.([]interface{}) {
		val := item.(*dynamic.Message)
		data[i].Id = fmt.Sprintf("%0x", val.GetFieldByName("id").(uint64))
		data[i].TableId = val.GetFieldByName("table_id").(uint32)
		data[i].Priority = val.GetFieldByName("priority").(uint32)
		// mask the lower 8 for the cookie, why?
		cookie := val.GetFieldByName("cookie").(uint64)
		if cookie == 0 {
			data[i].Cookie = "0"
		} else {
			data[i].Cookie = fmt.Sprintf("~%08x", val.GetFieldByName("cookie").(uint64)&0xffffffff)
		}
		match := val.GetFieldByName("match").(*dynamic.Message)
		fields := match.GetFieldByName("oxm_fields")
		for _, field := range fields.([]interface{}) {
			f := field.(*dynamic.Message)

			// Only support OFPXMC_OPENFLOW_BASIC (0x8000)
			if f.GetFieldByName("oxm_class").(int32) != 0x8000 {
				continue
			}

			basic := f.GetFieldByName("ofb_field").(*dynamic.Message)

			switch basic.GetFieldByName("type").(int32) {
			case 0: // IN_PORT
				outFields["InPort"] = 0
				data[i].InPort = fmt.Sprintf("%d", basic.GetFieldByName("port").(uint32))
			case 2: // METADATA
				outFields["Metadata"] = 0
				data[i].Metadata = fmt.Sprintf("0x%016x", basic.GetFieldByName("table_metadata").(uint64))
			case 5: // ETH_TYPE
				outFields["EthType"] = 0
				data[i].EthType = fmt.Sprintf("0x%04x", basic.GetFieldByName("eth_type").(uint32))
			case 6: // VLAN_ID
				outFields["VlanId"] = 0
				vid := basic.GetFieldByName("vlan_vid").(uint32)
				if vid == 0 {
					data[i].VlanId = "untagged"
				} else if vid&0x1000 > 0 {
					data[i].VlanId = fmt.Sprintf("%d", vid-4096)
				}
			case 10: // IP_PROTO
				outFields["IpProto"] = 0
				data[i].IpProto = fmt.Sprintf("%d", basic.GetFieldByName("ip_proto").(uint32))
			case 15: // UDP_SRC
				outFields["UdpSrc"] = 0
				data[i].UdpSrc = fmt.Sprintf("%d", basic.GetFieldByName("udp_src").(uint32))
			case 16: // UDP_DST
				outFields["UdpDst"] = 0
				data[i].UdpDst = fmt.Sprintf("%d", basic.GetFieldByName("udp_dst").(uint32))
			default:
				fmt.Printf("Unsupported match type: %d\n", basic.GetFieldByName("type").(int32))
			}
		}
		for _, instruction := range val.GetFieldByName("instructions").([]interface{}) {
			inst := instruction.(*dynamic.Message)
			switch inst.GetFieldByName("type").(uint32) {
			case 4: // APPLY_ACTIONS
				actions := inst.GetFieldByName("actions").(*dynamic.Message)
				for _, action := range actions.GetFieldByName("actions").([]interface{}) {
					a := action.(*dynamic.Message)
					switch a.GetFieldByName("type").(int32) {
					case 0: // OUTPUT
						outFields["Output"] = 0
						output := a.GetFieldByName("output").(*dynamic.Message)
						out := output.GetFieldByName("port").(uint32)
						switch out & 0x7fffffff {
						case 0:
							data[i].Output = "INVALID"
						case 0x7ffffff8:
							data[i].Output = "IN_PORT"
						case 0x7ffffff9:
							data[i].Output = "TABLE"
						case 0x7ffffffa:
							data[i].Output = "NORMAL"
						case 0x7ffffffb:
							data[i].Output = "FLOOD"
						case 0x7ffffffc:
							data[i].Output = "ALL"
						case 0x7ffffffd:
							data[i].Output = "CONTROLLER"
						case 0x7ffffffe:
							data[i].Output = "LOCAL"
						case 0x7fffffff:
							data[i].Output = "ANY"
						default:
							data[i].Output = fmt.Sprintf("%d", output.GetFieldByName("port").(uint32))
						}
					case 17: // PUSH_VLAN
						outFields["PushVlanId"] = 0
						push := a.GetFieldByName("push").(*dynamic.Message)
						data[i].PushVlanId = fmt.Sprintf("0x%x", push.GetFieldByName("ethertype").(uint32))
					case 18: // POP_VLAN
						outFields["PopVlan"] = 0
						data[i].PopVlan = "yes"
					case 25: // SET_FIELD

						set := a.GetFieldByName("set_field").(*dynamic.Message).GetFieldByName("field").(*dynamic.Message)
						if set.GetFieldByName("oxm_class").(int32) != 0x8000 {
							continue
						}

						basic := set.GetFieldByName("ofb_field").(*dynamic.Message)

						switch basic.GetFieldByName("type").(int32) {
						case 6: // VLAN_ID
							outFields["SetVlanId"] = 0
							vid := basic.GetFieldByName("vlan_vid").(uint32)
							if vid == 0 {
								data[i].SetVlanId = "untagged"
							} else if vid&0x1000 > 0 {
								data[i].SetVlanId = fmt.Sprintf("%d", vid-4096)
							}
						default:
							fmt.Printf("Unsupported SET_FIELD type: %d\n", basic.GetFieldByName("type").(int32))
						}

					default:
						fmt.Printf("Unsupporte ACTION TYPE: %d\n", a.GetFieldByName("type").(int32))
					}
				}
			default:
			}
		}
	}

	outputFormat := CharReplacer.Replace(options.Format)
	if options.Quiet {
		outputFormat = "{{.Id}}"
	} else if outputFormat == "" {
		outputFormat = buildOutputFormat(outFields)
	}

	result := CommandResult{
		Format:   format.Format(outputFormat),
		OutputAs: toOutputType(options.OutputAs),
		Data:     data,
	}
	GenerateOutput(&result)

	return nil
}
