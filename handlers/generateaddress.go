// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
	fct "github.com/FactomProject/factoid"
	"github.com/FactomProject/fctwallet/Wallet"
	"github.com/FactomProject/fctwallet/Wallet/Utility"
	"github.com/hoisie/web"
)

var _ = fct.Address{}

func HandleFactoidGenerateAddress(ctx *web.Context, name string) {
	if Utility.IsValidKey(name) == false {
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
	if Utility.IsValidKey(name) == false {
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

/*********************************************************************************************************/
/******************************************From Private Key***********************************************/
/*********************************************************************************************************/

func HandleFactoidGenerateAddressFromPrivateKey(ctx *web.Context, params string) {
	name := ctx.Params["name"]
	privateKey := ctx.Params["privateKey"]
	if Utility.IsValidKey(name) == false {
		reportResults(ctx, "Name provided is not valid", false)
		return
	}
	if len(privateKey) != 64 && len(privateKey) != 128 {
		reportResults(ctx, "Invalid private key length", false)
		return
	}
	if Utility.IsValidHex(privateKey) == false {
		reportResults(ctx, "Invalid private key format", false)
		return
	}

	adrstr, err := Wallet.GenerateAddressStringFromPrivateKey(name, privateKey)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, adrstr, true)
}

func HandleFactoidGenerateECAddressFromPrivateKey(ctx *web.Context, params string) {
	name := ctx.Params["name"]
	privateKey := ctx.Params["privateKey"]
	if Utility.IsValidKey(name) == false {
		reportResults(ctx, "Name provided is not valid", false)
		return
	}
	if len(privateKey) != 64 && len(privateKey) != 128 {
		reportResults(ctx, "Invalid private key length", false)
		return
	}
	if Utility.IsValidHex(privateKey) == false {
		reportResults(ctx, "Invalid private key format", false)
		return
	}

	adrstr, err := Wallet.GenerateECAddressStringFromPrivateKey(name, privateKey)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, adrstr, true)
}

/*********************************************************************************************************/
/********************************From human readable private key******************************************/
/*********************************************************************************************************/

func HandleFactoidGenerateAddressFromHumanReadablePrivateKey(ctx *web.Context, params string) {
	name := ctx.Params["name"]
	privateKey := ctx.Params["privateKey"]
	if Utility.IsValidKey(name) == false {
		reportResults(ctx, "Name provided is not valid", false)
		return
	}

	adrstr, err := Wallet.GenerateAddressStringFromHumanReadablePrivateKey(name, privateKey)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, adrstr, true)
}

func HandleFactoidGenerateECAddressFromHumanReadablePrivateKey(ctx *web.Context, params string) {
	name := ctx.Params["name"]
	privateKey := ctx.Params["privateKey"]
	if Utility.IsValidKey(name) == false {
		reportResults(ctx, "Name provided is not valid", false)
		return
	}

	adrstr, err := Wallet.GenerateECAddressStringFromHumanReadablePrivateKey(name, privateKey)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, adrstr, true)
}

/*********************************************************************************************************/
/*********************************************From mnemonic***********************************************/
/*********************************************************************************************************/

func HandleFactoidGenerateAddressFromMnemonic(ctx *web.Context, params string) {
	name := ctx.Params["name"]
	mnemonic := ctx.Params["mnemonic"]
	if Utility.IsValidKey(name) == false {
		reportResults(ctx, "Name provided is not valid", false)
		return
	}

	adrstr, err := Wallet.GenerateAddressStringFromMnemonic(name, mnemonic)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	reportResults(ctx, adrstr, true)
}

