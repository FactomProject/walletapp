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
	"strings"

	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
)

/*************************************************************************
 * Handler Functions
 *************************************************************************/

// Setup:  seed --
// Setup creates the 10 fountain Factoid Addresses, then sets address
// generation to be unique for this wallet.  You CAN call setup multiple
// times, but once the Fountain addresses are created, Setup only changes
// the seed.
//
// Setup must be called once before you do anything else with the wallet.
//
/*
func HandleFactoidSetup(ctx *web.Context, seed string) {
	// Make sure we have a seed.
	if len(seed) == 0 {
		msg := "You must supply some random seed. For example (don't use this!)\n" +
			"factom-cli setup 'woe!#in31!%234ng)%^&$%oeg%^&*^jp45694a;gmr@#t4 q34y'\n" +
			"would make a nice seed.  The more random the better.\n\n" +
			"Note that if you create an address before you call Setup, you must\n" +
			"use those address(s) as you access the fountians."

		reportResults(ctx, msg, false)
	}
	setFountian := false
	keys, _ := factoidState.GetDB().GetKeysValues([]byte(fct.W_NAME))
	if len(keys) == 0 {
		setFountian = true
		for i := 1; i <= 10; i++ {
			name := fmt.Sprintf("%02d-Fountain", i)
			_, err := factoidState.GetWallet().GenerateFctAddress([]byte(name), 1, 1)
			if err != nil {
				reportResults(ctx, err.Error(), false)
				return
			}
		}
	}

	seedprime := fct.Sha([]byte(fmt.Sprintf("%s%v", seed, time.Now().UnixNano()))).Bytes()
	factoidState.GetWallet().NewSeed(seedprime)

	if setFountian {
		reportResults(ctx, "New seed set, fountain addresses defined", true)
	} else {
		reportResults(ctx, "New seed set, no fountain addresses defined", true)
	}
}*/

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

	err := ValidateKey(key)
	if err != nil {
		return err
	}

	// Make sure we don't already have a transaction in process with this key
	t := factoidState.GetDB().GetRaw([]byte(fct.DB_BUILD_TRANS), []byte(key))
	if t != nil {
		return fmt.Errorf("Duplicate key: '", key, "'")
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

		if ins != outs + ecs {
            return 0, fmt.Errorf("Inputs and outputs don't add up")
		}
	}

	err := ValidateKey(key)
	if err != nil {
        return 0, err
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
/*
func HandleFactoidAddInput(ctx *web.Context, parms string) {
	trans, key, _, address, amount, ok := getParams_(ctx, parms, false)

	if !ok {
		return
	}
	err := ValidateKey(key)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	// First look if this is really an update
	for _, input := range trans.GetInputs() {
		if input.GetAddress().IsSameAs(address) {
			oldamt := input.GetAmount()
			input.SetAmount(uint64(amount))
			reportResults(ctx, fmt.Sprintf("Input was %s\n"+
				"Now is	%s",
				fct.ConvertDecimal(oldamt),
				fct.ConvertDecimal(uint64(amount))), true)
			return
		}
	}

	// Add our new input
	err = factoidState.GetWallet().AddInput(trans, address, uint64(amount))
	if err != nil {
		reportResults(ctx, "Failed to add input", false)
		return
	}

	// Update our map with our new transaction to the same key. Otherwise, all
	// of our work will go away!
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	reportResults(ctx, "Success adding Input", true)
}

func HandleFactoidAddOutput(ctx *web.Context, parms string) {
	trans, key, _, address, amount, ok := getParams_(ctx, parms, false)
	if !ok {
		return
	}

	err := ValidateKey(key)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	// First look if this is really an update
	for _, output := range trans.GetOutputs() {
		if output.GetAddress().IsSameAs(address) {
			oldamt := output.GetAmount()
			output.SetAmount(uint64(amount))
			reportResults(ctx, fmt.Sprintf("Input was %s\n"+
				"Now is	%s",
				fct.ConvertDecimal(oldamt),
				fct.ConvertDecimal(uint64(amount))), true)
			return
		}
	}
	// Add our new Output
	err = factoidState.GetWallet().AddOutput(trans, address, uint64(amount))
	if err != nil {
		reportResults(ctx, "Failed to add output", false)
		return
	}

	// Update our map with our new transaction to the same key.  Otherwise, all
	// of our work will go away!
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	reportResults(ctx, "Success adding output", true)
}

func HandleFactoidAddECOutput(ctx *web.Context, parms string) {
	trans, key, _, address, amount, ok := getParams_(ctx, parms, true)
	if !ok {
		return
	}

	err := ValidateKey(key)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}
	// First look if this is really an update
	for _, ecoutput := range trans.GetECOutputs() {
		if ecoutput.GetAddress().IsSameAs(address) {
			oldamt := ecoutput.GetAmount()
			ecoutput.SetAmount(uint64(amount))
			reportResults(ctx, fmt.Sprintf("Input was %s\n"+
				"Now is	%s",
				fct.ConvertDecimal(oldamt),
				fct.ConvertDecimal(uint64(amount))), true)
			return
		}
	}
	// Add our new Entry Credit Output
	err = factoidState.GetWallet().AddECOutput(trans, address, uint64(amount))
	if err != nil {
		reportResults(ctx, "Failed to add input", false)
		return
	}

	// Update our map with our new transaction to the same key.  Otherwise, all
	// of our work will go away!
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	reportResults(ctx, "Success adding Entry Credit Output", true)
}

func HandleFactoidSignTransaction(ctx *web.Context, key string) {

	err := ValidateKey(key)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	// Get the transaction
	trans, err := getTransaction(ctx, key)
	if err != nil {
		reportResults(ctx, "Failed to get the transaction", false)
		return
	}

	err = factoidState.GetWallet().Validate(1, trans)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	valid, err := factoidState.GetWallet().SignInputs(trans)
	if err != nil {
		reportResults(ctx, err.Error(), false)
	}
	if !valid {
		reportResults(ctx, "Not all inputs are signed", false)
	}
	// Update our map with our new transaction to the same key.  Otherwise, all
	// of our work will go away!
	factoidState.GetDB().PutRaw([]byte(fct.DB_BUILD_TRANS), []byte(key), trans)

	reportResults(ctx, "Success signing transaction", true)
}

func HandleFactoidSubmit(ctx *web.Context, jsonkey string) {
	type submitReq struct {
		Transaction string
	}

	in := new(submitReq)
	json.Unmarshal([]byte(jsonkey), in)

	key := in.Transaction
	// Get the transaction
	trans, err := getTransaction(ctx, key)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	err = factoidState.GetWallet().ValidateSignatures(trans)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	// Okay, transaction is good, so marshal and send to factomd!
	data, err := trans.MarshalBinary()
	if err != nil {
		reportResults(ctx, "Failed to marshal the transaction for factomd", false)
		return
	}

	transdata := string(hex.EncodeToString(data))

	s := struct{ Transaction string }{transdata}

	j, err := json.Marshal(s)
	if err != nil {
		reportResults(ctx, "Failed to marshal the transaction for factomd", false)
		return
	}

	resp, err := http.Post(
		fmt.Sprintf("http://%s/v1/factoid-submit/", ipaddressFD+portNumberFD),
		"application/json",
		bytes.NewBuffer(j))

	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		reportResults(ctx, err.Error(), false)
	}

	resp.Body.Close()

	r := new(Response)
	if err := json.Unmarshal(body, r); err != nil {
		reportResults(ctx, err.Error(), false)
	}

	if r.Success {
		factoidState.GetDB().DeleteKey([]byte(fct.DB_BUILD_TRANS), []byte(key))
		reportResults(ctx, r.Response, true)
	} else {
		reportResults(ctx, r.Response, false)
	}

}
*/
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

func GetAddresses() ([]string, []wallet.IWalletEntry) {
	keys, values := factoidState.GetDB().GetKeysValues([]byte(fct.W_NAME))
    answerWE:=[]wallet.IWalletEntry{}
    answerKeys:=[]string{}
	for i, k := range keys {
		we, ok := values[i].(wallet.IWalletEntry)
		if !ok {
			panic("Get Addresses finds the database corrupt. Shouldn't happen")
		}
        answerWE=append(answerWE, we)
        answerKeys = append(answerKeys, string(k))
	}
    return answerKeys, answerWE
}

func GetTransactions() ([]byte, error) {
	exch, err := GetFee()
	if err != nil {
		return nil, err
	}

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

	var out bytes.Buffer
	for i, key := range keys {
		if values[i] == nil {
			continue
		}
		trans := values[i].(fct.ITransaction)

		fee, _ := trans.CalculateFee(uint64(exch))
		cprt := ""
		cin, err := trans.TotalInputs()
		if err != nil {
			cprt = cprt + err.Error()
		}
		cout, err := trans.TotalOutputs()
		if err != nil {
			cprt = cprt + err.Error()
		}
		cecout, err := trans.TotalECs()
		if err != nil {
			cprt = cprt + err.Error()
		}

		if len(cprt) == 0 {
			v := int64(cin) - int64(cout) - int64(cecout)
			sign := ""
			if v < 0 {
				sign = "-"
				v = -v
			}
			cprt = fmt.Sprintf(" Currently will pay: %s%s",
				sign,
				strings.TrimSpace(fct.ConvertDecimal(uint64(v))))
			if sign == "-" || fee > uint64(v) {
				cprt = cprt + "\n\nWARNING: Currently your transaction fee may be too low"
			}
		}

		out.WriteString(fmt.Sprintf("\n%25s:  Fee Due: %s  %s\n\n%s\n",
			key,
			strings.TrimSpace(fct.ConvertDecimal(fee)),
			cprt,
			values[i].String()))
	}

	output := out.Bytes()
	// now look for the addresses, and replace them with our names. (the transactions
	// in flight also have a Factom address... We leave those alone.

	names, vs := factoidState.GetDB().GetKeysValues([]byte(fct.W_NAME))

	for i, name := range names {
		we, ok := vs[i].(wallet.IWalletEntry)
		if !ok {
			return nil, fmt.Errorf("Database is corrupt")
		}

		address, err := we.GetAddress()
		if err != nil {
			continue
		} // We shouldn't get any of these, but ignore them if we do.
		adrstr := []byte(hex.EncodeToString(address.Bytes()))

		output = bytes.Replace(output, adrstr, name, -1)
	}

	return output, nil
}
