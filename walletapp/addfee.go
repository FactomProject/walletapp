// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"strings"
	"bytes"
	"fmt"
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
	"github.com/FactomProject/fctwallet/Wallet/Utility"
	"strconv"
)


/************************************************************
 * AddFee
 ************************************************************/

type AddFee struct {
	ICommand
}

// AddFee <key> <name|address> amount
//
//

func (AddFee) Execute(state IState, args []string) (err error) {

	if len(args) != 3 && len(args) != 4 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	adr := args[2]
	rate := int64(0)
	if len(args) == 4 {
		srate, err := fct.ConvertFixedPoint(args[3])
		if err != nil {
			return fmt.Errorf("Could not parse exchange rate: %v",err)
		}
		rate, err = strconv.ParseInt(srate,10,64)
	}else{
		if rate, err = GetRate(state); err != nil {
			return fmt.Errorf("Could not reach the server to get the exchange rate")
		}
	}
			
	ib := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	trans, ok := ib.(fct.ITransaction)
	if ib == nil || !ok {
		return fmt.Errorf("Unknown Transaction")
	}

	var addr fct.IAddress
	
	if fct.ValidateFUserStr(adr) {
		addr = fct.NewAddress(fct.ConvertUserStrToAddress(adr))
	}else if Utility.IsValidHexAddress(adr) {
		badr,_ := hex.DecodeString(adr)
		addr = fct.NewAddress(badr)
	}else if Utility.IsValidNickname(adr) {
		we := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(adr))
		if we != nil {
			we2 := we.(wallet.IWalletEntry)
			addr, _ = we2.GetAddress()
			adr = hex.EncodeToString(addr.Bytes())
		} else {
			return fmt.Errorf("Name is undefined.")
		}
	}
	
	fee, err := trans.CalculateFee(uint64(rate))
	var tin, tout, tec uint64
	if tin,err = trans.TotalInputs(); err != nil {
		return err
	}
	if tout,err = trans.TotalOutputs(); err != nil {
		return err
	}
	if tec,err = trans.TotalECs(); err != nil {
		return err
	}
	
	if tin != tout+tec {
		msg := fmt.Sprintf("%s Total Inputs\n",fct.ConvertDecimal(tin))
		msg += fmt.Sprintf("%s Total Outputs and Entry Credits\n",fct.ConvertDecimal(tout+tec))
		msg += fmt.Sprintf("\nThe Inputs must match the outputs to use AddFee to add the fee to an input")
		return fmt.Errorf(msg)
	}
	
	
	for _,input := range trans.GetInputs() {
		if bytes.Equal(input.GetAddress().Bytes(), addr.Bytes()) {
		    input.SetAmount(input.GetAmount()+fee)
			fmt.Printf("Added fee of %v\n",strings.TrimSpace(fct.ConvertDecimal(fee)))
			break
		}
	}
	
	return nil
}

func (AddFee) Name() string {
	return "addfee"
}

func (AddFee) ShortHelp() string {
	return "AddFee <key> <name/address> -- Adds a fee to the given input.\n" +
		"                              Inputs must match Outputs, and the input address\n" +
		"                              must be one of the inputs to the transaction\n" 
}

func (AddFee) LongHelp() string {
	return `
AddFee <key> <name>                 Adds a fee to the given transaction. The name must give
                                    match an input to the transaction, and all the inputs
                                    must equal the outputs.
`
}

