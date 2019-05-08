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
	"github.com/ciena/voltctl/cli/version"
	"github.com/ciena/voltctl/format"
	"github.com/fullstorydev/grpcurl"
	flags "github.com/jessevdk/go-flags"
	"github.com/jhump/protoreflect/dynamic"
	"strings"
)

type VersionDetails struct {
	Version   string `json:"version"`
	GoVersion string `json:"goversion"`
	GitCommit string `json:"gitcommit"`
	BuildTime string `json:"buildtime"`
	Os        string `json:"os"`
	Arch      string `json:"arch"`
}

type VersionOutput struct {
	Client  VersionDetails `json:"client"`
	Cluster VersionDetails `json:"cluster"`
}

type VersionOpts struct {
	OutputOptions
}

var versionOpts = VersionOpts{}

var versionInfo = VersionOutput{
	Client: VersionDetails{
		Version:   version.Version,
		GoVersion: version.GoVersion,
		GitCommit: version.GitCommit,
		Os:        version.Os,
		Arch:      version.Arch,
		BuildTime: version.BuildTime,
	},
	Cluster: VersionDetails{
		Version:   "unknown-version",
		GoVersion: "unknown-goversion",
		GitCommit: "unknown-gitcommit",
		Os:        "unknown-os",
		Arch:      "unknown-arch",
		BuildTime: "unknown-buildtime",
	},
}

func RegisterVersionCommands(parent *flags.Parser) {
	parent.AddCommand("version", "display version", "Display client and server version", &versionOpts)
}

const DefaultFormat = `Client:
 Version        {{.Client.Version}}
 Go version:    {{.Client.GoVersion}}
 Git commit:    {{.Client.GitCommit}}
 Built:         {{.Client.BuildTime}}
 OS/Arch:       {{.Client.Os}}/{{.Client.Arch}}

Cluster:
 Version        {{.Cluster.Version}}
 Go version:    {{.Cluster.GoVersion}}
 Git commit:    {{.Cluster.GitCommit}}
 Built:         {{.Cluster.BuildTime}}
 OS/Arch:       {{.Cluster.Os}}/{{.Cluster.Arch}}
`

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

	versionInfo.Cluster.Version = strings.ReplaceAll(version.(string), "\n", "")

	outputFormat := CharReplacer.Replace(options.Format)
	if outputFormat == "" {
		outputFormat = DefaultFormat
	}
	if options.Quiet {
		outputFormat = "{{.Client.Version}}"
	}

	result := CommandResult{
		Format:    format.Format(outputFormat),
		OutputAs:  toOutputType(options.OutputAs),
		NameLimit: options.NameLimit,
		Data:      versionInfo,
	}

	GenerateOutput(&result)
	return nil
}
