// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.
package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	fct "github.com/FactomProject/factoid"
	"net/http"
)

/************************************************************
 * Submit
 ************************************************************/
type Submit struct {
	ICommand
}

// Submit <k>
//
// Submit the given transaction identified by the given key
func (Submit) Execute(state IState, args []string) error {

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

	err = state.GetFS().GetWallet().ValidateSignatures(trans)
	if err != nil {
		return err
	}

	// Okay, transaction is good, so marshal and send to factomd!
	data, err := trans.MarshalBinary()
	if err != nil {
		return err
	}

	transdata := string(hex.EncodeToString(data))

	s := struct{ Transaction string }{transdata}

	j, err := json.Marshal(s)
	if err != nil {
		return err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/v1/factoid-submit/", state.GetServer()),
		"application/json",
		bytes.NewBuffer(j))

	if err != nil {
		return fmt.Errorf("Error coming back from server ")
	}
	resp.Body.Close()

	// Clear out the transaction
	state.GetFS().GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), nil)

	fmt.Println("Transaction", key, "Submitted")

	return nil
}

func (Submit) Name() string {
	return "Submit"
}

func (Submit) ShortHelp() string {
	return "Submit <k> -- Submit the transaction given by the key <k>"
}

func (Submit) LongHelp() string {
	return `
Submit <key>                        Submits the transaction specified by the given key.
                                    Each input in the transaction must have  a valid
                                    signature, or Submit will reject the transaction.
`
}


