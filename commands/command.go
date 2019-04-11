package commands

import (
	"fmt"
)

type CommandNotFound struct {
	command string
}

func (c *CommandNotFound) Error() string {
	return fmt.Sprintf("Unable to locate command '%s'", c.command)
}

type CommandResult interface {
	ToJSON() string
	ToTable(format string, filter string) string
}

type Command struct {
	Invoke      func(args []string) (*CommandResult, error)
	subcommands map[string]*Command
}

func noCommand(args []string) (*CommandResult, error) {
	return nil, nil
}

var Commands = map[string]*Command{
	"device": {
		Invoke: noCommand,
		subcommands: map[string]*Command{
			"list": {
				Invoke:      listAllDevices,
				subcommands: nil,
			},
			"preprovision": {
				Invoke:      preprovisionDevice,
				subcommands: nil,
			},
			"enable": {
				Invoke:      enableDevice,
				subcommands: nil,
			},
		},
	},
}

func lookupCommand(base *Command, args []string) (*Command, error) {
	var cmd *Command
	var ok bool

	if args == nil || len(args) == 0 {
		return base, nil
	}

	fmt.Printf("Looking for '%s'\n", args[0])

	if base == nil {
		cmd, ok = Commands[args[0]]
		if !ok {
			return nil, &CommandNotFound{command: args[0]}
		}
		return lookupCommand(cmd, args[1:])
	}
	if base.subcommands != nil {
		cmd, ok = base.subcommands[args[0]]
		if ok {
			return lookupCommand(cmd, args[1:])
		}
	}

	return base, nil
}

func LookupCommand(args []string) (*Command, error) {
	return lookupCommand(nil, args)
}
