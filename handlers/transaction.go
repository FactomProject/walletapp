// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
	"github.com/hoisie/web"

	"github.com/FactomProject/fctwallet/Wallet"
)

/******************************************
 * Helper Functions
 ******************************************/

var badChar, _ = regexp.Compile("[^A-Za-z0-9_-]")
var badHexChar, _ = regexp.Compile("[^A-Fa-f0-9]")

type Response struct {
	Response string
	Success  bool
}

func ValidateKey(key string) (msg string, valid bool) {
	if len(key) > fct.ADDRESS_LENGTH {
		return "Key is too long.  Keys must be less than 32 characters", false
	}
	if badChar.FindStringIndex(key) != nil {
		str := fmt.Sprintf("The key or name '%s' contains invalid characters.\n"+
			"Keys and names are restricted to alphanumeric characters,\n"+
			"minuses (dashes), and underscores", key)
		return str, false
	}
	return "", true
}

// True is sccuess! False is failure.  The Response is what the CLI
// should report.
func reportResults(ctx *web.Context, response string, success bool) {
	b := Response{
		Response: response,
		Success:  success,
	}
	if p, err := json.Marshal(b); err != nil {

		ctx.WriteHeader(httpBad)
		return
	} else {
		ctx.Write(p)
	}
}

func getTransaction(ctx *web.Context, key string) (fct.ITransaction, error) {
	return Wallet.GetTransaction(key)
}

// &key=<key>&name=<name or address>&amount=<amount>
// If no amount is specified, a zero is returned.
func getParams_(ctx *web.Context, params string, ec bool) (
	trans fct.ITransaction,
	key string,
	name string,
	address fct.IAddress,
	amount int64,
	ok bool) {

	key = ctx.Params["key"]
	name = ctx.Params["name"]
	StrAmount := ctx.Params["amount"]

	if len(StrAmount) == 0 {
		StrAmount = "0"
	}

	if len(key) == 0 || len(name) == 0 {
		str := fmt.Sprintln("Missing Parameters: key='", key, "' name='", name, "' amount='", StrAmount, "'")
		reportResults(ctx, str, false)
		ok = false
		return
	}

	msg, valid := ValidateKey(key)
	if !valid {
		reportResults(ctx, msg, false)
		ok = false
		return
	}

	amount, err := strconv.ParseInt(StrAmount, 10, 64)
	if err != nil {
		str := fmt.Sprintln("Error parsing amount.\n", err)
		reportResults(ctx, str, false)
		ok = false
		return
	}

	// Get the transaction
	trans, err = getTransaction(ctx, key)
	if err != nil {
		reportResults(ctx, "Failure to locate the transaction", false)
		ok = false
		return
	}

	// Get the input/output/ec address.  Which could be a name.  First look and see if it is
	// a name.  If it isn't, then look and see if it is an address.  Someone could
	// do a weird Address as a name and fool the code, but that seems unlikely.
	// Could check for that some how, but there are many ways around such checks.

	if len(name) <= fct.ADDRESS_LENGTH {
		we := factoidState.GetDB().GetRaw([]byte(fct.W_NAME), []byte(name))
		if we != nil {
			address, err = we.(wallet.IWalletEntry).GetAddress()
			if err != nil || address == nil {
				reportResults(ctx, "Should not get an error geting a address from a Wallet Entry", false)
				ok = false
				return
			}
			ok = true
			return
		}
	}
	if (!ec && !fct.ValidateFUserStr(name)) || (ec && !fct.ValidateECUserStr(name)) {
		reportResults(ctx, fmt.Sprintf("The address specified isn't defined or is invalid: %s", name), false)
		ctx.WriteHeader(httpBad)
		ok = false
		return
	}
	baddr := fct.ConvertUserStrToAddress(name)

	address = fct.NewAddress(baddr)

	ok = true
	return
}

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
func HandleFactoidNewTransaction(ctx *web.Context, key string) {
	// Make sure we have a key
	if len(key) == 0 {
		reportResults(ctx, "Missing transaction key", false)
		return
	}

	msg, valid := ValidateKey(key)
	if !valid {
		reportResults(ctx, msg, false)
		return
	}

	err := Wallet.FactoidNewTransaction(key)
	if err != nil {
		reportResults(ctx, err.Error(), false)
	}

	reportResults(ctx, "Success building a transaction", true)
}

// Delete Transaction:  key --
// Remove a transaction rather than sign and submit the transaction.  Sometimes
// you just need to throw one a way, and rebuild it.
//
func HandleFactoidDeleteTransaction(ctx *web.Context, key string) {
	// Make sure we have a key
	if len(key) == 0 {
		reportResults(ctx, "Missing transaction key", false)
		return
	}
	err := Wallet.FactoidDeleteTransaction(key)
	if err != nil {
		reportResults(ctx, err.Error(), false)
	}
	reportResults(ctx, "Success deleting transaction", true)
}

