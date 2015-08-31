// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"fmt"
	fct "github.com/FactomProject/factoid"
	"strings"
)


/************************************************************
 * Print <v>
 ************************************************************/
type Print struct {
	ICommand
}

// Print <v1> <v2> ...
//
// Print Stuff.  We will add to this over time.  Right now, if <v> = a transaction
// key, it prints that transaction.

func (Print) Execute(state IState, args []string) error {
	fmt.Println()
	for i, v := range args {
		if i == 0 {
			continue
		}

		ib := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(v))
		trans, ok := ib.(fct.ITransaction)
		if ib != nil && ok {
			fmt.Println(trans)
			v, err := GetRate(state)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fee, err := trans.CalculateFee(uint64(v))
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Required Fee:       ", strings.TrimSpace(fct.ConvertDecimal(fee)))
			tin, err1 := trans.TotalInputs()
			tout, err2 := trans.TotalOutputs()
			if err1 == nil && err2 == nil {
				cfee := int64(tin) - int64(tout)
				sign := ""
				if cfee < 0 {
					sign = "-"
					cfee = -cfee
				}
				fmt.Print("Fee You are paying: ",
						  sign, strings.TrimSpace(fct.ConvertDecimal(uint64(cfee))), "\n")
			} else {
				if err1 != nil {
					fmt.Println("Inputs have an error: ", err1)
				}
				if err2 != nil {
					fmt.Println("Outputs have an error: ", err2)
				}
			}
			binary, err := trans.MarshalBinary()
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Transaction Size:   ", len(binary))
			continue
		}

		switch strings.ToLower(v) {
		case "currentblock":
			fmt.Println(state.GetFS().GetCurrentBlock())
		case "rate":
			v, err := GetRate(state)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println("Factoids to buy one Entry Credit: ",
				fct.ConvertDecimal(uint64(v)))
		case "height":
			fmt.Println("Directory block height is: ", state.GetFS().GetDBHeight())
		default:
			fmt.Println("Unknown: ", v)
		}
	}

	return nil
}

func (Print) Name() string {
	return "Print"
}

func (Print) ShortHelp() string {
	return "Print <v1> <v2> ...  Prints the specified transaction(s) or the exchange rate."
}

func (Print) LongHelp() string {
	return `
Print <v1> <v2> ...                 Prints the specified values.  If <v> is a key for 
                                    a transaction, it will print said transaction.
      Print rate                    Print the number of factoids required to buy one
                                    one entry credit
`
}
