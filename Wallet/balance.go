// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package Wallet

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/factoid/wallet"
	"io/ioutil"
	"net/http"
	"strconv"
)


func FactoidBalance(adr string) (int64, error) {
	if !fct.ValidateFUserStr(adr) {
		err := ValidateKey(adr)
		if err != nil {
			return 0, err
		}

		we := factoidState.GetDB().GetRaw([]byte(fct.W_NAME), []byte(adr))

		if we != nil {
			we2 := we.(wallet.IWalletEntry)
			addr, _ := we2.GetAddress()
			adr = hex.EncodeToString(addr.Bytes())
		}

	} else {
		baddr := fct.ConvertUserStrToAddress(adr)
		adr = hex.EncodeToString(baddr)
	}

	str := fmt.Sprintf("http://%s/v1/factoid-balance/%s", ipaddressFD+portNumberFD, adr)
	resp, err := http.Get(str)
	if err != nil {
		return 0, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()

	b := new(Response)
	if err := json.Unmarshal(body, b); err != nil {
		return 0, err
	}

	if !b.Success {
		return 0, fmt.Errorf("%s", b.Response)
	}

	v, err := strconv.ParseInt(b.Response, 10, 64)
	if err != nil {
		return 0, err
	}

	return v, nil

}

func ECBalance(adr string) (int64, error) {

	if !fct.ValidateECUserStr(adr) {
		err := ValidateKey(adr)
		if err != nil {
			return 0, err
		}

		we := factoidState.GetDB().GetRaw([]byte(fct.W_NAME), []byte(adr))

		if we != nil {
			we2 := we.(wallet.IWalletEntry)
			addr, _ := we2.GetAddress()
			adr = hex.EncodeToString(addr.Bytes())
		}

	} else {
		baddr := fct.ConvertUserStrToAddress(adr)
		adr = hex.EncodeToString(baddr)
	}

	str := fmt.Sprintf("http://%s/v1/entry-credit-balance/%s", ipaddressFD+portNumberFD, adr)
	resp, err := http.Get(str)
	if err != nil {
		return 0, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()

	b := new(Response)
	if err := json.Unmarshal(body, b); err != nil {
		return 0, err
	}

	if !b.Success {
		return 0, fmt.Errorf("%s", b.Response)
	}

	v, err := strconv.ParseInt(b.Response, 10, 64)
	if err != nil {
		return 0, err
	}

	return v, nil
}