func HandleFactoidAddFee(ctx *web.Context, parms string) {
	trans, key, _, address, _, ok := getParams_(ctx, parms, false)
	if !ok {
		return
	}

	name := ctx.Params["name"] // This is the name the user used.

	{
		ins, err := trans.TotalInputs()
		if err != nil {
			reportResults(ctx, err.Error(), false)
		}
		outs, err := trans.TotalOutputs()
		if err != nil {
			reportResults(ctx, err.Error(), false)
		}
		ecs, err := trans.TotalECs()
		if err != nil {
			reportResults(ctx, err.Error(), false)
		}

		if ins != outs+ecs {
			msg := fmt.Sprintf(
				"Addfee requires that all the inputs balance the outputs.\n"+
					"The total inputs of your transaction are              %s\n"+
					"The total outputs + ecoutputs of your transaction are %s",
				fct.ConvertDecimal(ins), fct.ConvertDecimal(outs+ecs))

			reportResults(ctx, msg, false)
			return
		}
	}

	transfee, err := Wallet.FactoidAddFee(trans, key, address, name)
	if err != nil {
		reportResults(ctx, err.Error(), false)
	}

	reportResults(ctx, fmt.Sprintf("Added %s to %s", fct.ConvertDecimal(uint64(transfee)), name), true)
	return
}

func HandleFactoidAddInput(ctx *web.Context, parms string) {
	trans, key, _, address, amount, ok := getParams_(ctx, parms, false)

	if !ok {
		return
	}

	err := Wallet.FactoidAddInput(trans, key, address, uint64(amount))
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, "Success adding Input", true)
}

func HandleFactoidAddOutput(ctx *web.Context, parms string) {
	trans, key, _, address, amount, ok := getParams_(ctx, parms, false)
	if !ok {
		return
	}

	err := Wallet.FactoidAddOutput(trans, key, address, uint64(amount))
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, "Success adding output", true)
}

func HandleFactoidAddECOutput(ctx *web.Context, parms string) {
	trans, key, _, address, amount, ok := getParams_(ctx, parms, true)
	if !ok {
		return
	}

	err := Wallet.FactoidAddECOutput(trans, key, address, uint64(amount))
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, "Success adding Entry Credit Output", true)
}

func HandleFactoidSignTransaction(ctx *web.Context, key string) {
	err := Wallet.FactoidSignTransaction(key)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, "Success signing transaction", true)
}

func HandleFactoidSubmit(ctx *web.Context, jsonkey string) {
	resp, err := Wallet.FactoidSubmit(jsonkey)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, resp, true)
}

func GetFee(ctx *web.Context) (int64, error) {
	return Wallet.GetFee()
}

func HandleGetFee(ctx *web.Context) {
	fee, err := Wallet.GetFee()
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	ctx.Write([]byte(fmt.Sprintf("{Fee: %d}", fee)))
}

func GetAddresses() []byte {
	keys, values := Wallet.GetAddresses()

	ecKeys := make([]string, 0, len(keys))
	fctKeys := make([]string, 0, len(keys))
	ecBalances := make([]string, 0, len(keys))
	fctBalances := make([]string, 0, len(keys))
	fctAddresses := make([]string, 0, len(keys))
	ecAddresses := make([]string, 0, len(keys))

	var maxlen int
	for i, k := range keys {
		if len(k) > maxlen {
			maxlen = len(k)
		}
		we := values[i]
		var adr string
		if we.GetType() == "ec" {
			address, err := we.GetAddress()
			if err != nil {
				continue
			}
			adr = fct.ConvertECAddressToUserStr(address)
			ecAddresses = append(ecAddresses, adr)
			ecKeys = append(ecKeys, k)
			bal, _ := ECBalance(adr)
			ecBalances = append(ecBalances, strconv.FormatInt(bal, 10))
		} else {
			address, err := we.GetAddress()
			if err != nil {
				continue
			}
			adr = fct.ConvertFctAddressToUserStr(address)
			fctAddresses = append(fctAddresses, adr)
			fctKeys = append(fctKeys, k)
			bal, _ := FctBalance(adr)
			sbal := fct.ConvertDecimal(uint64(bal))
			fctBalances = append(fctBalances, sbal)
		}
	}
	var out bytes.Buffer
	if len(fctKeys) > 0 {
		out.WriteString("\n  Factoid Addresses\n\n")
	}
	fstr := fmt.Sprintf("%s%vs    %s38s %s14s\n", "%", maxlen+4, "%", "%")
	for i, key := range fctKeys {
		str := fmt.Sprintf(fstr, key, fctAddresses[i], fctBalances[i])
		out.WriteString(str)
	}
	if len(ecKeys) > 0 {
		out.WriteString("\n  Entry Credit Addresses\n\n")
	}
	for i, key := range ecKeys {
		str := fmt.Sprintf(fstr, key, ecAddresses[i], ecBalances[i])
		out.WriteString(str)
	}

	return out.Bytes()
}

func GetTransactions(ctx *web.Context) ([]byte, error) {
	exch, err := GetFee(ctx)
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

func HandleGetAddresses(ctx *web.Context) {
	b := new(Response)
	b.Response = string(GetAddresses())
	b.Success = true
	j, err := json.Marshal(b)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}
	ctx.Write(j)
}

func HandleGetTransactions(ctx *web.Context) {
	b := new(Response)
	txt, err := GetTransactions(ctx)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}
	b.Response = string(txt)
	b.Success = true
	j, err := json.Marshal(b)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}
	ctx.Write(j)
}

func HandleFactoidValidate(ctx *web.Context) {
}

func HandleFactoidNewSeed(ctx *web.Context) {
}
