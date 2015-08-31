// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	fct "github.com/FactomProject/factoid"
)

/************************************************************
 * Sign
 ************************************************************/
type Sign struct {
	ICommand
}

// Sign <k>
//
// Sign the given transaction identified by the given key
func (Sign) Execute(state IState, args []string) error {
	if len(args) != 2 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	// Get the transaction
	ib := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	trans, ok := ib.(fct.ITransaction)
	if !ok {
		return fmt.Errorf("Invalid Parameters")
	}
	
	err := state.GetFS().GetWallet().Validate(1, trans)
	if err != nil {
		return err
	}
	
	ok, err = state.GetFS().GetWallet().SignInputs(trans)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("Error signing the transaction")
	}
	
	// Update our map with our new transaction to the same key.  Otherwise, all
	// of our work will go away!
	state.GetFS().GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)
	
	return nil
	
}

func (Sign) Name() string {
	return "Sign"
}

func (Sign) ShortHelp() string {
	return "Sign <k> -- Sign the transaction given by the key <k>"
}

func (Sign) LongHelp() string {
	return `
	Sign <key>                          Signs the transaction specified by the given key.
	Each input is found within the wallet, and if 
	we have the private key for that input, we 
	sign for that input.  
	
	Transctions can have inputs from multiple parties.
	In this case, the inputs can be signed by each
	party by first creating all the inputs and 
	outputs for a transaction.  Then signing your
	inputs.  Exporting the transaction.  Then
	sending it to the other party or parties for
	their signatures.
	`
}

