// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"fmt"
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
	"strconv"
)

/************************************************************
 * NewTransaction
 ************************************************************/

type NewTransaction struct {
	ICommand
}

// New Transaction:  key --
// We create a new transaction, and track it with the user supplied key.  The
// user can then use this key to make subsequent calls to add inputs, outputs,
// and to sign. Then they can submit the transaction.
//
// When the transaction is submitted, we clear it from our working memory.
// Multiple transactions can be under construction at one time, but they need
// their own keys. Once a transaction is either submitted or deleted, the key
// can be reused.
func (NewTransaction) Execute(state IState, args []string) error {

	if len(args) != 2 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]

	// Make sure we don't already have a transaction in process with this key
	t := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	if t != nil {
		return fmt.Errorf("Duplicate key: '%s'", key)
	}
	// Create a transaction
	t = state.GetFS().GetWallet().CreateTransaction(state.GetFS().GetTimeMilli())
	// Save it with the key
	state.GetFS().GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), t)

	fmt.Println("Beginning Transaction ", key)
	return nil
}

func (NewTransaction) Name() string {
	return "NewTransaction"
}

func (NewTransaction) ShortHelp() string {
	return "NewTransaction <key> -- Begins the construction of a transaction.\n" +
		"                        Subsequent modifications must reference the key."
}

func (NewTransaction) LongHelp() string {
	return `
NewTransaction <key>                Begins the construction of a transaction.
                                    The <key> is any token without whitespace up to
                                    32 characters in length that can be used in 
                                    AddInput, AddOutput, AddECOutput, Sign, and
                                    Submit commands to construct and submit 
                                    transactions.
`
}

/************************************************************
 * AddInput
 ************************************************************/

type AddInput struct {
	ICommand
}

// AddInput <key> <name|address> amount
//
//

func (AddInput) Execute(state IState, args []string) error {

	if len(args) != 4 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	adr := args[2]
	amt := args[3]

	ib := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	trans, ok := ib.(fct.ITransaction)
	if ib == nil || !ok {
		return fmt.Errorf("Unknown Transaction")
	}
		
	var addr fct.IAddress
	if !fct.ValidateFUserStr(adr) {
		if len(adr) != 64 {
			if len(adr) > 32 {
				return fmt.Errorf("Invalid Name.  Name is too long: %v characters", len(adr))
			}

			we := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(adr))

			if we != nil {
				we2 := we.(wallet.IWalletEntry)
				addr, _ = we2.GetAddress()
				adr = hex.EncodeToString(addr.Bytes())
			} else {
				return fmt.Errorf("Name is undefined.")
			}
		} else {
			badr,err := hex.DecodeString(adr)
			if err != nil {
				return fmt.Errorf("Invalid hex string: %s", err.Error())
			}
			addr = fct.NewAddress(badr)
		}
	} else {
		fmt.Printf("adr: %x\n",adr)
		addr = fct.NewAddress(fct.ConvertUserStrToAddress(adr))
	}
	amount, _ := fct.ConvertFixedPoint(amt)
	bamount, _ := strconv.ParseInt(amount, 10, 64)
	err := state.GetFS().GetWallet().AddInput(trans, addr, uint64(bamount))

	if err != nil {
		return err
	}

	fmt.Println("Added Input of ", amt, " to be paid from ", args[2],
		fct.ConvertFctAddressToUserStr(addr))
	return nil
}

func (AddInput) Name() string {
	return "AddInput"
}

func (AddInput) ShortHelp() string {
	return "AddInput <key> <name/address> <amount> -- Adds an input to a transaction.\n" +
		"                              the key should be created by NewTransaction, and\n" +
		"                              and the address and amount should come from your\n" +
		"                              wallet."
}

func (AddInput) LongHelp() string {
	return `
AddInput <key> <name|addr> <amt>    <key>       created by a previous NewTransaction call
                                    <name|addr> A Valid Name in your Factoid Address 
                                                book, or a valid Factoid Address
                                    <amt>       to be sent from the specified address to the
                                                outputs of this transaction.
`
}

/************************************************************
 * AddOutput
 ************************************************************/
type AddOutput struct {
	ICommand
}

// AddOutput <key> <name|address> amount
//
//

