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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

type Cli struct {
	version        string
	rootCommand    *Command
	defaultCommand *Command
	preRunCommand  func(context.Context, *Cli) error
	bannerFunction func(context.Context, *Cli) string
	errorHandler   func(string, error) error
	helpHandler    func(context.Context, *Cli) error
}

// NewCli - Creates a new Cli application object
func NewCli(name, description, version string) *Cli {
	result := &Cli{
		version:        version,
		bannerFunction: defaultBannerFunction,
	}
	result.rootCommand = NewCommand(name, description)
	result.rootCommand.app = result // the only place app is set
	return result
}

// Version - Get the Application version string.
func (c *Cli) Version() string {
	return c.version
}

// Name - Get the Application Name
func (c *Cli) Name() string {
	return c.rootCommand.name
}

// ShortDescription - Get the Application short description.
func (c *Cli) ShortDescription() string {
	return c.rootCommand.shortdescription
}

// PrintBanner - Prints the application banner!
func (c *Cli) PrintBanner(ctx context.Context) {
	out := Stdout(ctx)
	fmt.Fprintln(out, c.bannerFunction(ctx, c))
	fmt.Fprintln(out)
}

// PrintHelp - Prints the application's help.
func (c *Cli) PrintHelp(ctx context.Context) {
	c.rootCommand.PrintHelp(ctx)
}

// Run - Runs the application with the given arguments.
func (c *Cli) Run(ctx context.Context, args ...string) error {
	if c.preRunCommand != nil {
		err := c.preRunCommand(ctx, c)
		if err != nil {
			return err
		}
	}
	return c.rootCommand.run(ctx, args)
}

// NewSubCommand - Creates a new SubCommand for the application.
func (c *Cli) NewSubCommand(name, description string) *Command {
	return c.rootCommand.NewSubCommand(name, description)
}

// PreRun - Calls the given function before running the specific command.
func (c *Cli) PreRun(callback func(context.Context, *Cli) error) {
	c.preRunCommand = callback
}

// BoolFlag - Adds a boolean flag to the root command.
func (c *Cli) BoolFlag(name, description string, variable bool, ptr ...*bool) *Cli {
	c.rootCommand.BoolFlag(name, description, variable, ptr...)
	return c
}

// StringFlag - Adds a string flag to the root command.
func (c *Cli) StringFlag(name, description string, variable string, ptr ...*string) *Cli {
	c.rootCommand.StringFlag(name, description, variable, ptr...)
	return c
}

// IntFlag - Adds an int flag to the root command.
func (c *Cli) IntFlag(name, description string, variable int, ptr ...*int) *Cli {
	c.rootCommand.IntFlag(name, description, variable, ptr...)
	return c
}

// Action - Define an action from this command.
func (c *Cli) Action(callback Action) *Cli {
	c.rootCommand.Action(callback)
	return c
}

// LongDescription - Sets the long description for the command.
func (c *Cli) LongDescription(longdescription string) *Cli {
	c.rootCommand.LongDescription(longdescription)
	return c
}

// Command - Adds commands to the application.
func (c *Cli) Commands(commands ...*Command) *Cli {
	c.rootCommand.SubCommands(commands...)
	return c
}

// DefaultCommand - Sets the given command as the command to run when
// no other commands given.
func (c *Cli) DefaultCommand(defaultCommand *Command) *Cli {
	c.defaultCommand = defaultCommand
	return c
}

// BannerFunction - Set the function that is called
// to get the banner string.
func (c *Cli) BannerFunction(fn func(context.Context, *Cli) string) *Cli {
	c.bannerFunction = fn
	return c
}

// ErrorFunction - Set custom error message when undefined
// flags are used by the user. First argument is a string containing
// the commnad path used. Second argument is the undefined flag error.
func (c *Cli) ErrorFunction(fn func(string, error) error) *Cli {
	c.errorHandler = fn
	return c
}

// HelpHandler - Sets the help handler
func (c *Cli) HelpHandler(handler func(context.Context, *Cli) error) *Cli {
	c.helpHandler = handler
	return c
}

func (cli *Cli) RunBuffer(ctx context.Context, printsJson bool, args ...string) ([]byte, error) {
	ctx = context.WithValue(ctx, PrintJsonKey, printsJson)

	buf := new(bytes.Buffer)
	ctx = WithStdout(ctx, buf)
	err := cli.Run(ctx, args...)
	return buf.Bytes(), err
}

func (cli *Cli) RunLine(ctx context.Context, printsJson bool, line string) ([]byte, error) {
	words := strings.Fields(line)
	return cli.RunBuffer(ctx, printsJson, words...)
}

func (cli *Cli) RunUnmarshal(ctx context.Context, line string, ret interface{}) error {
	buf, err := cli.RunLine(ctx, true, line)
	if err == nil {
		err = json.Unmarshal(buf, ret)
	}
	return err
}
