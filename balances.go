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
	"github.com/FactomProject/factoid/wallet"
	"github.com/FactomProject/fctwallet/Wallet/Utility"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

/***********************************************
 * General Support Functions
 ***********************************************/

var badChar, _ = regexp.Compile("[^A-Za-z0-9_-]")
var badHexChar, _ = regexp.Compile("[^A-Fa-f0-9]")

func ValidName(name string) error {
	if len(name) > 32 {
		return fmt.Errorf("Name of address is too long.")
	}
	if badChar.FindStringIndex(name) != nil {
		return fmt.Errorf("Invalid name. Names must be alphanumeric, underscores, or hyphens.")
	}
	return nil
}

func GenAddress(state IState, adrType string, key string) error {
	validErr := ValidName(key)
	if validErr != nil {
		return validErr
	}
	switch strings.ToLower(adrType) {
	case "ec":
		adr, err := state.GetFS().GetWallet().GenerateECAddress([]byte(key))
		if err != nil {
			return err
		}
		fmt.Println(key, "=", fct.ConvertECAddressToUserStr(adr))
	case "fct":
		adr, err := state.GetFS().GetWallet().GenerateFctAddress([]byte(key), 1, 1)
		if err != nil {
			return err
		}
		fmt.Println(key, "=", fct.ConvertFctAddressToUserStr(adr))
	default:
		return fmt.Errorf("Invalid Parameters")
	}
	return nil
}

