// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package Wallet

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
	"github.com/FactomProject/fctwallet/Wallet/Utility"
)

// New Transaction:  key --
// We create a new transaction, and track it with the user supplied key.  The
// user can then use this key to make subsequent calls to add inputs, outputs,
// and to sign. Then they can submit the transaction.
//
// When the transaction is submitted, we clear it from our working memory.
// Multiple transactions can be under construction at one time, but they need
// their own keys. Once a transaction is either submitted or deleted, the key
// can be reused.
func FactoidNewTransaction(key string) error {
	// Make sure we have a key
	if len(key) == 0 {
		return fmt.Errorf("Missing transaction key")
	}

	ok := Utility.IsValidKey(key)
	if !ok {
		return  fmt.Errorf("Invalid name for transaction")
	}

	// Make sure we don't already have a transaction in process with this key
	t := factoidState.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	if t != nil {
		return fmt.Errorf("Duplicate key: '%s'", key)
	}
	// Create a transaction
	t = factoidState.GetWallet().CreateTransaction(factoidState.GetTimeMilli())
	// Save it with the key
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), t)

	return nil
}

// Delete Transaction:  key --
// Remove a transaction rather than sign and submit the transaction.  Sometimes
// you just need to throw one a way, and rebuild it.
//
func FactoidDeleteTransaction(key string) error {
	// Make sure we have a key
	if len(key) == 0 {
		return fmt.Errorf("Missing transaction key")
	}
	// Wipe out the key
	factoidState.GetDB().DeleteKey([]byte(fct.DB_BUILD_TRANS), []byte(key))
	return nil
}

func FactoidAddFee(trans fct.ITransaction, key string, address fct.IAddress, name string) (uint64, error) {
	{
		ins, err := trans.TotalInputs()
		if err != nil {
			return 0, err
		}
		outs, err := trans.TotalOutputs()
		if err != nil {
			return 0, err
		}
		ecs, err := trans.TotalECs()
		if err != nil {
			return 0, err
		}

		if ins != outs+ecs {
			return 0, fmt.Errorf("Inputs and outputs don't add up")
		}
	}

	ok := Utility.IsValidKey(key)
	if !ok {
		return 0, fmt.Errorf("Invalid name for transaction")
	}
	

	fee, err := GetFee()
	if err != nil {
		return 0, err
	}

	transfee, err := trans.CalculateFee(uint64(fee))
	if err != nil {
		return 0, err
	}

	adr, err := factoidState.GetWallet().GetAddressHash(address)
	if err != nil {
		return 0, err
	}

	for _, input := range trans.GetInputs() {
		if input.GetAddress().IsSameAs(adr) {
			amt, err := fct.ValidateAmounts(input.GetAmount(), transfee)
			if err != nil {
				return 0, err
			}
			input.SetAmount(amt)
			return transfee, nil
		}
	}
	return 0, fmt.Errorf("%s is not an input to the transaction.", key)
}

func FactoidAddInput(trans fct.ITransaction, key string, address fct.IAddress, amount uint64) error {
	ok := Utility.IsValidKey(key)
	if !ok {
		return fmt.Errorf("Invalid name for transaction")
	}
	

	// First look if this is really an update
	for _, input := range trans.GetInputs() {
		if input.GetAddress().IsSameAs(address) {
			input.SetAmount(amount)
			return nil
		}
	}

	// Add our new input
	err := factoidState.GetWallet().AddInput(trans, address, amount)
	if err != nil {
		return fmt.Errorf("Failed to add input")
	}

	// Update our map with our new transaction to the same key. Otherwise, all
	// of our work will go away!
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	return nil
}

func FactoidAddOutput(trans fct.ITransaction, key string, address fct.IAddress, amount uint64) error {
	ok := Utility.IsValidKey(key)
	if !ok {
		return fmt.Errorf("Invalid name for transaction")
	}

	// First look if this is really an update
	for _, output := range trans.GetOutputs() {
		if output.GetAddress().IsSameAs(address) {
			output.SetAmount(amount)
			return nil
		}
	}
	// Add our new Output
	err := factoidState.GetWallet().AddOutput(trans, address, uint64(amount))
	if err != nil {
		return fmt.Errorf("Failed to add output")
	}

	// Update our map with our new transaction to the same key.  Otherwise, all
	// of our work will go away!
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	return nil
}

