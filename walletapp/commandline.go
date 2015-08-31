// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/database"
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
	iroot := state.GetFS().GetWallet().GetDB().GetRaw([]byte(fct.W_SEEDS),fct.CURRENT_SEED[:])
	if iroot != nil {
		root := iroot.(database.IByteStore)
		rootkey := root.Bytes()[:32]
		state.GetFS().GetWallet().SetRoot(root.Bytes())
		nextseed := state.GetFS().GetWallet().GetDB().GetRaw([]byte(fct.W_SEED_HEADS),rootkey)
		state.GetFS().GetWallet().SetSeed(nextseed.(database.IByteStore).Bytes())
	}else{
		var seq [] int64
		var chars [] string
		reader := bufio.NewReader(os.Stdin)
		
		fmt.Print(`
          This is a new wallet
          ====================
          
In order to create secure addresses, we must generate a seed.  Please enter 10
random sequences of characters now:
`)
		for i := 1; i<=10; i++ {
			fmt.Printf("%2d: ",i)
			input, _ := reader.ReadString('\n')
			chars = append(chars,input)
			seq   = append(seq,time.Now().UnixNano())
			fmt.Println()
		}
		var randseq bytes.Buffer
		for _,v := range chars {
			randseq.WriteString(v)
		}
		for _,v := range seq {
			randseq.WriteString(fmt.Sprintf("%v",v))
		}
		state.GetFS().GetWallet().NewSeed(randseq.Bytes())
	}
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
