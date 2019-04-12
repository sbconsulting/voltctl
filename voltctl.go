package main

import (
	//"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	// "regexp"
	// "strings"
	//"text/tabwriter"
	//"text/template"
	//"text/template/parse"
	//"time"
	"github.com/ciena/voltctl/format"

	//"github.com/golang/protobuf/ptypes/empty"

	"github.com/ciena/voltctl/commands"
	"google.golang.org/grpc"
	"gopkg.in/yaml.v2"
	//"github.com/opencord/voltha/protos/go/voltha"
)

type config struct {
	ApiVersion string `yaml:"apiVersion"`
	Server     string `yaml:"server"`
}

var voltServer = flag.String("server", "", "Endpoint for VOLTHA API")
var voltConfigFile = flag.String("voltconfig", "", "Path to the voltconfig file to use for CLI request")
var specifiedFormat = flag.String("format", "", "Alternate format for table output")

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

	c, e := commands.LookupCommand(flag.Args())
	if e != nil {
		panic(e)
	}

	conn, err := grpc.Dial(cfg.Server, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Unable to connect: %v\n", err)
	}
	defer conn.Close()
	result, err := c.Invoke(conn, []string{})
	if err != nil {
		panic(err)
	}

	tableFormat := format.Format(result.Format)
	if *specifiedFormat != "" {
		tableFormat = format.Format(*specifiedFormat)
	}
	tableFormat.Execute(os.Stdout, true, result.Data)
}
