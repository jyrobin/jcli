package jcli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/jyrobin/goutil"
)

func ParseArgs(opts interface{}, args []string) ([]string, error) {
	return flags.NewParser(opts, parserOpts).ParseArgs(args)
}

func ToMap(v interface{}) (map[string]interface{}, error) {
	return goutil.StructToMap(v)
}

func ToContext(ctx context.Context, vals map[string]interface{}) context.Context {
	return goutil.ExtendContext(ctx, vals)
}

func Printf(ctx context.Context, format string, args ...interface{}) (int, error) {
	if w, ok := ctx.Value(STDOUT_KEY).(io.Writer); ok {
		return fmt.Fprintf(w, format, args...)
	}
	return fmt.Fprintf(os.Stdout, format, args...)
}

func Println(ctx context.Context, args ...interface{}) (int, error) {
	if w, ok := ctx.Value(STDOUT_KEY).(io.Writer); ok {
		return fmt.Fprintln(w, args...)
	}
	return fmt.Fprintln(os.Stdout, args...)
}
