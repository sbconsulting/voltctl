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
	"github.com/jhump/protoreflect/dynamic"
	"sort"
	"strings"
)

type FlowList struct {
	OutputOptions
	Args struct {
		Id string `positional-arg-name:"DEVICE_ID" required:"yes"`
	} `positional-args:"yes"`

	Method string
}

type FlowOpts struct {
	List FlowList `command:"list"`
}

/*
 * This is a partial list of OF match/action values. This list will be
 * expanded as new fields are needed within VOLTHA
 *
 * Strings are used in the output structure so that on output the table
 * can be "sparsely" populated with "empty" cells as opposed to 0 (zeros)
 * all over the place.
 */
type FlowOutput struct {
	Id                     string `json:"id"`
	TableId                uint32 `json:"tableid"`
	Priority               uint32 `json:"priority"`
	Cookie                 string `json:"cookie"`
	UnsupportedMatch       string `json:"unsupportedmatch,omitempty"`
	InPort                 string `json:"inport,omitempty"`
	EthType                string `json:"ethtype,omitempty"`
	VlanId                 string `json:"vlanid,omitempty"`
	IpProto                string `json:"ipproto,omitempty"`
	UdpSrc                 string `json:"udpsrc,omitempty"`
	UdpDst                 string `json:"dstsrc,omitempty"`
	Metadata               string `json:"metadata,omitempty"`
	UnsupportedInstruction string `json:"unsupportedinstruction,omitempty"`
	UnsupportedAction      string `json:"unsupportedaction,omitempty"`
	UnsupportedSetField    string `json:"unsupportedsetfield,omitempty"`
	SetVlanId              string `json:"setvlanid,omitempty"`
	PopVlan                string `json:"popvlan,omitempty"`
	PushVlanId             string `json:"pushvlanid,omitempty"`
	Output                 string `json:"output,omitempty"`
	GotoTable              string `json:"gototable,omitempty"`
}

var (
	flowOpts = FlowOpts{}

	// Used to sort the table colums in a consistent order
	SORT_ORDER = map[string]uint16{
		"Id":                     0,
		"TableId":                10,
		"Priority":               20,
		"Cookie":                 30,
		"UnsupportedMatch":       35,
		"InPort":                 40,
		"EthType":                50,
		"VlanId":                 60,
		"IpProto":                70,
		"UdpSrc":                 80,
		"UdpDst":                 90,
		"Metadata":               100,
		"UnsupportedInstruction": 102,
		"UnsupportedAction":      105,
		"UnsupportedSetField":    107,
		"SetVlanId":              110,
		"PopVlan":                120,
		"PushVlanId":             130,
		"Output":                 1000,
		"GotoTable":              1010,
	}
)

/*
 * Construct a template format string based on the fields required by the
 * results.
 */
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

func toVlanId(vid uint32) string {
	if vid == 0 {
		return "untagged"
	} else if vid&0x1000 > 0 {
		return fmt.Sprintf("%d", vid-4096)
	}
	return fmt.Sprintf("%d", vid)
}

func appendInt32(base string, val int32) string {
	if len(base) > 0 {
		return fmt.Sprintf("%s,%d", base, val)
	}
	return fmt.Sprintf("%d", val)
}

func appendUint32(base string, val uint32) string {
	if len(base) > 0 {
		return fmt.Sprintf("%s,%d", base, val)
	}
	return fmt.Sprintf("%d", val)
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

	switch options.Method {
	case "device-flow-list":
	case "logical-device-flow-list":
	default:
		panic(fmt.Errorf("Unknown method name: '%s'", options.Method))
	}

	descriptor, method, err := GetMethod(options.Method)
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

	// Always include Id, TableId, Priority, and Cookie in output
	outFields := map[string]byte{
		"Id":       0,
		"TableId":  0,
		"Priority": 0,
		"Cookie":   0,
	}

	// Walk the flows and populate the output table
	data := make([]FlowOutput, len(items.([]interface{})))
	for i, item := range items.([]interface{}) {
		val := item.(*dynamic.Message)
		data[i].Id = fmt.Sprintf("%016x", val.GetFieldByName("id").(uint64))
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
				data[i].VlanId = toVlanId(basic.GetFieldByName("vlan_vid").(uint32))
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
				/*
				 * For unsupported match types put them into an
				 * "Unsupported field so the table/json still
				 * outputs relatively correctly as opposed to
				 * having log messages.
				 */
				outFields["UnsupportedMatch"] = 0
				data[i].UnsupportedMatch = appendInt32(data[i].UnsupportedMatch, basic.GetFieldByName("type").(int32))
			}
		}
		for _, instruction := range val.GetFieldByName("instructions").([]interface{}) {
			inst := instruction.(*dynamic.Message)
			switch inst.GetFieldByName("type").(uint32) {
			case 1: // GOTO_TABLE
				outFields["GotoTable"] = 0
				goto_table := inst.GetFieldByName("goto_table").(*dynamic.Message)
				data[i].GotoTable = fmt.Sprintf("%d", goto_table.GetFieldByName("table_id").(uint32))
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

						// Only support OFPXMC_OPENFLOW_BASIC (0x8000)
						if set.GetFieldByName("oxm_class").(int32) != 0x8000 {
							continue
						}
						basic := set.GetFieldByName("ofb_field").(*dynamic.Message)

						switch basic.GetFieldByName("type").(int32) {
						case 6: // VLAN_ID
							outFields["SetVlanId"] = 0
							data[i].SetVlanId = toVlanId(basic.GetFieldByName("vlan_vid").(uint32))
						default: // Unsupported
							/*
							 * For unsupported match types put them into an
							 * "Unsupported field so the table/json still
							 * outputs relatively correctly as opposed to
							 * having log messages.
							 */
							outFields["UnsupportedSetField"] = 0
							data[i].UnsupportedSetField = appendInt32(data[i].UnsupportedSetField,
								basic.GetFieldByName("type").(int32))
						}
					default: // Unsupported
						/*
						 * For unsupported match types put them into an
						 * "Unsupported field so the table/json still
						 * outputs relatively correctly as opposed to
						 * having log messages.
						 */
						outFields["UnsupportedAction"] = 0
						data[i].UnsupportedAction = appendInt32(data[i].UnsupportedAction,
							a.GetFieldByName("type").(int32))
					}
				}
			default: // Unsupported
				/*
				 * For unsupported match types put them into an
				 * "Unsupported field so the table/json still
				 * outputs relatively correctly as opposed to
				 * having log messages.
				 */
				outFields["UnsupportedInstruction"] = 0
				data[i].UnsupportedInstruction = appendUint32(data[i].UnsupportedInstruction,
					inst.GetFieldByName("type").(uint32))
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
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      data,
	}
	GenerateOutput(&result)

	return nil
}