func FactoidAddECOutput(trans fct.ITransaction, key string, address fct.IAddress, amount uint64) error {
	ok := Utility.IsValidKey(key)
	if !ok {
		return fmt.Errorf("Invalid name for transaction")
	}
	// First look if this is really an update
	for _, ecoutput := range trans.GetECOutputs() {
		if ecoutput.GetAddress().IsSameAs(address) {
			ecoutput.SetAmount(amount)
			return nil
		}
	}
	// Add our new Entry Credit Output
	err := factoidState.GetWallet().AddECOutput(trans, address, uint64(amount))
	if err != nil {
		return fmt.Errorf("Failed to add Entry Credit Output")
	}

	// Update our map with our new transaction to the same key.  Otherwise, all
	// of our work will go away!
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	return nil
}

func FactoidSignTransaction(key string) error {
	ok := Utility.IsValidKey(key)
	if !ok {
		return fmt.Errorf("Invalid name for transaction")
	}

	// Get the transaction
	trans, err := GetTransaction(key)
	if err != nil {
		return fmt.Errorf("Failed to get the transaction")
	}

	err = factoidState.GetWallet().Validate(1, trans)
	if err != nil {
		return err
	}

	valid, err := factoidState.GetWallet().SignInputs(trans)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("Not all inputs are signed")
	}
	// Update our map with our new transaction to the same key.  Otherwise, all
	// of our work will go away!
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	return nil
}

func FactoidSubmit(jsonkey string) (string, error) {
	type submitReq struct {
		Transaction string
	}

	in := new(submitReq)
	json.Unmarshal([]byte(jsonkey), in)

	key := in.Transaction
	// Get the transaction
	trans, err := GetTransaction(key)
	if err != nil {
		return "", err
	}

	err = factoidState.GetWallet().ValidateSignatures(trans)
	if err != nil {
		return "", err
	}

	// Okay, transaction is good, so marshal and send to factomd!
	data, err := trans.MarshalBinary()
	if err != nil {
		return "", err
	}

	transdata := string(hex.EncodeToString(data))

	s := struct{ Transaction string }{transdata}

	j, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/v1/factoid-submit/", ipaddressFD+portNumberFD),
		"application/json",
		bytes.NewBuffer(j))

	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	resp.Body.Close()

	r := new(Response)
	if err := json.Unmarshal(body, r); err != nil {
		return "", err
	}

	if r.Success {
		factoidState.GetDB().DeleteKey([]byte(fct.DB_BUILD_TRANS), []byte(key))
		return "", nil
	} else {
		return "", fmt.Errorf(r.Response)
	}
	return r.Response, nil
}

func GetFee() (int64, error) {
	str := fmt.Sprintf("http://%s/v1/factoid-get-fee/", ipaddressFD+portNumberFD)
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

func GetAddresses() []wallet.IWalletEntry {
	_, values := factoidState.GetDB().GetKeysValues([]byte(fct.W_NAME))
	answerWE := []wallet.IWalletEntry{}
	for _, v := range values {
		we, ok := v.(wallet.IWalletEntry)
		if !ok {
			panic("Get Addresses finds the database corrupt. Shouldn't happen")
		}
		answerWE = append(answerWE, we)
	}
	return answerWE
}

func GetTransactions() ([][]byte, []fct.ITransaction, error) {
	// Get the transactions in flight.
	keys, values := factoidState.GetDB().GetKeysValues([]byte(fct.DB_BUILD_TRANS))

	for i := 0; i < len(keys)-1; i++ {
		for j := 0; j < len(keys)-i-1; j++ {
			if bytes.Compare(keys[j], keys[j+1]) > 0 {
				t := keys[j]
				keys[j] = keys[j+1]
				keys[j+1] = t
				t2 := values[j]
				values[j] = values[j+1]
				values[j+1] = t2
			}
		}
	}
	answer  := []fct.ITransaction{}
	theKeys := [][]byte{}
	

	
	for i, _ := range values {
		if values[i] == nil {
			continue
		}
		answer = append(answer, values[i].(fct.ITransaction))
		theKeys = append(theKeys,keys[i])
	}

	return theKeys, answer, nil
}

func GetWalletNames() (keys [][]byte, values []fct.IBlock) {
	return factoidState.GetDB().GetKeysValues([]byte(fct.W_NAME))
}

func GetRaw(bucket, key []byte) fct.IBlock {
	return factoidState.GetDB().GetRaw(bucket, key)
}

func GenerateFctAddress(name []byte, m int, n int) (hash fct.IAddress, err error) {
	return factoidState.GetWallet().GenerateFctAddress(name, m, n)
}

func NewSeed(data []byte) {
	factoidState.GetWallet().NewSeed(data)
}
