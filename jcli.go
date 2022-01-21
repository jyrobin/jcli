package jcli

// Adapted from clir https://github.com/leaanthony/clir
// MIT License, Copyright (c) 2019 Lea Anthony

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

const (
	FlagValuesKey = "__flag_values__"
	StdoutKey     = "__stdout__"
	PrintJsonKey  = "__print_json__"
)

// defaultBannerFunction prints a banner for the application.
// If version is a blank string, it is ignored.
func defaultBannerFunction(ctx context.Context, c *Cli) string {
	version := ""
	if len(c.Version()) > 0 {
		version = " " + c.Version()
	}
	return fmt.Sprintf("%s%s - %s", c.Name(), version, c.ShortDescription())
}

func getFlagValues(ctx context.Context) *flagValues {
	if flagVals, ok := ctx.Value(FlagValuesKey).(*flagValues); ok {
		return flagVals
	}
	return nil
}

func getValuePointer(ctx context.Context, name string) (interface{}, bool) {
	if flagVals := getFlagValues(ctx); flagVals != nil {
		if ptr, ok := flagVals.values[name]; ok {
			return ptr, true
		}
	}
	return nil, false
}

func IntFlag(ctx context.Context, name string, otherwise int) int {
	if ptr, ok := getValuePointer(ctx, name); ok {
		if ret, ok := ptr.(*int); ok {
			return *ret
		}
	}
	return otherwise
}

func FloatFlag(ctx context.Context, name string, otherwise float64) float64 {
	if ptr, ok := getValuePointer(ctx, name); ok {
		if ret, ok := ptr.(*float64); ok {
			return *ret
		}
	}
	return otherwise
}

func StringFlag(ctx context.Context, name, otherwise string) string {
	if ptr, ok := getValuePointer(ctx, name); ok {
		if ret, ok := ptr.(*string); ok {
			return *ret
		}
	}
	return otherwise
}

// StringFlags is a convenient function that calls StringFlag with multiple
// names and empty string as the default value
func StringFlags(ctx context.Context, names ...string) []string {
	ret := make([]string, 0, len(names))
	for _, name := range names {
		ret = append(ret, StringFlag(ctx, name, ""))
	}
	return ret
}

func BoolFlag(ctx context.Context, name string, otherwise bool) bool {
	if ptr, ok := getValuePointer(ctx, name); ok {
		if ret, ok := ptr.(*bool); ok {
			return *ret
		}
	}
	return otherwise
}

func HelpFlag(ctx context.Context) bool {
	return BoolFlag(ctx, "help", false)
}

func OtherArgs(ctx context.Context) []string {
	if flagVals := getFlagValues(ctx); flagVals != nil {
		return flagVals.flags.Args()
	}
	return nil
}

func Stdout(ctx context.Context) io.Writer {
	if w, ok := ctx.Value(StdoutKey).(io.Writer); ok && w != nil {
		return w
	}
	return os.Stdout
}

func PrintsJson(ctx context.Context) bool {
	b, ok := ctx.Value(PrintJsonKey).(bool)
	return ok && b
}

func WithStdout(ctx context.Context, w io.Writer) context.Context {
	return context.WithValue(ctx, StdoutKey, w)
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

func Printj(ctx context.Context, fmt string, val interface{}, rest ...interface{}) error {
	var err error
	if PrintsJson(ctx) {
		_, err = PrintJson(ctx, val)
	} else {
		_, err = Printf(ctx, fmt, append([]interface{}{val}, rest...))
	}
	return err
}
