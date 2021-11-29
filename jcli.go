package jcli

import (
	"context"
	"io"

	"github.com/jessevdk/go-flags"
)

const (
	parserOpts = flags.Default | flags.PassAfterNonOption
	STDOUT_KEY = "__stdout__"
)

type Cli struct {
	rootCtx context.Context
	Cmd     Cmd
}

type Cmd struct {
	Name  string
	Build func(parent *flags.Command, ctx context.Context) error
	Cmds  []Cmd
}

func (cmd Cmd) AddChild(cmds ...Cmd) {
	cmd.Cmds = append(cmd.Cmds, cmds...)
}

func New(ctx context.Context, cmds ...Cmd) *Cli {
	if ctx == nil {
		ctx = context.Background()
	}

	var cmd Cmd
	if len(cmds) > 0 {
		cmd = cmds[0]
	}

	return &Cli{ctx, cmd}
}

func (cli Cli) Context() context.Context {
	return cli.rootCtx
}

func (cli Cli) Is(key string) bool {
	b, ok := cli.rootCtx.Value(key).(bool)
	return ok && b
}

func (cli Cli) GetString(key string, defs ...string) string {
	if s, ok := cli.rootCtx.Value(key).(string); ok {
		return s
	} else if len(defs) > 0 {
		return defs[0]
	} else {
		return ""
	}
}

func (cli Cli) Execute(args []string) ([]string, error) {
	return cli.ExecuteContext(cli.Context(), args)
}

func (cli Cli) ExecuteMap(vals map[string]interface{}, args []string) ([]string, error) {
	return cli.ExecuteContext(ToContext(cli.Context(), vals), args)
}

func (cli Cli) StdoutContext(w io.Writer) context.Context {
	return context.WithValue(cli.rootCtx, STDOUT_KEY, w)
}

func (cli Cli) ExecuteOut(w io.Writer, args []string) ([]string, error) {
	return cli.ExecuteContext(cli.StdoutContext(w), args)
}

func (cli Cli) ExecuteContext(ctx context.Context, args []string) ([]string, error) {
	// build parser and commands
	parser := flags.NewParser(&struct{}{}, parserOpts)

	if err := cli.Cmd.Build(parser.Command, ctx); err != nil {
		return nil, err
	}

	return parser.ParseArgs(args)
}
