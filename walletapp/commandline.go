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
<<<<<<< HEAD
	"github.com/toqueteos/webbrowser"
=======
>>>>>>> 0c1aca35c0b5a6414d0243dc2b486561825a17b1
	// "golang.org/x/crypto/ssh/terminal"
)

var _ = fmt.Println
var _ fct.Transaction
var _ = time.Now

func main() {
	state := NewState("wallet_app_bolt.db")
    go startServer(state)
<<<<<<< HEAD
    webbrowser.Open("http://localhost:2337")
=======
>>>>>>> 0c1aca35c0b5a6414d0243dc2b486561825a17b1
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
