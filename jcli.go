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
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	FlagValuesKey = "__flag_values__"
	StdoutKey     = "__stdout__"
	PrintJsonKey  = "__print_json__"
	QuietKey      = "__quiet__"
)

var ErrHelp = errors.New("jcli: help requested")

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

func Quiet(ctx context.Context) bool {
	b, ok := ctx.Value(QuietKey).(bool)
	return ok && b
}

func PrintsJson(ctx context.Context) bool {
	b, ok := ctx.Value(PrintJsonKey).(bool)
	return ok && b
}

func WithStdout(ctx context.Context, w io.Writer) context.Context {
	return context.WithValue(ctx, StdoutKey, w)
}

func Printf(ctx context.Context, format string, args ...interface{}) error {
	var err error
	if w, ok := ctx.Value(StdoutKey).(io.Writer); ok {
		_, err = fmt.Fprintf(w, format, args...)
	} else {
		_, err = fmt.Fprintf(os.Stdout, format, args...)
	}
	return err
}

func Println(ctx context.Context, args ...interface{}) error {
	var err error
	if w, ok := ctx.Value(StdoutKey).(io.Writer); ok {
		_, err = fmt.Fprintln(w, args...)
	} else {
		_, err = fmt.Fprintln(os.Stdout, args...)
	}
	return err
}

func PrintJson(ctx context.Context, val interface{}, opts ...string) error {
	var buf []byte
	if len(opts) == 0 {
		buf, _ = json.Marshal(val)
	} else if len(opts) == 1 {
		buf, _ = json.MarshalIndent(val, "", opts[0])
	} else {
		buf, _ = json.MarshalIndent(val, opts[1], opts[0])
	}
	return Println(ctx, string(buf))
}

func Printj(ctx context.Context, fmt string, val interface{}, rest ...interface{}) error {
	if Quiet(ctx) {
		return nil
	} else if PrintsJson(ctx) || fmt == "" {
		return PrintJson(ctx, val, "  ")
	} else {
		return Printf(ctx, fmt, append([]interface{}{val}, rest...))
	}
}
