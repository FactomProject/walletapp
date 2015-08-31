// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	"encoding/hex"
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
)

/************************************************************
 * ExportKey
 ************************************************************/

type ExportKey struct {
	
}

var _ ICommand = (*ExportKey)(nil)

// ExportKey <name> 
func (ExportKey) Execute(state IState, args []string) error {

	if len(args) != 2 {
		return fmt.Errorf("Invalid Parameters")
	}
	name := args[1]
	
	weblk := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(name))
	if weblk == nil {
		return fmt.Errorf("Unknown address.  Check that you spelled the name correctly")
	}
	we := weblk.(wallet.IWalletEntry)
	public := we.GetKey(0)
	private := we.GetPrivKey(0)
	adrtype := we.GetType()
	
	binPublic := hex.EncodeToString(public)
	binPrivate := hex.EncodeToString(private[:32])
	var usrPublic, usrPrivate string
	if adrtype == "fct" {
		usrPublic = fct.ConvertFctAddressToUserStr(fct.NewAddress(public))
		usrPrivate = fct.ConvertFctPrivateToUserStr(fct.NewAddress(private))
	}else{
		usrPublic = fct.ConvertECAddressToUserStr(fct.NewAddress(public))
		usrPrivate = fct.ConvertECPrivateToUserStr(fct.NewAddress(private))
	}
	
	fmt.Println("Private Key:")
	fmt.Println("  ",usrPrivate)
	fmt.Println("  ",binPrivate)
	fmt.Println("Public Key:")
	fmt.Println("  ",usrPublic)
	fmt.Println("  ",binPublic)
	return nil
}
	

func (ExportKey) Name() string {
	return "exportkey"
}

func (ExportKey) ShortHelp() string {
	return "ExportKey <name>  -- Prints the private and public keys tied to this <name>"
	
}

func (ExportKey) LongHelp() string {
	return `
ExportKey <name>                    Prints the public and private keys tied to this <name>.
`
}



