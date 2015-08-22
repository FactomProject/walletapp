// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
	fct "github.com/FactomProject/factoid"
	"github.com/hoisie/web"
	"github.com/FactomProject/fctwallet/Wallet/Utility"
	"github.com/FactomProject/fctwallet/Wallet"
)

var _ = fct.Address{}

func HandleFactoidGenerateAddress(ctx *web.Context, name string) {
	ok := Utility.IsValidKey(name)
	if !ok {
		reportResults(ctx, "Name provided is not valid", false)
		return
	}

	adrstr, err := Wallet.GenerateAddressString(name)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, adrstr, true)
}

func HandleFactoidGenerateECAddress(ctx *web.Context, name string) {
	ok := Utility.IsValidKey(name)
	if !ok {
		reportResults(ctx, "Name provided is not valid", false)
		return
	}
	
	adrstr, err := Wallet.GenerateECAddressString(name)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, adrstr, true)
}
