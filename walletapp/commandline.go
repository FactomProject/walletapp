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
	"runtime"
	"strings"
	"time"
)

var _ = fmt.Println
var _ fct.Transaction
var _ = time.Now

func main() {
 	    var configDir string
        var staticDir string
        switch runtime.GOOS {
            case "windows":
                configDir = os.Getenv("HOME") + "\\.factom\\"
                staticDir = configDir + "walletapp/"
            case "darwin":
                configDir = os.Getenv("HOME") + "/.factom/"
                staticDir = "./staticfiles/"
            default:
                configDir = os.Getenv("HOME") + "/.factom/"
                staticDir = "/usr/share/factom/walletapp/"
        }
	    state := NewState(configDir + "factoid_wallet_bolt.db")
        go startServer(state, staticDir)
        Open("http://localhost:8093")
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
