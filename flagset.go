package jcli

import (
	"context"
	"flag"
	"fmt"
	"strings"
)

type flagValues struct {
	flags  *flag.FlagSet
	values map[string]interface{}
}

type flagProto struct {
	name        string
	description string
	value       interface{} // default value
	ptr         interface{} // type should match value
}

func (fp *flagProto) addFlag(flags *flag.FlagSet, vals map[string]interface{}) {
	switch v := fp.value.(type) {
	case string:
		if ptr, ok := fp.ptr.(*string); ok && ptr != nil {
			flags.StringVar(ptr, fp.name, v, fp.description)
			vals[fp.name] = ptr
		} else {
			vals[fp.name] = flags.String(fp.name, v, fp.description)
		}
	case int:
		if ptr, ok := fp.ptr.(*int); ok && ptr != nil {
			flags.IntVar(ptr, fp.name, v, fp.description)
			vals[fp.name] = ptr
		} else {
			vals[fp.name] = flags.Int(fp.name, v, fp.description)
		}
	case float64:
		if ptr, ok := fp.ptr.(*float64); ok && ptr != nil {
			flags.Float64Var(ptr, fp.name, v, fp.description)
			vals[fp.name] = ptr
		} else {
			vals[fp.name] = flags.Float64(fp.name, v, fp.description)
		}

	case bool:
		if ptr, ok := fp.ptr.(*bool); ok && ptr != nil {
			flags.BoolVar(ptr, fp.name, v, fp.description)
			vals[fp.name] = ptr
		} else {
			vals[fp.name] = flags.Bool(fp.name, v, fp.description)
		}
	}
}

type flagSet struct {
	protos map[string]*flagProto
}

func newFlagSet() *flagSet {
	return &flagSet{make(map[string]*flagProto)}
}

func (fs *flagSet) flagCount() int {
	return len(fs.protos)
}

func (fs *flagSet) addFlag(name, description string, val interface{}, ptr interface{}) {
	fs.protos[name] = &flagProto{name, description, val, ptr}
}

func (fs *flagSet) parseFlags(ctx context.Context, commandPath string, args []string) (context.Context, error) {
	flags := flag.NewFlagSet(commandPath, flag.ContinueOnError)
	vals := make(map[string]interface{})
	for _, proto := range fs.protos {
		proto.addFlag(flags, vals)
	}

	// add help flag here for the commandPath value; fix later
	vals["help"] = flags.Bool("help", false,
		"Get help on the '"+strings.ToLower(commandPath)+"' command.")

	flags.SetOutput(Stdout(ctx))
	if err := flags.Parse(args); err != nil {
		return ctx, err
	}

	return context.WithValue(ctx, FlagValuesKey, &flagValues{flags, vals}), nil
}

func (fs *flagSet) printDefaults(ctx context.Context) {
	if flagVals := getFlagValues(ctx); flagVals != nil {
		out := Stdout(ctx)
		fmt.Fprintln(out, "Flags:")
		fmt.Fprintln(out)
		// flagVals.flags.SetOutput(Stdout(ctx)) // set already
		flagVals.flags.PrintDefaults()
	}
}
