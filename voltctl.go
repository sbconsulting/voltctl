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
	"encoding/json"
	"fmt"
	"github.com/ciena/voltctl/commands"
	"github.com/ciena/voltctl/format"
	flags "github.com/jessevdk/go-flags"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"path"
)

type config struct {
	ApiVersion string `yaml:"apiVersion"`
	Server     string `yaml:"server"`
}

type GlobalOptions struct {
	Config string `short:"c" long:"config" env:"VOLTCONFIG" default:"" description:"Location of client config file"`
	Server string `short:"s" long:"server" default:"" description:"IP/Host and port of VOLTHA"`
	Debug  bool   `short:"d" long:"debug" description:"Enable debug mode"`
	UseTLS bool   `long:"tls" description:"Use TLS"`
	CACert string `long:"tlscacert" description:"Trust certs signed only by this CA"`
	Cert   string `long:"tlscert" description:"Path to TLS vertificate file"`
	Key    string `long:"tlskey" description:"Path to TLS key file"`
	Verify bool   `long:"tlsverify" description:"Use TLS and verify the remote"`
}

func main() {

	globalOpts := GlobalOptions{}

	parser := flags.NewNamedParser(path.Base(os.Args[0]), flags.Default|flags.PassAfterNonOption)
	_, err := parser.AddGroup("Application Options", "", &globalOpts)
	if err != nil {
		panic(err)
	}

	left, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		real := err.(*flags.Error)
		if real.Type == flags.ErrHelp {
			return
		}
		os.Exit(1)
	}

	if len(globalOpts.Config) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Unable to discover they users home directory: %s\n", err)
		}
		globalOpts.Config = fmt.Sprintf("%s/.volt/config", home)
	}

	var cfg config
	configFile, err := ioutil.ReadFile(globalOpts.Config)
	if err != nil {
		log.Printf("configFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	// Override from command line
	if globalOpts.Server != "" {
		cfg.Server = globalOpts.Server
	}

	context, err := commands.LookupCommand(left)
	if err != nil {
		panic(err)
	}

	conn, err := grpc.Dial(cfg.Server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect: %v\n", err)
	}
	defer conn.Close()
	result, err := context.Command.Invoke(conn, context)
	if err != nil {
		panic(err)
	}

	if result != nil && result.Data != nil {
		if result.OutputAs == commands.OUTPUT_TABLE {
			tableFormat := format.Format(result.Format)
			tableFormat.Execute(os.Stdout, true, result.Data)
		} else if result.OutputAs == commands.OUTPUT_JSON {
			asJson, err := json.Marshal(&result.Data)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s", asJson)
		} else if result.OutputAs == commands.OUTPUT_YAML {
			asYaml, err := yaml.Marshal(&result.Data)
			if err != nil {
				panic(err)
			}
			fmt.Printf("%s", asYaml)
		}
	}
}
