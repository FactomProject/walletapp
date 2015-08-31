// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/ed25519"
)

/************************************************************
 * Import
 ************************************************************/

type ImportKey struct {
	
}

var _ ICommand = (*ImportKey)(nil)

// ImportKey <name> <private key>
func (ImportKey) Execute(state IState, args []string) error {

	if len(args) != 3 {
		return fmt.Errorf("Invalid Parameters")
	}
	name := args[1]
	adr  := args[2]
	
	a := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(name))
	if a != nil {
		return fmt.Errorf("That address name is in use.  Specify a different name.")
	}
	
	fa := fct.ValidateFPrivateUserStr(adr) 
	ec := fct.ValidateECPrivateUserStr(adr) 
	if fa || ec {
		privateKey := fct.ConvertUserStrToAddress(adr)
		var fixed [64]byte
		copy(fixed[:],privateKey)
		publicKey := ed25519.GetPublicKey(&fixed)
		adrtype := "ec"
		if fa {
			adrtype = "fct"
		}
		_, err := state.GetFS().GetWallet().AddKeyPair(adrtype, []byte(name),publicKey[:],privateKey,false)
		if err != nil {
			return err
		}
		return nil
	}
	
	return fmt.Errorf("Not a valid Private Key; Check that your key is typed correctly")
}

func (ImportKey) Name() string {
	return "importkey"
}

func (ImportKey) ShortHelp() string {
	return "ImportKey <name> <private key> -- Create a new address <key> with the given <private key>"
	
}

func (ImportKey) LongHelp() string {
	return `
	ImportKey <name> <private key>     Create a new address name <addr> with the given <private key>
`
}



