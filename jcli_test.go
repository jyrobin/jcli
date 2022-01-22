// Copyright (c) 2021 Jing-Ying Chen. Subject to the MIT License.

package jcli

import (
	"bytes"
	"context"
	"reflect"
	"testing"
)

func TestBasic(t *testing.T) {
	var interactive bool
	var format string
	var args []string
	cli := NewCli("Basics", "Test basics", "0").
		StringFlag("fmt", "Format", "").
		BoolFlag("ui", "Interactive", false, &interactive).
		Action(func(ctx context.Context) error {
			format = StringFlag(ctx, "fmt", "???")
			args = OtherArgs(ctx)
			return nil
		})

	ctx := context.Background()
	line := "--ui --fmt json --xxx yyy hello --aaa bbb"
	_, err := cli.RunLine(ctx, false, line)
	if err == nil {
		t.Fatal("Should fail with unknown flag `xxx`")
	}

	line = "--ui --fmt json hello --aaa bbb"
	_, err = cli.RunLine(ctx, false, line)
	if err != nil {
		t.Fatal(err)
	}
	if format != "json" {
		t.Fatalf("expect format 'json', got '%s'", format)
	}
	if !interactive {
		t.Fatalf("expect interactive is true")
	}

	rest := []string{"hello", "--aaa", "bbb"}
	if !reflect.DeepEqual(args, rest) {
		t.Fatalf("Not the same: %+v vs. %+v", args, rest)
	}

	cli2 := NewCli("Hello", "Test sub command", "0").
		Action(func(ctx context.Context) error {
			Printf(ctx, "This is root")
			return nil
		})
	cli2.NewSubCommand("hello", "Hello").
		StringFlag("name", "Name", "").
		Action(func(ctx context.Context) error {
			Printf(ctx, "Hello %s", StringFlag(ctx, "name", "???"))
			return nil
		})

	buf := new(bytes.Buffer)
	ctx2 := WithStdout(ctx, buf)
	err = cli2.Run(ctx2, "hello", "--name", "you")
	if err != nil {
		t.Fatal(err)
	}
	if buf.String() != "Hello you" {
		t.Fatalf("Should be 'Hello you', got '%s'", buf.String())
	}

	ret, err := cli2.RunBuffer(ctx, false, "hello", "--name", "you")
	if err != nil {
		t.Fatal(err)
	}
	if string(ret) != "Hello you" {
		t.Fatalf("Should be 'Hello you', got '%s'", string(ret))
	}

	ret, err = cli2.RunBuffer(ctx, false)
	if string(ret) != "This is root" {
		t.Fatalf("Should be 'This is root', got '%s'", string(ret))
	}

	defCmd := cli2.NewSubCommand("default", "Default").
		Action(func(ctx context.Context) error {
			Printf(ctx, "This is default")
			return nil
		})
	cli2.Action(nil).DefaultCommand(defCmd)

	ret, err = cli2.RunBuffer(ctx, false)
	if string(ret) != "This is default" {
		t.Fatalf("Should be 'This is default', got '%s'", string(ret))
	}
}
