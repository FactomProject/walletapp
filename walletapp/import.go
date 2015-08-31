// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"encoding/hex"
	"fmt"
	"os"
	"io/ioutil"
	fct "github.com/FactomProject/factoid"
)

/************************************************************
 * Import
 ************************************************************/

type Import struct {
	
}

var _ ICommand = (*Import)(nil)

// New Transaction:  key --
// We create a new transaction, and track it with the user supplied key.  The
// user can then use this key to make subsequent calls to add inputs, outputs,
// and to sign. Then they can submit the transaction.
//
// When the transaction is submitted, we clear it from our working memory.
// Multiple transactions can be under construction at one time, but they need
// their own keys. Once a transaction is either submitted or deleted, the key
// can be reused.
func (Import) Execute(state IState, args []string) error {

	if len(args) != 3 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	filename := args[2]

	if _, err := os.Stat(filename); err != nil {
		return fmt.Errorf("Could not find the input file %s",filename)
	}
	
	{	// Doing a bit of variable scope management here, since I want t later.
		t := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
		if t != nil {
			return fmt.Errorf("That transaction already exists.  Specify a new one, or delete this one.")
		}
	}
	
	data, err := ioutil.ReadFile(filename)
	var hexdata []byte
	for _,b := range data {
		if b > 32 {
			hexdata = append(hexdata,b)
		}
	}
	
	bdata, err := hex.DecodeString(string(hexdata))
	if err != nil {
		return err 
	}
	
	t := new(fct.Transaction)
	err = t.UnmarshalBinary(bdata)
	if err != nil {
		return err 
	}
	
	state.GetFS().GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), t)
	
	fmt.Println("Transaction",filename,"has been imported")	
	return nil
}

func (Import) Name() string {
	return "import"
}

func (Import) ShortHelp() string {
	return "Import <key> <filename> -- Import the given transactiion from the given filename."
}

func (Import) LongHelp() string {
	return `
Import <key> <filename>             Import the given transaction to the given filename.
`
}



