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
package main

import (
	"github.com/ciena/voltctl/commands"
	flags "github.com/jessevdk/go-flags"
	"os"
	"path"
)

type config struct {
	ApiVersion string `yaml:"apiVersion"`
	Server     string `yaml:"server"`
}

type GlobalOptions struct {
	Config string `short:"c" long:"config" env:"VOLTCONFIG" value-name:"FILE" default:"~/.volt/config" description:"Location of client config file"`
	Server string `short:"s" long:"server" default:"" value-name:"SERVER:PORT" description:"IP/Host and port of VOLTHA"`
	Debug  bool   `short:"d" long:"debug" description:"Enable debug mode"`
	UseTLS bool   `long:"tls" description:"Use TLS"`
	CACert string `long:"tlscacert" description:"Trust certs signed only by this CA"`
	Cert   string `long:"tlscert" description:"Path to TLS vertificate file"`
	Key    string `long:"tlskey" description:"Path to TLS key file"`
	Verify bool   `long:"tlsverify" description:"Use TLS and verify the remote"`
}

func main() {

	parser := flags.NewNamedParser(path.Base(os.Args[0]), flags.Default|flags.PassAfterNonOption)
	_, err := parser.AddGroup("Global Options", "", &commands.GlobalOpts)
	if err != nil {
		panic(err)
	}
	commands.RegisterAdapterCommands(parser)
	commands.RegisterDeviceCommands(parser)
	commands.RegisterVersionCommands(parser)

	_, err = parser.ParseArgs(os.Args[1:])
	if err != nil {
		_, ok := err.(*flags.Error)
		if ok {
			real := err.(*flags.Error)
			if real.Type == flags.ErrHelp {
				return
			}
		} else {
			panic(err)
		}
		os.Exit(1)
	}
}
