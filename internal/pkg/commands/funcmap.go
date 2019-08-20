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
	"fmt"
	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/proto"
	descpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
)

type MethodNotFoundError struct {
	Name string
}

func (e *MethodNotFoundError) Error() string {
	return fmt.Sprintf("Method '%s' not found in function map", e.Name)
}

type MethodVersionNotFoundError struct {
	Name    string
	Version string
}

func (e *MethodVersionNotFoundError) Error() string {
	return fmt.Sprintf("Method '%s' does not have a verison for '%s' specfied in function map", e.Name, e.Version)
}

type DescriptorNotFoundError struct {
	Version string
}

func (e *DescriptorNotFoundError) Error() string {
	return fmt.Sprintf("Protocol buffer descriptor for API version '%s' not found", e.Version)
}

type UnableToParseDescriptorErrror struct {
	err     error
	Version string
}

func (e *UnableToParseDescriptorErrror) Error() string {
	return fmt.Sprintf("Unable to parse protocal buffer descriptor for version '%s': %s", e.Version, e.err)
}

var descriptorMap = map[string][]byte{
	"v1": V1Descriptor,
	"v2": V2Descriptor,
}

var functionMap = map[string]map[string]string{
	"version": {
		"v1": "voltha.VolthaGlobalService/GetVoltha",
		"v2": "voltha.VolthaService/GetVoltha",
	},
	"adapter-list": {
		"v1": "voltha.VolthaGlobalService/ListAdapters",
		"v2": "voltha.VolthaService/ListAdapters",
	},
	"device-list": {
		"v1": "voltha.VolthaGlobalService/ListDevices",
		"v2": "voltha.VolthaService/ListDevices",
	},
	"device-ports": {
		"v1": "voltha.VolthaGlobalService/ListDevicePorts",
		"v2": "voltha.VolthaService/ListDevicePorts",
	},
	"device-create": {
		"v1": "voltha.VolthaGlobalService/CreateDevice",
		"v2": "voltha.VolthaService/CreateDevice",
	},
	"device-delete": {
		"v1": "voltha.VolthaGlobalService/DeleteDevice",
		"v2": "voltha.VolthaService/DeleteDevice",
	},
	"device-enable": {
		"v1": "voltha.VolthaGlobalService/EnableDevice",
		"v2": "voltha.VolthaService/EnableDevice",
	},
	"device-disable": {
		"v1": "voltha.VolthaGlobalService/DisableDevice",
		"v2": "voltha.VolthaService/DisableDevice",
	},
	"device-reboot": {
		"v1": "voltha.VolthaGlobalService/RebootDevice",
		"v2": "voltha.VolthaService/RebootDevice",
	},
	"device-inspect": {
		"v1": "voltha.VolthaGlobalService/GetDevice",
		"v2": "voltha.VolthaService/GetDevice",
	},
	"device-flow-list": {
		"v1": "voltha.VolthaGlobalService/ListDeviceFlows",
		"v2": "voltha.VolthaService/ListDeviceFlows",
	},
	"logical-device-list": {
		"v1": "voltha.VolthaGlobalService/ListLogicalDevices",
		"v2": "voltha.VolthaService/ListLogicalDevices",
	},
	"logical-device-ports": {
		"v1": "voltha.VolthaGlobalService/ListLogicalDevicePorts",
		"v2": "voltha.VolthaService/ListLogicalDevicePorts",
	},
	"logical-device-flow-list": {
		"v1": "voltha.VolthaGlobalService/ListLogicalDeviceFlows",
		"v2": "voltha.VolthaService/ListLogicalDeviceFlows",
	},
	"logical-device-inspect": {
		"v1": "voltha.VolthaGlobalService/GetLogicalDevice",
		"v2": "voltha.VolthaService/GetLogicalDevice",
	},
	"devicegroup-list": {
		"v1": "voltha.VolthaGlobalService/ListDeviceGroups",
		"v2": "voltha.VolthaService/ListDeviceGroups",
	},
	"get-goroutine-count": {
		"v2": "afrouter.Configuration.GetGoroutineCount",
	},
}

func GetMethod(name string) (grpcurl.DescriptorSource, string, error) {
	version := GlobalConfig.ApiVersion
	f, ok := functionMap[name]
	if !ok {
		return nil, "", &MethodNotFoundError{name}
	}
	m, ok := f[version]
	if !ok {
		return nil, "", &MethodVersionNotFoundError{name, version}
	}
	filename, ok := descriptorMap[version]
	if !ok {
		return nil, "", &DescriptorNotFoundError{version}
	}

	var fds descpb.FileDescriptorSet
	err := proto.Unmarshal(filename, &fds)
	if err != nil {
		return nil, "", &UnableToParseDescriptorErrror{err, version}
	}
	desc, err := grpcurl.DescriptorSourceFromFileDescriptorSet(&fds)
	if err != nil {
		return nil, "", err
	}

	return desc, m, nil
}
