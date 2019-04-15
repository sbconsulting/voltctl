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
	//"context"
	"flag"
	"fmt"
	"github.com/ciena/voltctl/format"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/ciena/voltctl/commands"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
)

type config struct {
	ApiVersion string `yaml:"apiVersion"`
	Server     string `yaml:"server"`
}

var voltServer = flag.String("server", "", "Endpoint for VOLTHA API")
var voltConfigFile = flag.String("voltconfig", "", "Path to the voltconfig file to use for CLI request")
var specifiedFormat = flag.String("format", "", "Alternate format for table output")
var quiet = flag.Bool("q", false, "Only display ID in result table")

func main() {

	flag.Parse()

	if voltConfigFile == nil || len(*voltConfigFile) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Printf("Unable to discover they users home directory: %s\n", err)
		}
		defaultConfigFile := fmt.Sprintf("%s/.volt/config", home)
		voltConfigFile = &defaultConfigFile
	}

	var cfg config
	configFile, err := ioutil.ReadFile(*voltConfigFile)
	if err != nil {
		log.Printf("configFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(configFile, &cfg)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	// Override from command line
	if *voltServer != "" {
		cfg.Server = *voltServer
	}

	c, left, e := commands.LookupCommand(flag.Args())
	if e != nil {
		panic(e)
	}

	conn, err := grpc.Dial(cfg.Server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect: %v\n", err)
	}
	defer conn.Close()
	result, err := c.Invoke(conn, left)
	if err != nil {
		panic(err)
	}

	if result != nil && result.Data != nil {
		tableFormat := format.Format(result.Format)

		// Quiet overrides a specified format
		if *quiet {
			tableFormat = format.Format("{{.Id}}")
		} else if *specifiedFormat != "" {
			tableFormat = format.Format(strings.ReplaceAll(*specifiedFormat, "\\t", "\t"))
		}
		tableFormat.Execute(os.Stdout, true, result.Data)
	}
}