func (AddOutput) Execute(state IState, args []string) error {

	if len(args) != 4 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	adr := args[2]
	amt := args[3]

	ib := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	trans, ok := ib.(fct.ITransaction)
	if ib == nil || !ok {
		return fmt.Errorf("Unknown Transaction")
	}

	var addr fct.IAddress
	if !fct.ValidateFUserStr(adr) {
		if len(adr) != 64 {
			if len(adr) > 32 {
				return fmt.Errorf("Invalid Name.  Name is too long: %v characters", len(adr))
			}

			we := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(adr))

			if we != nil {
				we2 := we.(wallet.IWalletEntry)
				addr, _ = we2.GetAddress()
				adr = hex.EncodeToString(addr.Bytes())
			} else {
				return fmt.Errorf("Name is undefined.")
			}
		} else {
			if badHexChar.FindStringIndex(adr) != nil {
				return fmt.Errorf("Invalid Name.  Name is too long: %v characters", len(adr))
			}
		}
	} else {
		addr = fct.NewAddress(fct.ConvertUserStrToAddress(adr))
	}
	amount, _ := fct.ConvertFixedPoint(amt)
	bamount, _ := strconv.ParseInt(amount, 10, 64)
	err := state.GetFS().GetWallet().AddOutput(trans, addr, uint64(bamount))
	if err != nil {
		return err
	}

	fmt.Println("Added Output of ", amt, " to be paid to ", args[2],
		fct.ConvertFctAddressToUserStr(addr))

	return nil
}

func (AddOutput) Name() string {
	return "AddOutput"
}

func (AddOutput) ShortHelp() string {
	return "AddOutput <k> <n> <amount> -- Adds an output to a transaction.\n" +
		"                              the key <k> should be created by NewTransaction.\n" +
		"                              The address or name <n> can come from your address\n" +
		"                              book."
}

func (AddOutput) LongHelp() string {
	return `
AddOutput <key> <n|a> <amt>         <key>  created by a previous NewTransaction call
                                    <n|a>  A Valid Name in your Factoid Address 
                                           book, or a valid Factoid Address 
                                    <amt>  to be used to purchase Entry Credits at the
                                           current exchange rate.
`
}

/************************************************************
 * AddECOutput
 ************************************************************/
type AddECOutput struct {
	ICommand
}

// AddECOutput <key> <name|address> amount
//
// Buy Entry Credits

func (AddECOutput) Execute(state IState, args []string) error {

	if len(args) != 4 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	adr := args[2]
	amt := args[3]

	ib := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	trans, ok := ib.(fct.ITransaction)
	if ib == nil || !ok {
		return fmt.Errorf("Unknown Transaction")
	}

	var addr fct.IAddress
	if !fct.ValidateECUserStr(adr) {
		if len(adr) != 64 {
			if len(adr) > 32 {
				return fmt.Errorf("Invalid Name.  Name is too long: %v characters", len(adr))
			}

			we := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(adr))

			if we != nil {
				we2 := we.(wallet.IWalletEntry)
				addr, _ = we2.GetAddress()
				adr = hex.EncodeToString(addr.Bytes())
			} else {
				return fmt.Errorf("Name is undefined.")
			}
		} else {
			if badHexChar.FindStringIndex(adr) != nil {
				return fmt.Errorf("Invalid Name.  Name is too long: %v characters", len(adr))
			}
		}
	} else {
		addr = fct.NewAddress(fct.ConvertUserStrToAddress(adr))
	}
	amount, _ := fct.ConvertFixedPoint(amt)
	bamount, _ := strconv.ParseInt(amount, 10, 64)
	err := state.GetFS().GetWallet().AddECOutput(trans, addr, uint64(bamount))
	if err != nil {
		return err
	}

	fmt.Println("Added Output of ", amt, " to be paid to ", args[2],
		fct.ConvertECAddressToUserStr(addr))

	return nil
}

func (AddECOutput) Name() string {
	return "AddECOutput"
}

func (AddECOutput) ShortHelp() string {
	return "AddECOutput <k> <n> <amount> -- Adds an Entry Credit output (ecoutput)to a \n" +
		"                              transaction <k>.  The Entry Credits are assigned to\n" +
		"                              the address <n>.  The output <amount> is specified in\n" +
		"                              factoids, and purchases Entry Credits according to\n" +
		"                              the current exchange rage."
}

func (AddECOutput) LongHelp() string {
	return `
AddECOutput <key> <n|a> <amt>       <key>  created by a previous NewTransaction call
                                    <n|a>  Name or Address to hold the Entry Credits
                                    <amt>  Amount of Factoids to be used in this purchase.  Note
                                           that the exchange rate between Factoids and Entry
                                           Credits varies.
`
}




