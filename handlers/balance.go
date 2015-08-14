// Copyright 2015 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package handlers

import (
	"fmt"
	"github.com/hoisie/web"

	"github.com/FactomProject/fctwallet/Wallet"
)

func FctBalance(adr string) (int64, error) {
	return Wallet.FactoidBalance(adr)
}

func ECBalance(adr string) (int64, error) {
	return Wallet.ECBalance(adr)
}

func HandleEntryCreditBalance(ctx *web.Context, adr string) {
	v, err := ECBalance(adr)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}
	str := fmt.Sprintf("%d", v)
	reportResults(ctx, str, true)
}

func HandleFactoidBalance(ctx *web.Context, adr string) {
	v, err := FctBalance(adr)
	if err != nil {
		reportResults(ctx, err.Error(), false)
		return
	}

	str := fmt.Sprintf("%d", v)
	reportResults(ctx, str, true)
}
