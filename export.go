// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"bufio"
	"encoding/hex"
	"fmt"
	fct "github.com/FactomProject/factoid"
	"io/ioutil"
	"os"
)

/************************************************************
 * Export
 ************************************************************/

type Export struct {
}

var _ ICommand = (*Export)(nil)

// New Transaction:  key --
// We create a new transaction, and track it with the user supplied key.  The
// user can then use this key to make subsequent calls to add inputs, outputs,
// and to sign. Then they can submit the transaction.
//
// When the transaction is submitted, we clear it from our working memory.
// Multiple transactions can be under construction at one time, but they need
// their own keys. Once a transaction is either submitted or deleted, the key
// can be reused.
func (Export) Execute(state IState, args []string) error {

	if len(args) != 3 {
		return fmt.Errorf("Invalid Parameters")
	}
	key := args[1]
	filename := args[2]

	if _, err := os.Stat(filename); err == nil {
		fmt.Printf("File Exists.  Overwrite? (Y/N): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if input != "y\n" && input != "Y\n" {
			fmt.Println("answer: ", hex.EncodeToString([]byte(input)))
			return fmt.Errorf("Transaction not exported")
		}
	}

	t := state.GetFS().GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	if t == nil {
		return fmt.Errorf("Could not find the transaction")
	}

	data, err := t.MarshalBinary()
	if err != nil {
		return err
	}

	bytelen := 40
	var outdata []byte
	for len(data) > bytelen {
		outdata = append(outdata, []byte(hex.EncodeToString(data[:bytelen]))...)
		outdata = append(outdata, 10)
		data = data[bytelen:]
	}
	if len(data) > 0 {
		outdata = append(outdata, []byte(hex.EncodeToString(data))...)
		outdata = append(outdata, 10)
	}

	ioutil.WriteFile(filename, outdata, 0644)

	return nil
}

func (Export) Name() string {
	return "export"
}

func (Export) ShortHelp() string {
	return "Export <key> <filename> -- Export the given transactiion to the given filename."
}

func (Export) LongHelp() string {
	return `
Export <key> <filename>             Export the given transaction to the given filename.
`
}
