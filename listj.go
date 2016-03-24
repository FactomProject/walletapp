// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/fctwallet/Wallet/Utility"
)

/************************************************************
 * List
 ************************************************************/

type Listj struct {
}

var _ ICommand = (*List)(nil)

// List transactions <address list>
func (Listj) Execute(state IState, args []string) (err error) {
	if len(args) <= 1 {
		return fmt.Errorf("Nothing to list")
	}
	switch args[1] {
	case "all":
		fmt.Println("Listing all transactions: ")
		var list []byte
		if list, err = Utility.DumpTransactionsJSON(nil, 0, 0); err != nil {
			return err
		}
		fmt.Print(string(list))
		break
	default:
		var addresses [][]byte
		var adr string
		for i := 1; i < len(args); i++ {
			adr, err = LookupAddress(state, "FA", args[i])
			if err != nil {
				adr, err = LookupAddress(state, "EC", args[i])
				if err != nil {
					return fmt.Errorf("Could not understand address %s", args[i])
				}
			}
			badr, err := hex.DecodeString(adr)
			if err != nil {
				return fmt.Errorf("Could not understand address %s", args[i])
			}
			addresses = append(addresses, badr)
		}
		var list []byte
		if list, err = Utility.DumpTransactionsJSON(addresses, 0, 0); err != nil {
			return err
		}
		fmt.Print(string(list))
	}
	return nil
}

func (Listj) Name() string {
	return "Listj"
}

func (Listj) ShortHelp() string {
	return "Listj [all|address]  -- [all] prints all transactions in JSON, while specifying\n" +
		"                       limits transactions to those that involve the given\n" +
		"                       address.  All address types supported."

}

func (Listj) LongHelp() string {
	return `
Listj [all | address]               [all] prints all the factom transactions in JSON to date, 
                                    while address limits output to transactions that reference
                                    the address.  All address types are supported.
`
}
