// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	fct "github.com/FactomProject/factoid"
	"os"
	"time"
	// "golang.org/x/crypto/ssh/terminal"
)

var _ = fmt.Println
var _ fct.Transaction
var _ = time.Now

/*************************************************************
 * run a Script
 *************************************************************/

type Run struct {
}

var _ ICommand = (*Run)(nil)

func (r Run) Execute(state IState, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Wrong number of arguments")
	}

	f, err := os.Open(args[1])
	if err != nil {
		return err
	}

	run(state, bufio.NewReader(f), false)

	return nil
}

func (r Run) Name() string {
	return "run"
}

func (Run) ShortHelp() string {
	return "Run <filename>              -- Executes the script of the given filename"
}

func (Run) LongHelp() string {
	return `
Run <filename>                      Executes the script of the given filename
`
}
