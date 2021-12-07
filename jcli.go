package jcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jessevdk/go-flags"
	"github.com/peterh/liner"
	"github.com/spf13/viper"
)

const (
	parserOpts = flags.HelpFlag | flags.PassDoubleDash | flags.PassAfterNonOption

	ViperKey     = "__viper__"
	StdoutKey    = "__stdout__"
	PrintJsonKey = "__print_json__"
)

func ParseArgs(opts interface{}, args []string) ([]string, error) {
	return flags.NewParser(opts, parserOpts).ParseArgs(args)
}

type Cmd struct {
	Name    string
	Short   string
	Long    string
	Factory func(context.Context) interface{}
	Cmds    []Cmd
}

func (cmd Cmd) AddCommand(cmds ...Cmd) {
	cmd.Cmds = append(cmd.Cmds, cmds...)
}

type Cli struct {
	rootCtx    context.Context
	Cmds       []Cmd
	RemainArgs []string
}

func New(ctx context.Context, cmds ...Cmd) *Cli {
	if ctx == nil {
		ctx = context.Background()
	}
	return &Cli{ctx, cmds, nil}
}

func (cli Cli) AddCommand(cmds ...Cmd) {
	cli.Cmds = append(cli.Cmds, cmds...)
}

func (cli Cli) Context() context.Context {
	return cli.rootCtx
}

func (cli Cli) Is(key string) bool {
	b, ok := cli.rootCtx.Value(key).(bool)
	return ok && b
}

func (cli Cli) String(key string, defs ...string) string {
	if s, ok := cli.rootCtx.Value(key).(string); ok {
		return s
	} else if len(defs) > 0 {
		return defs[0]
	} else {
		return ""
	}
}

func (cli Cli) Execute(args []string) error {
	return cli.ExecuteContext(cli.Context(), args)
}

func (cli Cli) ExecuteContext(ctx context.Context, args []string) error {
	parser := flags.NewParser(&struct{}{}, parserOpts)
	err := buildCommands(cli.Cmds, parser.Command, ctx)
	if err == nil {
		_, err = parser.ParseArgs(args)
	}

	return err
}

func (cli *Cli) ExecuteBuffer(args []string, printsJson bool) ([]byte, error) {
	buf := new(bytes.Buffer)
	ctx := WithStdout(cli.Context(), buf)
	ctx = WithValue(ctx, PrintJsonKey, printsJson)
	err := cli.ExecuteContext(ctx, args)
	return buf.Bytes(), err
}

func (cli *Cli) ExecuteLine(line string, printsJson bool) ([]byte, error) {
	words := strings.Fields(line)
	return cli.ExecuteBuffer(words, printsJson)
}

func (cli *Cli) ExecuteUnmarshal(line string, ret interface{}) error {
	buf, err := cli.ExecuteLine(line, true)
	if err == nil {
		err = json.Unmarshal(buf, ret)
	}
	return err
}

func buildCommands(cmds []Cmd, parent *flags.Command, ctx context.Context) error {
	for _, cmd := range cmds {
		var fcmd interface{}
		if cmd.Factory != nil {
			fcmd = cmd.Factory(ctx)
		}
		if fcmd == nil {
			fcmd = &struct{}{}
		}

		chcmd, err := parent.AddCommand(cmd.Name, cmd.Short, cmd.Long, fcmd)
		if err != nil {
			return err
		}

		if err = buildCommands(cmd.Cmds, chcmd, ctx); err != nil {
			return err
		}
	}
	return nil
}

func (cli Cli) ExecuteLoop(prompt, historyPath string) error {
	line := liner.NewLiner()

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered", r)
			// fmt.Println(string(debug.Stack()))
			if line != nil {
				line.Close()
				line = nil
			}
		}
	}()

	defer func() {
		line.Close()
		line = nil
	}()

	line.SetCtrlCAborts(true)

	//if historyPath == "" {
	//	fname := fmt.Sprintf(".%s_history", cli.Name())
	//	historyPath = filepath.Join(os.TempDir(), fname)
	//}
	if historyPath != "" {
		if f, err := os.Open(historyPath); err == nil {
			_, _ = line.ReadHistory(f)
			f.Close()
		}
	}

	prompt = fmt.Sprintf("[%s] ", prompt)
	for {
		cmd, err := line.Prompt(prompt)
		if err == liner.ErrPromptAborted || err == io.EOF {
			fmt.Println("Bye")
			break
		}

		if err != nil {
			fmt.Println("Error reading line: ", err)
			continue
		}

		words := strings.Fields(cmd)
		if len(words) == 0 {
			continue
		}

		if words[0] == "exit" {
			fmt.Println("Bye")
			break
		}

		if err = cli.Execute(words); err != nil {
			fmt.Println(err)
		}

		line.AppendHistory(cmd)
	}

	if historyPath != "" {
		if f, err := os.Create(historyPath); err != nil {
			fmt.Print("Error writing history file: ", err)
		} else {
			_, _ = line.WriteHistory(f)
			f.Close()
		}
	}

	return nil
}

