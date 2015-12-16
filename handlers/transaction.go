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
	"time"
	"strings"

	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
	"github.com/hoisie/web"

	"github.com/FactomProject/fctwallet/Wallet"
	"github.com/FactomProject/fctwallet/Wallet/Utility"
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

// True is success! False is failure.  The Response is what the CLI
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
        ctx.ContentType("json")
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
		we := Wallet.GetRaw([]byte(fct.W_NAME), []byte(name))
		if we != nil {
			address, err = we.(wallet.IWalletEntry).GetAddress()
			if we.(wallet.IWalletEntry).GetType() == "ec" {
				if !ec {
					reportResults(ctx,"Was Expecting a Factoid Address",false)
					ok = false
					return
				}
			}else{
				if ec {
					reportResults(ctx,"Was Expecting an Entry Credit Address",false)
					ok = false
					return
				}	
			}
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

// Returns either an unbounded list of transactions, or the list of 
// transactions that involve a given address.
//
func HandleGetProcessedTransactions(ctx*web.Context, parms string) {
	cmd := ctx.Params["cmd"]
	adr := ctx.Params["address"]
	
	if cmd == "all" {
		list, err := Utility.DumpTransactions(nil)
		if err != nil {
			reportResults(ctx,err.Error(),false)
			return
		}
		reportResults(ctx,string(list),true)
	}else{
		
		adr, err := Wallet.LookupAddress("FA",adr)
		if err != nil {
			adr, err = Wallet.LookupAddress("EC",adr)
			if err != nil {
				reportResults(ctx,fmt.Sprintf("Could not understand address %s",adr),false)
				return
			}
		}
		badr,err := hex.DecodeString(adr)
		
		var adrs [][]byte
		adrs = append(adrs,badr)
		
		list, err := Utility.DumpTransactions(adrs)
		if err != nil {
			reportResults(ctx,err.Error(),false)
			return
		}
		reportResults(ctx,string(list),true)
	}
}

// Returns either an unbounded list of transactions, or the list of 
// transactions that involve a given address.
//
// Return in JSON
//
func HandleGetProcessedTransactionsj(ctx*web.Context, parms string) {
	cmd := ctx.Params["cmd"]
	adr := ctx.Params["address"]
	
	if cmd == "all" {
		list, err := Utility.DumpTransactionsJSON(nil)
		if err != nil {
			reportResults(ctx,err.Error(),false)
			return
		}
		reportResults(ctx,string(list),true)
	}else{
		
		adr, err := Wallet.LookupAddress("FA",adr)
		if err != nil {
			adr, err = Wallet.LookupAddress("EC",adr)
			if err != nil {
				reportResults(ctx,fmt.Sprintf("Could not understand address %s",adr),false)
				return
			}
		}
		badr,err := hex.DecodeString(adr)
		
		var adrs [][]byte
		adrs = append(adrs,badr)
		
		list, err := Utility.DumpTransactionsJSON(adrs)
		if err != nil {
			reportResults(ctx,err.Error(),false)
			return
		}
		reportResults(ctx,string(list),true)
	}
}



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
	keys, _ := Wallet.GetWalletNames()
	if len(keys) == 0 {
		setFountian = true
		for i := 1; i <= 10; i++ {
			name := fmt.Sprintf("%02d-Fountain", i)
			_, err := Wallet.GenerateFctAddress([]byte(name), 1, 1)
			if err != nil {
				reportResults(ctx, err.Error(), false)
				return
			}
		}
	}

	seedprime := fct.Sha([]byte(fmt.Sprintf("%s%v", seed, time.Now().UnixNano()))).Bytes()
	Wallet.NewSeed(seedprime)

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
		return
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
		return
	}
	reportResults(ctx, "Success deleting transaction", true)
}