// Get the Factoshis per Entry Credit Rate
func GetRate(state IState) (int64, error) {
	str := fmt.Sprintf("http://%s/v1/factoid-get-fee/", state.GetServer())
	resp, err := http.Get(str)
	if err != nil {
		return 0, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()

	type x struct{ Fee int64 }
	b := new(x)
	if err = json.Unmarshal(body, b); err != nil {
		return 0, err
	}

	return b.Fee, nil

}

func LookupAddress(state IState, adrType string, adr string) (string, error) {
	if Utility.IsValidAddress(adr) && strings.HasPrefix(adr, adrType) {
		baddr := fct.ConvertUserStrToAddress(adr)
		adr = hex.EncodeToString(baddr)
	} else if Utility.IsValidHexAddress(adr) {
		// the address is good enough.
	} else if Utility.IsValidNickname(adr) {
		we := state.GetFS().GetDB().GetRaw([]byte(fct.W_NAME), []byte(adr))

		if we != nil {
			we2 := we.(wallet.IWalletEntry)
			addr, _ := we2.GetAddress()
			adr = hex.EncodeToString(addr.Bytes())
		} else {
			return "", fmt.Errorf("Name %s is undefined.", adr)
		}
	} else {
		return "", fmt.Errorf("Invalid Name.  Check that you have entered the name correctly.")
	}

	return adr, nil
}

func FctBalance(state IState, adr string) (int64, error) {

	adr, err := LookupAddress(state, "FA", adr)
	if err != nil {
		return 0, err
	}

	str := fmt.Sprintf("http://%s/v1/factoid-balance/%s", state.GetServer(), adr)
	resp, err := http.Get(str)
	if err != nil {
		return 0, fmt.Errorf("Communication Error with Factom Client")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("Read Error with Factom Client")
	}
	resp.Body.Close()

	type Balance struct {
		Response string
		Success  bool
	}
	b := new(Balance)

	if err := json.Unmarshal(body, b); err != nil {
		return 0, fmt.Errorf("Parsing Error on data returned by Factom Client")
	}
	if !b.Success {
		return 0, fmt.Errorf(err.Error())
	}
	v, err := strconv.ParseInt(b.Response, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid balance returned by factomd")
	}
	return v, nil

}

func ECBalance(state IState, adr string) (int64, error) {

	adr, err := LookupAddress(state, "EC", adr)
	if err != nil {
		return 0, err
	}

	str := fmt.Sprintf("http://%s/v1/entry-credit-balance/%s", state.GetServer(), adr)
	resp, err := http.Get(str)
	if err != nil {
		return 0, fmt.Errorf("Communication Error with Factom Client")
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("Read Error with Factom Client")
	}
	resp.Body.Close()

	type Balance struct {
		Response string
		Success  bool
	}
	b := new(Balance)

	if err := json.Unmarshal(body, b); err != nil {
		return 0, fmt.Errorf("Parsing Error on data returned by Factom Client")
	}
	if !b.Success {
		return 0, fmt.Errorf(err.Error())
	}
	v, err := strconv.ParseInt(b.Response, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("Invalid balance returned by factomd")
	}
	return v, nil
}

func GetBalances(state IState) []byte {
	keys, values := state.GetFS().GetDB().GetKeysValues([]byte(fct.W_NAME))

	ecKeys := make([]string, 0, len(keys))
	fctKeys := make([]string, 0, len(keys))
	ecBalances := make([]string, 0, len(keys))
	fctBalances := make([]string, 0, len(keys))
	fctAddresses := make([]string, 0, len(keys))
	ecAddresses := make([]string, 0, len(keys))

	var maxlen int
	var connect = true
	for i, k := range keys {
		if len(k) > maxlen {
			maxlen = len(k)
		}
		we, ok := values[i].(wallet.IWalletEntry)
		if !ok {
			panic("Get Addresses finds the database corrupt.  Shouldn't happen")
		}
		var adr string
		if we.GetType() == "ec" {
			address, err := we.GetAddress()
			if err != nil {
				continue
			}
			adr = fct.ConvertECAddressToUserStr(address)
			ecAddresses = append(ecAddresses, adr)
			ecKeys = append(ecKeys, string(k))
			bal, err := ECBalance(state, adr)
			if err != nil {
				connect = false
			}
			if connect {
				ecBalances = append(ecBalances, strconv.FormatInt(bal, 10))
			} else {
				ecBalances = append(ecBalances, "-")
			}
		} else {
			address, err := we.GetAddress()
			if err != nil {
				continue
			}
			adr = fct.ConvertFctAddressToUserStr(address)
			fctAddresses = append(fctAddresses, adr)
			fctKeys = append(fctKeys, string(k))
			bal, err := FctBalance(state, adr)
			if err != nil {
				connect = false
			}
			sbal := fct.ConvertDecimal(uint64(bal))
			if connect {
				fctBalances = append(fctBalances, sbal)
			} else {
				fctBalances = append(fctBalances, "-")
			}
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
	if !connect {
		out.WriteString("Balances are unavailable;  Wallet is offline\n")
	}
	return out.Bytes()
}

/*************************************************************
 * Balance
 *************************************************************/

type Balance struct {
	ICommand
}

func (Balance) Execute(state IState, args []string) (err error) {
	if len(args) != 3 {
		return fmt.Errorf("Wrong number of parameters")
	}

	var bal int64
	switch strings.ToLower(args[1]) {
	case "ec":
		bal, err = ECBalance(state, args[2])
	case "fct":
		bal, err = FctBalance(state, args[2])
	default:
		return fmt.Errorf("Invalid parameters")
	}
	if err != nil {
		return err
	}
	if args[1] == "fct" {
		fmt.Println(args[2], "=", fct.ConvertDecimal(uint64(bal)))
	} else {
		fmt.Println(args[2], "=", bal)
	}
	return nil
}

func (Balance) Name() string {
	return "Balance"
}

func (Balance) ShortHelp() string {
	return "Balance <ec|fct> <name|address> -- Returns the Entry Credits or Factoids at the\n" +
		"                                   specified name or address"
}

func (Balance) LongHelp() string {
	return `
Balance <ec|fct> <name|address>     ec      -- an Entry Credit address balance
                                    fct     -- a Factoid address balance
                                    name    -- Look up address by its name
                                    address -- specify the address directly
`
}

/*************************************************************
 * New Address
 *************************************************************/

type NewAddress struct {
	ICommand
}

func (NewAddress) Execute(state IState, args []string) (err error) {

	if len(args) != 3 {
		return fmt.Errorf("Incorrect Number of Arguments")
	}
	if err := ValidName(args[2]); err != nil {
		return err
	}

	return GenAddress(state, args[1], args[2])
}

func (NewAddress) Name() string {
	return "NewAddress"
}

func (NewAddress) ShortHelp() string {
	return "NewAddress <ec|fct> <name> -- Returns a new Entry Credit or Factoid address"
}

func (NewAddress) LongHelp() string {
	return `
NewAddress <ec|fct> <name>          <ec>   Generates a new Entry Credit Address and
                                           saves it with the given <name>
                                    <fct>  Generates a new Factom Address and saves it
                                           with the given <name>
                                    <name> Names must be made up of alphanumeric 
                                           characters or underscores.  They cannot be
                                           more than 32 characters in length.  
`
}

/*************************************************************
 * Balances
 *************************************************************/

type Balances struct {
	ICommand
}

func (Balances) Execute(state IState, args []string) (err error) {

	if len(args) != 1 {
		return fmt.Errorf("Balances takes no arguments")
	}

	fmt.Println(string(GetBalances(state)))
	return nil
}

func (Balances) Name() string {
	return "Balances"
}

func (Balances) ShortHelp() string {
	return "Balances -- Returns the balances of the Factoids and Entry Credits" +
		"            in this wallet, or tracked by this wallet."
}

func (Balances) LongHelp() string {
	return `
Balances                            Returns the Factoid and Entry Credit names
                                    and balances for the address in or tracked by
                                    this wallet.
`
}
