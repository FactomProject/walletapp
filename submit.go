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
	"io/ioutil"
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

	err = isReasonableFee(state, trans)
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

func isReasonableFee(state IState, trans fct.ITransaction) error {
	feeRate, getErr := GetFee(state)
	if getErr != nil {
		return getErr
	}

	reqFee, err := trans.CalculateFee(uint64(feeRate))
	if err != nil {
		return err
	}

	sreqFee := int64(reqFee)

	tin, err := trans.TotalInputs()
	if err != nil {
		return err
	}

	tout, err := trans.TotalOutputs()
	if err != nil {
		return err
	}

	tec, err := trans.TotalECs()
	if err != nil {
		return err
	}

	cfee := int64(tin) - int64(tout) - int64(tec)

	if cfee >= (sreqFee * 10) {
		return fmt.Errorf("Unbalanced transaction (fee too high). Fee should be less than 10x the required fee.")
	}

	if cfee < sreqFee {
		return fmt.Errorf("Insufficient fee")
	}

	return nil
}

func GetFee(state IState) (int64, error) {
	str := fmt.Sprintf("http://%s/v1/factoid-get-fee/", state.GetServer())
	resp, err := http.Get(str)
	if err != nil {
		return 0, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body.Close()
		return 0, err
	}
	resp.Body.Close()

	type x struct{ Fee int64 }
	b := new(x)
	if err := json.Unmarshal(body, b); err != nil {
		return 0, err
	}

	return b.Fee, nil
}
