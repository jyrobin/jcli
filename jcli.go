package jcli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
)

var (
	ErrCommandNotFound = errors.New("unknown command")
)

type Cntx struct {
	cntx context.Context
	Body string
}

func NewCntx(body string, cntxs ...context.Context) *Cntx {
	var cntx context.Context = nil
	if len(cntxs) > 0 {
		cntx = cntxs[0]
	}
	if cntx == nil {
		cntx = context.Background()
	}
	return &Cntx{cntx, body}
}

func (c *Cntx) Context() context.Context {
	return c.cntx
}

func (c *Cntx) Value(key any) interface{} {
	return c.cntx.Value(key)
}

func (c *Cntx) With(cntx context.Context) *Cntx {
	ret := Cntx{}
	ret = *c
	ret.cntx = cntx
	return &ret
}

func (c *Cntx) WithValue(key, val any) *Cntx {
	return c.With(context.WithValue(c.cntx, key, val))
}

type Runner func(c *Cntx, args []string) (string, error)

type HelpFunc func() string

type Command interface {
	Run(c *Cntx, args []string) (string, error)
	Name() string
	Description() string
	Help() string
}

type Middleware func(*Cntx, []string) (*Cntx, []string)

type Cli struct {
	name    string
	desc    string
	cmds    map[string]Command
	middles []Middleware
	Helper  HelpFunc
}

func New(name, desc string, commands ...Command) *Cli {
	cli := &Cli{
		name:    name,
		desc:    desc,
		cmds:    map[string]Command{},
		middles: []Middleware{},
		Helper:  nil,
	}

	cli.Handle(SimpleCliHelper(cli, ""))
	for _, cmd := range commands {
		cli.Handle(cmd)
	}
	return cli
}

func (cli *Cli) Name() string {
	return cli.name
}

func (cli *Cli) Description() string {
	return cli.desc
}

func (cli *Cli) Help() string {
	if cli.Helper != nil {
		return cli.Helper()
	}
	return SimpleCliHelp(cli)
}

func (cli *Cli) Middleware(mids ...Middleware) *Cli {
	cli.middles = append(cli.middles, mids...)
	return cli
}

func (cli *Cli) Handle(cmd Command) *Cli {
	if cmd != nil {
		cli.cmds[cmd.Name()] = cmd
	}
	return cli
}

func (cli *Cli) HandleRunner(name, desc string, runner Runner) *Cli {
	return cli.Handle(NewCommand(name, desc, runner))
}

func (cli *Cli) DefCommand() Command {
	return cli.cmds[""]
}

func (cli *Cli) Run(c *Cntx, args []string) (string, error) {
	for _, mid := range cli.middles {
		c, args = mid(c, args)
	}

	if len(args) > 0 && cli.cmds[args[0]] != nil {
		return cli.cmds[args[0]].Run(c, args[1:])
	}

	if r := cli.DefCommand(); r != nil {
		return r.Run(c, args)
	}

	return "", ErrCommandNotFound
}

// example commands

type SimpleCommand struct {
	name   string
	desc   string
	Runner Runner
	Helper HelpFunc
}

func NewCommand(name, usage string, runner Runner) *SimpleCommand {
	return &SimpleCommand{name, usage, runner, nil}
}

func (c *SimpleCommand) Run(cntx *Cntx, args []string) (string, error) {
	return c.Runner(cntx, args)
}

func (c *SimpleCommand) Name() string {
	return c.name
}

func (c *SimpleCommand) Description() string {
	return c.desc
}

func (c *SimpleCommand) Help() string {
	if c.Helper != nil {
		return c.Helper()
	}
	return SimpleHeader(c.Name(), c.Description())
}

type FlagsCommand struct {
	name    string
	flagSet *flag.FlagSet
	Runner  Runner
	Helper  HelpFunc
}

func NewFlagsCommand(name string, flagSet *flag.FlagSet, runner Runner) *FlagsCommand {
	return &FlagsCommand{name, flagSet, runner, nil}
}