// utils

func WithValue(ctx context.Context, key string, v interface{}) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if v != nil {
		ctx = context.WithValue(ctx, key, v)
	}
	return ctx
}

func WithStdout(ctx context.Context, w io.Writer) context.Context {
	return WithValue(ctx, StdoutKey, w)
}
func WithViper(ctx context.Context, vip *viper.Viper) context.Context {
	return WithValue(ctx, ViperKey, vip)
}

func Printf(ctx context.Context, format string, args ...interface{}) (int, error) {
	if w, ok := ctx.Value(StdoutKey).(io.Writer); ok {
		return fmt.Fprintf(w, format, args...)
	}
	return fmt.Fprintf(os.Stdout, format, args...)
}

func Println(ctx context.Context, args ...interface{}) (int, error) {
	if w, ok := ctx.Value(StdoutKey).(io.Writer); ok {
		return fmt.Fprintln(w, args...)
	}
	return fmt.Fprintln(os.Stdout, args...)
}

func PrintJson(ctx context.Context, val interface{}) (int, error) {
	buf, _ := json.MarshalIndent(val, "", "  ")
	return Println(ctx, string(buf))
}

func GetViper(ctx context.Context) *viper.Viper {
	if vip, ok := ctx.Value(ViperKey).(*viper.Viper); ok {
		return vip
	}
	return nil
}

func GetValue(ctx context.Context, key string) interface{} {
	val := ctx.Value(key)
	if val == nil {
		if vip := GetViper(ctx); vip != nil {
			val = vip.Get(key)
		}
	}
	return val
}

func GetBool(ctx context.Context, key string) bool {
	b, ok := GetValue(ctx, key).(bool)
	return ok && b
}

func GetString(ctx context.Context, key string) string {
	if s, ok := ctx.Value(key).(string); ok {
		return s
	}
	return ""
}

func ValueOr(v interface{}, ctx context.Context, keys ...string) interface{} {
	if v != nil {
		return v
	}
	if ctx != nil {
		for _, key := range keys {
			if val := ctx.Value(key); val != nil {
				return val
			}
		}
	}
	return nil
}

func StringOr(v string, ctx context.Context, keys ...string) string {
	val := ValueOr(v, ctx, keys...)
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func ValueOrViper(v interface{}, vip *viper.Viper, keys ...string) interface{} {
	if v != nil {
		return v
	}
	if vip != nil {
		for _, key := range keys {
			if val := vip.Get(key); val != nil {
				return val
			}
		}
	}
	return nil
}

func StringOrViper(v string, vip *viper.Viper, keys ...string) string {
	val := ValueOrViper(v, vip, keys...)
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}

func GetStdout(ctx context.Context) io.Writer {
	if w, ok := ctx.Value(StdoutKey).(io.Writer); ok {
		return w
	}
	return nil
}

func PrintsJson(ctx context.Context) bool {
	return GetBool(ctx, PrintJsonKey)
}

type ViperConfig struct {
	ConfigFile string
	ConfigName string
	ConfigPath string
	ConfigType string
}

func NewViper(cfg ViperConfig) (*viper.Viper, error) {
	var vip *viper.Viper
	if cfg.ConfigFile != "" { // in cfg or from command flag
		vip = viper.New()
		vip.SetConfigFile(cfg.ConfigFile)
	} else if cfg.ConfigName != "" {
		vip = viper.New()
		home := cfg.ConfigPath
		var err error
		if home == "" {
			home, err = os.UserHomeDir()
			if err != nil {
				return nil, err
			}
		}
		vip.AddConfigPath(home)
		ct := cfg.ConfigType
		if ct == "" {
			ct = "yaml"
		}
		vip.SetConfigType(ct)
		vip.SetConfigName(cfg.ConfigName)
	} else {
		return nil, fmt.Errorf("Both ConfigFile and ViperName empty")
	}

	vip.AutomaticEnv()

	if err := vip.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	return vip, nil
}
