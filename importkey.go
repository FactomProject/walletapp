// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/ed25519"
	fct "github.com/FactomProject/factoid"
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
	adr := args[2]

	if err := ValidName(name); err != nil {
		return err
	}

	a := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(name))
	if a != nil {
		return fmt.Errorf("That address name is in use.  Specify a different name.")
	}

	fa := fct.ValidateFPrivateUserStr(adr)
	ec := fct.ValidateECPrivateUserStr(adr)
	b, err := hex.DecodeString(adr)
	if err == nil && len(b) != 32 {
		err = fmt.Errorf("wrong length")
	}
	if fa || ec || err == nil {
		var privateKey []byte
		if err != nil {
			privateKey = fct.ConvertUserStrToAddress(adr)
		} else {
			privateKey = b
		}
		var fixed [64]byte
		copy(fixed[:], privateKey)
		publicKey := ed25519.GetPublicKey(&fixed)
		adrtype := "ec"
		if fa {
			adrtype = "fct"
		}
		_, err := state.GetFS().GetWallet().AddKeyPair(adrtype, []byte(name), publicKey[:], privateKey, false)
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
ImportKey <name> <private key>      Create a new address name <addr> with the given <private key>
`
}
