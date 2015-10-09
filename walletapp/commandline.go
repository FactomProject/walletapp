// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"fmt"
	"io"
	fct "github.com/FactomProject/factoid"
	"os"
	"strings"
	"time"
	// "golang.org/x/crypto/ssh/terminal"
)

var _ = fmt.Println
var _ fct.Transaction
var _ = time.Now

func main() {
	    state := NewState("wallet_app_bolt.db")
        go startServer(state)
        Open("http://localhost:2337")
	    run(state, os.Stdin,true)
}
	
var fsprompt string = "===============> "	
	
func run(state IState, reader io.Reader, prompt bool){	
	r := bufio.NewScanner(reader)
	if prompt {
		fmt.Print(fsprompt)
	}
	for r.Scan() {
		line := r.Text()
		args := strings.Fields(string(line))
		err := state.Execute(args)
		if err != nil {
			fmt.Println(err)
			c, _ := state.GetCommand(args)
			if c != nil {
				fmt.Println(c.ShortHelp())
			}
		}
		if prompt {
			fmt.Print(fsprompt)
		}
	}
	if prompt {
		fmt.Println()
	}
}
