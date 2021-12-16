// Copyright (c) 2021 Jing-Ying Chen
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

	_, err := ParseArgs(&opts, words)
	if err == nil {
		t.Fatal("Should fail with unknown flag `xxx`")
	}

	line = "--ui --fmt json hello --aaa bbb"
	words = strings.Fields(line)

	var args []string
	if args, err = ParseArgs(&opts, words); err != nil {
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
	cli := New(ctx, &Cmd{
		Name:  "hello",
		Short: "Hello",
		Long:  "Hello",
		Factory: func(ctx context.Context) interface{} {
			return &helloCmd{ctx: ctx}
		},
	}, &Cmd{
		Name:  "",
		Short: "root",
		Long:  "Root Command",
		Factory: func(ctx context.Context) interface{} {
			return &rootCmd{ctx: ctx}
		},
	})

	if !cli.Is("Interactive") || !ctx.Value("Interactive").(bool) {
		t.Fatal("Should be interactive")
	}
	if cli.String("Format") != "json" || "json" != ctx.Value("Format").(string) {
		t.Fatal("Should be json")
	}

	buf := new(bytes.Buffer)
	ctx1 := WithStdout(ctx, buf)
	cli.ExecuteContext(ctx1, []string{"hello", "--name", "you"})
	if buf.String() != "Hello you" {
		t.Fatalf("Should be '%s', got '%s'", "Hello you", buf.String())
	}

	buf = new(bytes.Buffer)
	ctx1 = WithStdout(ctx, buf)
	cli.ExecuteContext(ctx1, []string{})
	if buf.String() != "This is root" {
		t.Fatalf("Should be '%s', got '%s'", "This is root", buf.String())
	}
}

type helloCmd struct {
	ctx  context.Context
	Name string `long:"name"`
}

func (c *helloCmd) Execute([]string) error {
	Printf(c.ctx, "Hello %s", c.Name)
	return nil
}

type rootCmd struct {
	ctx context.Context
}

func (c *rootCmd) Execute([]string) error {
	Printf(c.ctx, "This is root")
	return nil
}
