package jcli

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/jyrobin/goutil"
)

func TestBasic(t *testing.T) {
	opts := struct {
		Interactive bool   `long:"ui"`
		Format      string `long:"fmt"`
	}{}

	line := "--ui --fmt json --xxx yyy hello --aaa bbb"
	words := strings.Fields(line)

	args, err := ParseArgs(&opts, words)
	if err == nil {
		t.Fatal("Should fail with unknown flag `xxx`")
	}

	line = "--ui --fmt json hello --aaa bbb"
	words = strings.Fields(line)

	args, err = ParseArgs(&opts, words)
	if err != nil {
		t.Fatal(err)
	}

	rest := []string{"hello", "--aaa", "bbb"}
	if !reflect.DeepEqual(args, rest) {
		t.Fatalf("Not the same: %+v vs. %+v", args, rest)
	}

	vals, err := goutil.StructToMap(opts)
	exp := map[string]interface{}{
		"Interactive": true,
		"Format":      "json",
	}
	if err != nil || !reflect.DeepEqual(vals, exp) {
		t.Fatalf("Should equal: %+v vs. %+v", vals, exp)
	}

	ctx := goutil.ContextWithMap(nil, vals)
	cli := New(ctx, Cmd{
		Name:    "hello",
		Short:   "Hello",
		Long:    "Hello",
		Factory: HelloCmd,
	})

	if !cli.Is("Interactive") || !ctx.Value("Interactive").(bool) {
		t.Fatal("Should be interactive")
	}
	if cli.String("Format") != "json" || "json" != ctx.Value("Format").(string) {
		t.Fatal("Should be json")
	}

	buf := new(bytes.Buffer)
	ctx = WithStdout(ctx, buf)
	cli.ExecuteContext(ctx, []string{"hello", "--name", "you"})
	if buf.String() != "Hello you" {
		t.Fatalf("Should be '%s', got '%s'", buf.String(), "Hello you")
	}
}

type helloCmd struct {
	ctx  context.Context
	Name string `long:"name"`
}

func HelloCmd(ctx context.Context) interface{} {
	return &helloCmd{ctx: ctx}
}

func (c *helloCmd) Execute(args []string) error {
	Printf(c.ctx, "Hello %s", c.Name)
	return nil
}
