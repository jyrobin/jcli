// Copyright (c) 2021 Jing-Ying Chen. Subject to the MIT License.

package jcli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/peterh/liner"
)

func RunLoop(cli *Cli, ctx context.Context, prompt, historyPath string) error {
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

		if words[0] == "exit" || words[0] == "quit" {
			fmt.Println("Bye")
			break
		}

		if err = cli.Run(ctx, words...); err != nil {
			if err != ErrHelp {
				fmt.Println(err)
			}
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
