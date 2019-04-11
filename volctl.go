package main

import (
	"fmt"
	"os"

	"github.com/ciena/volctl/commands"
)

func main() {
	c, e := commands.LookupCommand(os.Args[1:])
	if e != nil {
		panic(e)
	}
	fmt.Printf("%+v\n", c)
	c.Invoke([]string{"hello", "world"})
}
