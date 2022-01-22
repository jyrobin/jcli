// MIT License
//
// Copyright (c) 2019 Lea Anthony
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// See https://github.com/leaanthony/clir
//
// - added context.Context support by Jing-Ying Chen, 2022

package jcli

import (
	"context"
	"fmt"
	"strings"
)

const (
	maxDepth = 10
)

// Command represents a command that may be run by the user
type Command struct {
	app              *Cli     // only root command has non-nil app (i.e. when parent == nil)
	parent           *Command // filled when parent.AddCommand(this)
	name             string
	shortdescription string
	longdescription  string
	subCommands      []*Command
	subCommandsMap   map[string]*Command
	actionCallback   Action
	hidden           bool
	flags            *flagSet
}

// NewCommand creates a new Command
func NewCommand(name string, description string) *Command {
	result := &Command{
		name:             name,
		shortdescription: description,
		subCommandsMap:   make(map[string]*Command),
		hidden:           false,
		flags:            newFlagSet(),
	}
	return result
}

func (c *Command) commandPath() string {
	pth := c.name
	for i := maxDepth; i > 0 && c.parent != nil; i-- {
		c = c.parent
		if c.name != "" {
			pth = c.name + " " + pth
		}
	}
	return pth
}

func (c *Command) longestSubcommand() int {
	var longest int
	for _, subcommand := range c.subCommands {
		if n := len(subcommand.name); n > longest {
			longest = n
		}
	}
	return longest
}

func (c *Command) getCli() *Cli {
	for i := maxDepth; i > 0 && c != nil; i-- {
		if c.app != nil {
			return c.app
		}
		c = c.parent
	}
	return nil
}

// Run - Runs the Command with the given arguments
func (c *Command) run(ctx context.Context, args []string) error {
	app := c.getCli()
	if app == nil {
		return fmt.Errorf("Command not setup correctly")
	}

	var err error

	// If we have arguments, process them
	if len(args) > 0 {
		// Check for subcommand
		subcommand := c.subCommandsMap[args[0]]
		if subcommand != nil {
			return subcommand.run(ctx, args[1:])
		}

		// Parse flags
		commandPath := c.commandPath()
		ctx, err = c.flags.parseFlags(ctx, commandPath, args)
		if err != nil {
			if app.errorHandler != nil {
				return app.errorHandler(c.commandPath(), err)
			}
			return fmt.Errorf("Error: %s\nSee '%s --help' for usage", err, commandPath)
		}

		// Help takes precedence
		if HelpFlag(ctx) {
			c.PrintHelp(ctx)
			return nil
		}
	}

	// Do we have an action?
	if c.actionCallback != nil {
		return c.actionCallback(ctx)
	}

	// If we haven't specified a subcommand
	// check for an app level default command
	if app.defaultCommand != nil {
		// Prevent recursion!
		if app.defaultCommand != c {
			// only run default command if no args passed
			if len(args) == 0 {
				return app.defaultCommand.run(ctx, args)
			}
		}
	}

	// Nothing left we can do
	c.PrintHelp(ctx)

	return nil
}

// Action - Define an action from this command
func (c *Command) Action(callback Action) *Command {
	c.actionCallback = callback
	return c
}

// Command - Adds subcommands to this command
func (c *Command) SubCommands(commands ...*Command) *Command {
	for _, command := range commands {
		c.AddCommand(command)
	}
	return c
}

// PrintHelp - Output the help text for this command
func (c *Command) PrintHelp(ctx context.Context) {
	app := c.getCli()
	if app != nil {
		app.PrintBanner(ctx)
	}

	commandPath := c.commandPath()
	commandTitle := commandPath
	if c.shortdescription != "" {
		commandTitle += " - " + c.shortdescription
	}
	// Ignore root command
	if commandPath != c.name {
		fmt.Println(commandTitle)
	}
	if c.longdescription != "" {
		fmt.Println(c.longdescription + "\n")
	}
	if len(c.subCommands) > 0 {
		fmt.Println("Available commands:")
		fmt.Println("")
		longest := c.longestSubcommand()
		for _, subcommand := range c.subCommands {
			if subcommand.isHidden() {
				continue
			}
			spacer := strings.Repeat(" ", 3+longest-len(subcommand.name))
			isDefault := ""
			if subcommand.isDefaultCommand() {
				isDefault = "[default]"
			}
			fmt.Printf("   %s%s%s %s\n", subcommand.name, spacer, subcommand.shortdescription, isDefault)
		}
		fmt.Println("")
	}
	if c.flags.flagCount() > 0 {
		c.flags.printDefaults(ctx)
	}
	fmt.Fprintln(Stdout(ctx))
}

// isDefaultCommand returns true if called on the default command
func (c *Command) isDefaultCommand() bool {
	app := c.getCli()
	return app != nil && app.defaultCommand == c
}

// isHidden returns true if the command is a hidden command
func (c *Command) isHidden() bool {
	return c.hidden
}

// Hidden hides the command from the Help system
func (c *Command) Hidden() {
	c.hidden = true
}

// NewSubCommand - Creates a new subcommand
func (c *Command) NewSubCommand(name, description string) *Command {
	result := NewCommand(name, description)
	c.AddCommand(result)
	return result
}

// AddCommand - Adds a subcommand, which should be non-nil
func (c *Command) AddCommand(command *Command) {
	// if command == nil {
	// 	return
	// }

	command.parent = c // the only place parent is set
	name := command.name
	c.subCommands = append(c.subCommands, command)
	c.subCommandsMap[name] = command
}

// BoolFlag - Adds a boolean flag to the command. Use the first pointer in ptrs, if given,
// for storage, which is shared and not suitable for concurrent execution.
func (c *Command) BoolFlag(name, description string, val bool, ptrs ...*bool) *Command {
	if len(ptrs) > 0 {
		c.flags.addFlag(name, description, val, ptrs[0])
	} else {
		c.flags.addFlag(name, description, val, nil)
	}
	return c
}

// StringFlag - Adds a string flag to the command
func (c *Command) StringFlag(name, description string, val string, ptrs ...*string) *Command {
	if len(ptrs) > 0 {
		c.flags.addFlag(name, description, val, ptrs[0])
	} else {
		c.flags.addFlag(name, description, val, nil)
	}
	return c
}

// IntFlag - Adds an int flag to the command
func (c *Command) IntFlag(name, description string, val int, ptrs ...*int) *Command {
	if len(ptrs) > 0 {
		c.flags.addFlag(name, description, val, ptrs[0])
	} else {
		c.flags.addFlag(name, description, val, nil)
	}
	return c
}

// FloatFlag - Adds a float flag to the command
func (c *Command) FloatFlag(name, description string, val float64, ptrs ...*float64) *Command {
	if len(ptrs) > 0 {
		c.flags.addFlag(name, description, val, ptrs[0])
	} else {
		c.flags.addFlag(name, description, val, nil)
	}
	return c
}

// LongDescription - Sets the long description for the command
func (c *Command) LongDescription(longdescription string) *Command {
	c.longdescription = longdescription
	return c
}
