/*************************************************************************
 * Handler Functions
 *************************************************************************/

// Setup:  seed --
// Setup creates the 10 fountain Factoid Addresses, then sets address
// generation to be unique for this wallet.  You CAN call setup multiple
// times, but once the Fountain addresses are created, Setup only changes
// the seed.
//
// Setup must be called once before you do anything else with the wallet.
//

// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package main

/*
import (
	"fmt"
	fct "github.com/FactomProject/factoid"
	"time"
	// "golang.org/x/crypto/ssh/terminal"
)

var _ = fmt.Println
var _ fct.Transaction
var _ = time.Now
*/
/*************************************************************
 * run a Script
 *************************************************************/
/*
type Setup struct {
}

var _ ICommand = (*Setup)(nil)

func (r Setup) Execute(state IState, args []string) error {
	// Make sure we have a seed.
	if len(args) != 2 {
		msg := "You must supply some random seed. For example (don't use this!)\n" +
		"factom-cli setup 'woe!#in31!%234ng)%^&$%oeg%^&*^jp45694a;gmr@#t4 q34y'\n" +
		"would make a nice seed.  The more random the better.\n\n" +
		"Note that if you create an address before you call Setup, you must\n" +
		"use those address(s) as you access the fountians."

		return fmt.Errorf(msg)
	}
	setFountian := false
	keys, _ := state.GetFS().GetWallet().GetDB().GetKeysValues([]byte(fct.W_NAME))
	if len(keys) == 0 {
		setFountian = true
		for i := 1; i <= 10; i++ {
			name := fmt.Sprintf("%02d-Fountain", i)
			err := GenAddress(state,"fct",name)
			if err != nil {
				fmt.Println(err)
				return nil
			}
		}
	}

	seedprime := fct.Sha([]byte(fmt.Sprintf("%s%v", args[1], time.Now().UnixNano()))).Bytes()
	NewSeed(state, seedprime)

	if setFountian {
		fmt.Println("New seed set, fountain addresses defined")
	} else {
		fmt.Println("New seed set, no fountain addresses defined")
	}
	return nil
}

func (r Setup) Name() string {
	return "setup"
}

func (Setup) ShortHelp() string {
	return "setup <seed>                -- Sets up Fountain addresses, and seeds the wallet\n"+
	       "                               This only works if no addresses have been created.\n"+
		   "                               If addresses exist, no Fountain addresses are created.\n"+
		   "                               But the wallet gets a new seed.\n"
}

func (Setup) LongHelp() string {
	return `
Setup <seed>                        Sets up Fountain addresses, and seeds the wallet.
                                    This only works if no addresses have been created.
                                    If addresses exist, no Fountain addresses are created.
                                    But the wallet gets a new seed in any case.
`
}

func NewSeed(state IState, data []byte) {
	state.GetFS().GetWallet().NewSeed(data)
}
*/