func (c *FlagsCommand) Name() string {
	return c.name
}
func (c *FlagsCommand) Description() string {
	return c.flagSet.Name()
}
func (c *FlagsCommand) Help() string {
	if c.Helper != nil {
		return c.Helper()
	}
	return fmt.Sprintf("%s\n%s",
		SimpleHeader(c.Name(), c.Description()),
		FlagSetUsage(c.flagSet, "  "))
}

func (c *FlagsCommand) Run(cntx *Cntx, args []string) (string, error) {
	if err := c.flagSet.Parse(args); err != nil {
		return "", err
	}
	return c.Runner(cntx, c.flagSet.Args())
}

// example command supporting concurrency

type Flagger func() (*flag.FlagSet, map[string]interface{})

type FlaggerCommand struct {
	name    string
	flagger Flagger
	Runner  Runner
	Helper  HelpFunc
}

func NewFlaggerCommand(name string, flagger Flagger, runner Runner) *FlaggerCommand {
	return &FlaggerCommand{name, flagger, runner, nil}
}

func (cmd *FlaggerCommand) Run(cntx *Cntx, args []string) (string, error) {
	flagSet, params := cmd.flagger()
	if err := flagSet.Parse(args); err != nil {
		return "", err
	}
	cx := cntx.Context()
	for name, param := range params {
		cx = context.WithValue(cx, name, param)
	}
	return cmd.Runner(cntx.With(cx), flagSet.Args())
}

func (cmd *FlaggerCommand) Name() string {
	return cmd.name
}
func (cmd *FlaggerCommand) Description() string {
	flags, _ := cmd.flagger()
	return flags.Name()
}
func (cmd *FlaggerCommand) Help() string {
	if cmd.Helper != nil {
		return cmd.Helper()
	}
	flagSet, _ := cmd.flagger()
	return fmt.Sprintf("%s\n%s", SimpleHeader(cmd.name, flagSet.Name()),
		FlagSetUsage(flagSet, "  "))
}

// utils

func SimpleHeader(name, desc string) string {
	if name == "" {
		return desc
	} else if desc == "" {
		return name
	}
	return name + ": " + desc
}

func FlagSetUsage(flagSet *flag.FlagSet, prefix string) string {
	var ret strings.Builder
	flagSet.VisitAll(func(f *flag.Flag) {
		var b strings.Builder
		fmt.Fprintf(&b, "%s-%s", prefix, f.Name)
		name, usage := flag.UnquoteUsage(f)
		if len(name) > 0 {
			b.WriteString(" ")
			b.WriteString(name)
		}
		if b.Len() <= 4 {
			b.WriteString("  ")
		} else {
			b.WriteString("\n")
			b.WriteString(prefix)
			b.WriteString("  ")
		}
		b.WriteString(strings.ReplaceAll(usage, "\n", "\n"+prefix+"  "))
		if ret.Len() > 0 {
			ret.WriteString("\n")
		}
		ret.WriteString(b.String())
	})
	return ret.String()
}

func SimpleCliHeader(cli *Cli) string {
	if cmd := cli.DefCommand(); cmd != nil {
		return cli.Name() + ": " + cmd.Help()
	}
	return SimpleHeader(cli.Name(), cli.Description())
}

func SimpleCliUsage(cli *Cli, prefix string) string {
	var ret strings.Builder
	for name, cmd := range cli.cmds {
		if name != "" {
			if ret.Len() == 0 {
				ret.WriteString("\nSubcommands:")
			}
			ret.WriteString("\n" + prefix)
			ret.WriteString(SimpleHeader(name, cmd.Description()))
		}
	}
	return ret.String()
}

func SimpleCliHelp(cli *Cli) string {
	return SimpleCliHeader(cli) + SimpleCliUsage(cli, "  ")
}

func SimpleCliHelper(cli *Cli, desc string) *SimpleCommand {
	if desc == "" {
		desc = "print help message"
	}
	cmd := NewCommand("help", desc, nil)
	cmd.Runner = func(c *Cntx, args []string) (string, error) {
		var child Command = nil
		if len(args) > 0 {
			child = cli.cmds[args[0]]
		}
		if child != nil && child != cmd {
			return child.Help(), nil
		}

		return cli.Help(), nil
	}
	return cmd
}
