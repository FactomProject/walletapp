// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
	"github.com/FactomProject/ed25519"
)

/************************************************************
 * AddressFromWords
 ************************************************************/

type AddressFromWords struct {
	
}

var _ ICommand = (*AddressFromWords)(nil)

// New Transaction:  key --
// We create a new transaction, and track it with the user supplied key.  The
// user can then use this key to make subsequent calls to add inputs, outputs,
// and to sign. Then they can submit the transaction.
//
// When the transaction is submitted, we clear it from our working memory.
// Multiple transactions can be under construction at one time, but they need
// their own keys. Once a transaction is either submitted or deleted, the key
// can be reused.
func (AddressFromWords) Execute(state IState, args []string) error {

	if len(args) != 14 {
		return fmt.Errorf("Invalid Parameters")
	}
	name := args[1]
	
	na := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(name))
	if na != nil {
		return fmt.Errorf("The name %s is already in use.  Names must be unique.",name)
	}
	
	var mnstr string
	for i:=2;i<14;i++ {
		if len(args[i])==0 {
			return fmt.Errorf("Invalid mnemonic; the %d word has issues",i+1)
		}
		mnstr = mnstr+args[i]+" "
	}
	
	privateKey,err :=  wallet.MnemonicStringToPrivateKey(mnstr) 
	if err != nil {
		return err
	}
	
	var fixed [64]byte
	copy(fixed[:],privateKey)
	publicKey := ed25519.GetPublicKey(&fixed)
	
	state.GetFS().GetWallet().AddKeyPair("fct",[]byte(name),publicKey[:],fixed[:],false)

	return nil
	
}

func (AddressFromWords) Name() string {
	return "addressfromwords"
}

func (AddressFromWords) ShortHelp() string {
	return "AddressFromWords <name> <12 words> -- Export the given transactiion to the given filename."
}

func (AddressFromWords) LongHelp() string {
	return `
AddressFromWords <name> <12 words>  Compute an address from 12 words, and assign to <name>.
`
}



