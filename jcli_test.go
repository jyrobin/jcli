package jcli

import (
	"bytes"
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/jessevdk/go-flags"
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

	vals, err := ToMap(opts)
	exp := map[string]interface{}{
		"Interactive": true,
		"Format":      "json",
	}
	if err != nil || !reflect.DeepEqual(vals, exp) {
		t.Fatalf("Should equal: %+v vs. %+v", vals, exp)
	}

	ctx := ToContext(nil, vals)
	cli := New(ctx, Cmd{
		Build: buildHelloCmd,
	})

	if !cli.Is("Interactive") || !ctx.Value("Interactive").(bool) {
		t.Fatal("Should be interactive")
	}
	if cli.GetString("Format") != "json" || "json" != ctx.Value("Format").(string) {
		t.Fatal("Should be json")
	}

	buf := new(bytes.Buffer)
	cli.ExecuteOut(buf, []string{"hello", "--name", "you"})
	if buf.String() != "Hello you" {
		t.Fatalf("Should be '%s', got '%s'", buf.String(), "Hello you")
	}
}

func buildHelloCmd(parent *flags.Command, ctx context.Context) error {
	parent.AddCommand("hello", "Hi", "Hello", &helloCmd{ctx, ""})
	return nil
}

type helloCmd struct {
	ctx  context.Context
	Name string `long:"name"`
}

func (c *helloCmd) Execute(args []string) error {
	Printf(c.ctx, "Hello %s", c.Name)
	return nil
}