func HandleProperties (ctx *web.Context){
	prop,err := Wallet.GetProperties()
	if err != nil {
		reportResults(ctx, "Failed to retrieve properties",false)
		return
	}
	
	top := prop.Protocol_Version/1000000
	mid := (prop.Protocol_Version%1000000)/1000
	low := prop.Protocol_Version%1000
	
	ret :=      fmt.Sprintf("Protocol Version:   %d.%d.%d\n",top,mid,low)
	
	top  = prop.Factomd_Version/1000000
	mid  = (prop.Factomd_Version%1000000)/1000
	low  = prop.Factomd_Version%1000
	
	ret = ret+ fmt.Sprintf("factomd Version:    %d.%d.%d\n",top,mid,low)

	top  = prop.Fctwallet_Version/1000000
	mid  = (prop.Fctwallet_Version%1000000)/1000
	low  = prop.Fctwallet_Version%1000
	
	ret = ret+ fmt.Sprintf("fctwallet Version:  %d.%d.%d\n",top,mid,low)

	reportResults(ctx, ret, true)
	
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
			return
		}
		outs, err := trans.TotalOutputs()
		if err != nil {
			reportResults(ctx, err.Error(), false)
			return
		}
		ecs, err := trans.TotalECs()
		if err != nil {
			reportResults(ctx, err.Error(), false)
			return
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
		return
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
	_, err := Wallet.FactoidSubmit(jsonkey)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, "Success Submitting transaction", true)
}

func GetFee(ctx *web.Context) (int64, error) {
	return Wallet.GetFee()
}

func HandleGetFee(ctx *web.Context, k string) {
	
	var trans fct.ITransaction
	var err error
	
	key := ctx.Params["key"]
	
	fmt.Println("getfee", key)
	 
	if len(key) > 0 {
		trans, err = getTransaction(ctx, key)
		if err != nil {
			reportResults(ctx, "Failure to locate the transaction", false)
			return
		}
	}
	
	fee, err := Wallet.GetFee()
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}
	
	if trans != nil {
		ufee, _ := trans.CalculateFee(uint64(fee))
		fee = int64(ufee)
	}
	
	reportResults(ctx, fmt.Sprintf("%s",strings.TrimSpace(fct.ConvertDecimal(uint64(fee)))), true)
}

func GetAddresses() []byte {
	values := Wallet.GetAddresses()

	ecKeys := make([]string, 0, len(values))
	fctKeys := make([]string, 0, len(values))
	ecBalances := make([]string, 0, len(values))
	fctBalances := make([]string, 0, len(values))
	fctAddresses := make([]string, 0, len(values))
	ecAddresses := make([]string, 0, len(values))

	var maxlen int
	for _, we := range values {
		if len(we.GetName()) > maxlen {
			maxlen = len(we.GetName())
		}
		var adr string
		if we.GetType() == "ec" {
			address, err := we.GetAddress()
			if err != nil {
				continue
			}
			adr = fct.ConvertECAddressToUserStr(address)
			ecAddresses = append(ecAddresses, adr)
			ecKeys = append(ecKeys, string(we.GetName()))
			bal, _ := ECBalance(adr)
			ecBalances = append(ecBalances, strconv.FormatInt(bal, 10))
		} else {
			address, err := we.GetAddress()
			if err != nil {
				continue
			}
			adr = fct.ConvertFctAddressToUserStr(address)
			fctAddresses = append(fctAddresses, adr)
			fctKeys = append(fctKeys, string(we.GetName()))
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

	keys, transactions, err := Wallet.GetTransactions()
	if err != nil {
		return nil, err
	}

	var out bytes.Buffer
	for i, trans := range transactions {
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
			keys[i],
			strings.TrimSpace(fct.ConvertDecimal(fee)),
			cprt,
			transactions[i].String()))
	}

	output := out.Bytes()
	// now look for the addresses, and replace them with our names. (the transactions
	// in flight also have a Factom address... We leave those alone.

	names, vs := Wallet.GetWalletNames()

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
    ctx.ContentType("json")
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
    ctx.ContentType("json")
	ctx.Write(j)
}

func HandleFactoidValidate(ctx *web.Context) {
}

func HandleFactoidNewSeed(ctx *web.Context) {
}
