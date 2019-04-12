package commands

import (
	"fmt"
	"github.com/ciena/voltctl/format"
	"google.golang.org/grpc"
)

type CommandNotFound struct {
	command string
}

func (c *CommandNotFound) Error() string {
	return fmt.Sprintf("Unable to locate command '%s'", c.command)
}

type Anonymous struct{}
type CommandResult struct {
	Format format.Format
	Data   interface{}
}

type Command struct {
	Invoke      func(conn *grpc.ClientConn, args []string) (*CommandResult, error)
	subcommands map[string]*Command
}

func noCommand(conn *grpc.ClientConn, args []string) (*CommandResult, error) {
	return nil, fmt.Errorf("Unknown command specified")
}

var Commands = map[string]*Command{
	"adapter": {
		Invoke: noCommand,
		subcommands: map[string]*Command{
			"list": {
				Invoke:      listAllAdapters,
				subcommands: nil,
			},
		},
	},
	"device": {
		Invoke: noCommand,
		subcommands: map[string]*Command{
			"create": {
				Invoke:      createDevice,
				subcommands: nil,
			},
			"delete": {
				Invoke:      deleteDevice,
				subcommands: nil,
			},
			"list": {
				Invoke:      listAllDevices,
				subcommands: nil,
			},
			"enable": {
				Invoke:      enableDevice,
				subcommands: nil,
			},
			"disable": {
				Invoke:      disableDevice,
				subcommands: nil,
			},
		},
	},
}

func lookupCommand(base *Command, args []string) (*Command, []string, error) {
	var cmd *Command
	var ok bool

	if args == nil || len(args) == 0 {
		return base, args, nil
	}

	if base == nil {
		cmd, ok = Commands[args[0]]
		if !ok {
			return nil, args, &CommandNotFound{command: args[0]}
		}
		return lookupCommand(cmd, args[1:])
	}
	if base.subcommands != nil {
		cmd, ok = base.subcommands[args[0]]
		if ok {
			return lookupCommand(cmd, args[1:])
		}
	}

	return base, args, nil
}

func LookupCommand(args []string) (*Command, []string, error) {
	return lookupCommand(nil, args)
}
